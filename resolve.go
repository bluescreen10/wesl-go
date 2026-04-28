package wesl

import "github.com/bluescreen10/wesl-go/ast"

// ResolveFile walks f and evaluates every @if/@else node against defines,
// replacing each conditional node with its chosen branch (or nothing when the
// condition is false and there is no @else).  The result is a new *File that
// contains no IfAttr* nodes.
func ResolveFile(f *ast.File, defines map[string]bool) *ast.File {
	out := &ast.File{}
	for _, d := range f.Decls {
		if r := resolveDecl(d, defines); r != nil {
			out.Decls = append(out.Decls, r)
		}
	}
	return out
}

// resolveDecl returns the resolved form of a declaration, or nil if the
// declaration is guarded by a false @if with no @else.
func resolveDecl(d ast.Decl, defines map[string]bool) ast.Decl {
	switch d := d.(type) {
	case *ast.IfAttrDecl:
		if evalCondition(d.Cond, defines) {
			return resolveDecl(d.Then, defines)
		}
		if d.Else != nil {
			return resolveDecl(d.Else, defines)
		}
		return nil

	case *ast.FuncDecl:
		return resolveFuncDecl(d, defines)

	case *ast.StructDecl:
		return resolveStructDecl(d, defines)

	// Declarations with no nested conditionals are returned as-is.
	default:
		return d
	}
}

// resolveFuncDecl resolves @if nodes inside a function body.
func resolveFuncDecl(d *ast.FuncDecl, defines map[string]bool) *ast.FuncDecl {
	out := &ast.FuncDecl{Attrs: d.Attrs, Name: d.Name, ReturnAttrs: d.ReturnAttrs, ReturnType: d.ReturnType}
	for _, p := range d.Params {
		switch p := p.(type) {
		case *ast.IfAttrParam:
			if evalCondition(p.Cond, defines) {
				out.Params = append(out.Params, p.Then)
			} else if p.Else != nil {
				out.Params = append(out.Params, p.Else)
			}
		default:
			out.Params = append(out.Params, p)
		}
	}
	out.Body = resolveCompoundStmt(d.Body, defines)
	return out
}

// resolveStructDecl resolves @if members inside a struct body.
func resolveStructDecl(d *ast.StructDecl, defines map[string]bool) *ast.StructDecl {
	out := &ast.StructDecl{Attrs: d.Attrs, Name: d.Name}
	for _, m := range d.Members {
		switch m := m.(type) {
		case *ast.IfAttrStructField:
			if evalCondition(m.Cond, defines) {
				out.Members = append(out.Members, m.Then)
			} else if m.Else != nil {
				out.Members = append(out.Members, m.Else)
			}
		default:
			out.Members = append(out.Members, m)
		}
	}
	return out
}

// ── Statement resolution ──────────────────────────────────────────────────────

func resolveStmt(s ast.Stmt, defines map[string]bool) ast.Stmt {
	switch s := s.(type) {
	case *ast.IfAttrStmt:
		if evalCondition(s.Cond, defines) {
			return resolveStmt(s.Then, defines)
		}
		if s.Else != nil {
			return resolveStmt(s.Else, defines)
		}
		return nil

	case *ast.CompoundStmt:
		return resolveCompoundStmt(s, defines)

	case *ast.IfStmt:
		return resolveIfStmt(s, defines)

	case *ast.ForStmt:
		out := *s
		out.Body = resolveCompoundStmt(s.Body, defines)
		return &out

	case *ast.WhileStmt:
		out := *s
		out.Body = resolveCompoundStmt(s.Body, defines)
		return &out

	case *ast.LoopStmt:
		out := *s
		out.Body = resolveCompoundStmt(s.Body, defines)
		// if s.Continuing != nil {
		// 	cont := *s.Continuing
		// 	cont.Body = resolveCompoundStmt(s.Continuing.Body, defines)
		// 	out.Continuing = &cont
		// }
		return &out

	case *ast.ContinuingStmt:
		out := *s
		out.Body = resolveCompoundStmt(s.Body, defines)
		return &out

	case *ast.SwitchStmt:
		return resolveSwitchStmt(s, defines)

	default:
		return s
	}
}

func resolveCompoundStmt(s *ast.CompoundStmt, defines map[string]bool) *ast.CompoundStmt {
	if s == nil {
		return nil
	}
	return &ast.CompoundStmt{Attrs: s.Attrs, Stmts: resolveStmtSlice(s.Stmts, defines)}
}

func resolveStmtSlice(stmts []ast.Stmt, defines map[string]bool) []ast.Stmt {
	var out []ast.Stmt
	for _, s := range stmts {
		if r := resolveStmt(s, defines); r != nil {
			out = append(out, r)
		}
	}
	return out
}

func resolveIfStmt(s *ast.IfStmt, defines map[string]bool) *ast.IfStmt {
	out := *s
	out.Then = resolveCompoundStmt(s.Then, defines)
	if s.ElseIf != nil {
		out.ElseIf = resolveIfStmt(s.ElseIf, defines)
	}
	if s.Else != nil {
		out.Else = resolveCompoundStmt(s.Else, defines)
	}
	return &out
}

func resolveSwitchStmt(s *ast.SwitchStmt, defines map[string]bool) *ast.SwitchStmt {
	out := &ast.SwitchStmt{Attrs: s.Attrs, Expr: s.Expr}
	for _, c := range s.Clauses {
		switch c := c.(type) {
		case *ast.IfAttrClause:
			if evalCondition(c.Cond, defines) {
				out.Clauses = append(out.Clauses, resolveClause(c.Then, defines))
			} else {
				if c.Else != nil {
					out.Clauses = append(out.Clauses, resolveClause(c.Else, defines))
				}
			}
		default:
			out.Clauses = append(out.Clauses, resolveClause(c, defines))
		}
	}
	return out
}

func resolveClause(c ast.SwitchClause, defines map[string]bool) ast.SwitchClause {
	switch c := c.(type) {
	case *ast.CaseClause:
		return &ast.CaseClause{
			Attrs:     c.Attrs,
			Selectors: c.Selectors,
			Body:      resolveCompoundStmt(c.Body, defines),
		}
	default:
		panic("unknown clause")
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
		return defines[e.Name] // missing key → false

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
