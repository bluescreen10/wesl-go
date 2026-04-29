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
