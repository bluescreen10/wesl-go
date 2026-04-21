package parser

import "github.com/bluescreen10/wesl-go/ast"

// ── Directives ────────────────────────────────────────────────────────────────

// parseDiagnosticDirective parses:
//
//	attribute* 'diagnostic' diagnostic_control ';'
//
//	diagnostic_control :
//	  '(' severity_control_name ',' diagnostic_rule_name ','? ')'
//
//	severity_control_name : 'error' | 'warning' | 'info' | 'off'
//	diagnostic_rule_name  : ident ( '.' ident )?
func (p *parser) parseDiagnosticDirective(attrs []ast.Attribute) *ast.DiagnosticDirective {
	p.expect(tokenDiagnostic)
	ctrl := p.parseDiagnosticControl()
	p.expect(tokenSemicolon)
	return &ast.DiagnosticDirective{Attrs: attrs, Control: ctrl}
}

func (p *parser) parseDiagnosticControl() ast.DiagnosticControl {
	p.expect(tokenLParen)

	severity := p.nextNonTrivia()
	if severity.typ != tokenIdent {
		p.unexpected(severity)
	}
	p.expect(tokenComma)

	// diagnostic_rule_name: ident ( '.' ident )?
	name := p.expect(tokenIdent)
	ruleName := name.val
	if p.at(tokenDot) {
		p.nextNonTrivia() // consume '.'
		sub := p.expect(tokenIdent)
		ruleName += "." + sub.val
	}

	if p.peekNonTrivia().typ == tokenComma {
		p.nextNonTrivia()
	}
	p.expect(tokenRParen)

	return ast.DiagnosticControl{Severity: severity.val, RuleName: ruleName}
}

// parseEnableDirective parses:
//
//	attribute* 'enable' enable_extension_list ';'
//
//	enable_extension_list : ident ( ',' ident )* ','?
func (p *parser) parseEnableDirective(attrs []ast.Attribute) *ast.EnableDirective {
	p.expect(tokenEnable)
	extensions := p.parseIdentList()
	p.expect(tokenSemicolon)
	return &ast.EnableDirective{Attrs: attrs, Extensions: extensions}
}

// parseRequiresDirective parses:
//
//	attribute* 'requires' software_extension_list ';'
//
//	software_extension_list : ident ( ',' ident )* ','?
func (p *parser) parseRequiresDirective(attrs []ast.Attribute) *ast.RequiresDirective {
	p.expect(tokenRequires)
	extensions := p.parseIdentList()
	p.expect(tokenSemicolon)
	return &ast.RequiresDirective{Attrs: attrs, Extensions: extensions}
}

// parseIdentList parses a non-empty comma-separated list of identifiers with
// an optional trailing comma.
func (p *parser) parseIdentList() []string {
	var idents []string
	for {
		name := p.expect(tokenIdent)
		idents = append(idents, name.val)
		if !p.at(tokenComma) {
			break
		}
		p.nextNonTrivia() // consume ','
		// Trailing comma: stop if the next token is not an ident.
		if p.peekNonTrivia().typ != tokenIdent {
			break
		}
	}
	return idents
}

// ── struct_decl ───────────────────────────────────────────────────────────────

// parseStructDecl parses:
//
//	attribute* 'struct' ident struct_body_decl
//
//	struct_body_decl : '{' struct_member* '}'
//	struct_member    : attribute* member_ident ':' type_specifier
func (p *parser) parseStructDecl(attrs []ast.Attribute) *ast.StructDecl {
	p.expect(tokenStruct)
	name := p.expect(tokenIdent)

	p.expect(tokenLBrace)
	var members []ast.StructMember
	for !p.at(tokenRBrace) {
		if p.at(tokenIfAttr) {
			members = append(members, p.parseIfAttrStructField())
		} else {
			members = append(members, p.parseStructField())
		}
		// Struct members are separated by (optional) commas in WGSL.
		if p.at(tokenComma) {
			p.nextNonTrivia()
		}
	}
	p.expect(tokenRBrace)

	return &ast.StructDecl{Attrs: attrs, Name: name.val, Members: members}
}

func (p *parser) parseStructField() ast.StructField {
	attrs := p.parseAttributes()
	name := p.expect(tokenIdent)
	p.expect(tokenColon)
	typ := p.parseTypeSpecifier()
	return ast.StructField{Attrs: attrs, Name: name.val, Type: typ}
}

func (p *parser) parseIfAttrStructField() ast.IfAttrStructField {
	p.nextNonTrivia() // consume @if token
	cond := p.parseExpression()
	then := p.parseStructField()

	node := ast.IfAttrStructField{Cond: cond, Then: then}
	if p.at(tokenElseAttr) {
		p.nextNonTrivia() // consume @else token
		els := p.parseStructField()
		node.Else = &els
	}
	return node
}

// ── type_alias_decl ───────────────────────────────────────────────────────────

// parseTypeAliasDecl parses:
//
//	attribute* 'alias' ident '=' type_specifier
func (p *parser) parseTypeAliasDecl(attrs []ast.Attribute) *ast.TypeAliasDecl {
	p.expect(tokenAlias)
	name := p.expect(tokenIdent)
	p.expect(tokenEqual)
	ts := p.parseTypeSpecifier()
	p.expect(tokenSemicolon)
	return &ast.TypeAliasDecl{Attrs: attrs, Name: name.val, Type: ts}
}

// ── variable_decl ─────────────────────────────────────────────────────────────

// parseVariableDecl parses the core variable declaration form:
//
//	attribute* 'var' _disambiguate_template template_list? optionally_typed_ident
//
// The attrs argument contains attributes already consumed by the caller.
// The 'var' keyword is consumed here.
func (p *parser) parseVariableDecl(attrs []ast.Attribute) ast.VariableDecl {
	p.expect(tokenVar)
	var templateArgs []ast.Expr
	if p.at(tokenLAngle) {
		templateArgs = p.parseTemplateList()
	}
	return ast.VariableDecl{Attrs: attrs, TemplateArgs: templateArgs, Ident: p.parseOptionallyTypedIdent()}
}

// ── global_variable_decl ──────────────────────────────────────────────────────

// parseGlobalVariableDecl parses:
//
//	variable_decl ( '=' expression )?
func (p *parser) parseGlobalVariableDecl(attrs []ast.Attribute) *ast.GlobalVariableDecl {
	decl := p.parseVariableDecl(attrs)
	gvd := &ast.GlobalVariableDecl{Decl: decl}
	if p.at(tokenEqual) {
		p.nextNonTrivia()
		gvd.Init = p.parseExpression()
	}
	p.expect(tokenSemicolon)
	return gvd
}

// ── global_value_decl ─────────────────────────────────────────────────────────

// parseGlobalValueDecl parses:
//
//	attribute* 'const'    optionally_typed_ident '=' expression
//	attribute* 'override' optionally_typed_ident ( '=' expression )?
func (p *parser) parseGlobalValueDecl(attrs []ast.Attribute) *ast.GlobalValueDecl {
	kw := p.nextNonTrivia() // 'const' or 'override'
	if kw.typ != tokenConst && kw.typ != tokenOverride {
		p.unexpected(kw)
	}

	ident := p.parseOptionallyTypedIdent()
	gvd := &ast.GlobalValueDecl{Attrs: attrs, Keyword: kw.val, Ident: ident}

	switch kw.typ {
	case tokenConst:
		// Initialiser is mandatory for const.
		p.expect(tokenEqual)
		gvd.Init = p.parseExpression()

	case tokenOverride:
		// Initialiser is optional for override.
		if p.at(tokenEqual) {
			p.nextNonTrivia()
			gvd.Init = p.parseExpression()
		}
	}

	p.expect(tokenSemicolon)
	return gvd
}

func (p *parser) parseGlobalConstAssert(attrs []ast.Attribute) *ast.ConstAssertStmt {
	stmt := p.parseConstAssertStatement(attrs)
	p.expect(tokenSemicolon)
	return stmt
}
