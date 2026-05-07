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
	sym  string
	path []string // raw path segments (may contain "package"/"super") from the source
}

type module struct {
	filePath string
	file     *ast.File
	imports  map[string]importEntry  // alias → import info
	symbols  map[string]ast.Decl     // local symbol table
	used     map[string]bool         // symbols marked as used
	order    []string                // symbols in order first marked used
	renames  map[*ast.Ident]struct{} // idents to rename in the apply pass
}

type Resolver struct {
	files     map[string]*ast.File
	defines   map[string]bool
	resolved  map[string]*module
	order     []string // files in order of first load (pre-order)
	rootFile  string
	namespace map[string]bool       // all output names claimed so far
	names     map[fileSymbol]string // (file,sym) → assigned output name
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

	r.rootFile = filename
	mod := r.loadModule(filename)

	// Seed the namespace with root-module symbol names so imported symbols
	// that mangle to the same name get a numeric suffix instead of colliding.
	r.namespace = make(map[string]bool)
	r.names = make(map[fileSymbol]string)
	for _, d := range mod.file.Decls {
		if n := getDeclName(d); n != "" {
			r.namespace[n] = true
		}
	}

	// resolveRefs marks used symbols and records which idents need renaming.
	r.resolveRefs(mod)
	for _, filePath := range r.order {
		m := r.resolved[filePath]
		for ident := range m.renames {
			ident.Val = r.applyRename(m, ident.Val)
		}
	}

	decls := mod.file.Decls

	// Emit imported modules in load order (index 0 is root, skip it).
	for _, filePath := range r.order[1:] {
		m := r.resolved[filePath]
		// ConstAssertStmts have no name and are always emitted.
		for _, d := range m.file.Decls {
			if _, ok := d.(*ast.ConstAssertStmt); ok {
				decls = append(decls, d)
			}
		}
		// Named symbols in dependency-traversal order.
		for _, name := range m.order {
			d := m.symbols[name]
			if d == nil {
				continue
			}
			setDeclName(d, r.nameFor(filePath, name))
			decls = append(decls, d)
		}
	}

	return &ast.File{Decls: decls}, nil
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
		renames:  make(map[*ast.Ident]struct{}),
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
			if evalCondition(n.Cond, r.defines) {
				return n.Then
			} else {
				return n.Else
			}
		case *ast.IfAttrStmt:
			if evalCondition(n.Cond, r.defines) {
				return n.Then
			} else {
				return n.Else
			}
		case *ast.IfAttrParam:
			if evalCondition(n.Cond, r.defines) {
				return n.Then
			} else {
				return n.Else
			}
		case *ast.IfAttrClause:
			if evalCondition(n.Cond, r.defines) {
				return n.Then
			} else {
				return n.Else
			}
		case *ast.IfAttrStructMember:
			if evalCondition(n.Cond, r.defines) {
				return n.Then
			} else {
				return n.Else
			}
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
		default:
			if name := getDeclName(d); name != "" {
				mod.symbols[name] = d
			}
			mod.file.Decls[j] = d
			j++
		}

	}
	mod.file.Decls = mod.file.Decls[:j]
}

func (r *Resolver) resolveRefs(mod *module) {
	scope := newScopeStack()

	for _, d := range mod.file.Decls {
		r.resolveDecl(mod, d, scope)
	}
}

func (r *Resolver) resolveDecl(mod *module, d ast.Decl, scope *scopeStack) {
	switch d := d.(type) {
	case *ast.FuncDecl:
		r.resolveFuncDecl(mod, d, scope)
	case *ast.StructDecl:
		r.resolveStructDecl(mod, d, scope)
	case *ast.TypeAliasDecl:
		r.resolveTypeAliasDecl(mod, d, scope)
	case *ast.VarStmt:
		r.resolveVarDecl(mod, d, scope)
	case *ast.ValStmt:
		r.resolveValDecl(mod, d, scope)
	}
}

func (r *Resolver) resolveFuncDecl(mod *module, f *ast.FuncDecl, scope *scopeStack) {
	for _, p := range f.Params {
		if fp, ok := p.(*ast.FuncParam); ok {
			r.resolveType(mod, fp.Type, scope)
		}
	}
	if f.ReturnType != nil {
		r.resolveType(mod, f.ReturnType, scope)
	}
	if f.Body != nil {
		w := &refWalker{r: r, mod: mod, scope: scope}
		ast.Walk(f.Body, w.walk)
	}
}

// resolveType resolves a TypeSpecifier and marks any needed renames.
func (r *Resolver) resolveType(mod *module, typ *ast.TypeSpecifier, scope *scopeStack) {
	for _, arg := range typ.TemplateArgs {
		ast.Walk(arg, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.Ident:
				r.resolveExprName(mod, n.Val, scope)
				if r.needsRename(mod, n.Val, scope) {
					mod.renames[n] = struct{}{}
				}
			}
			return true
		})
	}
	r.resolveName(mod, typ.Name.Val, scope)
	if r.needsRename(mod, typ.Name.Val, scope) {
		mod.renames[typ.Name] = struct{}{}
	}
}

// resolveName marks name as used in mod, loading external modules as needed,
// and recurses into the declaration's own dependencies. The used map acts as a
// cycle guard so mutual references terminate.
func (r *Resolver) resolveName(mod *module, name string, scope *scopeStack) {
	if name == "" || scope.has(name) {
		return
	}

	// Module symbols take precedence over builtins (e.g. alias f32 = ...).
	if d, ok := mod.symbols[name]; ok {
		if mod.used[name] {
			return
		}
		mod.used[name] = true
		mod.order = append(mod.order, name)
		r.resolveDecl(mod, d, scope)
		return
	}

	if isBuiltinType(name) {
		return
	}

	if i, ok := mod.imports[name]; ok {
		segs := r.resolvePathSegs(i.path, mod.filePath)
		file := r.lookupFile(segs)
		if file == "" {
			// Fallback: sym may be the last path segment of a module file path.
			file = r.lookupFile(append(segs, i.sym))
		}
		m := r.loadModule(file)
		if m == nil || m.used[i.sym] {
			return
		}
		m.used[i.sym] = true
		m.order = append(m.order, i.sym)
		if d := m.symbols[i.sym]; d != nil {
			r.resolveDecl(m, d, newScopeStack())
		}
	}
}

// resolveExprName handles both plain names and inline "a::b::sym" references
// for the mark-used pass only (no renaming here).
func (r *Resolver) resolveExprName(mod *module, name string, scope *scopeStack) {
	if !strings.Contains(name, "::") {
		r.resolveName(mod, name, scope)
		return
	}
	filePath, sym := r.resolveQualifiedName(mod, name, mod.filePath)
	if filePath == "" {
		return
	}
	m := r.loadModule(filePath)
	if m == nil || m.used[sym] {
		return
	}
	m.used[sym] = true
	m.order = append(m.order, sym)
	if d := m.symbols[sym]; d != nil {
		r.resolveDecl(m, d, newScopeStack())
	}
}

// needsRename reports whether name should be added to the rename table.
// Locals (scope.has) and builtins are never renamed.
func (r *Resolver) needsRename(mod *module, name string, scope *scopeStack) bool {
	if name == "" || scope.has(name) {
		return false
	}
	if strings.Contains(name, "::") {
		return true
	}
	// Module symbols take precedence over builtins (e.g. alias f32 = ...).
	if _, ok := mod.symbols[name]; ok {
		return mod.filePath != r.rootFile
	}
	if isBuiltinType(name) {
		return false
	}
	_, ok := mod.imports[name]
	return ok
}

// applyRename computes the output name for an ident whose original value is name.
// Called during the apply pass; scope filtering has already been done at record time.
func (r *Resolver) applyRename(mod *module, name string) string {
	if strings.Contains(name, "::") {
		filePath, sym := r.resolveQualifiedName(mod, name, mod.filePath)
		if filePath == "" {
			return name
		}
		if filePath == r.rootFile {
			return sym
		}
		return r.nameFor(filePath, sym)
	}
	// Module symbols take precedence over builtins.
	if _, ok := mod.symbols[name]; ok {
		if mod.filePath == r.rootFile {
			return name
		}
		return r.nameFor(mod.filePath, name)
	}
	if i, ok := mod.imports[name]; ok {
		segs := r.resolvePathSegs(i.path, mod.filePath)
		depFile := r.lookupFile(segs)
		if depFile == "" {
			depFile = r.lookupFile(append(segs, i.sym))
		}
		if depFile == "" {
			return name
		}
		if depFile == r.rootFile {
			return i.sym
		}
		return r.nameFor(depFile, i.sym)
	}
	return name
}

// nameFor returns the collision-free output name for (file, sym), allocating
// one on first call and caching it for consistent reuse.
func (r *Resolver) nameFor(file, sym string) string {
	key := fileSymbol{file, sym}
	if n, ok := r.names[key]; ok {
		return n
	}
	base := mangleName(file, sym)
	n := base
	for i := 0; r.namespace[n]; i++ {
		n = fmt.Sprintf("%s_%d", base, i)
	}
	r.namespace[n] = true
	r.names[key] = n
	return n
}

func (r *Resolver) resolveStructDecl(mod *module, s *ast.StructDecl, scope *scopeStack) {
	for _, m := range s.Members {
		if sf, ok := m.(*ast.StructMember); ok {
			r.resolveType(mod, sf.Type, scope)
		}
	}
}

func (r *Resolver) resolveTypeAliasDecl(mod *module, a *ast.TypeAliasDecl, scope *scopeStack) {
	r.resolveType(mod, a.Type, scope)
}

func (r *Resolver) resolveVarDecl(mod *module, v *ast.VarStmt, scope *scopeStack) {
	if v.Type != nil {
		r.resolveType(mod, v.Type, scope)
	}
	if v.Init != nil {
		ast.Walk(v.Init, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.CallExpr:
				r.resolveExprName(mod, n.Callee.Val, scope)
				if r.needsRename(mod, n.Callee.Val, scope) {
					mod.renames[n.Callee] = struct{}{}
				}
			case *ast.Ident:
				r.resolveExprName(mod, n.Val, scope)
				if r.needsRename(mod, n.Val, scope) {
					mod.renames[n] = struct{}{}
				}
			}
			return true
		})
	}
}

func (r *Resolver) resolveValDecl(mod *module, v *ast.ValStmt, scope *scopeStack) {
	if v.Type != nil {
		r.resolveType(mod, v.Type, scope)
	}
	if v.Init != nil {
		ast.Walk(v.Init, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.CallExpr:
				r.resolveExprName(mod, n.Callee.Val, scope)
				if r.needsRename(mod, n.Callee.Val, scope) {
					mod.renames[n.Callee] = struct{}{}
				}
			case *ast.Ident:
				r.resolveExprName(mod, n.Val, scope)
				if r.needsRename(mod, n.Val, scope) {
					mod.renames[n] = struct{}{}
				}
			}
			return true
		})
	}
}

// ── Symbol/file lookup ────────────────────────────────────────────────────────

// resolvePathSegs converts an import prefix (containing package/super/path segments)
// into root-relative path segments, resolving package/super relative to sourceFile.
func (r *Resolver) resolvePathSegs(prefix []string, sourceFile string) []string {
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
	return segs
}

// resolveQualifiedName resolves an inline qualified name like "foo::bar" or
// "package::file::sym" to (filePath, sym) using the module map or path resolution.
func (r *Resolver) resolveQualifiedName(mod *module, name string, sourceFile string) (string, string) {
	parts := strings.Split(name, "::")
	sym := parts[len(parts)-1]
	root := parts[0]

	if entry, ok := mod.imports[root]; ok {
		p := entry.path
		p = append(p, parts[1:len(parts)-1]...)
		p = r.resolvePathSegs(p, sourceFile)
		file := r.lookupFile(p)
		if file == "" {
			// Fallback: the import alias may point to a module file whose name
			// is entry.sym (e.g. import package::dir::modname → file "dir/modname").
			file = r.lookupFile(append(p, entry.sym))
		}
		return file, sym
	}
	p := r.resolvePathSegs(parts[:len(parts)-1], sourceFile)
	return r.lookupFile(p), sym
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

// ── AST helpers ───────────────────────────────────────────────────────────────

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

// refWalker carries the state needed to resolve and rename references inside a
// function body. Using a struct with a method avoids a self-referential closure.
type refWalker struct {
	r     *Resolver
	mod   *module
	scope *scopeStack
}

func (w *refWalker) walk(n ast.Node) bool {
	switch n := n.(type) {
	case *ast.BlockStmt:
		w.scope.push()
		for _, s := range n.Stmts {
			ast.Walk(s, w.walk)
		}
		w.scope.pop()
		return false
	case *ast.VarStmt:
		if n.Type != nil {
			w.r.resolveType(w.mod, n.Type, w.scope)
		}
		w.scope.add(n.Name.Val)
	case *ast.ValStmt:
		if n.Type != nil {
			w.r.resolveType(w.mod, n.Type, w.scope)
		}
		w.scope.add(n.Name.Val)
	case *ast.CallExpr:
		w.r.resolveExprName(w.mod, n.Callee.Val, w.scope)
		if w.r.needsRename(w.mod, n.Callee.Val, w.scope) {
			w.mod.renames[n.Callee] = struct{}{}
		}
	case *ast.Ident:
		w.r.resolveExprName(w.mod, n.Val, w.scope)
		if w.r.needsRename(w.mod, n.Val, w.scope) {
			w.mod.renames[n] = struct{}{}
		}
	}
	return true
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
	for i := len(s.blocks) - 1; i >= 0; i-- {
		if _, ok := s.blocks[i][name]; ok {
			return true
		}
	}
	return false
}

func getDeclName(d ast.Decl) string {
	switch d := d.(type) {
	case *ast.FuncDecl:
		return d.Name.Val
	case *ast.StructDecl:
		return d.Name.Val
	case *ast.ValStmt:
		return d.Name.Val
	case *ast.VarStmt:
		return d.Name.Val
	case *ast.TypeAliasDecl:
		return d.Name.Val
	default:
		return ""
	}
}

func setDeclName(d ast.Decl, name string) {
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

func mangleName(file, sym string) string {
	return "package_" + strings.ReplaceAll(file, "/", "_") + "_" + sym
}
