package ast

// WalkExpr calls fn on e and, if fn returns true, recurses into its children.
func WalkExpr(e Expr, fn func(Expr) bool) {
	if e == nil || !fn(e) {
		return
	}
	switch ex := e.(type) {
	case *BinaryExpr:
		WalkExpr(ex.Left, fn)
		WalkExpr(ex.Right, fn)
	case *UnaryExpr:
		WalkExpr(ex.Operand, fn)
	case *CallExpr:
		for _, a := range ex.Args {
			WalkExpr(a, fn)
		}
		for _, a := range ex.TemplateArgs {
			WalkExpr(a, fn)
		}
	case *IndexExpr:
		WalkExpr(ex.Base, fn)
		WalkExpr(ex.Index, fn)
	case *MemberExpr:
		WalkExpr(ex.Base, fn)
	case *AddrOfExpr:
		WalkExpr(ex.Operand, fn)
	case *DerefExpr:
		WalkExpr(ex.Operand, fn)
	case *ParenExpr:
		WalkExpr(ex.Inner, fn)
		// Ident, LitExpr: leaves, no children
	}
}

// WalkStmt calls stmtFn on s and, if stmtFn returns true, recurses into child
// statements and calls exprFn (via WalkExpr) on all contained expressions.
// Either callback may be nil to skip that node type.
func WalkStmt(s Stmt, stmtFn func(Stmt) bool, exprFn func(Expr) bool) {
	if s == nil || stmtFn != nil && !stmtFn(s) {
		return
	}
	expr := func(e Expr) {
		if exprFn != nil {
			WalkExpr(e, exprFn)
		}
	}
	stmt := func(child Stmt) { WalkStmt(child, stmtFn, exprFn) }

	switch st := s.(type) {
	case *AssignmentStmt:
		expr(st.LHS)
		expr(st.RHS)
	case *ReturnStmt:
		expr(st.Value)
	case *VarStmt:
		expr(st.Init)
	case *ValStmt:
		expr(st.Init)
	case *FuncCallStmt:
		expr(&st.Call)
	case *IncDecStmt:
		expr(st.LHS)
	case *CompoundStmt:
		for _, s2 := range st.Stmts {
			stmt(s2)
		}
	case *IfStmt:
		expr(st.Cond)
		if st.Then != nil {
			stmt(st.Then)
		}
		if st.ElseIf != nil {
			stmt(st.ElseIf)
		}
		if st.Else != nil {
			stmt(st.Else)
		}
	case *ForStmt:
		if st.Init != nil {
			stmt(st.Init)
		}
		expr(st.Cond)
		if st.Update != nil {
			stmt(st.Update)
		}
		if st.Body != nil {
			stmt(st.Body)
		}
	case *WhileStmt:
		expr(st.Cond)
		if st.Body != nil {
			stmt(st.Body)
		}
	case *LoopStmt:
		if st.Body != nil {
			stmt(st.Body)
		}
	case *ContinuingStmt:
		if st.Body != nil {
			stmt(st.Body)
		}
	case *SwitchStmt:
		expr(st.Expr)
		for _, cl := range st.Clauses {
			if cc, ok := cl.(*CaseClause); ok {
				stmt(cc.Body)
			}
		}
	}
}

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
