package parser

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/bluescreen10/wesl-go/ast"
)

type tree struct {
}

type parser struct {
	input         string
	pos           int
	toks          [3]token
	peekCount     int
	templateDepth int

	lex *lexer

	ast ast.File

	buf []token //FIXME: remove
}

func Parse(input string) (*ast.File, error) {
	p := &parser{input: input, lex: lex(input)}
	return p.Parse()
}

func (p *parser) Parse() (decls *ast.File, err error) {
	defer func() {
		if e := recover(); e != nil {
			switch e := e.(type) {
			case error:
				err = e
			case string:
				err = errors.New(e)
			default:
				panic(e)
			}
		}
	}()
	for !p.at(tokenEOF) {
		decl := p.parseTopLevelDecl()
		switch any(decl).(type) {
		case []ast.ImportDecl:
			p.ast.Imports = append(p.ast.Imports, decl.([]ast.ImportDecl)...)
		default:
			p.ast.Decls = append(p.ast.Decls, decl)
		}
	}
	return &p.ast, nil
}

func (p *parser) parseTopLevelDecl() ast.Decl {
	if p.at(tokenIfAttr) {
		return p.parseIfAttrDecl()
	}

	attrs := p.parseAttributes()
	tok := p.peekNonTrivia()
	switch tok.typ {
	case tokenDiagnostic:
		return p.parseDiagnosticDirective(attrs)
	case tokenEnable:
		return p.parseEnableDirective(attrs)
	case tokenRequires:
		return p.parseRequiresDirective(attrs)
	case tokenConstAssert:
		return p.parseGlobalConstAssert(attrs)
	case tokenStruct:
		return p.parseStructDecl(attrs)
	case tokenAlias:
		return p.parseTypeAliasDecl(attrs)
	case tokenVar:
		return p.parseGlobalVariableDecl(attrs)
	case tokenConst, tokenOverride:
		return p.parseGlobalValueDecl(attrs)
	case tokenFn:
		return p.parseFnDecl(attrs)
	case tokenImport:
		return p.parseImportDecl()
	default:
		p.unexpected(tok)
	}
	panic("unreachable")
}

func (p *parser) next() token {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.toks[0] = p.lex.nextToken()
	}
	return p.toks[p.peekCount]
}

func (p *parser) peek() token {
	if p.peekCount > 0 {
		return p.toks[p.peekCount-1]
	}
	p.peekCount = 1
	p.toks[0] = p.lex.nextToken()
	return p.toks[0]
}

func (p *parser) peekNonTrivia() token {
	tok := p.nextNonTrivia()
	p.backup()
	return tok
}

func (p *parser) backup() {
	p.peekCount++
}

func (p *parser) nextNonTrivia() token {
	var tok token
	for {
		tok = p.next()
		if tok.typ != tokenSpace && tok.typ != tokenComment {
			break
		}
	}
	return tok
}

func (p *parser) at(typ tokenType) bool {
	tok := p.peekNonTrivia()
	return tok.typ == typ
}

func (p *parser) expect(expected tokenType) token {
	tok := p.nextNonTrivia()
	if tok.typ != expected {
		p.unexpected(tok)
	}
	return tok
}

func (p *parser) expectOneOf(expected ...tokenType) token {
	tok := p.nextNonTrivia()
	if !slices.Contains(expected, tok.typ) {
		p.unexpected(tok)
	}
	return tok
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
