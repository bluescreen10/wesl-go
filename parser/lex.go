package parser

import (
	"fmt"
	"strings"
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
	// special
	tokenError tokenType = iota
	tokenEOF
	tokenSpace

	// punctuation
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

	// WGSL
	tokenComment
	tokenIdent
	tokenAttr
	tokenFn
	tokenArrow
	tokenDot
	tokenTrue
	tokenFalse
	tokenNumber
	tokenDiagnostic
	tokenEnable
	tokenRequires
	tokenStruct
	tokenAlias
	tokenVar
	tokenLet
	tokenConst
	tokenOverride
	tokenUnderscore

	//WGSL Flow Control
	tokenLoop
	tokenFor
	tokenWhile
	tokenBreak
	tokenContinue
	tokenContinuing
	tokenDiscard
	tokenIf
	tokenElse
	tokenReturn
	tokenConstAssert
	tokenSwitch
	tokenCase
	tokenDefault

	//WGSL Asignment
	tokenPlusEq
	tokenMinusEq
	tokenStarEq
	tokenSlashEq
	tokenPercentEq
	tokenAmpEq
	tokenPipeEq
	tokenCaretEq
	tokenLtLtEq
	tokenGtGtEq

	// Operators
	tokenPipePipe
	tokenAmpAmp
	tokenPlusPlus
	tokenMinusMinus
	tokenPipe
	tokenCaret
	tokenAmp
	tokenEqualEqual
	tokenBangEqual
	tokenBang
	tokenLAngle
	tokenRAngle
	tokenLtEqual
	tokenGtEqual
	tokenLtLt
	tokenGtGt
	tokenPlus
	tokenMinus
	tokenStar
	tokenSlash
	tokenPercent
	tokenTilde
	tokenEqual

	// Wesl
	tokenImport
	tokenAs
	tokenSuper
	tokenIfAttr
	tokenElseAttr
)

const eof = -1

var key = map[string]tokenType{
	"import":       tokenImport,
	"as":           tokenAs,
	"super":        tokenSuper,
	"fn":           tokenFn,
	"true":         tokenTrue,
	"false":        tokenFalse,
	"diagnostic":   tokenDiagnostic,
	"enable":       tokenEnable,
	"requires":     tokenRequires,
	"struct":       tokenStruct,
	"alias":        tokenAlias,
	"var":          tokenVar,
	"let":          tokenLet,
	"const":        tokenConst,
	"override":     tokenOverride,
	"loop":         tokenLoop,
	"for":          tokenFor,
	"while":        tokenWhile,
	"if":           tokenIf,
	"else":         tokenElse,
	"return":       tokenReturn,
	"continue":     tokenContinue,
	"continuing":   tokenContinuing,
	"discard":      tokenDiscard,
	"const_assert": tokenConstAssert,
	"_":            tokenUnderscore,
	"switch":       tokenSwitch,
	"case":         tokenCase,
	"default":      tokenDefault,
}

var attr = map[string]tokenType{
	"@if":   tokenIfAttr,
	"@else": tokenElseAttr,
}

var operators = map[string]tokenType{
	"||":  tokenPipePipe,
	"&&":  tokenAmpAmp,
	"++":  tokenPlusPlus,
	"--":  tokenMinusMinus,
	"|":   tokenPipe,
	"^":   tokenCaret,
	"&":   tokenAmp,
	"==":  tokenEqualEqual,
	"!=":  tokenBangEqual,
	"<":   tokenLAngle,
	">":   tokenRAngle,
	"<=":  tokenLtEqual,
	">=":  tokenGtEqual,
	"<<":  tokenLtLt,
	">>":  tokenGtGt,
	"+":   tokenPlus,
	"-":   tokenMinus,
	"*":   tokenStar,
	"/":   tokenSlash,
	"%":   tokenPercent,
	"!":   tokenBang,
	"=":   tokenEqual,
	"+=":  tokenPlusEq,
	"-=":  tokenMinusEq,
	"*=":  tokenStarEq,
	"/=":  tokenSlashEq,
	"%=":  tokenPercentEq,
	"&=":  tokenAmpEq,
	"|=":  tokenPipeEq,
	"^=":  tokenCaretEq,
	"<<=": tokenLtLtEq,
	">>=": tokenGtGtEq,
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

func (l *lexer) nextToken() token {
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

func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
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
	case r == '.':
		return l.emit(tokenDot)
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
		next := l.peek()
		if next == '/' || next == '*' {
			l.backup()
			return lexComment
		}
		return l.errorf("unexpected character: %#U", r)
	case r == '@':
		next := l.peek()
		if isAlphaNumeric(next) {
			l.backup()
			return lexAttribute
		}
		return l.errorf("unexpected character: %#U", r)
	case r == '+' || r == '-':
		next := l.peek()
		l.backup()
		if isNumber(next) {
			return lexNumber
		} else {
			return lexOperator
		}
	case isNumber(r):
		l.backup()
		return lexNumber
	case isOperator(r):
		l.backup()
		return lexOperator
	case isAlphaNumeric(r):
		l.backup()
		return lexIdent
	default:
		return l.errorf("unrecognized character: %#U", r)
	}
}

func lexNumber(l *lexer) stateFn {
	digits := "0123456789"

	l.accept("+-")

	if l.accept("0") {
		if l.accept("xX") {
			digits = "0123456789abcdefABCDEF"
		}
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if len(digits) == 10+1 && l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}

	l.acceptRun("iuhf")

	if isAlphaNumeric(l.peek()) {
		l.next()
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	return l.emit(tokenNumber)
}

func lexOperator(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case isOperator(r):
			// consume it
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			if typ, ok := operators[word]; ok {
				return l.emit(typ)
			}
			return l.errorf("unrecognized operator: %s", word)
		}
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
			//FIXME: WGSL allows nested comments
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

func lexAttribute(l *lexer) stateFn {
	l.next()

	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):

		default:
			l.backup()
			word := l.input[l.start:l.pos]
			if typ, ok := attr[word]; ok {
				return l.emit(typ)
			}
			return l.emit(tokenAttr)
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
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

func isNewline(r rune) bool {
	return r == -1 || r == '\n'
}

func isOperator(r rune) bool {
	return r == '-' ||
		r == '+' ||
		r == '<' ||
		r == '>' ||
		r == '=' ||
		r == '!' ||
		r == '/' ||
		r == '%' ||
		r == '|' ||
		r == '%' ||
		r == '*' ||
		r == '^' ||
		r == '~'
}

func isNumber(r rune) bool {
	return ('0' <= r && r <= '9')
}
