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
	filename string // path of the source file that defines the symbol
	sym      string // local name of the symbol within that file
}

// importEntry records one explicit import from a module's import list.
type importEntry struct {
	path     []string // path segments before the final symbol name (including anchors)
	sym      string   // unaliased symbol name being imported
	filename string   // resolved source file path for the imported symbol; empty until resolved
}

// module holds all resolver state for a single source file.
type module struct {
	filename     string                 // path of the source file this module represents
	file         *ast.File              // AST of the source file (possibly rewritten)
	imports      map[string]importEntry // alias → import info for all imports in this file
	symbols      map[string]ast.Decl    // local symbol table mapping name → declaration
	used         map[string]bool        // symbols reachable from the entry point
	order        []string               // names in the order they were first marked used
	renames      map[*ast.Ident]bool    // idents that must be renamed during the apply pass
	constAsserts []*ast.ConstAssertStmt // const_assert statements collected from this module
}

// Resolver resolves imports and symbol references across a set of WESL source
// files, producing a single merged AST.
type Resolver struct {
	files    map[string]*ast.File // all known source files keyed by their path
	defines  map[string]bool      // compile-time feature flags used to evaluate @if conditions
	resolved map[string]*module   // cache of already-loaded modules keyed by file path
	order    []string             // file paths in pre-order load sequence (root is index 0)
}

// ResolveFile is a convenience function that constructs a Resolver and resolves
// the given entry-point file. files is a map of all source files the resolver
// may load, and defines is the set of active compile-time feature flags.
// It returns the merged AST or an error if resolution fails.
func ResolveFile(filename string, files map[string]*ast.File, defines map[string]bool) (*ast.File, error) {
	r := New(files, defines)
	return r.ResolveFile(filename)
}

// New creates a Resolver that can resolve imports across the given set of
// source files. files maps each file path to its parsed AST, and defines
// supplies the compile-time feature flags used to evaluate @if conditions.
func New(files map[string]*ast.File, defines map[string]bool) *Resolver {
	return &Resolver{
		files:    files,
		defines:  defines,
		resolved: make(map[string]*module),
	}
}

// ResolveFile resolves all imports and symbol references reachable from the
// named entry-point file and returns a single merged AST containing only the
// declarations that are actually used. It returns an error if any import or
// symbol cannot be found, or if a panic occurs during resolution.
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

// emitFile assembles the final merged AST from root and all transitively
// loaded modules. Declarations from imported modules are appended in load
// order after the root module's own declarations.
func (r *Resolver) emitFile(root *module) *ast.File {
	decls := root.file.Decls

	// Emit imported modules in load order (index 0 is root, skip it).
	for _, filename := range r.order[1:] {
		m := r.resolved[filename]

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

// renameSymbols walks every ident that was flagged for renaming and replaces
// its value with a collision-free output name. Root-module symbols keep their
// original names; imported symbols are mangled to avoid clashes.
func (r *Resolver) renameSymbols(root *module) {
	// Seed names with root-module symbols so imported symbols that mangle to
	// the same name get a numeric suffix instead of colliding.
	names := make(map[fileSymbol]string)
	for n := range root.symbols {
		names[fileSymbol{root.filename, n}] = n
	}

	// rename refs
	for _, filename := range r.order {
		m := r.resolved[filename]
		for ident := range m.renames {
			ident.Val = r.applyRename(m, ident, names, root.filename)
			ident.Path = nil
		}
	}
}

// loadModule loads and initializes the module for the given file path, or
// returns the already-cached module if it was loaded before. It returns nil
// when the file path is not found in the resolver's file map.
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
		filename: filename,
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

// resolveConditionals rewrites f by evaluating every @if conditional node
// against the resolver's defines map, replacing each conditional with the
// appropriate branch.
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

// buildSymbolAndImports populates mod.symbols and mod.imports by scanning the
// top-level declarations of the module's file. Import declarations are
// consumed and not retained in the file's declaration list.
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

// resolveRef marks root and all declarations it transitively depends on as
// used within mod, loading external modules as needed. scope tracks local
// variable bindings introduced by the current declaration so they are not
// mistaken for module-level symbol references.
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

		case *ast.ValStmt:
			scope.add(n.Name.Val)

		case *ast.BlockStmt:
			scope.push()
			ast.WalkList(n.Stmts, walk)
			scope.pop()
			return false

		case *ast.TypeSpecifier:
			if r.resolveName(mod, n.Name, scope) {
				mod.renames[n.Name] = true
			}
			for _, arg := range n.TemplateArgs {
				ast.Walk(arg, walk)
			}
			return false

		case *ast.Ident:
			if r.resolveName(mod, n, scope) {
				mod.renames[n] = true
			}
		}
		return true
	}

	ast.Walk(root, walk)
}

// lookupImportFile resolves the source file path for an importEntry relative
// to mod's own file. It first tries the import's path prefix alone, then
// falls back to appending the symbol name as the last path segment.
func (r *Resolver) lookupImportFile(mod *module, i importEntry) string {
	if filename := r.lookupPath(i.path, mod.filename); filename != "" {
		return filename
	}
	// Fallback: sym may be the last path segment of a module file path.
	return r.lookupPath(append(i.path, i.sym), mod.filename)
}

// resolveName marks ident as used in mod, loading external modules as needed,
// and recurses into the declaration's own dependencies. Handles both plain
// names and inline qualified references like "package::foo::MyType".
// The used map acts as a cycle guard so mutual references terminate.
// It returns true when ident was resolved to a module-level symbol that
// requires renaming.
func (r *Resolver) resolveName(mod *module, ident *ast.Ident, scope *scopeStack) bool {
	if ident == nil || ident.Val == "" {
		return false
	}

	if len(ident.Path) > 0 {
		filename, sym := r.resolveQualifiedName(mod, ident)
		if filename == "" {
			return false
		}
		m := r.loadModule(filename)
		if m != nil && !m.used[sym] {
			if d := m.symbols[sym]; d != nil {
				r.resolveRef(m, d)
			} else {
				panic("symbol not found")
			}
		}
		return true
	}

	name := ident.Val
	if scope.has(name) {
		return false
	}

	if d, ok := mod.symbols[name]; ok {
		if !mod.used[name] {
			r.resolveRef(mod, d)
		}
		return true
	}

	if ast.IsBuiltinType(name) {
		return false
	}

	if i, ok := mod.imports[name]; ok {
		filename := r.lookupImportFile(mod, i)
		m := r.loadModule(filename)

		i.filename = filename
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

	return false
}

// applyRename computes the collision-free output name for ident within mod.
// rootFilename is the entry-point file's path; symbols defined there keep
// their original names while symbols from other files are mangled.
func (r *Resolver) applyRename(mod *module, ident *ast.Ident, names map[fileSymbol]string, rootFilename string) string {
	if len(ident.Path) > 0 {
		filename, sym := r.resolveQualifiedName(mod, ident)
		if filename == "" {
			return ident.Val
		}
		if filename == rootFilename {
			return sym
		}
		return r.nameFor(filename, sym, names)
	}

	name := ident.Val
	if _, ok := mod.symbols[name]; ok {
		if mod.filename == rootFilename {
			return name
		}
		return r.nameFor(mod.filename, name, names)
	}

	if i, ok := mod.imports[name]; ok {
		if i.filename == "" {
			return name
		}
		if i.filename == rootFilename {
			return i.sym
		}
		return r.nameFor(i.filename, i.sym, names)
	}
	return name
}

// nameFor returns the collision-free output name for the (filename, sym) pair,
// allocating a mangled name on first call and caching it for consistent reuse
// across all references to the same symbol.
func (r *Resolver) nameFor(filename, sym string, names map[fileSymbol]string) string {
	key := fileSymbol{filename, sym}
	if n, ok := names[key]; ok {
		return n
	}
	base := mangleName(filename, sym)
	n := base
	for i := 0; r.nameTaken(n, names); i++ {
		n = fmt.Sprintf("%s_%d", base, i)
	}
	names[key] = n
	return n
}

// nameTaken reports whether name is already assigned to any symbol in names.
func (r *Resolver) nameTaken(name string, names map[fileSymbol]string) bool {
	for _, v := range names {
		if v == name {
			return true
		}
	}
	return false
}

// resolveQualifiedName resolves a qualified ident (whose Path is non-empty) to
// a (filename, sym) pair. It consults the module's import table to expand
// import aliases before performing a path lookup.
func (r *Resolver) resolveQualifiedName(mod *module, ident *ast.Ident) (string, string) {
	sym := ident.Val
	root := ident.Path[0]
	rest := ident.Path[1:]

	if entry, ok := mod.imports[root]; ok {
		p := append(append([]string{}, entry.path...), rest...)
		if filename := r.lookupPath(p, mod.filename); filename != "" {
			return filename, sym
		}
		// Fallback: the import alias may point to a module file whose name
		// is entry.sym (e.g. import package::dir::modname → file "dir/modname").
		return r.lookupPath(append(p, entry.sym), mod.filename), sym
	}
	return r.lookupPath(ident.Path, mod.filename), sym
}

// lookupFile searches the resolver's file map for the longest prefix of segs
// that matches a known file path, returning that path or "" if none is found.
func (r *Resolver) lookupFile(segs []string) string {
	for i := len(segs); i >= 1; i-- {
		candidate := strings.Join(segs[:i], "/")
		if _, ok := r.files[candidate]; ok {
			return candidate
		}
	}
	return ""
}

// lookupPath resolves an import path prefix (which may contain "package" and
// "super" anchors) relative to sourceFilename and returns the matching file
// path, or "" if no registered file matches.
func (r *Resolver) lookupPath(prefix []string, sourceFilename string) string {
	dir := path.Dir(sourceFilename)
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

// scopeStack tracks lexically scoped local variable names introduced by
// let/var declarations within a function body.
type scopeStack struct {
	blocks []map[string]struct{} // stack of block-level name sets; index 0 is the outermost scope
}

// newScopeStack creates a scopeStack with a single empty outermost scope.
func newScopeStack() *scopeStack {
	return &scopeStack{blocks: []map[string]struct{}{make(map[string]struct{})}}
}

// push adds a new inner scope to the stack, used when entering a block statement.
func (s *scopeStack) push() {
	s.blocks = append(s.blocks, make(map[string]struct{}))
}

// pop removes the innermost scope from the stack, used when leaving a block statement.
func (s *scopeStack) pop() {
	s.blocks = s.blocks[:len(s.blocks)-1]
}

// add registers name in the innermost scope of the stack.
func (s *scopeStack) add(name string) {
	s.blocks[len(s.blocks)-1][name] = struct{}{}
}

// has reports whether name appears in any scope except the outermost (global)
// scope, which holds module-level symbols rather than local variables.
func (s scopeStack) has(name string) bool {
	// Don't check the global namespace
	for i := len(s.blocks) - 1; i > 0; i-- {
		if _, ok := s.blocks[i][name]; ok {
			return true
		}
	}
	return false
}

// getNodeName returns the name identifier for named declaration nodes
// (functions, structs, type aliases, let/var statements). It returns nil for
// node types that do not carry a name.
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

// mangleName produces a unique output name for sym defined in filename by
// encoding the file path into the identifier, replacing path separators with
// underscores.
func mangleName(filename, sym string) string {
	return "package_" + strings.ReplaceAll(filename, "/", "_") + "_" + sym
}

// pickBranch evaluates cond against defines and returns then if the condition
// is true, or els otherwise. T must be an ast.Node so that the result can be
// used directly as a replacement node in a Rewrite callback.
func pickBranch[T ast.Node](cond ast.Expr, then, els T, defines map[string]bool) T {
	if evalCondition(cond, defines) {
		return then
	}
	return els
}

// evalCondition evaluates a compile-time boolean expression against the
// provided defines map and returns the result. Identifiers are looked up in
// defines (missing keys evaluate to false). Supported operators are !, &&,
// ||, ==, and !=. Parenthesized and literal expressions are also handled.
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
