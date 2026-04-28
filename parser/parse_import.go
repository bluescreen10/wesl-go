package parser

import "github.com/bluescreen10/wesl-go/ast"

func (p *parser) parseImportDecl() *ast.ImportDecl {
	p.expect(tokenImport)
	decl := p.parseImportPath()
	p.expect(tokenSemicolon)
	return decl
}

// parseImportPath builds an ImportDecl from the path after the 'import' keyword.
// It does not consume the trailing semicolon.
func (p *parser) parseImportPath() *ast.ImportDecl {
	decl := &ast.ImportDecl{}
	allowSpecial := true // package/super are only valid before any regular ident

	for {
		tok := p.nextNonTrivia()

		switch tok.typ {
		case tokenLBrace:
			decl.Items = p.parseImportItemList()
			p.expect(tokenRBrace)
			return decl

		case tokenPackage, tokenSuper:
			if !allowSpecial {
				p.unexpected(tok)
			}
			if p.peekNonTrivia().typ != tokenColonColon {
				p.unexpected(tok)
			}
			p.nextNonTrivia() // consume ::
			decl.Path = append(decl.Path, tok.val)

		case tokenIdent:
			allowSpecial = false
			next := p.peekNonTrivia()
			if next.typ == tokenColonColon {
				p.nextNonTrivia() // consume ::
				decl.Path = append(decl.Path, tok.val)
			} else {
				item := ast.ImportItem{Path: []string{tok.val}}
				if next.typ == tokenAs {
					p.nextNonTrivia() // consume 'as'
					item.Alias = p.expect(tokenIdent).val
				}
				decl.Items = []ast.ImportItem{item}
				return decl
			}

		default:
			p.unexpected(tok)
		}
	}
}

func (p *parser) parseImportItemList() []ast.ImportItem {
	var items []ast.ImportItem
	for {
		items = append(items, p.parseImportItem(nil)...)

		tok := p.nextNonTrivia()
		switch tok.typ {
		case tokenComma:
			if p.peekNonTrivia().typ == tokenRBrace {
				return items // trailing comma
			}
		case tokenRBrace:
			p.backup()
			return items
		default:
			p.unexpected(tok)
		}
	}
}

// parseImportItem parses one item (or a nested brace group) from a brace list,
// prepending prefix to every resulting item's path. Returns one or more items
// because a nested "d::{ e, f }" expands inline.
func (p *parser) parseImportItem(prefix []string) []ast.ImportItem {
	tok := p.expect(tokenIdent)
	next := p.peekNonTrivia()

	if next.typ == tokenColonColon {
		p.nextNonTrivia() // consume ::
		if p.peekNonTrivia().typ == tokenLBrace {
			// Nested brace group: d::{ e, f } — expand inline
			p.nextNonTrivia() // consume {
			sub := p.parseImportItemList()
			p.expect(tokenRBrace)
			newPrefix := append(append([]string{}, prefix...), tok.val)
			result := make([]ast.ImportItem, len(sub))
			for i, s := range sub {
				result[i] = ast.ImportItem{
					Path:  append(append([]string{}, newPrefix...), s.Path...),
					Alias: s.Alias,
				}
			}
			return result
		}
		// More path segments — recurse with extended prefix
		return p.parseImportItem(append(append([]string{}, prefix...), tok.val))
	}

	// Terminal segment, possibly with alias
	item := ast.ImportItem{Path: append(append([]string{}, prefix...), tok.val)}
	if next.typ == tokenAs {
		p.nextNonTrivia() // consume 'as'
		item.Alias = p.expect(tokenIdent).val
	}
	return []ast.ImportItem{item}
}
