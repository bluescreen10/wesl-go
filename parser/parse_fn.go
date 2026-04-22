package parser

import "github.com/bluescreen10/wesl-go/ast"

func (p *parser) parseFnDecl(attrs []ast.Attribute) *ast.FnDecl {
	p.expect(tokenFn)
	name := p.expect(tokenIdent)

	p.expect(tokenLParen)
	var params []ast.Param

	for !p.at(tokenRParen) {
		if p.at(tokenIfAttr) {
			params = append(params, p.parseIfAttrParam())
		} else {
			params = append(params, p.parseParam())
		}
		if p.at(tokenComma) {
			p.nextNonTrivia()
		}
	}

	p.expect(tokenRParen)

	var retAttrs []ast.Attribute
	var retType *ast.TypeSpecifier
	if p.at(tokenArrow) {
		p.nextNonTrivia()
		retAttrs = p.parseAttributes()
		typ := p.parseTypeSpecifier()
		retType = &typ
	}

	body := p.parseCompoundStatement(nil)

	return &ast.FnDecl{
		Attrs:       attrs,
		Name:        name.val,
		Params:      params,
		ReturnAttrs: retAttrs,
		ReturnType:  retType,
		Body:        body,
	}
}

func (p *parser) parseIfAttrParam() ast.IfAttrParam {
	p.nextNonTrivia() // consume @if token
	cond := p.parseExpression()
	then := p.parseParam()

	node := ast.IfAttrParam{Cond: cond, Then: then}
	if p.at(tokenElseAttr) {
		p.nextNonTrivia() // consume @else token
		els := p.parseParam()
		node.Else = &els
	}
	return node
}

func (p *parser) parseParam() ast.FnParam {
	paramAttrs := p.parseAttributes()
	paramName := p.expect(tokenIdent)
	p.expect(tokenColon)
	paramType := p.parseTypeSpecifier()
	return ast.FnParam{
		Name:  paramName.val,
		Type:  paramType,
		Attrs: paramAttrs,
	}
}
