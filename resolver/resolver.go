package resolver

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/bluescreen10/wesl-go/ast"
)

// fileSymbol uniquely identifies a symbol within a specific source file.
type fileSymbol struct {
	file string
	sym  string
}

// importEntry is one explicit import in a module's import list.
type importEntry struct {
	path []string
	sym  string
	file string
}

type module struct {
	filePath     string
	file         *ast.File
	imports      map[string]importEntry // alias → import info
	symbols      map[string]ast.Decl    // local symbol table
	used         map[string]bool        // symbols marked as used
	order        []string               // symbols in order first marked used
	renames      map[*ast.Ident]bool    // idents to rename in the apply pass
	constAsserts []*ast.ConstAssertStmt
}

type Resolver struct {
	files    map[string]*ast.File
	defines  map[string]bool
	resolved map[string]*module
	order    []string // files in order of first load (pre-order)
}

func ResolveFile(filename string, files map[string]*ast.File, defines map[string]bool) (*ast.File, error) {
	r := New(files, defines)
	return r.ResolveFile(filename)
}

func New(files map[string]*ast.File, defines map[string]bool) *Resolver {
	return &Resolver{
		files:    files,
		defines:  defines,
		resolved: make(map[string]*module),
	}
}

func (r *Resolver) ResolveFile(filename string) (f *ast.File, err error) {
	defer func() {
		if e := recover(); e != nil {
			switch e := e.(type) {
			case error:
				err = e
			case string:
				err = errors.New(e)
			default:
				panic(e)
			}
		}
	}()

	mod := r.loadModule(filename)
	r.resolveRef(mod, mod.file)
	r.renameSymbols(mod)
	return r.emitFile(mod), nil
}

func (r *Resolver) emitFile(root *module) *ast.File {
	decls := root.file.Decls

	// Emit imported modules in load order (index 0 is root, skip it).
	for _, filePath := range r.order[1:] {
		m := r.resolved[filePath]

		// ConstAssertStmts have no name and are always emitted.
		for _, d := range m.constAsserts {
			decls = append(decls, d)
		}

		// Named symbols in dependency-traversal order.
		for _, name := range m.order {
			d := m.symbols[name]
			decls = append(decls, d)
		}
	}

	return &ast.File{Decls: decls}
}

func (r *Resolver) renameSymbols(root *module) {
	// Seed names with root-module symbols so imported symbols that mangle to
	// the same name get a numeric suffix instead of colliding.
	names := make(map[fileSymbol]string)
	for n := range root.symbols {
		names[fileSymbol{root.filePath, n}] = n
	}

	// rename refs
	for _, filePath := range r.order {
		m := r.resolved[filePath]
		for ident := range m.renames {
			ident.Val = r.applyRename(m, ident.Val, names, root.filePath)
		}
	}
}

func (r *Resolver) loadModule(filename string) *module {
	if mod := r.resolved[filename]; mod != nil {
		return mod
	}

	file := r.files[filename]
	if file == nil {
		return nil
	}

	file = r.resolveConditionals(file.Clone())
	mod := &module{
		filePath: filename,
		file:     file,
		used:     make(map[string]bool),
		symbols:  make(map[string]ast.Decl),
		imports:  make(map[string]importEntry),
		renames:  make(map[*ast.Ident]bool),
	}

	r.buildSymbolAndImports(mod)
	r.resolved[filename] = mod
	r.order = append(r.order, filename)
	return mod
}

func (r *Resolver) resolveConditionals(f *ast.File) *ast.File {
	out := ast.Rewrite(f, func(n ast.Node) ast.Node {
		switch n := n.(type) {
		case *ast.IfAttrDecl:
			return pickBranch(n.Cond, n.Then, n.Else, r.defines)
		case *ast.IfAttrStmt:
			return pickBranch(n.Cond, n.Then, n.Else, r.defines)
		case *ast.IfAttrParam:
			return pickBranch(n.Cond, n.Then, n.Else, r.defines)
		case *ast.IfAttrClause:
			return pickBranch(n.Cond, n.Then, n.Else, r.defines)
		case *ast.IfAttrStructMember:
			return pickBranch(n.Cond, n.Then, n.Else, r.defines)
		default:
			return n
		}
	})
	return out.(*ast.File)
}

func (r *Resolver) buildSymbolAndImports(mod *module) {
	var j int

	for _, d := range mod.file.Decls {
		switch d := d.(type) {
		case *ast.ImportDecl:
			for _, i := range d.Imports {
				alias := i.Alias
				name := i.Path[len(i.Path)-1]
				if alias == "" {
					alias = name
				}
				mod.imports[alias] = importEntry{
					sym:  name,
					path: i.Path[:len(i.Path)-1],
				}
			}
		case *ast.ConstAssertStmt:
			mod.constAsserts = append(mod.constAsserts, d)
			mod.file.Decls[j] = d
			j++
		default:
			if i := getNodeName(d); i != nil {
				name := i.Val
				mod.symbols[name] = d
			}
			mod.file.Decls[j] = d
			j++
		}

	}
	mod.file.Decls = mod.file.Decls[:j]
}

func (r *Resolver) resolveRef(mod *module, root ast.Node) {
	scope := newScopeStack()

	if ident := getNodeName(root); ident != nil {
		name := ident.Val
		mod.used[name] = true
		mod.order = append(mod.order, name)
		mod.renames[ident] = true
	}

	var walk func(n ast.Node) bool

	walk = func(n ast.Node) bool {
		switch n := n.(type) {

		case *ast.VarStmt:
			scope.add(n.Name.Val)
			ast.Walk(n.Type, walk)
			ast.WalkList(n.TemplateArgs, walk)
			ast.Walk(n.Init, walk)
			return false

		case *ast.ValStmt:
			scope.add(n.Name.Val)
			ast.Walk(n.Type, walk)
			ast.Walk(n.Init, walk)
			return false

		case *ast.BlockStmt:
			scope.push()
			ast.WalkList(n.Stmts, walk)
			scope.pop()
			return false

		case *ast.TypeSpecifier:
			// Refs need renaming
			if r.resolveName(mod, n.Name.Val, scope) {
				mod.renames[n.Name] = true
			}

			for _, arg := range n.TemplateArgs {
				ast.Walk(arg, walk)
			}
			return false

		case *ast.Ident:
			// Refs need renaming
			if r.resolveName(mod, n.Val, scope) {
				mod.renames[n] = true
			}
		}
		return true
	}

	ast.Walk(root, walk)
}

// lookupImportFile resolves the file path for an importEntry relative to mod.
func (r *Resolver) lookupImportFile(mod *module, i importEntry) string {
	if file := r.lookupPath(i.path, mod.filePath); file != "" {
		return file
	}
	// Fallback: sym may be the last path segment of a module file path.
	return r.lookupPath(append(i.path, i.sym), mod.filePath)
}

// resolveName marks name as used in mod, loading external modules as needed,
// and recurses into the declaration's own dependencies. Handles both plain
// names and inline qualified references like "package::foo::MyType".
// The used map acts as a cycle guard so mutual references terminate.
func (r *Resolver) resolveName(mod *module, name string, scope *scopeStack) bool {
	// Do nothing for empty or locally defined symbols
	if name == "" || scope.has(name) {
		return false
	}

	// Look for module symbol table
	if d, ok := mod.symbols[name]; ok {
		if !mod.used[name] {
			r.resolveRef(mod, d)
		}
		return true
	}

	// Do nothing for built-ins
	if isBuiltinType(name) {
		return false
	}

	// Look for imported symbols
	if i, ok := mod.imports[name]; ok {
		file := r.lookupImportFile(mod, i)
		m := r.loadModule(file)

		i.file = file
		mod.imports[name] = i

		if m != nil && !m.used[i.sym] {
			if d := m.symbols[i.sym]; d != nil {
				r.resolveRef(m, d)
			} else {
				panic("symbol not found")
			}
		}

		return true
	}

	// Look for fully qualified names
	if strings.Contains(name, "::") {
		filePath, sym := r.resolveQualifiedName(mod, name, mod.filePath)
		if filePath == "" {
			return false
		}
		m := r.loadModule(filePath)
		if m != nil && !m.used[sym] {
			if d := m.symbols[sym]; d != nil {
				r.resolveRef(m, d)
			} else {
				panic("symbol not found")
			}
		}
		return true
	}

	return false
}

// applyRename computes the output name for an ident whose original value is name.
// Called during the apply pass; scope filtering has already been done at record time.
func (r *Resolver) applyRename(mod *module, name string, names map[fileSymbol]string, rootFile string) string {
	if strings.Contains(name, "::") {
		filePath, sym := r.resolveQualifiedName(mod, name, mod.filePath)
		if filePath == "" {
			return name
		}
		if filePath == rootFile {
			return sym
		}
		return r.nameFor(filePath, sym, names)
	}

	// Module symbols take precedence over builtins.
	if _, ok := mod.symbols[name]; ok {
		if mod.filePath == rootFile {
			return name
		}
		return r.nameFor(mod.filePath, name, names)
	}

	if i, ok := mod.imports[name]; ok {
		if i.file == "" {
			return name
		}
		if i.file == rootFile {
			return i.sym
		}
		return r.nameFor(i.file, i.sym, names)
	}
	return name
}

// nameFor returns the collision-free output name for (file, sym), allocating
// one on first call and caching it for consistent reuse.
func (r *Resolver) nameFor(file, sym string, names map[fileSymbol]string) string {
	key := fileSymbol{file, sym}
	if n, ok := names[key]; ok {
		return n
	}
	base := mangleName(file, sym)
	n := base
	for i := 0; r.nameTaken(n, names); i++ {
		n = fmt.Sprintf("%s_%d", base, i)
	}
	names[key] = n
	return n
}

func (r *Resolver) nameTaken(name string, names map[fileSymbol]string) bool {
	for _, v := range names {
		if v == name {
			return true
		}
	}
	return false
}

// resolveQualifiedName resolves an inline qualified name like "foo::bar" or
// "package::file::sym" to (filePath, sym) using the module map or path resolution.
func (r *Resolver) resolveQualifiedName(mod *module, name string, sourceFile string) (string, string) {
	parts := strings.Split(name, "::")
	sym := parts[len(parts)-1]
	root := parts[0]

	if entry, ok := mod.imports[root]; ok {
		p := append(entry.path, parts[1:len(parts)-1]...)
		if file := r.lookupPath(p, sourceFile); file != "" {
			return file, sym
		}
		// Fallback: the import alias may point to a module file whose name
		// is entry.sym (e.g. import package::dir::modname → file "dir/modname").
		return r.lookupPath(append(p, entry.sym), sourceFile), sym
	}
	return r.lookupPath(parts[:len(parts)-1], sourceFile), sym
}

func (r *Resolver) lookupFile(segs []string) string {
	for i := len(segs); i >= 1; i-- {
		candidate := strings.Join(segs[:i], "/")
		if _, ok := r.files[candidate]; ok {
			return candidate
		}
	}
	return ""
}

// lookupPath resolves an import prefix (containing package/super/path segments)
// relative to sourceFile and returns the matching file path, or "" if not found.
func (r *Resolver) lookupPath(prefix []string, sourceFile string) string {
	dir := path.Dir(sourceFile)
	var segs []string
	for _, p := range prefix {
		switch p {
		case "package":
			dir = "."
		case "super":
			if parent := path.Dir(dir); parent != dir {
				dir = parent
			}
		default:
			segs = append(segs, p)
		}
	}
	trimmed := strings.TrimPrefix(dir, "./")
	if trimmed != "." && trimmed != "" {
		segs = append(strings.Split(trimmed, "/"), segs...)
	}
	return r.lookupFile(segs)
}

func isBuiltinType(typ string) bool {
	switch typ {
	case "array", "atomic", "bool",
		"f16", "f32", "i32", "u32",
		"mat2x2", "mat2x3", "mat2x4",
		"mat3x2", "mat3x3", "mat3x4",
		"mat4x2", "mat4x3", "mat4x4",
		"ptr",
		"sampler", "sampler_comparison",
		"texture_1d", "texture_2d", "texture_2d_array",
		"texture_3d", "texture_cube", "texture_cube_array",
		"texture_depth_2d", "texture_depth_2d_array",
		"texture_depth_cube", "texture_depth_cube_array",
		"texture_depth_multisampled_2d",
		"texture_multisampled_2d",
		"texture_storage_1d", "texture_storage_2d",
		"texture_storage_2d_array", "texture_storage_3d",
		"vec2", "vec3", "vec4",
		"binding_array":
		return true
	default:
		return false
	}
}

type scopeStack struct {
	blocks []map[string]struct{}
}

func newScopeStack() *scopeStack {
	return &scopeStack{blocks: []map[string]struct{}{make(map[string]struct{})}}
}

func (s *scopeStack) push() {
	s.blocks = append(s.blocks, make(map[string]struct{}))
}

func (s *scopeStack) pop() {
	s.blocks = s.blocks[:len(s.blocks)-1]
}

func (s *scopeStack) add(name string) {
	s.blocks[len(s.blocks)-1][name] = struct{}{}
}

func (s scopeStack) has(name string) bool {
	// Don't check the global namespace
	for i := len(s.blocks) - 1; i > 0; i-- {
		if _, ok := s.blocks[i][name]; ok {
			return true
		}
	}
	return false
}

func getNodeName(d ast.Node) *ast.Ident {
	switch d := d.(type) {
	case *ast.FuncDecl:
		return d.Name
	case *ast.StructDecl:
		return d.Name
	case *ast.ValStmt:
		return d.Name
	case *ast.VarStmt:
		return d.Name
	case *ast.TypeAliasDecl:
		return d.Name
	default:
		return nil
	}
}

func setNodeName(d ast.Decl, name string) {
	switch d := d.(type) {
	case *ast.FuncDecl:
		d.Name.Val = name
	case *ast.StructDecl:
		d.Name.Val = name
	case *ast.ValStmt:
		d.Name.Val = name
	case *ast.VarStmt:
		d.Name.Val = name
	case *ast.TypeAliasDecl:
		d.Name.Val = name
	}
}

func mangleName(file, sym string) string {
	return "package_" + strings.ReplaceAll(file, "/", "_") + "_" + sym
}

func pickBranch[T ast.Node](cond ast.Expr, then, els T, defines map[string]bool) T {
	if evalCondition(cond, defines) {
		return then
	}
	return els
}

func evalCondition(expr ast.Expr, defines map[string]bool) bool {
	switch e := expr.(type) {

	case *ast.LitExpr:
		switch e.Val {
		case "true":
			return true
		case "false":
			return false
		}

	case *ast.Ident:
		return defines[e.Val] // missing key → false

	case *ast.UnaryExpr:
		if e.Op == "!" {
			return !evalCondition(e.Operand, defines)
		}

	case *ast.BinaryExpr:
		switch e.Op {
		case "&&":
			return evalCondition(e.Left, defines) && evalCondition(e.Right, defines)
		case "||":
			return evalCondition(e.Left, defines) || evalCondition(e.Right, defines)
		case "==":
			return evalCondition(e.Left, defines) == evalCondition(e.Right, defines)
		case "!=":
			return evalCondition(e.Left, defines) != evalCondition(e.Right, defines)
		}

	case *ast.ParenExpr:
		return evalCondition(e.Inner, defines)
	}

	return false
}
