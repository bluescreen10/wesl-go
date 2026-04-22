package parser

import "github.com/bluescreen10/wesl-go/ast"

func (p *parser) parseAttributes() []ast.Attribute {
	var attrs []ast.Attribute

	for p.at(tokenAttr) {
		attr := p.parseAttribute()
		attrs = append(attrs, attr)
	}

	return attrs
}

func (p *parser) parseAttribute() ast.Attribute {
	tok := p.nextNonTrivia()

	// switch tok.typ {
	// case tokenIfAttr:
	// 	attr := p.parseIfAttrStmt()
	// 	attr.Then = p.parseStatement()
	// 	if p.peekNonTrivia().typ == tokenElseAttr {
	// 		attr.Else = p.parseStatement()
	// 	}
	// 	return attr
	// case
	if tok.typ == tokenAttr {
		var args []ast.Expr
		if p.at(tokenLParen) {
			args = p.parseAttributeExpressionList()
		}
		attr := ast.Attribute{
			Name: tok.val,
			Args: args,
		}
		return attr
	}

	p.unexpected(tok)
	panic("unreachable")
}

func (p *parser) parseAttributeExpressionList() []ast.Expr {
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

func (p *parser) parseIfAttrDecl() *ast.IfAttrDecl {
	p.nextNonTrivia() // consume @if token
	cond := p.parseExpression()
	then := p.parseTopLevelDecl()

	node := &ast.IfAttrDecl{Cond: cond, Then: then}
	if p.at(tokenElseAttr) {
		p.nextNonTrivia() // consume @else token
		node.Else = p.parseTopLevelDecl()
	}
	return node
}

func (p *parser) parseIfAttrStmt() *ast.IfAttrStmt {
	p.expect(tokenIfAttr)
	p.expect(tokenLParen)
	cond := p.parseExpression()
	if p.peekNonTrivia().typ == tokenComma {
		p.next() // trailing comman
	}
	p.expect(tokenRParen)

	then := p.parseStatement()

	var els ast.Stmt
	if p.at(tokenElseAttr) {
		p.nextNonTrivia()
		els = p.parseStatement()
	}

	return &ast.IfAttrStmt{Cond: cond, Then: then, Else: els}
}
