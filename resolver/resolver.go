package resolver

import (
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

type resolvedModule struct {
	filePath string
	file     *ast.File
	imports  map[string]importEntry // alias → import info
	symbols  map[string]ast.Decl    // local symbol table
	used     map[string]bool        // symbols marked as used
	order    []string               // symbols in order first marked used
}

type Resolver struct {
	files     map[string]*ast.File
	defines   map[string]bool
	resolved  map[string]*resolvedModule
	loadOrder []string // files in order of first load (pre-order)
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
		resolved: make(map[string]*resolvedModule),
	}
}

func mangleName(file, sym string) string {
	return "package_" + strings.ReplaceAll(file, "/", "_") + "_" + sym
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

func (r *Resolver) ResolveFile(filename string) (f *ast.File, err error) {
	// defer func() {
	// 	if e := recover(); e != nil {
	// 		switch e := e.(type) {
	// 		case error:
	// 			err = e
	// 		case string:
	// 			err = errors.New(e)
	// 		default:
	// 			panic(e)
	// 		}
	// 	}
	// }()

	r.rootFile = filename
	mod := r.loadModule(filename)

	// Seed the namespace with root-module symbol names so imported symbols
	// that mangle to the same name get a numeric suffix instead of colliding.
	r.namespace = make(map[string]bool)
	r.names = make(map[fileSymbol]string)
	for _, d := range mod.file.Decls {
		if n := d.GetName(); n != "" {
			r.namespace[n] = true
		}
	}

	// resolveRefs marks used symbols AND renames references in-place.
	r.resolveRefs(mod)

	decls := mod.file.Decls

	// Emit imported modules in load order (index 0 is root, skip it).
	for _, filePath := range r.loadOrder[1:] {
		m := r.resolved[filePath]
		// ConstAssertDecls have no name and are always emitted.
		for _, d := range m.file.Decls {
			if _, ok := d.(*ast.ConstAssertDecl); ok {
				decls = append(decls, d)
			}
		}
		// Named symbols in dependency-traversal order.
		for _, name := range m.order {
			d := m.symbols[name]
			if d == nil {
				continue
			}
			cloned := cloneDecl(d)
			cloned.SetName(r.nameFor(filePath, name))
			decls = append(decls, cloned)
		}
	}

	return &ast.File{Decls: decls}, nil
}

func (r *Resolver) loadModule(filename string) *resolvedModule {
	if mod := r.resolved[filename]; mod != nil {
		return mod
	}

	file := r.files[filename]
	if file == nil {
		return nil
	}

	file = r.ResolveConditionals(file)
	mod := &resolvedModule{
		filePath: filename,
		file:     file,
		used:     make(map[string]bool),
		symbols:  make(map[string]ast.Decl),
		imports:  make(map[string]importEntry),
	}

	r.buildSymbolAndImports(mod)
	r.resolved[filename] = mod
	r.loadOrder = append(r.loadOrder, filename)
	return mod
}

func (r *Resolver) buildSymbolAndImports(mod *resolvedModule) {
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
			if name := d.GetName(); name != "" {
				mod.symbols[name] = d
			}

			// remove import declarations
			mod.file.Decls[j] = d
			j++
		}
	}
	mod.file.Decls = mod.file.Decls[:j]
}

func (r *Resolver) resolveRefs(mod *resolvedModule) {
	scope := newScopeStack()

	for _, d := range mod.file.Decls {
		r.resolveRefDecl(mod, d, scope)
	}
}

func (r *Resolver) resolveRefDecl(mod *resolvedModule, d ast.Decl, scope scopeStack) {
	switch d := d.(type) {
	case *ast.FuncDecl:
		r.resolveRefFuncDecl(mod, d, scope)
	case *ast.StructDecl:
		r.resolveRefStructDecl(mod, d, scope)
	case *ast.TypeAliasDecl:
		r.resolveRefTypeAliasDecl(mod, d, scope)
	case *ast.GlobalVarDecl:
		r.resolveRefGlobalVarDecl(mod, d, scope)
	case *ast.GlobalValDecl:
		r.resolveRefGlobalValDecl(mod, d, scope)
	}
}

func (r *Resolver) resolveRefFuncDecl(mod *resolvedModule, f *ast.FuncDecl, scope scopeStack) {
	for _, p := range f.Params {
		if fp, ok := p.(*ast.FuncParam); ok {
			r.resolveRefType(mod, fp.Type, scope)
		}
	}
	if f.ReturnType != nil {
		r.resolveRefType(mod, f.ReturnType, scope)
	}
	if f.Body != nil {
		w := &refWalker{r: r, mod: mod, scope: &scope}
		ast.Walk(f.Body, w.walk)
	}
}

// resolveRefType resolves and renames a TypeSpecifier in-place.
func (r *Resolver) resolveRefType(mod *resolvedModule, typ *ast.TypeSpecifier, scope scopeStack) {
	for _, arg := range typ.TemplateArgs {
		ast.Walk(arg, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.Ident:
				orig := n.Val
				r.resolveRefExprName(mod, orig, scope)
				n.Val = r.getExprRename(mod, orig, scope)
			}
			return true
		})
	}
	r.resolveRefName(mod, typ.Name, scope)
	typ.Name = r.getRenameName(mod, typ.Name, scope)
}

// resolveRefName marks name as used in mod, loading external modules as needed,
// and recurses into the declaration's own dependencies. The used map acts as a
// cycle guard so mutual references terminate.
func (r *Resolver) resolveRefName(mod *resolvedModule, name string, scope scopeStack) {
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
		r.resolveRefDecl(mod, d, scope)
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
			r.resolveRefDecl(m, d, newScopeStack())
		}
	}
}

// resolveRefExprName handles both plain names and inline "a::b::sym" references
// for the mark-used pass only (no renaming here).
func (r *Resolver) resolveRefExprName(mod *resolvedModule, name string, scope scopeStack) {
	if !strings.Contains(name, "::") {
		r.resolveRefName(mod, name, scope)
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
		r.resolveRefDecl(m, d, newScopeStack())
	}
}

// getExprRename returns the output name for name (plain or qualified) in mod context.
func (r *Resolver) getExprRename(mod *resolvedModule, name string, scope scopeStack) string {
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
	return r.getRenameName(mod, name, scope)
}

// getRenameName returns the output name for a plain (non-qualified) name.
func (r *Resolver) getRenameName(mod *resolvedModule, name string, scope scopeStack) string {
	if name == "" || scope.has(name) {
		return name
	}
	// Module symbols take precedence over builtins (e.g. alias f32 = ...).
	if _, ok := mod.symbols[name]; ok {
		if mod.filePath == r.rootFile {
			return name
		}
		return r.nameFor(mod.filePath, name)
	}
	if isBuiltinType(name) {
		return name
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

func (r *Resolver) resolveRefStructDecl(mod *resolvedModule, s *ast.StructDecl, scope scopeStack) {
	for _, m := range s.Members {
		if sf, ok := m.(*ast.StructMember); ok {
			r.resolveRefType(mod, sf.Type, scope)
		}
	}
}

func (r *Resolver) resolveRefTypeAliasDecl(mod *resolvedModule, a *ast.TypeAliasDecl, scope scopeStack) {
	r.resolveRefType(mod, a.Type, scope)
}

func (r *Resolver) resolveRefGlobalVarDecl(mod *resolvedModule, v *ast.GlobalVarDecl, scope scopeStack) {
	if v.Type != nil {
		r.resolveRefType(mod, v.Type, scope)
	}
	if v.Init != nil {
		ast.Walk(v.Init, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.CallExpr:
				orig := n.Callee.Val
				r.resolveRefExprName(mod, orig, scope)
				n.Callee.Val = r.getExprRename(mod, orig, scope)
			case *ast.Ident:
				orig := n.Val
				r.resolveRefExprName(mod, orig, scope)
				n.Val = r.getExprRename(mod, orig, scope)
			}
			return true
		})
	}
}

func (r *Resolver) resolveRefGlobalValDecl(mod *resolvedModule, v *ast.GlobalValDecl, scope scopeStack) {
	if v.Type != nil {
		r.resolveRefType(mod, v.Type, scope)
	}
	if v.Init != nil {
		ast.Walk(v.Init, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.CallExpr:
				orig := n.Callee.Val
				r.resolveRefExprName(mod, orig, scope)
				n.Callee.Val = r.getExprRename(mod, orig, scope)
			case *ast.Ident:
				orig := n.Val
				r.resolveRefExprName(mod, orig, scope)
				n.Val = r.getExprRename(mod, orig, scope)
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
func (r *Resolver) resolveQualifiedName(mod *resolvedModule, name string, sourceFile string) (string, string) {
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

func cloneDecl(d ast.Decl) ast.Decl {
	switch dd := d.(type) {
	case *ast.FuncDecl:
		c := *dd
		return &c
	case *ast.StructDecl:
		c := *dd
		return &c
	case *ast.GlobalValDecl:
		c := *dd
		return &c
	case *ast.GlobalVarDecl:
		c := *dd
		return &c
	case *ast.TypeAliasDecl:
		c := *dd
		return &c
	case *ast.ConstAssertDecl:
		c := *dd
		return &c
	}
	return d
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

// refWalker carries the state needed to resolve and rename references inside a
// function body. Using a struct with a method avoids a self-referential closure.
type refWalker struct {
	r     *Resolver
	mod   *resolvedModule
	scope *scopeStack
}

func (w *refWalker) walk(n ast.Node) bool {
	switch n := n.(type) {
	case *ast.CompoundStmt:
		w.scope.push()
		for _, s := range n.Stmts {
			ast.Walk(s, w.walk)
		}
		w.scope.pop()
		return false
	case *ast.VarStmt:
		if n.Type != nil {
			w.r.resolveRefType(w.mod, n.Type, *w.scope)
		}
		w.scope.add(n.Name.Val)
	case *ast.ValStmt:
		if n.Type != nil {
			w.r.resolveRefType(w.mod, n.Type, *w.scope)
		}
		w.scope.add(n.Name.Val)
	case *ast.CallExpr:
		orig := n.Callee.Val
		w.r.resolveRefExprName(w.mod, orig, *w.scope)
		n.Callee.Val = w.r.getExprRename(w.mod, orig, *w.scope)
	case *ast.Ident:
		orig := n.Val
		w.r.resolveRefExprName(w.mod, orig, *w.scope)
		n.Val = w.r.getExprRename(w.mod, orig, *w.scope)
	}
	return true
}

type scopeStack struct {
	blocks []map[string]struct{}
}

func newScopeStack() scopeStack {
	return scopeStack{blocks: []map[string]struct{}{make(map[string]struct{})}}
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
