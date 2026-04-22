package parser

import "github.com/bluescreen10/wesl-go/ast"

// parseStatement parses a single statement.
//
//	statement :
//	  ';'
//	| return_statement ';'
//	| if_statement
//	| switch_statement
//	| loop_statement
//	| for_statement
//	| while_statement
//	| func_call_statement ';'
//	| variable_or_value_statement ';'
//	| break_statement ';'
//	| continue_statement ';'
//	| discard_statement ';'
//	| variable_updating_statement ';'
//	| compound_statement
//	| const_assert_statement ';'
//
// Attributes consumed before this call are passed in via attrs.
func (p *parser) parseStatement() ast.Stmt {
	attrs := p.parseAttributes()
	return p.parseStatementBody(attrs)
}

// parseStatementBody dispatches on the next keyword after any leading
// attributes have already been consumed.
func (p *parser) parseStatementBody(attrs []ast.Attribute) ast.Stmt {
	t := p.peekNonTrivia()

	switch t.typ {
	case tokenIfAttr:
		return p.parseIfAttrStmt()
	// ';'
	case tokenSemicolon:
		p.nextNonTrivia()
		return &ast.EmptyStmt{}

	// compound_statement
	case tokenLBrace:
		return p.parseCompoundStatement(attrs)

	// return_statement ';'
	case tokenReturn:
		s := p.parseReturnStatement(attrs)
		p.expect(tokenSemicolon)
		return s

	// break_statement ';'  |  break_if_statement ';'
	case tokenBreak:
		s := p.parseBreakOrBreakIf(attrs)
		p.expect(tokenSemicolon)
		return s

	// continue_statement ';'
	case tokenContinue:
		s := p.parseContinueStatement(attrs)
		p.expect(tokenSemicolon)
		return s

	// continuing_statement  (no trailing ';' — the compound_statement handles it)
	case tokenContinuing:
		return p.parseContinuingStatement(attrs)

	// discard_statement ';'
	case tokenDiscard:
		s := p.parseDiscardStatement(attrs)
		p.expect(tokenSemicolon)
		return s

	// if_statement
	case tokenIf:
		return p.parseIfStatement(attrs)

	// switch_statement
	case tokenSwitch:
		return p.parseSwitchStatement(attrs)

	// loop_statement
	case tokenLoop:
		return p.parseLoopStatement(attrs)

	// for_statement
	case tokenFor:
		return p.parseForStatement(attrs)

	// while_statement
	case tokenWhile:
		return p.parseWhileStatement(attrs)

	// variable_or_value_statement ';'  (var / let / const)
	case tokenVar, tokenLet, tokenConst:
		s := p.parseVarOrValueStatement(attrs)
		p.expect(tokenSemicolon)
		return s

	// const_assert_statement ';'
	case tokenConstAssert:
		s := p.parseConstAssertStatement(attrs)
		p.expect(tokenSemicolon)
		return s

	// '_' '=' expression ';'  (blank assignment)
	case tokenUnderscore:
		s := p.parseBlankAssignment(attrs)
		p.expect(tokenSemicolon)
		return s

	// Identifiers, '*', '&', '(' can start either a func_call_statement or a
	// variable_updating_statement.  Parse the LHS / call expression first, then
	// decide based on what follows.
	case tokenIdent, tokenStar, tokenAmp, tokenLParen, tokenPackage, tokenSuper:
		s := p.parseExpressionStatement(attrs)
		p.expect(tokenSemicolon)
		return s

	default:
		p.unexpected(t)
		return nil
	}
}

// ── Individual statement parsers ──────────────────────────────────────────────

// parseReturnStatement parses (without the trailing ';'):
//
//	attribute* 'return' expression?
func (p *parser) parseReturnStatement(attrs []ast.Attribute) *ast.ReturnStmt {
	p.expect(tokenReturn)
	var value ast.Expr
	// An expression follows unless the next token is ';' or '}'.
	t := p.peekNonTrivia()
	if t.typ != tokenSemicolon && t.typ != tokenRBrace {
		value = p.parseExpression()
	}
	return &ast.ReturnStmt{Attrs: attrs, Value: value}
}

// parseBreakOrBreakIf disambiguates between break_statement and break_if_statement
// (both start with 'break').
//
//	break_statement    : attribute* 'break'
//	break_if_statement : attribute* 'break' 'if' expression ';'
//
// The trailing ';' is consumed by the caller.
func (p *parser) parseBreakOrBreakIf(attrs []ast.Attribute) ast.Stmt {
	p.expect(tokenBreak)

	// 'break' followed by 'if' → break_if_statement.
	if p.peekNonTrivia().typ == tokenIf {
		p.nextNonTrivia() // consume 'if'
		cond := p.parseExpression()
		return &ast.BreakIfStmt{Attrs: attrs, Cond: cond}
	}
	return &ast.BreakStmt{Attrs: attrs}
}

// parseContinueStatement parses (without the trailing ';'):
//
//	attribute* 'continue'
func (p *parser) parseContinueStatement(attrs []ast.Attribute) *ast.ContinueStmt {
	p.expect(tokenContinue)
	return &ast.ContinueStmt{Attrs: attrs}
}

// parseContinuingStatement parses:
//
//	attribute* 'continuing' continuing_compound_statement
//
// continuing_compound_statement has the same shape as compound_statement but
// may additionally contain break_if_statement; we parse it identically and
// let the semantic pass enforce the distinction.
func (p *parser) parseContinuingStatement(attrs []ast.Attribute) *ast.ContinuingStmt {
	p.expect(tokenContinuing)
	body := p.parseCompoundStatement(nil)
	return &ast.ContinuingStmt{Attrs: attrs, Body: body}
}

// parseDiscardStatement parses (without the trailing ';'):
//
//	attribute* 'discard'
func (p *parser) parseDiscardStatement(attrs []ast.Attribute) *ast.DiscardStmt {
	p.expect(tokenDiscard)
	return &ast.DiscardStmt{Attrs: attrs}
}

// parseConstAssertStatement parses (without the trailing ';'):
//
//	attribute* 'const_assert' expression
func (p *parser) parseConstAssertStatement(attrs []ast.Attribute) *ast.ConstAssertStmt {
	p.expect(tokenConstAssert)
	expr := p.parseExpression()
	return &ast.ConstAssertStmt{Attrs: attrs, Expr: expr}
}

// parseBlankAssignment parses (without the trailing ';'):
//
//	attribute* '_' '=' expression
func (p *parser) parseBlankAssignment(attrs []ast.Attribute) *ast.AssignmentStmt {
	p.expect(tokenUnderscore)
	p.expect(tokenEqual)
	rhs := p.parseExpression()
	return &ast.AssignmentStmt{Attrs: attrs, LHS: nil, Op: "=", RHS: rhs}
}

// parseExpressionStatement handles the ambiguous case where a statement begins
// with an expression (ident / '*' / '&' / '(').  It parses the expression,
// then inspects what follows to decide whether this is:
//
//   - A func_call_statement   (the expression must be a CallExpr)
//   - An assignment_statement ('=' or compound-assignment operator follows)
//   - An increment_statement  ('++' follows)
//   - A decrement_statement   ('--' follows)
//
// The trailing ';' is consumed by the caller.
func (p *parser) parseExpressionStatement(attrs []ast.Attribute) ast.Stmt {
	// Parse as a postfix expression — this covers both call phrases
	// (foo(), vec3<f32>(...)) and lhs expressions (x, a[i], p.m).
	// No lookahead needed upfront; we dispatch purely on what follows.
	expr := p.parsePostfixExpr()

	switch p.peekNonTrivia().typ {
	case tokenPlusPlus:
		p.nextNonTrivia()
		return &ast.IncrementStmt{Attrs: attrs, LHS: expr}

	case tokenMinusMinus:
		p.nextNonTrivia()
		return &ast.DecrementStmt{Attrs: attrs, LHS: expr}

	case tokenEqual:
		p.nextNonTrivia()
		return &ast.AssignmentStmt{Attrs: attrs, LHS: expr, Op: "=", RHS: p.parseExpression()}

	default:
		if op, ok := p.isCompoundAssignOp(); ok {
			p.nextNonTrivia()
			return &ast.AssignmentStmt{Attrs: attrs, LHS: expr, Op: op, RHS: p.parseExpression()}
		}
		// If none of the above matched, expr must be a call.
		call, ok := expr.(*ast.CallExpr)
		if !ok {
			p.unexpected(p.peekNonTrivia())
		}
		return &ast.FnCallStmt{Attrs: attrs, Call: *call}
	}
}

func (p *parser) isCompoundAssignOp() (string, bool) {
	t := p.peekNonTrivia()
	switch t.typ {
	case tokenPlusEq, tokenMinusEq, tokenStarEq, tokenSlashEq, tokenPercentEq,
		tokenAmpEq, tokenPipeEq, tokenCaretEq, tokenLtLtEq, tokenGtGtEq:
		return t.val, true
	}
	return "", false
}

// ── variable_or_value_statement ───────────────────────────────────────────────

// parseVarOrValueStatement parses (without the trailing ';'):
//
//	variable_or_value_statement :
//	  variable_decl
//	| variable_decl '=' expression
//	| attribute* 'let'   optionally_typed_ident '=' expression
//	| attribute* 'const' optionally_typed_ident '=' expression
func (p *parser) parseVarOrValueStatement(attrs []ast.Attribute) *ast.VarOrValueStmt {
	t := p.peekNonTrivia()

	switch t.typ {
	case tokenVar:
		decl := p.parseVariableDecl(attrs)
		//FIXME: Adding keyword "var" generates "var var"
		stmt := &ast.VarOrValueStmt{Attrs: attrs, Keyword: "", Decl: &decl}
		if p.at(tokenEqual) {
			p.nextNonTrivia()
			stmt.Init = p.parseExpression()
		}
		return stmt

	case tokenLet:
		p.nextNonTrivia()
		ident := p.parseOptionallyTypedIdent()
		p.expect(tokenEqual)
		init := p.parseExpression()
		return &ast.VarOrValueStmt{
			Attrs:   attrs,
			Keyword: "let",
			Ident:   &ident,
			Init:    init,
		}

	case tokenConst:
		p.nextNonTrivia()
		ident := p.parseOptionallyTypedIdent()
		p.expect(tokenEqual)
		init := p.parseExpression()
		return &ast.VarOrValueStmt{
			Attrs:   attrs,
			Keyword: "const",
			Ident:   &ident,
			Init:    init,
		}

	default:
		p.unexpected(t)
		return nil
	}
}

// ── compound_statement ────────────────────────────────────────────────────────

// parseCompoundStatement parses:
//
//	attribute* '{' statement* '}'
func (p *parser) parseCompoundStatement(attrs []ast.Attribute) *ast.CompoundStmt {
	p.expect(tokenLBrace)
	var stmts []ast.Stmt
	for !p.at(tokenRBrace) {
		if p.at(tokenEOF) {
			t := p.peekNonTrivia()
			p.unexpected(t)
		}
		stmts = append(stmts, p.parseStatement())
	}
	p.expect(tokenRBrace)
	return &ast.CompoundStmt{Attrs: attrs, Stmts: stmts}
}

// ── case_clause and default_alone_clause ─────────────────────────────────────

// parseCaseClause parses:
//
//	attribute* 'case' case_selectors ':'? compound_statement
//
// Leading attrs must have been consumed by the caller.
func (p *parser) parseSwitchClause(attrs []ast.Attribute) ast.SwitchClause {
	tok := p.peekNonTrivia()
	switch tok.typ {
	case tokenCase:
		return p.parseCaseClause(attrs)
	case tokenDefault:
		return p.parseDefaultAloneClause(attrs)
	default:
		p.unexpected(tok)
	}
	panic("unreachable")
}

func (p *parser) parseCaseClause(attrs []ast.Attribute) *ast.CaseClause {
	p.expect(tokenCase)
	selectors := p.parseCaseSelectors()
	if p.at(tokenColon) {
		p.nextNonTrivia()
	}

	body := p.parseCompoundStatement(nil)
	return &ast.CaseClause{Attrs: attrs, Selectors: selectors, Body: body}
}

// parseDefaultAloneClause parses:
//
//	attribute* 'default' ':'? compound_statement
//
// Leading attrs must have been consumed by the caller.
func (p *parser) parseDefaultAloneClause(attrs []ast.Attribute) *ast.DefaultAloneClause {
	p.expect(tokenDefault)
	if p.at(tokenColon) {
		p.nextNonTrivia()
	}
	body := p.parseCompoundStatement(nil)
	return &ast.DefaultAloneClause{Attrs: attrs, Body: body}
}

func (p *parser) parseIfAttrClause() *ast.IfAttrClause {
	p.expect(tokenIfAttr)
	cond := p.parseExpression()
	attrs := p.parseAttributes()
	then := p.parseSwitchClause(attrs)

	node := &ast.IfAttrClause{Cond: cond, Then: then}
	if p.at(tokenElseAttr) {
		p.expect(tokenElseAttr)
		attrs := p.parseAttributes()
		els := p.parseSwitchClause(attrs)
		node.Else = els
	}
	return node
}

// parseCaseSelectors parses the comma-separated list of case expressions:
//
//	case_selectors : case_selector ( ',' case_selector )* ','?
//
// In WGSL a case_selector is either 'default' or an expression.
func (p *parser) parseCaseSelectors() []ast.Expr {
	var selectors []ast.Expr
	for {
		t := p.peekNonTrivia()
		if t.typ == tokenDefault {
			p.nextNonTrivia()
			selectors = append(selectors, &ast.Ident{Name: "default"})
		} else {
			selectors = append(selectors, p.parseExpression())
		}
		if !p.at(tokenComma) {
			break
		}
		p.nextNonTrivia() // consume ','
		// Trailing comma — peek to see if another selector follows.
		next := p.peekNonTrivia()
		if next.typ == tokenColon || next.typ == tokenLBrace {
			break
		}
	}
	return selectors
}

// ── Complex statements (full sub-grammar not provided; stubs below) ───────────

// parseIfStatement parses:
//
//	attribute* 'if' expression compound_statement
//	  ( 'else' ( if_statement | compound_statement ) )?
func (p *parser) parseIfStatement(attrs []ast.Attribute) *ast.IfStmt {
	p.expect(tokenIf)
	cond := p.parseExpression()
	then := p.parseCompoundStatement(nil)
	stmt := &ast.IfStmt{Attrs: attrs, Cond: cond, Then: then}

	if p.at(tokenElse) {
		p.nextNonTrivia() // consume 'else'
		if p.at(tokenIf) {
			// else-if chain
			innerAttrs := p.parseAttributes()
			stmt.ElseIf = p.parseIfStatement(innerAttrs)
		} else {
			stmt.Else = p.parseCompoundStatement(nil)
		}
	}
	return stmt
}

// parseSwitchStatement parses:
//
//	attribute* 'switch' expression switch_body_attributes? '{' switch_clause* '}'
func (p *parser) parseSwitchStatement(attrs []ast.Attribute) *ast.SwitchStmt {
	p.expect(tokenSwitch)
	expr := p.parseExpression()

	// Optional attributes on the switch body itself.
	_ = p.parseAttributes()

	p.expect(tokenLBrace)
	var clauses []ast.SwitchClause
	for !p.at(tokenRBrace) {
		clauseAttrs := p.parseAttributes()
		t := p.peekNonTrivia()
		switch t.typ {
		case tokenIfAttr:
			clauses = append(clauses, p.parseIfAttrClause())
		default:
			clauses = append(clauses, p.parseSwitchClause(clauseAttrs))
		}
	}
	p.expect(tokenRBrace)
	return &ast.SwitchStmt{Attrs: attrs, Expr: expr, Clauses: clauses}
}

// parseLoopStatement parses:
//
//	attribute* 'loop' attribute* '{' statement* ( continuing_statement )? '}'
func (p *parser) parseLoopStatement(attrs []ast.Attribute) *ast.LoopStmt {
	p.expect(tokenLoop)
	bodyAttrs := p.parseAttributes()
	p.expect(tokenLBrace)

	var stmts []ast.Stmt

	for !p.at(tokenRBrace) {
		if p.at(tokenEOF) {
			p.unexpected(p.peekNonTrivia())
		}
		// A continuing statement ends the loop body.
		innerAttrs := p.parseAttributes()
		// if p.at(tokenContinuing) {
		// 	cont = p.parseContinuingStatement(innerAttrs)
		// 	break
		// }
		stmts = append(stmts, p.parseStatementBody(innerAttrs))
	}
	p.expect(tokenRBrace)
	return &ast.LoopStmt{Attrs: attrs, BodyAttrs: bodyAttrs, Body: &ast.CompoundStmt{Stmts: stmts}}
}

// parseForStatement parses:
//
//	attribute* 'for' '(' for_init? ';' expression? ';' for_update? ')' compound_statement
func (p *parser) parseForStatement(attrs []ast.Attribute) *ast.ForStmt {
	p.expect(tokenFor)
	p.expect(tokenLParen)

	// for_init: variable_or_value_statement | variable_updating_statement | func_call_statement
	var init ast.Stmt
	if !p.at(tokenSemicolon) {
		initAttrs := p.parseAttributes()
		t := p.peekNonTrivia()
		switch t.typ {
		case tokenVar, tokenLet, tokenConst:
			init = p.parseVarOrValueStatement(initAttrs)
		default:
			init = p.parseExpressionStatement(initAttrs)
		}
	}
	p.expect(tokenSemicolon)

	// for_condition
	var cond ast.Expr
	if !p.at(tokenSemicolon) {
		cond = p.parseExpression()
	}
	p.expect(tokenSemicolon)

	// for_update: variable_updating_statement | func_call_statement
	var update ast.Stmt
	if !p.at(tokenRParen) {
		updateAttrs := p.parseAttributes()
		update = p.parseExpressionStatement(updateAttrs)
	}
	p.expect(tokenRParen)

	body := p.parseCompoundStatement(nil)
	return &ast.ForStmt{Attrs: attrs, Init: init, Cond: cond, Update: update, Body: body}
}

// parseWhileStatement parses:
//
//	attribute* 'while' expression compound_statement
func (p *parser) parseWhileStatement(attrs []ast.Attribute) *ast.WhileStmt {
	p.expect(tokenWhile)
	cond := p.parseExpression()
	body := p.parseCompoundStatement(nil)
	return &ast.WhileStmt{Attrs: attrs, Cond: cond, Body: body}
}
