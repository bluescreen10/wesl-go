package ast

import "reflect"

func Rewrite(n Node, fn func(Node) Node) Node {
	if v := reflect.ValueOf(n); v.Kind() == reflect.Pointer && v.IsNil() {
		n = nil
	}

	r := fn(n)
	if r == nil {
		return nil
	}

	if n != r {
		Rewrite(r, fn)
	}

	switch r := r.(type) {

	// Decls
	case *File:
		r.Decls = RewriteList(r.Decls, fn)
	case *RequiresDirective:
		r.Attrs = RewriteList(r.Attrs, fn)
	case *EnableDirective:
		r.Attrs = RewriteList(r.Attrs, fn)
	case *DiagnosticDirective:
		r.Attrs = RewriteList(r.Attrs, fn)
	case *IfAttrDecl:
		r.Cond = wrap[Expr](Rewrite(r.Cond, fn))
		r.Then = wrap[Decl](Rewrite(r, fn))
		r.Else = wrap[Decl](Rewrite(r, fn))
	case *FuncDecl:
		r.Attrs = RewriteList(r.Attrs, fn)
		r.Params = RewriteList(r.Params, fn)
		r.ReturnType = wrap[*TypeSpecifier](Rewrite(r.ReturnType, fn))
		r.Body = wrap[*BlockStmt](Rewrite(r.Body, fn))
	case *StructDecl:
		r.Attrs = RewriteList(r.Attrs, fn)
		r.Members = RewriteList(r.Members, fn)

	// Stmt
	case *AssignmentStmt:
		r.Attrs = RewriteList(r.Attrs, fn)
		r.LHS = wrap[Expr](Rewrite(r.LHS, fn))
		r.RHS = wrap[Expr](Rewrite(r.RHS, fn))
	case *ReturnStmt:
		r.Attrs = RewriteList(r.Attrs, fn)
		r.Value = wrap[Expr](Rewrite(r.Value, fn))
	case *VarStmt:
		r.Attrs = RewriteList(r.Attrs, fn)
		r.TemplateArgs = RewriteList(r.TemplateArgs, fn)
		r.Name = wrap[*Ident](Rewrite(r.Name, fn))
		r.Type = wrap[*TypeSpecifier](Rewrite(r.Type, fn))
		r.Init = wrap[Expr](Rewrite(r.Init, fn))
	case *ValStmt:
		r.Attrs = RewriteList(r.Attrs, fn)
		r.Name = wrap[*Ident](Rewrite(r.Name, fn))
		r.Type = wrap[*TypeSpecifier](Rewrite(r.Type, fn))
		r.Init = wrap[Expr](Rewrite(r.Init, fn))
	case *FuncCallStmt:
		r.Attrs = RewriteList(r.Attrs, fn)
		r.Call = wrap[*CallExpr](Rewrite(r.Call, fn))
	case *IncDecStmt:
		r.Attrs = RewriteList(r.Attrs, fn)
		r.LHS = wrap[Expr](Rewrite(r.LHS, fn))
	case *BlockStmt:
		r.Attrs = RewriteList(r.Attrs, fn)
		r.Stmts = RewriteList(r.Stmts, fn)
	case *IfAttrStmt:
		r.Cond = wrap[Expr](Rewrite(r.Cond, fn))
		r.Then = wrap[Stmt](Rewrite(r.Then, fn))
		r.Else = wrap[Stmt](Rewrite(r.Else, fn))
	case *IfStmt:
		r.Attrs = RewriteList(r.Attrs, fn)
		r.Cond = wrap[Expr](Rewrite(r.Cond, fn))
		r.Then = wrap[*BlockStmt](Rewrite(r.Then, fn))
		r.ElseIf = wrap[*IfStmt](Rewrite(r.ElseIf, fn))
		r.Else = wrap[*BlockStmt](Rewrite(r.Else, fn))
	case *ForStmt:
		r.Attrs = RewriteList(r.Attrs, fn)
		r.Init = wrap[Stmt](Rewrite(r.Init, fn))
		r.Cond = wrap[Expr](Rewrite(r.Cond, fn))
		r.Update = wrap[Stmt](Rewrite(r.Update, fn))
		r.Body = wrap[*BlockStmt](Rewrite(r.Body, fn))
	case *WhileStmt:
		r.Attrs = RewriteList(r.Attrs, fn)
		r.Cond = wrap[Expr](Rewrite(r.Cond, fn))
		r.Body = wrap[*BlockStmt](Rewrite(r.Body, fn))
	case *LoopStmt:
		r.Attrs = RewriteList(r.Attrs, fn)
		r.Body = wrap[*BlockStmt](Rewrite(r.Body, fn))
	case *ContinuingStmt:
		r.Attrs = RewriteList(r.Attrs, fn)
		r.Body = wrap[*BlockStmt](Rewrite(r.Body, fn))
	case *SwitchStmt:
		r.Attrs = RewriteList(r.Attrs, fn)
		r.Expr = wrap[Expr](Rewrite(r.Expr, fn))
		r.Clauses = RewriteList(r.Clauses, fn)

	// Expr
	case *BinaryExpr:
		r.Left = wrap[Expr](Rewrite(r.Left, fn))
		r.Right = wrap[Expr](Rewrite(r.Right, fn))
	case *UnaryExpr:
		r.Operand = wrap[Expr](Rewrite(r.Operand, fn))
	case *CallExpr:
		r.Callee = wrap[*Ident](Rewrite(r.Callee, fn))
		r.TemplateArgs = RewriteList(r.TemplateArgs, fn)
		r.Args = RewriteList(r.Args, fn)
	case *IndexExpr:
		r.Base = wrap[Expr](Rewrite(r.Base, fn))
		r.Index = wrap[Expr](Rewrite(r.Base, fn))
	case *MemberExpr:
		r.Base = wrap[Expr](Rewrite(r.Base, fn))
	case *AddrOfExpr:
		r.Operand = wrap[Expr](Rewrite(r.Operand, fn))
	case *DerefExpr:
		r.Operand = wrap[Expr](Rewrite(r.Operand, fn))
	case *ParenExpr:
		r.Inner = wrap[Expr](Rewrite(r.Inner, fn))
		// Ident, LitExpr: leaves, no children

	}

	return r
}

func RewriteList[T Node](items []T, fn func(Node) Node) []T {
	var j int
	for _, item := range items {
		if r := Rewrite(item, fn); r != nil {
			items[j] = r.(T)
			j++
		}
	}
	return items[:j]
}

func wrap[T Node](val Node) T {
	if val == nil {
		var zero T
		return zero
	}
	return val.(T)
}
