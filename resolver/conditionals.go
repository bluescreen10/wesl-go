package resolver

import "github.com/bluescreen10/wesl-go/ast"

func (r *Resolver) ResolveConditionals(f *ast.File) *ast.File {
	out := ast.Rewrite(f, func(n ast.Node) ast.Node {
		switch n := n.(type) {
		case *ast.IfAttrDecl:
			if r.evalCondition(n.Cond) {
				return n.Then
			} else {
				return n.Else
			}
		case *ast.IfAttrStmt:
			if r.evalCondition(n.Cond) {
				return n.Then
			} else {
				return n.Else
			}
		case *ast.IfAttrParam:
			if r.evalCondition(n.Cond) {
				return n.Then
			} else {
				return n.Else
			}
		case *ast.IfAttrClause:
			if r.evalCondition(n.Cond) {
				return n.Then
			} else {
				return n.Else
			}
		case *ast.IfAttrStructMember:
			if r.evalCondition(n.Cond) {
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
		return r.defines[e.Val] // missing key → false

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
