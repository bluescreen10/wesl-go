package parser

import "github.com/bluescreen10/wesl-go/ast"

func isTemplateableIdent(name string) bool {
	switch name {
	case "array", "atomic",
		"bool",
		"f16", "f32",
		"i32", "u32",
		"mat2x2", "mat2x3", "mat2x4",
		"mat3x2", "mat3x3", "mat3x4",
		"mat4x2", "mat4x3", "mat4x4",
		"ptr",
		"sampler", "sampler_comparison",
		"texture_1d", "texture_2d", "texture_2d_array",
		"texture_3d", "texture_cube", "texture_cube_array",
		"texture_depth_2d", "texture_depth_2d_array",
		"texture_depth_cube", "texture_depth_cube_array",
		"texture_depth_multisampled_2d",
		"texture_multisampled_2d",
		"texture_storage_1d", "texture_storage_2d",
		"texture_storage_2d_array", "texture_storage_3d",
		"vec2", "vec3", "vec4",
		"binding_array": // WESL extension
		return true
	}
	return false
}

// ── Precedence table ─────────────────────────────────────────────────────────
//
// WGSL forbids mixing bitwise (|, ^, &) with logical (||, &&) operators in
// the same expression without explicit parentheses.  That constraint is
// enforced by a later semantic pass; here we assign them the numeric
// precedences from the WGSL specification so that the Pratt parser produces
// the correct tree when they do appear in valid isolation.
//
//   1  ||
//   2  &&
//   3  |
//   4  ^
//   5  &
//   6  == !=
//   7  < > <= >=
//   8  << >>
//   9  + -
//  10  * / %

func (p *parser) infixPrec(t token) (prec int) {
	if p.templateDepth > 0 {
		switch t.typ {
		case tokenLAngle, tokenRAngle, tokenGtGt, tokenLtLt:
			return 0
		}
	}
	switch t.typ {
	case tokenPipePipe:
		return 1
	case tokenAmpAmp:
		return 2
	case tokenPipe:
		return 3
	case tokenCaret:
		return 4
	case tokenAmp:
		return 5
	case tokenEqualEqual, tokenBangEqual:
		return 6
	case tokenLAngle, tokenRAngle, tokenLtEqual, tokenGtEqual:
		return 7
	case tokenLtLt, tokenGtGt:
		return 8
	case tokenPlus, tokenMinus:
		return 9
	case tokenStar, tokenSlash, tokenPercent:
		return 10
	}
	return 0
}

// ── Expression entry point ────────────────────────────────────────────────────

// parseExpression parses a full expression.
func (p *parser) parseExpression() ast.Expr {
	return p.parseExprPrec(0)
}

// parseExprPrec is the Pratt precedence-climbing core.
// It consumes the longest expression whose outermost binary operator has
// precedence strictly greater than minPrec.
//
// Left-associativity: when recursing for the RHS we pass the same prec as
// minPrec so that a same-precedence operator on the right breaks the loop
// (the loop condition is  prec <= minPrec, i.e. equal prec stops here).
func (p *parser) parseExprPrec(minPrec int) ast.Expr {
	left := p.parseUnaryExpr()
	for {
		op := p.peekNonTrivia()
		prec := p.infixPrec(op)
		if prec <= minPrec {
			break
		}
		p.nextNonTrivia()
		right := p.parseExprPrec(prec)
		left = &ast.BinaryExpr{Op: op.val, Left: left, Right: right}
	}
	return left
}

// ── Unary expressions ─────────────────────────────────────────────────────────

// parseUnaryExpr parses a unary prefix expression.
//
//	unary_expression :
//	  '!' unary_expression
//	| '~' unary_expression
//	| '-' unary_expression
//	| '*' unary_expression      -- pointer dereference
//	| '&' unary_expression      -- address-of
//	| singular_expression       -- postfix / primary
func (p *parser) parseUnaryExpr() ast.Expr {
	t := p.peekNonTrivia()
	switch t.typ {
	case tokenBang, tokenTilde, tokenMinus:
		p.nextNonTrivia()
		return &ast.UnaryExpr{Op: t.val, Operand: p.parseUnaryExpr()}
	case tokenStar:
		p.nextNonTrivia()
		return &ast.DerefExpr{Operand: p.parseUnaryExpr()}
	case tokenAmp:
		p.nextNonTrivia()
		return &ast.AddrOfExpr{Operand: p.parseUnaryExpr()}
	}
	return p.parsePostfixExpr()
}

// ── Postfix / primary expressions ────────────────────────────────────────────

// parsePostfixExpr applies zero or more postfix operations (indexing and
// member access) to a primary expression.
//
//	singular_expression :
//	  primary_expression component_or_swizzle_specifier*
//
//	component_or_swizzle_specifier :
//	  '[' expression ']'
//	| '.' ident
func (p *parser) parsePostfixExpr() ast.Expr {
	expr := p.parsePrimaryExpr()
	for {
		switch p.peekNonTrivia().typ {
		case tokenLBracket:
			p.expect(tokenLBracket)
			idx := p.parseExpression()
			p.expect(tokenRBracket)
			expr = &ast.IndexExpr{Base: expr, Index: idx}

		case tokenDot:
			p.nextNonTrivia()
			member := p.expect(tokenIdent)
			expr = &ast.MemberExpr{Base: expr, Member: member.val}

		default:
			return expr
		}
	}
}

// parsePrimaryExpr parses literals, parenthesised expressions, identifiers,
// and call expressions (including template-parameterised calls).
//
//	primary_expression :
//	  int_literal | float_literal | 'true' | 'false'
//	| '(' expression ')'
//	| ident
//	| ident '(' argument_expression_list? ')'              -- plain call
//	| ident '<' template_list '>' '(' argument_expression_list? ')' -- templated call
func (p *parser) parsePrimaryExpr() ast.Expr {
	t := p.nextNonTrivia()
	switch t.typ {
	case tokenNumber, tokenTrue, tokenFalse:
		return &ast.LitExpr{Val: t.val}

	case tokenLParen:
		inner := p.parseExpression()
		p.expect(tokenRParen)
		return &ast.ParenExpr{Inner: inner}
	case tokenPackage, tokenSuper:
		ident := t.val
		p.expect(tokenDoubleColon)
		for p.at(tokenIdent) {
			tok := p.nextNonTrivia()
			ident += "::" + tok.val
			if p.at(tokenDoubleColon) {
				p.nextNonTrivia()
			}
		}
		if p.at(tokenLParen) {
			return &ast.CallExpr{Callee: ident, Args: p.parseArgumentExpressionList()}
		}
		return &ast.Ident{Name: ident}

	case tokenIdent:
		ident := t.val
		// Consume namespace separators: foo::bar::baz
		for p.at(tokenDoubleColon) {
			p.nextNonTrivia() // consume ::
			seg := p.expect(tokenIdent)
			ident += "::" + seg.val
		}
		if ident != t.val {
			// Namespace-qualified path; template args not supported here.
			if p.at(tokenLParen) {
				return &ast.CallExpr{Callee: ident, Args: p.parseArgumentExpressionList()}
			}
			return &ast.Ident{Name: ident}
		}
		// Unqualified identifier: check for template args on built-in types.
		if isTemplateableIdent(t.val) && p.at(tokenLAngle) {
			targs := p.parseTemplateList()
			if p.at(tokenLParen) {
				return &ast.CallExpr{Callee: t.val, TemplateArgs: targs, Args: p.parseArgumentExpressionList()}
			}
			// Template-only (type specifier used as expression): vec3<f32> with no call.
			return &ast.CallExpr{Callee: t.val, TemplateArgs: targs}
		}
		if p.at(tokenLParen) {
			return &ast.CallExpr{Callee: t.val, Args: p.parseArgumentExpressionList()}
		}
		return &ast.Ident{Name: t.val}

	default:
		p.unexpected(t)
		return nil
	}
}

// isTypeConstructorKeyword returns true for keyword tokens that can appear as
// type names in call position (e.g. array, bool, i32, u32, f32, f16, …).
// In many WGSL lexers these are emitted as tokenIdent; this guard handles
// lexers that promote them to keyword tokens.
func isTypeConstructorKeyword(t token) bool {
	switch t.val {
	case "array", "atomic",
		"bool",
		"f16", "f32",
		"i32", "u32",
		"mat2x2", "mat2x3", "mat2x4",
		"mat3x2", "mat3x3", "mat3x4",
		"mat4x2", "mat4x3", "mat4x4",
		"ptr",
		"sampler", "sampler_comparison",
		"texture_1d", "texture_2d", "texture_2d_array",
		"texture_3d", "texture_cube", "texture_cube_array",
		"texture_depth_2d", "texture_depth_2d_array",
		"texture_depth_cube", "texture_depth_cube_array",
		"texture_depth_multisampled_2d",
		"texture_multisampled_2d",
		"texture_storage_1d", "texture_storage_2d",
		"texture_storage_2d_array", "texture_storage_3d",
		"vec2", "vec3", "vec4":
		return true
	}
	return false
}

// ── Template disambiguation (_disambiguate_template) ─────────────────────────

func (p *parser) parseTemplateList() []ast.Expr {
	p.expect(tokenLAngle)
	p.templateDepth++
	defer func() { p.templateDepth-- }()

	closeTemplate := func() bool {
		switch p.peekNonTrivia().typ {
		case tokenRAngle:
			p.nextNonTrivia()
			return true
		case tokenGtGt:
			p.nextNonTrivia()
			// Split: give the remaining '>' back to the enclosing template.
			p.buf = append([]token{{typ: tokenRAngle, val: ">"}}, p.buf...)
			return true
		}
		return false
	}

	var args []ast.Expr
	for {
		if closeTemplate() {
			return args
		}
		if len(args) > 0 {
			p.expect(tokenComma)
			if closeTemplate() { // trailing comma
				return args
			}
		}
		args = append(args, p.parseTemplateArg())
	}
}

// parseTemplateArg parses one argument inside a template list.
//
// Template args are either type specifiers (vec3<f32>, ptr<storage,T>) or
// const-expressions (integer literals, named constants, arithmetic on them).
// Identifiers are always parsed as type specifiers because:
//   - A named constant like 'N' becomes TypeSpecifier{Name:"N"}.asExpr() = Ident
//   - A nested type like 'vec3<f32>' is consumed greedily by parseTypeSpecifier,
//     which treats '<' as a template opener unconditionally — avoiding the need
//     for templateDepth to suppress '<' for that nested parse.
func (p *parser) parseTemplateArg() ast.Expr {
	if p.peekNonTrivia().typ == tokenIdent {
		return p.parseTypeSpecifier().AsExpr()
	}
	return p.parseExpression()
}

// ── Type specifier ────────────────────────────────────────────────────────────

// parseTypeSpecifier parses a type reference:
//
//	type_specifier : ident template_list?
func (p *parser) parseTypeSpecifier() ast.TypeSpecifier {
	name := p.nextNonTrivia()
	if name.typ != tokenIdent {
		p.unexpected(name)
	}
	var args []ast.Expr
	if p.at(tokenLAngle) { // type context: '<' is always a template opener
		args = p.parseTemplateList()
	}
	return ast.TypeSpecifier{Name: name.val, TemplateArgs: args}
}

// ── Optionally-typed identifier ───────────────────────────────────────────────

// parseOptionallyTypedIdent parses:
//
//	optionally_typed_ident : ident ( ':' type_specifier )?
func (p *parser) parseOptionallyTypedIdent() ast.OptionallyTypedIdent {
	name := p.expect(tokenIdent)
	oti := ast.OptionallyTypedIdent{Name: name.val}
	if p.at(tokenColon) {
		p.nextNonTrivia() // consume ':'
		ts := p.parseTypeSpecifier()
		oti.Type = &ts
	}
	return oti
}

// ── LHS expression ────────────────────────────────────────────────────────────

// parseLhsExpression parses an lhs_expression — the left-hand side of an
// assignment.  Unlike a general expression, this cannot contain binary
// operators or function calls at the top level.
//
//	lhs_expression :
//	  core_lhs_expression component_or_swizzle_specifier*
//	| '*' lhs_expression
//	| '&' lhs_expression
//
//	core_lhs_expression :
//	  ident
//	| '(' lhs_expression ')'
func (p *parser) parseLhsExpression() ast.Expr {
	t := p.peekNonTrivia()
	switch t.typ {
	case tokenStar:
		p.nextNonTrivia()
		return &ast.DerefExpr{Operand: p.parseLhsExpression()}
	case tokenAmp:
		p.nextNonTrivia()
		return &ast.AddrOfExpr{Operand: p.parseLhsExpression()}
	case tokenLParen:
		p.nextNonTrivia()
		inner := p.parseLhsExpression()
		p.expect(tokenRParen)
		return p.parseLhsPostfix(&ast.ParenExpr{Inner: inner})
	case tokenIdent:
		p.nextNonTrivia()
		return p.parseLhsPostfix(&ast.Ident{Name: t.val})
	default:
		p.unexpected(t)
		return nil
	}
}

// parseLhsPostfix applies member-access and index operations to an already-
// parsed lhs core expression.
func (p *parser) parseLhsPostfix(base ast.Expr) ast.Expr {
	for {
		switch p.peekNonTrivia().typ {
		case tokenLBracket:
			p.nextNonTrivia()
			idx := p.parseExpression()
			p.expect(tokenRBracket)
			base = &ast.IndexExpr{Base: base, Index: idx}
		case tokenDot:
			p.nextNonTrivia()
			member := p.expect(tokenIdent)
			base = &ast.MemberExpr{Base: base, Member: member.val}
		default:
			return base
		}
	}
}

// ── Call phrase ───────────────────────────────────────────────────────────────

// parseCallPhrase parses a call_phrase used in func_call_statement:
//
//	call_phrase : ident template_list? argument_expression_list
func (p *parser) parseCallPhrase() ast.CallExpr {
	name := p.expect(tokenIdent)
	var targs []ast.Expr
	if isTemplateableIdent(name.val) && p.at(tokenLAngle) {
		targs = p.parseTemplateList()
	}
	args := p.parseArgumentExpressionList()
	return ast.CallExpr{Callee: name.val, TemplateArgs: targs, Args: args}
}

func (p *parser) parseArgumentExpressionList() []ast.Expr {
	p.expect(tokenLParen)
	var args []ast.Expr
	for !p.at(tokenRParen) {
		if len(args) > 0 {
			p.expect(tokenComma)
			continue
		}
		args = append(args, p.parseExpression())
	}
	p.expect(tokenRParen)
	return args
}
