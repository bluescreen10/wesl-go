package parser

import "github.com/bluescreen10/wesl-go/ast"

func (p *parser) parseImportDecl() []ast.ImportDecl {
	p.expect(tokenImport)
	decl := p.parseImportPath("", true, true)
	p.expect(tokenSemicolon)
	return decl
}

func (p *parser) parseImportPath(lineage string, allowSuper, allowBrace bool) []ast.ImportDecl {
	tok := p.nextNonTrivia()

	switch tok.typ {
	case tokenLBrace:
		if !allowBrace {
			p.unexpected(tok)
		}
		decls := p.parseImportList(lineage)
		p.expect(tokenRBrace)
		return decls

	case tokenSuper:
		if !allowSuper || p.peekNonTrivia().typ != tokenDoubleColon {
			p.unexpected(tok)
		}
		next := p.nextNonTrivia()
		return p.parseImportPath(lineage+tok.val+next.val, true, true)

	case tokenIdent:
		alias := tok.val
		next := p.peekNonTrivia()

		if next.typ == tokenDoubleColon {
			p.nextNonTrivia()
			return p.parseImportPath(lineage+tok.val+next.val, false, true)
		}

		if next.typ == tokenAs {
			p.nextNonTrivia()
			alias = p.expect(tokenIdent).val
			next = p.peekNonTrivia()
		}

		if next.typ == tokenRBrace || next.typ == tokenComma || next.typ == tokenSemicolon {
			return []ast.ImportDecl{{
				Symbol: tok.val,
				Path:   lineage,
				Alias:  alias,
			}}
		} else {
			p.unexpected(next)
		}

	default:
		p.unexpected(tok)
	}
	panic("unreachable")
}

func (p *parser) parseImportList(lineage string) []ast.ImportDecl {
	var decls []ast.ImportDecl
	for {
		decls = append(decls, p.parseImportPath(lineage, false, false)...)
		tok := p.nextNonTrivia()
		switch tok.typ {
		case tokenComma:
			continue
		case tokenRBrace:
			p.backup()
			return decls
		default:
			p.unexpected(tok)
		}
	}
}
