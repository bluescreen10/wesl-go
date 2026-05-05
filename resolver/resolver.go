package resolver

import (
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
	filePath   string
	file       *ast.File
	imports    map[string]importEntry // alias → import info
	symbols    map[string]ast.Decl    // local symbol table
	used       map[string]bool        // symbols marked as used
	inlineRefs map[string]fileSymbol  // qualified "a::b::c" → resolved (file, sym)
	order      []string
}

func (m *resolvedModule) buildSymbolTable() {
	for _, d := range m.file.Decls {
		if name := d.GetName(); name != "" {
			m.symbols[name] = d
		}
	}
}

func (m *resolvedModule) buildImportList() {
	for _, d := range m.file.Decls {
		imp, ok := d.(*ast.ImportDecl)
		if !ok {
			continue
		}
		for _, i := range imp.Imports {
			alias := i.Alias
			name := i.Path[len(i.Path)-1]
			if alias == "" {
				alias = name
			}
			m.imports[alias] = importEntry{
				sym:  name,
				path: i.Path[:len(i.Path)-1],
			}
		}
	}
}

type Resolver struct {
	files     map[string]*ast.File
	defines   map[string]bool
	resolved  map[string]*resolvedModule
	loadOrder []string // files in order of first load (pre-order)
	rootFile  string
}

func ResolveFile(filename string, files map[string]*ast.File, defines map[string]bool) *ast.File {
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

func (r *Resolver) ResolveFile(filename string) *ast.File {
	r.rootFile = filename
	mod := r.loadModule(filename)

	r.resolveRefs(mod)

	var decls []ast.Decl

	// Emit root module's local decls first.
	rootRenames := r.moduleRenameMap(filename)
	for _, d := range mod.file.Decls {
		if _, ok := d.(*ast.ImportDecl); ok {
			continue
		}
		cloned := cloneDecl(d)
		rewriteDeclRefs(cloned, rootRenames)
		decls = append(decls, cloned)
	}

	// Emit imported modules in load order (index 0 is root, skip it).
	for _, filePath := range r.loadOrder[1:] {
		m := r.resolved[filePath]
		renames := r.moduleRenameMap(filePath)
		// ConstAssertDecls have no name and are always emitted.
		for _, d := range m.file.Decls {
			if _, ok := d.(*ast.ConstAssertDecl); ok {
				decls = append(decls, d)
			}
		}
		// Named symbols in dependency-traversal order (primary first, deps after).
		for _, name := range m.order {
			d := m.symbols[name]
			// if d == nil {
			// 	continue
			// }
			//cloned := cloneDecl(d)
			d.SetName(mangleName(filePath, name))
			rewriteDeclRefs(d, renames)
			decls = append(decls, d)
		}
	}

	return &ast.File{Decls: decls}
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
		filePath:   filename,
		file:       file,
		used:       make(map[string]bool),
		symbols:    make(map[string]ast.Decl),
		imports:    make(map[string]importEntry),
		inlineRefs: make(map[string]fileSymbol),
	}

	mod.buildSymbolTable()
	mod.buildImportList()
	//r.registerModuleImports(mod)
	r.resolved[filename] = mod
	r.loadOrder = append(r.loadOrder, filename)
	return mod
}

// // registerModuleImports populates moduleMap for any import declarations that
// // refer to whole modules rather than specific symbols.
// func (r *Resolver) registerModuleImports(mod *resolvedModule) {
// 	for _, d := range mod.file.Decls {
// 		imp, ok := d.(*ast.ImportDecl)
// 		if !ok {
// 			continue
// 		}
// 		for _, i := range imp.Imports {
// 			prefix := i.Path[:len(i.Path)-1]
// 			sym := i.Path[len(i.Path)-1]
// 			segs := r.resolvePathSegs(prefix, mod.filePath)
// 			if fp := r.lookupFile(append(segs, sym)); fp != "" {
// 				r.moduleMap[sym] = fp
// 			}
// 		}
// 	}
// }

// moduleRenameMap builds the rename map used during emit for a given module:
// maps every locally-visible name to its output name.
func (r *Resolver) moduleRenameMap(filePath string) map[string]string {
	mod := r.resolved[filePath]
	renames := make(map[string]string)
	isRoot := filePath == r.rootFile

	for name := range mod.symbols {
		if isRoot {
			renames[name] = name
		} else {
			renames[name] = mangleName(filePath, name)
		}
	}

	for alias, entry := range mod.imports {
		segs := r.resolvePathSegs(entry.path, filePath)
		depFile := r.lookupFile(segs)
		if depFile == "" {
			continue
		}
		if depFile == r.rootFile {
			renames[alias] = entry.sym
		} else {
			renames[alias] = mangleName(depFile, entry.sym)
		}
	}

	for qualName, fs := range mod.inlineRefs {
		if fs.file == r.rootFile {
			renames[qualName] = fs.sym
		} else {
			renames[qualName] = mangleName(fs.file, fs.sym)
		}
	}

	return renames
}

// ── resolveRefs: mark used symbols across modules ─────────────────────────────

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
			r.resolveRefType(mod, &fp.Type, scope)
		}
	}
	if f.ReturnType != nil {
		r.resolveRefType(mod, f.ReturnType, scope)
	}
	if f.Body != nil {
		scope.push()
		for _, s := range f.Body.Stmts {
			ast.WalkStmt(s, func(s ast.Stmt) bool {
				switch st := s.(type) {
				case *ast.VarStmt:
					if st.Type != nil {
						r.resolveRefType(mod, st.Type, scope)
						scope.add(st.Name)
					}
				case *ast.ValStmt:
					if st.Type != nil {
						r.resolveRefType(mod, st.Type, scope)
						scope.add(st.Name)
					}
				case *ast.CompoundStmt:
					scope.push()
				}
				return true
			}, func(e ast.Expr) bool {
				r.resolveRefExpr(mod, e, scope)
				return true
			})
		}
	}
}

func (r *Resolver) resolveRefType(mod *resolvedModule, typ *ast.TypeSpecifier, scope scopeStack) {
	for _, arg := range typ.TemplateArgs {
		ast.WalkExpr(arg, func(e ast.Expr) bool {
			r.resolveRefExpr(mod, e, scope)
			return true
		})
	}
	if isBuiltinType(typ.Name) {
		return
	}
	r.resolveRefName(mod, typ.Name, scope)
}

// resolveRefName marks name as used in mod, loading external modules as needed,
// and recurses into the declaration's own dependencies. The used map acts as a
// cycle guard so mutual references terminate.
func (r *Resolver) resolveRefName(mod *resolvedModule, name string, scope scopeStack) {
	if name == "" {
		return
	}

	if scope.has(name) {
		return
	}

	if d, ok := mod.symbols[name]; ok {
		// Local symbol takes precedence over builtins (e.g. alias f32 = ...).
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
		m := r.loadModule(file)
		if m == nil || m.used[i.sym] {
			return
		}
		m.used[i.sym] = true
		m.order = append(m.order, i.sym)
		if d := m.symbols[i.sym]; d != nil {
			r.resolveRefDecl(m, d, scope)
		}
	}
}

// resolveRefExpr is the per-node visitor used inside WalkExpr/WalkStmt.
func (r *Resolver) resolveRefExpr(mod *resolvedModule, e ast.Expr, scope scopeStack) {
	switch ex := e.(type) {
	case *ast.CallExpr:
		r.resolveRefExprName(mod, ex.Callee, scope)
	case *ast.Ident:
		r.resolveRefExprName(mod, ex.Name, scope)
	}
}

// resolveRefExprName handles both plain names and inline "a::b::sym" references.
func (r *Resolver) resolveRefExprName(mod *resolvedModule, name string, scope scopeStack) {
	if !strings.Contains(name, "::") {
		r.resolveRefName(mod, name, scope)
		return
	}
	filePath, sym := r.resolveQualifiedName(mod, name, mod.filePath)
	if filePath == "" {
		return
	}
	mod.inlineRefs[name] = fileSymbol{filePath, sym}
	m := r.loadModule(filePath)
	if m == nil || m.used[sym] {
		return
	}
	m.used[sym] = true
	m.order = append(m.order, sym)
	if d := m.symbols[sym]; d != nil {
		r.resolveRefDecl(m, d, scope)
	}
}

func (r *Resolver) resolveRefStructDecl(mod *resolvedModule, s *ast.StructDecl, scope scopeStack) {
	for _, m := range s.Members {
		if sf, ok := m.(*ast.StructMember); ok {
			r.resolveRefType(mod, &sf.Type, scope)
		}
	}
}

func (r *Resolver) resolveRefTypeAliasDecl(mod *resolvedModule, a *ast.TypeAliasDecl, scope scopeStack) {
	r.resolveRefType(mod, &a.Type, scope)
}

func (r *Resolver) resolveRefGlobalVarDecl(mod *resolvedModule, v *ast.GlobalVarDecl, scope scopeStack) {
	if v.Type != nil {
		r.resolveRefType(mod, v.Type, scope)
	}
	if v.Init != nil {
		ast.WalkExpr(v.Init, func(e ast.Expr) bool {
			r.resolveRefExpr(mod, e, scope)
			return true
		})
	}
}

func (r *Resolver) resolveRefGlobalValDecl(mod *resolvedModule, v *ast.GlobalValDecl, scope scopeStack) {
	if v.Type != nil {
		r.resolveRefType(mod, v.Type, scope)
	}
	if v.Init != nil {
		ast.WalkExpr(v.Init, func(e ast.Expr) bool {
			r.resolveRefExpr(mod, e, scope)
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

	// it's registered import
	if entry, ok := mod.imports[root]; ok {
		path := entry.path
		path = append(path, parts[:len(parts)-1]...)
		path = r.resolvePathSegs(path, sourceFile)
		return r.lookupFile(path), sym
	} else {
		path := r.resolvePathSegs(parts[:len(parts)-1], sourceFile)
		return r.lookupFile(path), sym
	}

	// pathParts := parts[:len(parts)-1]
	// if len(pathParts) == 0 {
	// 	return "", sym
	// }
	// // Module alias: single-segment prefix already registered in moduleMap.
	// if len(pathParts) == 1 {
	// 	if fp, ok := r.moduleMap[pathParts[0]]; ok {
	// 		return fp, sym
	// 	}
	// }
	// segs := r.resolvePathSegs(pathParts, sourceFile)
	// return r.lookupFile(segs), sym
	return "", sym
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

func rewriteDeclRefs(d ast.Decl, renames map[string]string) {
	if len(renames) == 0 {
		return
	}
	renameStr := func(s string) string {
		if v, ok := renames[s]; ok {
			return v
		}
		if idx := strings.LastIndex(s, "::"); idx >= 0 {
			if v, ok := renames[s[idx+2:]]; ok {
				return v
			}
		}
		return s
	}
	renameTS := func(ts *ast.TypeSpecifier) {
		if ts != nil {
			ts.Name = renameStr(ts.Name)
		}
	}
	rewriteExpr := func(e ast.Expr) bool {
		switch ex := e.(type) {
		case *ast.CallExpr:
			ex.Callee = renameStr(ex.Callee)
		case *ast.Ident:
			ex.Name = renameStr(ex.Name)
		}
		return true
	}
	rewriteStmt := func(s ast.Stmt) bool {
		switch sv := s.(type) {
		case *ast.VarStmt:
			renameTS(sv.Type)
		case *ast.ValStmt:
			renameTS(sv.Type)
		}
		return true
	}

	switch dd := d.(type) {
	case *ast.FuncDecl:
		for _, p := range dd.Params {
			if fp, ok := p.(*ast.FuncParam); ok {
				renameTS(&fp.Type)
			}
		}
		renameTS(dd.ReturnType)
		if dd.Body != nil {
			for _, s := range dd.Body.Stmts {
				ast.WalkStmt(s, rewriteStmt, rewriteExpr)
			}
		}
	case *ast.StructDecl:
		for _, m := range dd.Members {
			if sf, ok := m.(*ast.StructMember); ok {
				renameTS(&sf.Type)
			}
		}
	case *ast.GlobalValDecl:
		renameTS(dd.Type)
		ast.WalkExpr(dd.Init, rewriteExpr)
	case *ast.GlobalVarDecl:
		renameTS(dd.Type)
		if dd.Init != nil {
			ast.WalkExpr(dd.Init, rewriteExpr)
		}
	case *ast.TypeAliasDecl:
		renameTS(&dd.Type)
	}
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

func newScopeStack() scopeStack {
	return scopeStack{make(map[string]struct{})}
}

type scopeStack []map[string]struct{}

func (s scopeStack) push() {
	s = append(s, make(map[string]struct{}))
}

func (s scopeStack) add(name string) {
	s[len(s)-1][name] = struct{}{}
}

func (s scopeStack) has(name string) bool {
	for i := len(s) - 1; i > 0; i-- {
		if _, ok := s[i][name]; ok {
			return true
		}
	}
	return false
}
