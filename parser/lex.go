package parser

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type token struct {
	typ tokenType
	pos int
	val string
}

func (i token) String() string {
	switch {
	case i.typ == tokenEOF:
		return "EOF"
	case i.typ == tokenError:
		return i.val
	case i.typ > tokenComment:
		return fmt.Sprintf("<%s>", i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

type tokenType int

const (
	tokenError tokenType = iota
	tokenEOF
	tokenSpace
	tokenComma
	tokenDoubleColon
	tokenColon
	tokenSemicolon
	tokenLBrace
	tokenRBrace
	tokenLParen
	tokenRParen
	tokenLBracket
	tokenRBracket
	tokenComment
	tokenIdent
	tokenImport
	tokenAs
	tokenSuper
	tokenAtIf
)

const eof = -1

var key = map[string]tokenType{
	"import": tokenImport,
	"as":     tokenAs,
	"@if":    tokenAtIf,
	"super":  tokenSuper,
}

type stateFn func(*lexer) stateFn

type lexer struct {
	input string
	pos   int
	start int
	token *token
	line  int
	atEOF bool
}

func lex(input string) *lexer {
	l := &lexer{
		input: input,
	}
	return l
}

func (l *lexer) nexttoken() token {
	l.token = nil
	state := lexDecl

	for l.pos <= len(l.input) && state != nil {
		state = state(l)
	}

	return *l.token
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

func (l *lexer) emit(typ tokenType) stateFn {
	l.token = &token{
		typ: typ,
		pos: l.start,
		val: l.input[l.start:l.pos],
	}
	l.start = l.pos
	return nil
}

func (l *lexer) errorf(format string, args ...any) stateFn {
	l.token = &token{
		typ: tokenError,
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
		l.emit(tokenEOF)
		return nil
	case isSpace(r):
		l.backup()
		return lexSpace
	case r == ':':
		if l.peek() == ':' {
			l.next()
			return l.emit(tokenDoubleColon)
		}
		return l.emit(tokenColon)
	case r == ',':
		return l.emit(tokenComma)
	case r == ';':
		return l.emit(tokenSemicolon)
	case r == '{':
		return l.emit(tokenLBrace)
	case r == '}':
		return l.emit(tokenRBrace)
	case r == '(':
		return l.emit(tokenLParen)
	case r == ')':
		return l.emit(tokenRParen)
	case r == '[':
		return l.emit(tokenLBracket)
	case r == ']':
		return l.emit(tokenRBracket)
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
	return l.emit(tokenComment)
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
			return l.emit(tokenIdent)
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

	return l.emit(tokenSpace)
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
