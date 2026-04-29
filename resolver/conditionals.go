package resolver

import "github.com/bluescreen10/wesl-go/ast"

func (r *Resolver) ResolveConditionals(f *ast.File) *ast.File {
	out := &ast.File{}
	for _, d := range f.Decls {
		if r := r.resolveCondDecl(d); r != nil {
			out.Decls = append(out.Decls, r)
		}
	}
	return out
}

func (r *Resolver) resolveCondDecl(d ast.Decl) ast.Decl {
	switch d := d.(type) {
	case *ast.IfAttrDecl:
		if r.evalCondition(d.Cond) {
			return r.resolveCondDecl(d.Then)
		}
		if d.Else != nil {
			return r.resolveCondDecl(d.Else)
		}
		return nil
	case *ast.FuncDecl:
		return r.resolveCondFuncDecl(d)

	case *ast.StructDecl:
		return r.resolveCondStructDecl(d)
	default:
		return d
	}
}

// resolveFuncDecl resolves @if nodes inside a function body.
func (r *Resolver) resolveCondFuncDecl(d *ast.FuncDecl) *ast.FuncDecl {
	out := &ast.FuncDecl{Attrs: d.Attrs, Name: d.Name, ReturnAttrs: d.ReturnAttrs, ReturnType: d.ReturnType}
	for _, p := range d.Params {
		switch p := p.(type) {
		case *ast.IfAttrParam:
			if r.evalCondition(p.Cond) {
				out.Params = append(out.Params, p.Then)
			} else if p.Else != nil {
				out.Params = append(out.Params, p.Else)
			}
		default:
			out.Params = append(out.Params, p)
		}
	}
	out.Body = r.resolveCondCompoundStmt(d.Body)
	return out
}

// resolveStructDecl resolves @if members inside a struct body.
func (r *Resolver) resolveCondStructDecl(d *ast.StructDecl) *ast.StructDecl {
	out := &ast.StructDecl{Attrs: d.Attrs, Name: d.Name}
	for _, m := range d.Members {
		switch m := m.(type) {
		case *ast.IfAttrStructMember:
			if r.evalCondition(m.Cond) {
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

func (r *Resolver) resolveCondStmt(s ast.Stmt) ast.Stmt {
	switch s := s.(type) {
	case *ast.IfAttrStmt:
		if r.evalCondition(s.Cond) {
			return r.resolveCondStmt(s.Then)
		}
		if s.Else != nil {
			return r.resolveCondStmt(s.Else)
		}
		return nil

	case *ast.CompoundStmt:
		return r.resolveCondCompoundStmt(s)

	case *ast.IfStmt:
		return r.resolveCondIfStmt(s)

	case *ast.ForStmt:
		out := *s
		out.Body = r.resolveCondCompoundStmt(s.Body)
		return &out

	case *ast.WhileStmt:
		out := *s
		out.Body = r.resolveCondCompoundStmt(s.Body)
		return &out

	case *ast.LoopStmt:
		out := *s
		out.Body = r.resolveCondCompoundStmt(s.Body)
		return &out

	case *ast.ContinuingStmt:
		out := *s
		out.Body = r.resolveCondCompoundStmt(s.Body)
		return &out

	case *ast.SwitchStmt:
		return r.resolveCondSwitchStmt(s)

	default:
		return s
	}
}

func (r *Resolver) resolveCondCompoundStmt(s *ast.CompoundStmt) *ast.CompoundStmt {
	if s == nil {
		return nil
	}
	var stmts []ast.Stmt
	for _, s := range s.Stmts {
		if r := r.resolveCondStmt(s); r != nil {
			stmts = append(stmts, r)
		}
	}
	return &ast.CompoundStmt{Attrs: s.Attrs, Stmts: stmts}
}

func (r *Resolver) resolveCondIfStmt(s *ast.IfStmt) *ast.IfStmt {
	out := *s
	out.Then = r.resolveCondCompoundStmt(s.Then)
	if s.ElseIf != nil {
		out.ElseIf = r.resolveCondIfStmt(s.ElseIf)
	}
	if s.Else != nil {
		out.Else = r.resolveCondCompoundStmt(s.Else)
	}
	return &out
}

func (r *Resolver) resolveCondSwitchStmt(s *ast.SwitchStmt) *ast.SwitchStmt {
	out := &ast.SwitchStmt{Attrs: s.Attrs, Expr: s.Expr}
	for _, c := range s.Clauses {
		switch c := c.(type) {
		case *ast.IfAttrClause:
			if r.evalCondition(c.Cond) {
				out.Clauses = append(out.Clauses, r.resolveCondClause(c.Then))
			} else {
				if c.Else != nil {
					out.Clauses = append(out.Clauses, r.resolveCondClause(c.Else))
				}
			}
		default:
			out.Clauses = append(out.Clauses, r.resolveCondClause(c))
		}
	}
	return out
}

func (r *Resolver) resolveCondClause(c ast.Clause) ast.Clause {
	switch c := c.(type) {
	case *ast.CaseClause:
		return &ast.CaseClause{
			Attrs:     c.Attrs,
			Selectors: c.Selectors,
			Body:      r.resolveCondCompoundStmt(c.Body),
		}
	default:
		panic("unknown clause")
	}
}

func (r *Resolver) evalCondition(expr ast.Expr) bool {
	switch e := expr.(type) {

	case *ast.LitExpr:
		switch e.Val {
		case "true":
			return true
		case "false":
			return false
		}

	case *ast.Ident:
		return r.defines[e.Name] // missing key → false

	case *ast.UnaryExpr:
		if e.Op == "!" {
			return !r.evalCondition(e.Operand)
		}

	case *ast.BinaryExpr:
		switch e.Op {
		case "&&":
			return r.evalCondition(e.Left) && r.evalCondition(e.Right)
		case "||":
			return r.evalCondition(e.Left) || r.evalCondition(e.Right)
		case "==":
			return r.evalCondition(e.Left) == r.evalCondition(e.Right)
		case "!=":
			return r.evalCondition(e.Left) != r.evalCondition(e.Right)
		}

	case *ast.ParenExpr:
		return r.evalCondition(e.Inner)
	}

	return false
}
