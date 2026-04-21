package parser

import "github.com/bluescreen10/wesl-go/ast"

func (p *parser) parseFnDecl(attrs []ast.Attribute) *ast.FnDecl {
	p.expect(tokenFn)
	name := p.expect(tokenIdent)

	p.expect(tokenLParen)
	var params []ast.Param

	for !p.at(tokenRParen) {
		if len(params) > 0 {
			p.expect(tokenComma)
			continue
		}

		paramAttrs := p.parseAttributes()
		paramName := p.expect(tokenIdent)
		paramType := p.parseTypeSpecifier()
		param := ast.Param{
			Name:  paramName.val,
			Type:  paramType,
			Attrs: paramAttrs,
		}
		params = append(params, param)
	}

	p.expect(tokenRParen)

	var retAttrs []ast.Attribute
	var retType ast.TypeSpecifier
	if p.at(tokenArrow) {
		p.nextNonTrivia()
		retAttrs = p.parseAttributes()
		retType = p.parseTypeSpecifier()
	}

	p.parseCompoundStatement(nil)

	return &ast.FnDecl{
		Attrs:       attrs,
		Name:        name.val,
		Params:      params,
		ReturnAttrs: retAttrs,
		ReturnType:  retType,
		Body:        nil,
	}
}
