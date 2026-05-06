package ast

func Walk(n Node, fn func(Node) bool) {
	if n == nil || !fn(n) {
		return
	}

	switch n := n.(type) {

	// Decls
	case *File:
		for _, d := range n.Decls {
			Walk(d, fn)
		}
	case *GlobalValDecl:
		for _, a := range n.Attrs {
			Walk(a, fn)
		}
		Walk(n.Type, fn)
		Walk(n.Init, fn)
	case *GlobalVarDecl:
		for _, a := range n.Attrs {
			Walk(a, fn)
		}
		Walk(n.Type, fn)
		Walk(n.Init, fn)
	case *RequiresDirective:
		for _, a := range n.Attrs {
			Walk(a, fn)
		}
	case *EnableDirective:
		for _, a := range n.Attrs {
			Walk(a, fn)
		}
	case *ConstAssertDecl:
		Walk(n.Assert, fn)
	case *DiagnosticDirective:
		for _, a := range n.Attrs {
			Walk(a, fn)
		}
	case *IfAttrDecl:
		Walk(n.Cond, fn)
		Walk(n.Then, fn)
		Walk(n.Else, fn)
	case *FuncDecl:
		for _, a := range n.Attrs {
			Walk(a, fn)
		}
		for _, p := range n.Params {
			Walk(p, fn)
		}

		Walk(n.ReturnType, fn)
	case *StructDecl:
		for _, a := range n.Attrs {
			Walk(a, fn)
		}
		for _, m := range n.Members {
			Walk(m, fn)
		}

	// Stmt
	case *AssignmentStmt:
		Walk(n.LHS, fn)
		Walk(n.RHS, fn)
	case *ReturnStmt:
		Walk(n.Value, fn)
	case *VarStmt:
		Walk(n.Init, fn)
	case *ValStmt:
		Walk(n.Init, fn)
	case *FuncCallStmt:
		Walk(&n.Call, fn)
	case *IncDecStmt:
		Walk(n.LHS, fn)
	case *CompoundStmt:
		for _, s := range n.Stmts {
			Walk(s, fn)
		}
	case *IfAttrStmt:
		Walk(n.Cond, fn)
		Walk(n.Then, fn)
		Walk(n.Else, fn)
	case *IfStmt:
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
		if n.Init != nil {
			Walk(n.Init, fn)
		}
		Walk(n.Cond, fn)
		if n.Update != nil {
			Walk(n.Update, fn)
		}
		if n.Body != nil {
			Walk(n.Body, fn)
		}
	case *WhileStmt:
		Walk(n.Cond, fn)
		if n.Body != nil {
			Walk(n.Body, fn)
		}
	case *LoopStmt:
		if n.Body != nil {
			Walk(n.Body, fn)
		}
	case *ContinuingStmt:
		if n.Body != nil {
			Walk(n.Body, fn)
		}
	case *SwitchStmt:
		Walk(n.Expr, fn)
		for _, cl := range n.Clauses {
			Walk(cl, fn)
		}

	// Expr
	case *BinaryExpr:
		Walk(n.Left, fn)
		Walk(n.Right, fn)
	case *UnaryExpr:
		Walk(n.Operand, fn)
	case *CallExpr:
		for _, a := range n.Args {
			Walk(a, fn)
		}
		for _, a := range n.TemplateArgs {
			Walk(a, fn)
		}
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
