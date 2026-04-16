package wesl

import (
	"errors"
	"fmt"
	"strings"
)

type tree struct {
	root *listNode
}

type parser struct {
	input     string
	token     [3]item
	peekCount int
	lex       *lexer
}

func parse(input string) (*tree, error) {
	p := &parser{input: input, lex: lex(input)}
	return p.Parse()
}

func (p *parser) Parse() (t *tree, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
			fmt.Println(err)
		}
	}()

	for {
		item := p.next()
		switch item.typ {
		case itemEOF:
			return nil, nil
		case itemError:
			return nil, errors.New(item.val)
		case itemImport:
			p.parseImport()
		}
	}
}

func (p *parser) next() item {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.token[0] = p.lex.nextItem()
	}
	return p.token[p.peekCount]
}

func (p *parser) peek() item {
	if p.peekCount > 0 {
		return p.token[p.peekCount-1]
	}
	p.peekCount = 1
	p.token[0] = p.lex.nextItem()
	return p.token[0]
}

func (p *parser) backup() {
	p.peekCount++
}

func (p *parser) nextNonSpace() item {
	var token item
	for {
		token = p.next()
		if token.typ != itemSpace {
			break
		}
	}
	return token
}

func (p *parser) expect(expected itemType) item {
	token := p.nextNonSpace()
	if token.typ != expected {
		p.unexpected(token)
	}
	return token
}

func (p *parser) expectOneOf(expected ...itemType) item {
	token := p.nextNonSpace()
	for _, e := range expected {
		if token.typ == e {
			return token
		}
	}
	p.unexpected(token)
	return token
}

func (p *parser) parseImport() *tree {
	switch token := p.expectOneOf(itemText, itemLeftDelim); {
	case token.typ == itemText:
		if p.peek().typ == itemNsSep {
			p.backup2()

		}
	}
	p.expect(itemStmtSep)
	return nil
}

func (p *parser) unexpected(token item) {
	p.errorf(token, "unexpected %s", token)
}

func (p *parser) errorf(token item, format string, args ...any) {

	var lineStart, lineEnd int
	pos := token.pos

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
