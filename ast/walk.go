package ast

func Walk(n Node, fn func(Node) bool) {
	if n == nil || !fn(n) {
		return
	}

	switch n := n.(type) {

	// Decls
	case *File:
		WalkList(n.Decls, fn)
	case *GlobalValDecl:
		WalkList(n.Attrs, fn)
		Walk(n.Name, fn)
		Walk(n.Type, fn)
		Walk(n.Init, fn)
	case *GlobalVarDecl:
		WalkList(n.Attrs, fn)
		Walk(n.Name, fn)
		Walk(n.Type, fn)
		Walk(n.Init, fn)
	case *RequiresDirective:
		WalkList(n.Attrs, fn)
	case *EnableDirective:
		WalkList(n.Attrs, fn)
	case *ConstAssertDecl:
		Walk(n.Assert, fn)
	case *DiagnosticDirective:
		WalkList(n.Attrs, fn)
	case *IfAttrDecl:
		Walk(n.Cond, fn)
		Walk(n.Then, fn)
		Walk(n.Else, fn)
	case *FuncDecl:
		WalkList(n.Attrs, fn)
		Walk(n.Name, fn)
		WalkList(n.Params, fn)
		WalkList(n.ReturnAttrs, fn)
		Walk(n.ReturnType, fn)
		Walk(n.Body, fn)
	case *StructDecl:
		WalkList(n.Attrs, fn)
		WalkList(n.Members, fn)

	// Stmt
	case *AssignmentStmt:
		WalkList(n.Attrs, fn)
		Walk(n.LHS, fn)
		Walk(n.RHS, fn)
	case *ReturnStmt:
		WalkList(n.Attrs, fn)
		Walk(n.Value, fn)
	case *VarStmt:
		WalkList(n.Attrs, fn)
		Walk(n.Init, fn)
	case *ValStmt:
		WalkList(n.Attrs, fn)
		Walk(n.Init, fn)
	case *FuncCallStmt:
		WalkList(n.Attrs, fn)
		Walk(n.Call, fn)
	case *IncDecStmt:
		WalkList(n.Attrs, fn)
		Walk(n.LHS, fn)
	case *BlockStmt:
		WalkList(n.Attrs, fn)
		WalkList(n.Stmts, fn)
	case *IfAttrStmt:
		Walk(n.Cond, fn)
		Walk(n.Then, fn)
		Walk(n.Else, fn)
	case *IfStmt:
		WalkList(n.Attrs, fn)
		Walk(n.Cond, fn)
		if n.Then != nil {
			Walk(n.Then, fn)
		}
		if n.ElseIf != nil {
			Walk(n.ElseIf, fn)
		}
		if n.Else != nil {
			Walk(n.Else, fn)
		}
	case *ForStmt:
		WalkList(n.Attrs, fn)
		if n.Init != nil {
			Walk(n.Init, fn)
		}
		if n.Cond != nil {
			Walk(n.Cond, fn)
		}
		if n.Update != nil {
			Walk(n.Update, fn)
		}
		Walk(n.Body, fn)
	case *WhileStmt:
		WalkList(n.Attrs, fn)
		Walk(n.Cond, fn)
		Walk(n.Body, fn)
	case *LoopStmt:
		WalkList(n.Attrs, fn)
		Walk(n.Body, fn)
	case *ContinuingStmt:
		WalkList(n.Attrs, fn)
		Walk(n.Body, fn)
	case *SwitchStmt:
		WalkList(n.Attrs, fn)
		Walk(n.Expr, fn)
		WalkList(n.Clauses, fn)

	// Expr
	case *BinaryExpr:
		Walk(n.Left, fn)
		Walk(n.Right, fn)
	case *UnaryExpr:
		Walk(n.Operand, fn)
	case *CallExpr:
		Walk(n.Callee, fn)
		WalkList(n.Args, fn)
		WalkList(n.TemplateArgs, fn)
	case *IndexExpr:
		Walk(n.Base, fn)
		Walk(n.Index, fn)
	case *MemberExpr:
		Walk(n.Base, fn)
	case *AddrOfExpr:
		Walk(n.Operand, fn)
	case *DerefExpr:
		Walk(n.Operand, fn)
	case *ParenExpr:
		Walk(n.Inner, fn)
		// Ident, LitExpr: leaves, no children
	}
}

func WalkList[T Node](items []T, fn func(Node) bool) {
	for _, i := range items {
		Walk(i, fn)
	}
}
