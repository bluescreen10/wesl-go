package resolver

import (
	"fmt"
	"maps"
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
	sym        string
	outputName string
}

type Resolver struct {
	files              map[string]*ast.File
	defines            map[string]bool
	resolved           map[string]bool             // files that have had conditionals resolved
	taken              map[string]bool             // names in the output namespace
	assigned           map[fileSymbol]string       // (file,sym) -> output name; also cycle guard
	depMap             map[fileSymbol][]fileSymbol // pre-mutation dep cache populated by assignName
	emitted            map[fileSymbol]bool
	moduleMap          map[string]string // module alias -> file path
	constAssertedFiles map[string]bool
}

func ResolveFile(fileName string, files map[string]*ast.File, defines map[string]bool) *ast.File {
	r := New(files, defines)
	return r.ResolveFile(fileName)
}

func New(files map[string]*ast.File, defines map[string]bool) *Resolver {
	return &Resolver{
		files:              maps.Clone(files),
		defines:            defines,
		resolved:           make(map[string]bool), // true if conditionals have been resolved
		taken:              make(map[string]bool),
		assigned:           make(map[fileSymbol]string),
		depMap:             make(map[fileSymbol][]fileSymbol),
		emitted:            make(map[fileSymbol]bool),
		moduleMap:          make(map[string]string),
		constAssertedFiles: make(map[string]bool),
	}
}

func (r *Resolver) ensureResolved(path string) {
	if r.resolved[path] {
		return
	}
	r.resolved[path] = true
	if f, ok := r.files[path]; ok {
		r.files[path] = r.ResolveConditionals(f)
	}
}

// ── Entry point ───────────────────────────────────────────────────────────────

func (r *Resolver) ResolveFile(fileName string) *ast.File {
	r.ensureResolved(fileName)

	mainAST := r.files[fileName]
	if mainAST == nil {
		return &ast.File{}
	}

	// Seed taken with names defined locally in main.
	for _, d := range mainAST.Decls {
		if _, ok := d.(*ast.ImportDecl); ok {
			continue
		}
		if n := d.GetName(); n != "" {
			r.taken[n] = true
		}
	}

	entries, inlineRenames := r.collectImports(fileName)

	// Phase 1: assign output names for all reachable symbols.
	for _, e := range entries {
		r.assignName(e.srcFile, e.sym, e.outputName)
	}

	// Build rename map for main file: local alias -> final output name.
	// Iterating all entries (including duplicates for the same symbol under
	// different aliases) naturally handles the multi-alias case.
	mainRenames := inlineRenames
	for _, e := range entries {
		actualFile, actualSym := r.resolveSymbol(e.srcFile, e.sym)
		if actualFile == "" {
			continue
		}
		if out := r.assigned[fileSymbol{actualFile, actualSym}]; out != "" {
			mainRenames[e.outputName] = out
		}
	}

	var localDecls []ast.Decl
	for _, d := range mainAST.Decls {
		if _, ok := d.(*ast.ImportDecl); ok {
			continue
		}
		cloned := cloneDecl(d)
		rewriteDeclRefs(cloned, mainRenames)
		localDecls = append(localDecls, cloned)
	}

	// Phase 2: emit primaries first, then their transitive deps.
	// Two loops preserve the invariant: all directly-imported symbols appear
	// before any of their support functions/types.
	var importedDecls []ast.Decl
	for _, e := range entries {
		r.emitPrimary(e.srcFile, e.sym, &importedDecls)
	}
	for _, e := range entries {
		r.emitDeps(e.srcFile, e.sym, &importedDecls)
	}

	return &ast.File{Decls: append(localDecls, importedDecls...)}
}

// ── Phase 1: name assignment ──────────────────────────────────────────────────

// assignName ensures every symbol transitively reachable from (srcFile, sym)
// has a unique output name. The assigned map doubles as a cycle guard.
func (r *Resolver) assignName(srcFile, sym, preferredName string) {
	actualFile, actualSym := r.resolveSymbol(srcFile, sym)
	if actualFile == "" {
		return
	}
	actualKey := fileSymbol{actualFile, actualSym}
	if _, ok := r.assigned[actualKey]; ok {
		return
	}

	chosen := preferredName
	if chosen == "" {
		chosen = actualSym
	}
	chosen = r.makeUnique(chosen)
	r.assigned[actualKey] = chosen
	r.taken[chosen] = true

	decl := r.findDeclInFile(r.files[actualFile], actualSym)
	if decl == nil {
		return
	}
	// Cache deps before any buildDecl call can mutate the body through shallow clones.
	deps := r.depsOf(decl, actualFile)
	r.depMap[actualKey] = deps
	for _, dep := range deps {
		r.assignName(dep.file, dep.sym, "")
	}
}

// ── Phase 2: emission ─────────────────────────────────────────────────────────

// emitPrimary emits the decl for (srcFile, sym) without recursing into deps.
func (r *Resolver) emitPrimary(srcFile, sym string, output *[]ast.Decl) {
	actualFile, actualSym := r.resolveSymbol(srcFile, sym)
	if actualFile == "" {
		return
	}
	key := fileSymbol{actualFile, actualSym}
	if r.emitted[key] {
		return
	}
	r.emitted[key] = true

	decl := r.findDeclInFile(r.files[actualFile], actualSym)
	if decl == nil {
		return
	}
	r.emitConstAsserts(actualFile, output)
	*output = append(*output, r.buildDecl(decl, actualFile, actualSym, r.depMap[key]))
}

// emitDeps recursively emits transitive dependencies of (srcFile, sym).
// Uses depMap to read pre-mutation deps cached during assignName.
func (r *Resolver) emitDeps(srcFile, sym string, output *[]ast.Decl) {
	actualFile, actualSym := r.resolveSymbol(srcFile, sym)
	if actualFile == "" {
		return
	}
	for _, dep := range r.depMap[fileSymbol{actualFile, actualSym}] {
		if r.emitted[dep] {
			continue
		}
		r.emitted[dep] = true
		depDecl := r.findDeclInFile(r.files[dep.file], dep.sym)
		if depDecl == nil {
			continue
		}
		r.emitConstAsserts(dep.file, output)
		*output = append(*output, r.buildDecl(depDecl, dep.file, dep.sym, r.depMap[dep]))
		r.emitDeps(dep.file, dep.sym, output)
	}
}

func (r *Resolver) buildDecl(decl ast.Decl, actualFile, actualSym string, deps []fileSymbol) ast.Decl {
	outputName := r.assigned[fileSymbol{actualFile, actualSym}]
	renames := map[string]string{}
	for _, dep := range deps {
		if out := r.assigned[dep]; out != "" && out != dep.sym {
			renames[dep.sym] = out
		}
	}
	cloned := cloneDecl(decl)
	cloned.SetName(outputName)
	rewriteDeclRefs(cloned, renames)
	return cloned
}

func (r *Resolver) emitConstAsserts(filePath string, output *[]ast.Decl) {
	if r.constAssertedFiles[filePath] {
		return
	}
	r.constAssertedFiles[filePath] = true
	for _, d := range r.files[filePath].Decls {
		if _, ok := d.(*ast.ConstAssertDecl); ok {
			*output = append(*output, d)
		}
	}
}

// ── Import collection ─────────────────────────────────────────────────────────

func (r *Resolver) collectImports(filePath string) ([]importEntry, map[string]string) {
	f := r.files[filePath]
	renames := map[string]string{}
	var entries []importEntry

	for _, d := range f.Decls {
		imp, ok := d.(*ast.ImportDecl)
		if !ok {
			continue
		}
		for _, item := range imp.Items {
			srcFile, origSym := r.resolveImportItem(imp.Path, item)
			if srcFile == "" {
				r.registerModuleImport(imp.Path, item)
				continue
			}
			desiredName := item.Alias
			if desiredName == "" {
				desiredName = item.Path[len(item.Path)-1]
			}
			entries = append(entries, importEntry{srcFile, origSym, desiredName})
		}
	}

	// Inline package::file::sym and module::sym references in non-import decls.
	for _, d := range f.Decls {
		if _, ok := d.(*ast.ImportDecl); ok {
			continue
		}
		r.collectInlineRefs(d, &entries, renames)
	}

	return entries, renames
}

func (r *Resolver) collectInlineRefs(d ast.Decl, entries *[]importEntry, renames map[string]string) {
	visit := func(e ast.Expr) bool {
		var name string
		switch ex := e.(type) {
		case *ast.CallExpr:
			name = ex.Callee
		case *ast.Ident:
			name = ex.Name
		default:
			return true
		}
		if !strings.Contains(name, "::") {
			return true
		}
		parts := strings.Split(name, "::")
		sym := parts[len(parts)-1]
		var filePath string
		if parts[0] == "package" || parts[0] == "super" {
			if len(parts) >= 3 {
				filePath = r.lookupFile(parts[1 : len(parts)-1])
			}
		} else if fp, ok := r.moduleMap[parts[0]]; ok {
			filePath = fp
		}
		if filePath != "" {
			*entries = append(*entries, importEntry{filePath, sym, sym})
			renames[name] = sym
		}
		return true
	}

	switch dd := d.(type) {
	case *ast.FuncDecl:
		if dd.Body != nil {
			for _, s := range dd.Body.Stmts {
				ast.WalkStmt(s, func(ast.Stmt) bool { return true }, visit)
			}
		}
	case *ast.GlobalValDecl:
		ast.WalkExpr(dd.Init, visit)
	case *ast.GlobalVarDecl:
		if dd.Init != nil {
			ast.WalkExpr(dd.Init, visit)
		}
	}
}

func (r *Resolver) registerModuleImport(prefix []string, item ast.ImportItem) {
	modName := item.Path[len(item.Path)-1]
	var segs []string
	for _, s := range prefix {
		if s != "package" && s != "super" {
			segs = append(segs, s)
		}
	}
	segs = append(segs, item.Path...)
	candidate := "./" + strings.Join(segs, "/") + ".wgsl"
	if _, ok := r.files[candidate]; ok {
		r.moduleMap[modName] = candidate
		return
	}
	if len(segs) > 0 {
		candidate2 := "./" + strings.Join(segs[:len(segs)-1], "/") + ".wgsl"
		if _, ok := r.files[candidate2]; ok {
			r.moduleMap[modName] = candidate2
		}
	}
}

// ── Symbol/file lookup ────────────────────────────────────────────────────────

func (r *Resolver) lookupFile(segs []string) string {
	for i := len(segs); i >= 1; i-- {
		//FIXME: remove extensions
		candidate := "./" + strings.Join(segs[:i], "/") + ".wgsl"
		if _, ok := r.files[candidate]; ok {
			return candidate
		}
	}
	return ""
}

func (r *Resolver) resolveImportItem(prefix []string, item ast.ImportItem) (string, string) {
	sym := item.Path[len(item.Path)-1]
	var segs []string
	for _, s := range prefix {
		if s != "package" && s != "super" {
			segs = append(segs, s)
		}
	}
	segs = append(segs, item.Path[:len(item.Path)-1]...)
	if fp := r.lookupFile(segs); fp != "" {
		return fp, sym
	}
	// If appending sym finds a file, it's a module import (no specific symbol).
	if r.lookupFile(append(segs, sym)) != "" {
		return "", ""
	}
	return "", ""
}

func (r *Resolver) resolveSymbol(srcFilePath, sym string) (string, string) {
	r.ensureResolved(srcFilePath)
	srcFile := r.files[srcFilePath]
	if srcFile == nil {
		return "", ""
	}
	if r.findDeclInFile(srcFile, sym) != nil {
		return srcFilePath, sym
	}
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
				if fp, origSym := r.resolveImportItem(imp.Path, item); fp != "" {
					return fp, origSym
				}
			}
		}
	}
	return "", ""
}

// ── Dependency collection ─────────────────────────────────────────────────────

func (r *Resolver) depsOf(decl ast.Decl, srcFilePath string) []fileSymbol {
	names := r.referencedNames(decl)
	seen := map[string]bool{}
	var deps []fileSymbol
	for _, name := range names {
		if seen[name] {
			continue
		}
		seen[name] = true
		if fp, sym := r.resolveSymbol(srcFilePath, name); fp != "" {
			deps = append(deps, fileSymbol{fp, sym})
		}
	}
	return deps
}

// scopeStack tracks locally-defined names so local bindings are not reported
// as external dependencies.
type scopeStack []map[string]bool

func newScopeStack() scopeStack          { return scopeStack{make(map[string]bool)} }
func (ss scopeStack) push() scopeStack   { return append(ss, make(map[string]bool)) }
func (ss scopeStack) define(name string) { ss[len(ss)-1][name] = true }
func (ss scopeStack) has(name string) bool {
	for i := len(ss) - 1; i >= 0; i-- {
		if ss[i][name] {
			return true
		}
	}
	return false
}

func (r *Resolver) referencedNames(decl ast.Decl) []string {
	var names []string
	addName := func(name string, sc scopeStack) {
		if !strings.Contains(name, "::") && !sc.has(name) {
			names = append(names, name)
		}
	}

	// Scope is constant within an expression (only statements introduce bindings),
	// so WalkExpr captures sc by value safely.
	walkExpr := func(e ast.Expr, sc scopeStack) {
		ast.WalkExpr(e, func(ex ast.Expr) bool {
			switch ex := ex.(type) {
			case *ast.Ident:
				addName(ex.Name, sc)
			case *ast.CallExpr:
				callee := ex.Callee
				if idx := strings.LastIndex(callee, "::"); idx >= 0 {
					callee = callee[idx+2:]
				}
				addName(callee, sc)
			}
			return true
		})
	}

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
		case *ast.FuncCallStmt:
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
			walkExpr(st.Init, *sc)
			if st.Type != nil {
				addName(st.Type.Name, *sc)
			}
			(*sc).define(st.Name)
		case *ast.CompoundStmt:
			nested := (*sc).push()
			walkStmts(st.Stmts, &nested)
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
				cc := cl.(*ast.CaseClause)
				nested := (*sc).push()
				walkStmts(cc.Body.Stmts, &nested)
			}
		case *ast.IncDecStmt:
			walkExpr(st.LHS, *sc)
		}
	}

	switch dd := decl.(type) {
	case *ast.FuncDecl:
		sc := newScopeStack()
		for _, p := range dd.Params {
			if fp, ok := p.(*ast.FuncParam); ok {
				addName(fp.Type.Name, sc)
				for _, ta := range fp.Type.TemplateArgs {
					walkExpr(ta, sc)
				}
				sc.define(fp.Name)
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
			if sf, ok := m.(*ast.StructMember); ok {
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

// ── AST helpers ───────────────────────────────────────────────────────────────

func (r *Resolver) findDeclInFile(f *ast.File, sym string) ast.Decl {
	if f == nil {
		return nil
	}
	for _, d := range f.Decls {
		if d.GetName() == sym {
			return d
		}
	}
	return nil
}

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

// ── Unique naming ─────────────────────────────────────────────────────────────

func (r *Resolver) makeUnique(base string) string {
	if !r.taken[base] {
		return base
	}
	for i := 0; ; i++ {
		if candidate := fmt.Sprintf("%s%d", base, i); !r.taken[candidate] {
			return candidate
		}
	}
}
