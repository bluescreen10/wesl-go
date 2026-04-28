package wesl

import (
	"fmt"
	"strings"

	"github.com/bluescreen10/wesl-go/ast"
)

// fileSymbol uniquely identifies a symbol within a specific source file.
type fileSymbol struct {
	file string
	sym  string
}

// importEntry is one unit of work: bring symbol from srcFile into output as outputName.
type importEntry struct {
	srcFile    string
	sym        string // original name in srcFile
	outputName string // desired name in output (alias or original)
}

type Resolver struct {
	files              map[string]*ast.File
	defines            map[string]bool
	symbols            map[string]SymbolTable
	taken              map[string]bool             // names currently in output namespace
	assigned           map[fileSymbol]string       // (file,sym) -> final output name; "" = in-progress (cycle guard)
	depMap             map[fileSymbol][]fileSymbol // cached deps per (actualFile,actualSym) from Phase 1
	emitted            map[fileSymbol]bool         // (file,sym) -> already written to output
	moduleMap          map[string]string           // local module alias -> file path (for module-style imports)
	constAssertedFiles map[string]bool             // files whose const_asserts have been emitted
	mainFile           string
}

func NewResolver(files map[string]*ast.File, defines map[string]bool) *Resolver {
	return &Resolver{
		files:              files,
		defines:            defines,
		symbols:            make(map[string]SymbolTable),
		taken:              make(map[string]bool),
		assigned:           make(map[fileSymbol]string),
		depMap:             make(map[fileSymbol][]fileSymbol),
		emitted:            make(map[fileSymbol]bool),
		moduleMap:          make(map[string]string),
		constAssertedFiles: make(map[string]bool),
	}
}

// ── Entry point ───────────────────────────────────────────────────────────────

func (r *Resolver) Resolve(mainFile string) *ast.File {
	r.mainFile = mainFile

	// 1. Resolve @if/@else in every file.
	for path, f := range r.files {
		r.files[path] = ResolveFile(f, r.defines)
	}

	// 2. Build per-file symbol tables (after conditional resolution).
	for path, f := range r.files {
		r.symbols[path] = BuildSymbolTable(f, path)
	}

	mainAST := r.files[mainFile]
	if mainAST == nil {
		return &ast.File{}
	}

	// 3. Seed 'taken' with every name defined locally in main (non-import decls).
	for _, d := range mainAST.Decls {
		if _, ok := d.(*ast.ImportDecl); ok {
			continue
		}
		if n := r.declName(d); n != "" {
			r.taken[n] = true
		}
	}

	// 4. Collect all imports: explicit ImportDecl items + inline package:: refs.
	entries, mainRenames, deferredAliases := r.collectImports(mainFile)

	// 5. Phase 1 – assign output names for primaries and all transitive deps.
	for _, e := range entries {
		r.assignName(e.srcFile, e.sym, e.outputName)
	}

	// 6. Build the complete rename map for main file references.
	//    Merge assigned names for explicit imports into mainRenames.
	for _, e := range entries {
		key := fileSymbol{e.srcFile, e.sym}
		if out, ok := r.assigned[key]; ok && out != "" {
			mainRenames[e.outputName] = out
		}
	}
	// Resolve deferred aliases: same symbol imported under two different names.
	for _, da := range deferredAliases {
		if out, ok := r.assigned[da.key]; ok && out != "" {
			mainRenames[da.alias] = out
		}
	}

	// 7. Rewrite + keep local (non-import) decls from main.
	var localDecls []ast.Decl
	for _, d := range mainAST.Decls {
		if _, ok := d.(*ast.ImportDecl); ok {
			continue
		}
		cloned := r.cloneDecl(d)
		r.rewriteDeclRefs(cloned, mainRenames)
		localDecls = append(localDecls, cloned)
	}

	// 8. Emit primaries first, then deps.
	var importedDecls []ast.Decl

	for _, e := range entries {
		key := fileSymbol{e.srcFile, e.sym}
		if r.emitted[key] {
			continue
		}
		r.emitted[key] = true
		r.emitFileConstAsserts(e.srcFile, &importedDecls)
		decl := r.buildDecl(e.srcFile, e.sym)
		if decl != nil {
			importedDecls = append(importedDecls, decl)
		}
	}
	for _, e := range entries {
		r.emitDeps(e.srcFile, e.sym, &importedDecls)
	}

	return &ast.File{Decls: append(localDecls, importedDecls...)}
}

// ── Import collection ─────────────────────────────────────────────────────────

// deferredAlias records an alias for a symbol that was imported more than once.
// Its mapping must be resolved after Phase 1 (assignName) completes.
type deferredAlias struct {
	key   fileSymbol
	alias string
}

// collectImports returns importEntry list, a rename map for the main file, and
// deferred aliases for duplicate imports (same symbol imported under two names).
func (r *Resolver) collectImports(filePath string) ([]importEntry, map[string]string, []deferredAlias) {
	f := r.files[filePath]
	renames := map[string]string{}
	var entries []importEntry
	var deferred []deferredAlias
	seen := map[fileSymbol]bool{}

	for _, d := range f.Decls {
		imp, ok := d.(*ast.ImportDecl)
		if !ok {
			continue
		}
		for _, item := range imp.Items {
			srcFile, origSym := r.resolveImportItem(imp.Path, item)
			if srcFile == "" {
				// Try as module import (no specific symbol, just a namespace).
				r.registerModuleImport(imp.Path, item)
				continue
			}
			desiredName := item.Alias
			if desiredName == "" {
				desiredName = item.Path[len(item.Path)-1]
			}
			key := fileSymbol{srcFile, origSym}
			if seen[key] {
				// Same symbol imported twice: record the alias to resolve after Phase 1.
				deferred = append(deferred, deferredAlias{key, desiredName})
				continue
			}
			seen[key] = true
			entries = append(entries, importEntry{srcFile, origSym, desiredName})
		}
	}

	// Inline references: package::file::sym or super::file::sym anywhere in non-import decls.
	for _, d := range f.Decls {
		if _, ok := d.(*ast.ImportDecl); ok {
			continue
		}
		r.collectInlineRefs(d, &entries, renames, seen)
	}

	return entries, renames, deferred
}

// registerModuleImport records a module-style import: "import super::file1" or
// "import package::file1::foo" where file1/foo.wgsl exists.
func (r *Resolver) registerModuleImport(prefix []string, item ast.ImportItem) {
	modName := item.Path[len(item.Path)-1]
	// Build the file path for the module.
	var segs []string
	for _, s := range prefix {
		if s != "package" && s != "super" {
			segs = append(segs, s)
		}
	}
	for _, s := range item.Path {
		segs = append(segs, s)
	}
	candidate := "./" + strings.Join(segs, "/") + ".wgsl"
	if _, ok := r.files[candidate]; ok {
		r.moduleMap[modName] = candidate
		return
	}
	// Also try without the last segment as the file name, the last is the module alias.
	if len(segs) > 0 {
		candidate2 := "./" + strings.Join(segs[:len(segs)-1], "/") + ".wgsl"
		if _, ok := r.files[candidate2]; ok {
			r.moduleMap[modName] = candidate2
		}
	}
}

// collectInlineRefs walks a decl looking for qualified names (package::x::y or module::y)
// and registers them as additional imports.
func (r *Resolver) collectInlineRefs(d ast.Decl, entries *[]importEntry, renames map[string]string, seen map[fileSymbol]bool) {
	var walkExpr func(e ast.Expr)
	walkExpr = func(e ast.Expr) {
		if e == nil {
			return
		}
		switch ex := e.(type) {
		case *ast.CallExpr:
			r.tryAddInlineRef(ex.Callee, entries, renames, seen)
			for _, a := range ex.Args {
				walkExpr(a)
			}
			for _, a := range ex.TemplateArgs {
				walkExpr(a)
			}
		case *ast.Ident:
			r.tryAddInlineRef(ex.Name, entries, renames, seen)
		case *ast.BinaryExpr:
			walkExpr(ex.Left)
			walkExpr(ex.Right)
		case *ast.UnaryExpr:
			walkExpr(ex.Operand)
		case *ast.IndexExpr:
			walkExpr(ex.Base)
			walkExpr(ex.Index)
		case *ast.MemberExpr:
			walkExpr(ex.Base)
		case *ast.AddrOfExpr:
			walkExpr(ex.Operand)
		case *ast.DerefExpr:
			walkExpr(ex.Operand)
		case *ast.ParenExpr:
			walkExpr(ex.Inner)
		}
	}
	var walkStmt func(s ast.Stmt)
	walkStmt = func(s ast.Stmt) {
		if s == nil {
			return
		}
		switch st := s.(type) {
		case *ast.FnCallStmt:
			walkExpr(&st.Call)
		case *ast.AssignmentStmt:
			walkExpr(st.LHS)
			walkExpr(st.RHS)
		case *ast.ReturnStmt:
			walkExpr(st.Value)
		case *ast.VarStmt:
			walkExpr(st.Init)
		case *ast.ValStmt:
			walkExpr(st.Init)
		case *ast.CompoundStmt:
			for _, s2 := range st.Stmts {
				walkStmt(s2)
			}
		case *ast.IfStmt:
			walkExpr(st.Cond)
			for _, s2 := range st.Then.Stmts {
				walkStmt(s2)
			}
			if st.ElseIf != nil {
				walkStmt(st.ElseIf)
			}
			if st.Else != nil {
				for _, s2 := range st.Else.Stmts {
					walkStmt(s2)
				}
			}
		case *ast.ForStmt:
			walkStmt(st.Init)
			walkExpr(st.Cond)
			walkStmt(st.Update)
			if st.Body != nil {
				for _, s2 := range st.Body.Stmts {
					walkStmt(s2)
				}
			}
		case *ast.WhileStmt:
			walkExpr(st.Cond)
			for _, s2 := range st.Body.Stmts {
				walkStmt(s2)
			}
		case *ast.LoopStmt:
			for _, s2 := range st.Body.Stmts {
				walkStmt(s2)
			}
		case *ast.SwitchStmt:
			walkExpr(st.Expr)
			for _, cl := range st.Clauses {
				cl := cl.(*ast.CaseClause)
				for _, s2 := range cl.Body.Stmts {
					walkStmt(s2)
				}
			}
		case *ast.IncDecStmt:
			walkExpr(st.LHS)
		}
	}

	switch dd := d.(type) {
	case *ast.FuncDecl:
		if dd.Body != nil {
			for _, s := range dd.Body.Stmts {
				walkStmt(s)
			}
		}
	case *ast.GlobalValDecl:
		walkExpr(dd.Init)
	case *ast.GlobalVarDecl:
		if dd.Init != nil {
			walkExpr(dd.Init)
		}
	}
}

func (r *Resolver) tryAddInlineRef(name string, entries *[]importEntry, renames map[string]string, seen map[fileSymbol]bool) {
	if !strings.Contains(name, "::") {
		return
	}
	parts := strings.Split(name, "::")
	// Check if first segment is "package" or "super" – direct file reference.
	if parts[0] == "package" || parts[0] == "super" {
		if len(parts) < 3 {
			return
		}
		sym := parts[len(parts)-1]
		fileParts := parts[1 : len(parts)-1]
		filePath := r.lookupFile(fileParts)
		if filePath == "" {
			return
		}
		key := fileSymbol{filePath, sym}
		if !seen[key] {
			seen[key] = true
			*entries = append(*entries, importEntry{filePath, sym, sym})
		}
		renames[name] = sym
		return
	}
	// Check if first segment is a known module alias.
	if filePath, ok := r.moduleMap[parts[0]]; ok {
		sym := parts[len(parts)-1]
		key := fileSymbol{filePath, sym}
		if !seen[key] {
			seen[key] = true
			*entries = append(*entries, importEntry{filePath, sym, sym})
		}
		renames[name] = sym
	}
}

// ── File lookup ───────────────────────────────────────────────────────────────

// lookupFile finds the file path for a slice of path segments (without package/super).
func (r *Resolver) lookupFile(segs []string) string {
	// Try from most specific to least.
	for i := len(segs); i >= 1; i-- {
		candidate := "./" + strings.Join(segs[:i], "/") + ".wgsl"
		if _, ok := r.files[candidate]; ok {
			return candidate
		}
	}
	return ""
}

// resolveImportItem returns (filePath, originalSymbolName) for an import item.
// Returns ("","") when the item refers to a module (no specific symbol).
func (r *Resolver) resolveImportItem(prefix []string, item ast.ImportItem) (string, string) {
	sym := item.Path[len(item.Path)-1]

	// Build file path segments: strip "package"/"super" from prefix, add item sub-path.
	var segs []string
	for _, s := range prefix {
		if s != "package" && s != "super" {
			segs = append(segs, s)
		}
	}
	// item.Path[0..len-2] are sub-directories; last is the symbol.
	segs = append(segs, item.Path[:len(item.Path)-1]...)

	filePath := r.lookupFile(segs)
	if filePath != "" {
		return filePath, sym
	}
	// If no file found yet, try treating the symbol itself as a file segment.
	// (e.g., import super::file1 where "file1" is the module)
	segs2 := append(segs, sym)
	if fp := r.lookupFile(segs2); fp != "" {
		// This is a module import, not a symbol import.
		return "", ""
	}
	return "", ""
}

// ── Name assignment (Phase 1) ─────────────────────────────────────────────────

// assignName ensures (srcFile, sym) has a unique output name.
// preferredName is the desired name (alias or original); empty means use sym.
func (r *Resolver) assignName(srcFile, sym, preferredName string) string {
	key := fileSymbol{srcFile, sym}

	if existing, ok := r.assigned[key]; ok {
		return existing // already assigned (or in-progress cycle guard)
	}
	// Cycle guard: mark as in-progress before any recursion.
	r.assigned[key] = ""

	// Find the actual decl (might be in srcFile's own decls or via its imports).
	actualFile, actualSym := r.resolveSymbol(srcFile, sym)
	if actualFile == "" {
		return ""
	}
	actualKey := fileSymbol{actualFile, actualSym}

	// When resolved through an import, actualKey differs from key.
	if actualKey != key {
		if existing, ok := r.assigned[actualKey]; ok {
			r.assigned[key] = existing
			return existing
		}
		r.assigned[actualKey] = "" // cycle guard for the resolved key too
	}

	chosen := preferredName
	if chosen == "" {
		chosen = actualSym
	}
	chosen = r.makeUnique(chosen)
	r.assigned[key] = chosen
	r.assigned[actualKey] = chosen
	r.taken[chosen] = true

	// Cache deps and recursively assign names. Caching here (before recursion) lets
	// emitDeps use the original dep names even after buildDecl mutates the cloned AST.
	decl := r.findDeclInFile(r.files[actualFile], actualSym)
	if decl != nil {
		deps := r.collectDeps(decl, r.files[actualFile], actualFile)
		r.depMap[actualKey] = deps
		for _, dep := range deps {
			r.assignName(dep.file, dep.sym, "")
		}
	}

	return chosen
}

// resolveSymbol looks up sym in srcFile: first in its own decls, then through
// its import declarations. Returns (filePath, originalSymbolName).
func (r *Resolver) resolveSymbol(srcFilePath, sym string) (string, string) {
	srcFile := r.files[srcFilePath]
	if srcFile == nil {
		return "", ""
	}
	if r.findDeclInFile(srcFile, sym) != nil {
		return srcFilePath, sym
	}
	// Check srcFile's import declarations.
	for _, d := range srcFile.Decls {
		imp, ok := d.(*ast.ImportDecl)
		if !ok {
			continue
		}
		for _, item := range imp.Items {
			importedName := item.Alias
			if importedName == "" {
				importedName = item.Path[len(item.Path)-1]
			}
			if importedName == sym {
				fp, origSym := r.resolveImportItem(imp.Path, item)
				if fp != "" {
					return fp, origSym
				}
			}
		}
	}
	return "", ""
}

// ── Dependency collection ─────────────────────────────────────────────────────

// collectDeps returns all (file, sym) pairs that decl depends on.
func (r *Resolver) collectDeps(decl ast.Decl, srcFile *ast.File, srcFilePath string) []fileSymbol {
	names := r.referencedNames(decl)
	var deps []fileSymbol
	seen := map[string]bool{}
	for _, name := range names {
		if seen[name] {
			continue
		}
		seen[name] = true
		fp, sym := r.resolveSymbol(srcFilePath, name)
		if fp != "" {
			deps = append(deps, fileSymbol{fp, sym})
		}
	}
	return deps
}

// scopeStack tracks locally-defined names for scope-aware dependency collection.
// Each entry is one block scope; the last entry is the innermost.
type scopeStack []map[string]bool

func newScopeStack() scopeStack          { return scopeStack{make(map[string]bool)} }
func (ss scopeStack) push() scopeStack   { return append(ss, make(map[string]bool)) }
func (ss scopeStack) pop() scopeStack    { return ss[:len(ss)-1] }
func (ss scopeStack) define(name string) { ss[len(ss)-1][name] = true }
func (ss scopeStack) has(name string) bool {
	for i := len(ss) - 1; i >= 0; i-- {
		if ss[i][name] {
			return true
		}
	}
	return false
}

// referencedNames collects all names a decl externally depends on.
// For function bodies it is scope-aware: names defined by local var/let/const
// statements (or parameters) shadow outer symbols and are not reported as deps.
func (r *Resolver) referencedNames(decl ast.Decl) []string {
	var names []string

	// addName adds name as a potential dep only when it is not locally shadowed.
	// scopes is passed by value so closures capture a snapshot of the current stack.
	addName := func(name string, scopes scopeStack) {
		if strings.Contains(name, "::") {
			return // qualified names handled elsewhere
		}
		if !scopes.has(name) {
			names = append(names, name)
		}
	}

	var walkExpr func(e ast.Expr, sc scopeStack)
	walkExpr = func(e ast.Expr, sc scopeStack) {
		if e == nil {
			return
		}
		switch ex := e.(type) {
		case *ast.Ident:
			addName(ex.Name, sc)
		case *ast.CallExpr:
			callee := ex.Callee
			if strings.Contains(callee, "::") {
				parts := strings.Split(callee, "::")
				callee = parts[len(parts)-1]
			}
			addName(callee, sc)
			for _, a := range ex.Args {
				walkExpr(a, sc)
			}
			for _, a := range ex.TemplateArgs {
				walkExpr(a, sc)
			}
		case *ast.BinaryExpr:
			walkExpr(ex.Left, sc)
			walkExpr(ex.Right, sc)
		case *ast.UnaryExpr:
			walkExpr(ex.Operand, sc)
		case *ast.IndexExpr:
			walkExpr(ex.Base, sc)
			walkExpr(ex.Index, sc)
		case *ast.MemberExpr:
			walkExpr(ex.Base, sc)
		case *ast.AddrOfExpr:
			walkExpr(ex.Operand, sc)
		case *ast.DerefExpr:
			walkExpr(ex.Operand, sc)
		case *ast.ParenExpr:
			walkExpr(ex.Inner, sc)
		}
	}

	// walkStmts and walkStmt are mutually recursive; declare before assigning.
	var walkStmt func(s ast.Stmt, sc *scopeStack)
	var walkStmts func(stmts []ast.Stmt, sc *scopeStack)
	walkStmts = func(stmts []ast.Stmt, sc *scopeStack) {
		for _, s := range stmts {
			walkStmt(s, sc)
		}
	}
	walkStmt = func(s ast.Stmt, sc *scopeStack) {
		if s == nil {
			return
		}
		switch st := s.(type) {
		case *ast.FnCallStmt:
			walkExpr(&st.Call, *sc)
		case *ast.AssignmentStmt:
			walkExpr(st.LHS, *sc)
			walkExpr(st.RHS, *sc)
		case *ast.ReturnStmt:
			walkExpr(st.Value, *sc)
		case *ast.VarStmt:
			walkExpr(st.Init, *sc)
			if st.Type != nil {
				addName(st.Type.Name, *sc)
			}
			(*sc).define(st.Name)
		case *ast.ValStmt:
			// Process the init and type first (they see the scope BEFORE the new binding).
			walkExpr(st.Init, *sc)
			if st.Type != nil {
				addName(st.Type.Name, *sc)
			}
			(*sc).define(st.Name)
		case *ast.CompoundStmt:
			// Nested block: new scope, but share sequential side-effects within it.
			nested := (*sc).push()
			walkStmts(st.Stmts, &nested)
			// Don't pop sc; the nested scope is local to this block.
		case *ast.IfStmt:
			walkExpr(st.Cond, *sc)
			if st.Then != nil {
				nested := (*sc).push()
				walkStmts(st.Then.Stmts, &nested)
			}
			if st.ElseIf != nil {
				walkStmt(st.ElseIf, sc)
			}
			if st.Else != nil {
				nested := (*sc).push()
				walkStmts(st.Else.Stmts, &nested)
			}
		case *ast.ForStmt:
			// for loop gets its own scope (init var is local to the loop).
			loopSc := (*sc).push()
			walkStmt(st.Init, &loopSc)
			walkExpr(st.Cond, loopSc)
			walkStmt(st.Update, &loopSc)
			if st.Body != nil {
				bodySc := loopSc.push()
				walkStmts(st.Body.Stmts, &bodySc)
			}
		case *ast.WhileStmt:
			walkExpr(st.Cond, *sc)
			if st.Body != nil {
				nested := (*sc).push()
				walkStmts(st.Body.Stmts, &nested)
			}
		case *ast.LoopStmt:
			if st.Body != nil {
				nested := (*sc).push()
				walkStmts(st.Body.Stmts, &nested)
			}
		case *ast.ContinuingStmt:
			if st.Body != nil {
				nested := (*sc).push()
				walkStmts(st.Body.Stmts, &nested)
			}
		case *ast.SwitchStmt:
			walkExpr(st.Expr, *sc)
			for _, cl := range st.Clauses {
				cl := cl.(*ast.CaseClause)
				nested := (*sc).push()
				walkStmts(cl.Body.Stmts, &nested)
			}
		case *ast.IncDecStmt:
			walkExpr(st.LHS, *sc)
		}
	}

	switch dd := decl.(type) {
	case *ast.FuncDecl:
		// Function scope: parameters are defined before the body.
		sc := newScopeStack()
		for _, p := range dd.Params {
			if fp, ok := p.(*ast.FuncParam); ok {
				addName(fp.Type.Name, sc)
				for _, ta := range fp.Type.TemplateArgs {
					walkExpr(ta, sc)
				}
				sc.define(fp.Name) // param name shadows anything from outer scope
			}
		}
		if dd.ReturnType != nil {
			addName(dd.ReturnType.Name, sc)
			for _, ta := range dd.ReturnType.TemplateArgs {
				walkExpr(ta, sc)
			}
		}
		if dd.Body != nil {
			bodySc := sc.push()
			walkStmts(dd.Body.Stmts, &bodySc)
		}

	case *ast.StructDecl:
		sc := newScopeStack()
		for _, m := range dd.Members {
			if sf, ok := m.(*ast.StructField); ok {
				addName(sf.Type.Name, sc)
				for _, ta := range sf.Type.TemplateArgs {
					walkExpr(ta, sc)
				}
			}
		}

	case *ast.GlobalValDecl:
		sc := newScopeStack()
		if dd.Type != nil {
			addName(dd.Type.Name, sc)
		}
		walkExpr(dd.Init, sc)

	case *ast.GlobalVarDecl:
		sc := newScopeStack()
		if dd.Type != nil {
			addName(dd.Type.Name, sc)
		}
		if dd.Init != nil {
			walkExpr(dd.Init, sc)
		}

	case *ast.TypeAliasDecl:
		sc := newScopeStack()
		addName(dd.Type.Name, sc)
	}
	return names
}

// ── Emission (Phase 2) ────────────────────────────────────────────────────────

// buildDecl returns the decl for (srcFile, sym) with its output name and dep renames applied.
func (r *Resolver) buildDecl(srcFile, sym string) ast.Decl {
	actualFile, actualSym := r.resolveSymbol(srcFile, sym)
	if actualFile == "" {
		return nil
	}
	decl := r.findDeclInFile(r.files[actualFile], actualSym)
	if decl == nil {
		return nil
	}
	actualKey := fileSymbol{actualFile, actualSym}
	outputName := r.assigned[actualKey]

	// Build rename map from cached deps (avoids re-deriving from a possibly-mutated decl).
	depRenames := map[string]string{}
	for _, dep := range r.depMap[actualKey] {
		depOut := r.assigned[dep]
		if depOut != "" && depOut != dep.sym {
			depRenames[dep.sym] = depOut
		}
	}

	cloned := r.cloneDecl(decl)
	r.renameDeclHeader(cloned, outputName)
	r.rewriteDeclRefs(cloned, depRenames)
	return cloned
}

// emitFileConstAsserts emits all const_assert declarations from filePath into output,
// but only once per file (tracked by constAssertedFiles).
func (r *Resolver) emitFileConstAsserts(filePath string, output *[]ast.Decl) {
	if r.constAssertedFiles[filePath] {
		return
	}
	r.constAssertedFiles[filePath] = true
	f := r.files[filePath]
	if f == nil {
		return
	}
	for _, d := range f.Decls {
		if _, ok := d.(*ast.ConstAssertDecl); ok {
			*output = append(*output, d)
		}
	}
}

// emitDeps emits transitive dependencies of (srcFile, sym) into output (after primaries).
func (r *Resolver) emitDeps(srcFile, sym string, output *[]ast.Decl) {
	actualFile, actualSym := r.resolveSymbol(srcFile, sym)
	if actualFile == "" {
		return
	}
	actualKey := fileSymbol{actualFile, actualSym}
	for _, dep := range r.depMap[actualKey] {
		if r.emitted[dep] {
			continue
		}
		r.emitted[dep] = true
		r.emitFileConstAsserts(dep.file, output)
		built := r.buildDecl(dep.file, dep.sym)
		if built != nil {
			*output = append(*output, built)
		}
		r.emitDeps(dep.file, dep.sym, output)
	}
}

// ── Unique naming ─────────────────────────────────────────────────────────────

func (r *Resolver) makeUnique(base string) string {
	if !r.taken[base] {
		return base
	}
	for i := 0; ; i++ {
		candidate := fmt.Sprintf("%s%d", base, i)
		if !r.taken[candidate] {
			return candidate
		}
	}
}

// ── AST helpers ───────────────────────────────────────────────────────────────

func (r *Resolver) declName(d ast.Decl) string {
	switch dd := d.(type) {
	case *ast.FuncDecl:
		return dd.Name
	case *ast.StructDecl:
		return dd.Name
	case *ast.GlobalValDecl:
		return dd.Name
	case *ast.GlobalVarDecl:
		return dd.Name
	case *ast.TypeAliasDecl:
		return dd.Name
	}
	return ""
}

func (r *Resolver) findDeclInFile(f *ast.File, sym string) ast.Decl {
	if f == nil {
		return nil
	}
	for _, d := range f.Decls {
		if r.declName(d) == sym {
			return d
		}
	}
	return nil
}

// cloneDecl does a shallow structural copy sufficient for renaming.
func (r *Resolver) cloneDecl(d ast.Decl) ast.Decl {
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

// renameDeclHeader sets the declaration's own name to outputName.
func (r *Resolver) renameDeclHeader(d ast.Decl, outputName string) {
	switch dd := d.(type) {
	case *ast.FuncDecl:
		dd.Name = outputName
	case *ast.StructDecl:
		dd.Name = outputName
	case *ast.GlobalValDecl:
		dd.Name = outputName
	case *ast.GlobalVarDecl:
		dd.Name = outputName
	case *ast.TypeAliasDecl:
		dd.Name = outputName
	}
}

// rewriteDeclRefs rewrites all name references inside a decl using renames map.
func (r *Resolver) rewriteDeclRefs(d ast.Decl, renames map[string]string) {
	if len(renames) == 0 {
		return
	}
	renameStr := func(s string) string {
		if v, ok := renames[s]; ok {
			return v
		}
		// Also handle qualified names: "pkg::file::sym" -> lookup sym or full name.
		if strings.Contains(s, "::") {
			parts := strings.Split(s, "::")
			last := parts[len(parts)-1]
			if v, ok := renames[s]; ok {
				return v
			}
			if v, ok := renames[last]; ok {
				return v
			}
		}
		return s
	}
	renameTS := func(ts *ast.TypeSpecifier) {
		if ts == nil {
			return
		}
		ts.Name = renameStr(ts.Name)
	}

	var rewriteExpr func(e ast.Expr)
	rewriteExpr = func(e ast.Expr) {
		if e == nil {
			return
		}
		switch ex := e.(type) {
		case *ast.CallExpr:
			ex.Callee = renameStr(ex.Callee)
			for _, a := range ex.Args {
				rewriteExpr(a)
			}
			for _, a := range ex.TemplateArgs {
				rewriteExpr(a)
			}
		case *ast.Ident:
			ex.Name = renameStr(ex.Name)
		case *ast.BinaryExpr:
			rewriteExpr(ex.Left)
			rewriteExpr(ex.Right)
		case *ast.UnaryExpr:
			rewriteExpr(ex.Operand)
		case *ast.IndexExpr:
			rewriteExpr(ex.Base)
			rewriteExpr(ex.Index)
		case *ast.MemberExpr:
			rewriteExpr(ex.Base)
		case *ast.AddrOfExpr:
			rewriteExpr(ex.Operand)
		case *ast.DerefExpr:
			rewriteExpr(ex.Operand)
		case *ast.ParenExpr:
			rewriteExpr(ex.Inner)
		}
	}

	var rewriteStmt func(s ast.Stmt)
	rewriteStmt = func(s ast.Stmt) {
		if s == nil {
			return
		}
		switch st := s.(type) {
		case *ast.FnCallStmt:
			rewriteExpr(&st.Call)
		case *ast.AssignmentStmt:
			rewriteExpr(st.LHS)
			rewriteExpr(st.RHS)
		case *ast.ReturnStmt:
			rewriteExpr(st.Value)
		case *ast.VarStmt:
			rewriteExpr(st.Init)
			renameTS(st.Type)
		case *ast.ValStmt:
			rewriteExpr(st.Init)
			renameTS(st.Type)
		case *ast.CompoundStmt:
			for _, s2 := range st.Stmts {
				rewriteStmt(s2)
			}
		case *ast.IfStmt:
			rewriteExpr(st.Cond)
			for _, s2 := range st.Then.Stmts {
				rewriteStmt(s2)
			}
			if st.ElseIf != nil {
				rewriteStmt(st.ElseIf)
			}
			if st.Else != nil {
				for _, s2 := range st.Else.Stmts {
					rewriteStmt(s2)
				}
			}
		case *ast.ForStmt:
			rewriteStmt(st.Init)
			rewriteExpr(st.Cond)
			rewriteStmt(st.Update)
			if st.Body != nil {
				for _, s2 := range st.Body.Stmts {
					rewriteStmt(s2)
				}
			}
		case *ast.WhileStmt:
			rewriteExpr(st.Cond)
			for _, s2 := range st.Body.Stmts {
				rewriteStmt(s2)
			}
		case *ast.LoopStmt:
			for _, s2 := range st.Body.Stmts {
				rewriteStmt(s2)
			}
		case *ast.SwitchStmt:
			rewriteExpr(st.Expr)
			for _, cl := range st.Clauses {
				cl := cl.(*ast.CaseClause)
				for _, s2 := range cl.Body.Stmts {
					rewriteStmt(s2)
				}

			}
		case *ast.IncDecStmt:
			rewriteExpr(st.LHS)
		}
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
				rewriteStmt(s)
			}
		}
	case *ast.StructDecl:
		for _, m := range dd.Members {
			if sf, ok := m.(*ast.StructField); ok {
				renameTS(&sf.Type)
			}
		}
	case *ast.GlobalValDecl:
		renameTS(dd.Type)
		rewriteExpr(dd.Init)
	case *ast.GlobalVarDecl:
		renameTS(dd.Type)
		if dd.Init != nil {
			rewriteExpr(dd.Init)
		}
	case *ast.TypeAliasDecl:
		renameTS(&dd.Type)
	}
}
