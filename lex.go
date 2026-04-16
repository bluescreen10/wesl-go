package wesl

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type item struct {
	typ itemType
	pos int
	val string
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case i.typ > itemComment:
		return fmt.Sprintf("<%s>", i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

type itemType int

const (
	itemError itemType = iota
	itemEOF
	itemSpace
	itemSep
	itemNsSep
	itemTypeSep
	itemStmtSep
	itemLeftDelim
	itemRightDelim
	itemComment
	itemImport
	itemIf
	itemText
)

const eof = -1

var key = map[string]itemType{
	"import": itemImport,
	"@if":    itemIf,
}

type stateFn func(*lexer) stateFn

type lexer struct {
	input string
	pos   int
	start int
	item  *item
	line  int
	atEOF bool
}

func lex(input string) *lexer {
	l := &lexer{
		input: input,
	}
	return l
}

func (l *lexer) nextItem() item {
	l.item = nil
	state := lexDecl

	for l.pos <= len(l.input) && state != nil {
		state = state(l)
	}

	return *l.item
}

func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.atEOF = true
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += w
	if r == '\n' {
		l.line++
	}
	return r
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) backup() {
	if !l.atEOF && l.pos > 0 {
		r, w := utf8.DecodeLastRuneInString(l.input[:l.pos])
		l.pos -= w
		if r == '\n' {
			l.line--
		}
	}

	l.atEOF = false
}

func (l *lexer) emit(typ itemType) stateFn {
	l.item = &item{
		typ: typ,
		pos: l.start,
		val: l.input[l.start:l.pos],
	}
	l.start = l.pos
	return nil
}

func (l *lexer) errorf(format string, args ...any) stateFn {
	l.item = &item{
		typ: itemError,
		pos: l.start,
		val: fmt.Sprintf(format, args...),
	}
	l.start = 0
	l.pos = 0
	l.input = l.input[:0]
	return nil
}

func lexDecl(l *lexer) stateFn {
	switch r := l.next(); {
	case r == eof:
		l.emit(itemEOF)
		return nil
	case isSpace(r):
		l.backup()
		return lexSpace
	case r == ':':
		if l.peek() == ':' {
			l.next()
			return l.emit(itemNsSep)
		}
		return l.emit(itemTypeSep)
	case r == ',':
		return l.emit(itemSep)
	case r == ';':
		return l.emit(itemStmtSep)
	case r == '{':
		return l.emit(itemLeftDelim)
	case r == '}':
		return l.emit(itemRightDelim)
	case r == '/':
		if next := l.peek(); next == '/' || next == '*' {
			l.backup()
			l.backup()
			return lexComment
		}
		return l.errorf("unexpected character: %#U", r)
	case isAlphaNumeric(r):
		l.backup()
		return lexIdent
	default:
		fmt.Println(l.input[l.start:])
		return l.errorf("unrecognized character: %#U", r)
	}
}

func lexComment(l *lexer) stateFn {
	l.next()

	var r rune

	if l.next() == '/' {
		// inline comment
		for {
			r = l.next()
			if isNewline(r) {
				break
			}
		}
	} else {
		// comment block
		for {
			r = l.next()
			if r == '*' {
				if l.peek() == '/' {
					l.next()
					break
				}
			}
		}
	}
	return l.emit(itemComment)
}

func lexIdent(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):

		default:
			l.backup()
			word := l.input[l.start:l.pos]
			if typ, ok := key[word]; ok {
				return l.emit(typ)
			}
			return l.emit(itemText)
		}
	}
}

func lexSpace(l *lexer) stateFn {
	var r rune

	for {
		r = l.peek()
		if !isSpace(r) {
			break
		}
		l.next()
	}

	return l.emit(itemSpace)
}

func isAlphaNumeric(r rune) bool {
	return r == '@' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

func isNewline(r rune) bool {
	return r == -1 || r == '\n'
}
