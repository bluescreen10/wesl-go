package parser

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bluescreen10/wesl-go/ast"
)

type tree struct {
}

type parser struct {
	input     string
	toks      [3]token
	peekCount int
	lex       *lexer

	imports []*ast.ImportDecl
}

func Parse(input string) (*tree, error) {
	p := &parser{input: input, lex: lex(input)}
	tree, err := p.Parse()
	return tree, err
}

func (p *parser) Parse() (t *tree, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	p.parseGlobalDecl()
	return nil, nil
}

func (p *parser) parseGlobalDecl() {
	for {
		tok := p.next()
		switch tok.typ {
		case tokenEOF:
			return
		case tokenError:
			panic(tok.val)
		case tokenImport:
			p.parseImportDecl()
		}
	}
}

func (p *parser) next() token {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.toks[0] = p.lex.nexttoken()
	}
	return p.toks[p.peekCount]
}

func (p *parser) peek() token {
	if p.peekCount > 0 {
		return p.toks[p.peekCount-1]
	}
	p.peekCount = 1
	p.toks[0] = p.lex.nexttoken()
	return p.toks[0]
}

func (p *parser) peekNonSpace() token {
	tok := p.nextNonSpace()
	p.backup()
	return tok
}

func (p *parser) backup() {
	p.peekCount++
}

func (p *parser) nextNonSpace() token {
	var tok token
	for {
		tok = p.next()
		if tok.typ != tokenSpace {
			break
		}
	}
	return tok
}

func (p *parser) expect(expected tokenType) token {
	tok := p.nextNonSpace()
	if tok.typ != expected {
		p.unexpected(tok)
	}
	return tok
}

func (p *parser) expectOneOf(expected ...tokenType) token {
	tok := p.nextNonSpace()
	if !slices.Contains(expected, tok.typ) {
		p.unexpected(tok)
	}
	return tok
}

func (p *parser) parseImportDecl() {
	p.parseImportPath("", true, true)
	p.expect(tokenSemicolon)
}

func (p *parser) parseImportPath(lineage string, allowSuper, allowBrace bool) {
	tok := p.nextNonSpace()

	switch tok.typ {
	case tokenLBrace:
		if !allowBrace {
			p.unexpected(tok)
		}
		p.parseImportList(lineage)
		p.expect(tokenRBrace)

	case tokenSuper:
		if !allowSuper || p.peekNonSpace().typ != tokenDoubleColon {
			p.unexpected(tok)
		}
		next := p.nextNonSpace()
		p.parseImportPath(lineage+tok.val+next.val, true, true)

	case tokenIdent:
		alias := tok.val
		next := p.peekNonSpace()

		if next.typ == tokenDoubleColon {
			p.nextNonSpace()
			p.parseImportPath(lineage+tok.val+next.val, false, true)
			return
		}

		if next.typ == tokenAs {
			p.nextNonSpace()
			alias = p.expect(tokenIdent).val
			next = p.peekNonSpace()
		}

		if next.typ == tokenRBrace || next.typ == tokenComma || next.typ == tokenSemicolon {
			p.imports = append(p.imports, &ast.ImportDecl{
				Symbol: tok.val,
				Path:   lineage,
				Alias:  alias,
			})
		} else {
			p.unexpected(next)
		}

	default:
		p.unexpected(tok)
	}
}

func (p *parser) parseImportList(lineage string) {
	for {
		p.parseImportPath(lineage, false, false)
		tok := p.nextNonSpace()
		switch tok.typ {
		case tokenComma:
			continue
		case tokenRBrace:
			p.backup()
			return
		default:
			p.unexpected(tok)
		}
	}
}

func (p *parser) unexpected(tok token) {
	p.errorf(tok, "unexpected %s", tok)
}

func (p *parser) errorf(tok token, format string, args ...any) {
	var lineStart, lineEnd int
	pos := tok.pos

	if pos >= len(p.input) {
		pos = len(p.input) - 1
	}

	for lineStart = pos; lineStart > 0; lineStart-- {
		if p.input[lineStart] == '\n' {
			lineStart++
			break
		}
	}

	for lineEnd = pos; lineEnd < len(p.input); lineEnd++ {
		if p.input[lineEnd] == '\n' {
			lineEnd--
			break
		}
	}

	errStr := fmt.Sprintf(format, args...)
	errLine := p.input[lineStart:lineEnd]
	errLocation := strings.Repeat(" ", pos-lineStart) + "^"
	panic(fmt.Errorf("%s\n%s\n%s", errStr, errLine, errLocation))
}
