package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bluescreen10/wesl-go/ast"
)

type parser struct {
	input   string
	tok     token
	nextTok token
	lex     *lexer

	templateDepth int
	pendingClose  int // pending '>' from a split '>>' token inside a template
}

func Parse(input string) (*ast.File, error) {
	p := &parser{input: input, lex: lex(input)}
	p.init()
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

	ast := ast.File{}

	for !p.at(tokenEOF) {
		ast.Decls = append(ast.Decls, p.parseTopLevelDecl())
	}
	return &ast, nil
}

func (p *parser) init() {
	p.tok = p.next0()
	p.nextTok = p.next0()
}

func (p *parser) peek() token {
	return p.tok
}

func (p *parser) next() token {
	tok := p.tok
	p.tok = p.nextTok
	p.nextTok = p.next0()
	return tok
}

func (p *parser) next0() token {
	for {
		tok := p.lex.nextToken()
		if tok.typ != tokenSpace && tok.typ != tokenComment {
			return tok
		}
	}
}

func (p *parser) accept(typ tokenType) bool {
	if p.at(typ) {
		p.next()
		return true
	}
	return false
}

func (p *parser) at(typ tokenType) bool {
	return p.tok.typ == typ
}

func (p *parser) expect(expected tokenType) token {
	tok := p.next()
	if tok.typ != expected {
		p.unexpected(tok)
	}
	return tok
}

func (p *parser) expectOneOf(expected ...tokenType) token {
	tok := p.next()
	for _, e := range expected {
		if tok.typ == e {
			return tok
		}
	}
	p.unexpected(tok)
	panic("unreachable")
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

func (p *parser) parseAttributes() []ast.Attribute {
	var attrs []ast.Attribute

	for p.at(tokenAttr) {
		attr := p.parseAttribute()
		attrs = append(attrs, attr)
	}

	return attrs
}

func (p *parser) parseAttribute() ast.Attribute {
	tok := p.expect(tokenAttr)

	var args []ast.Expr
	if p.at(tokenLParen) {
		args = p.parseAttributeExpressionList()
	}

	attr := ast.Attribute{
		Name: tok.val,
		Args: args,
	}

	return attr
}

func (p *parser) parseAttributeExpressionList() []ast.Expr {
	p.expect(tokenLParen)

	var args []ast.Expr
	for !p.at(tokenRParen) {
		if len(args) > 0 {
			p.expect(tokenComma)
			continue
		}

		args = append(args, p.parseExpression())
	}

	p.expect(tokenRParen)
	return args
}

func (p *parser) parseIfAttrDecl() *ast.IfAttrDecl {
	p.expect(tokenIfAttr)
	cond := p.parseExpression()
	then := p.parseTopLevelDecl()

	var els ast.Decl
	if p.accept(tokenElseAttr) {
		els = p.parseTopLevelDecl()
	}

	return &ast.IfAttrDecl{Cond: cond, Then: then, Else: els}
}

func (p *parser) parseIfAttrStmt() *ast.IfAttrStmt {
	p.expect(tokenIfAttr)
	cond := p.parseExpression()
	then := p.parseStatement()

	var els ast.Stmt
	if p.accept(tokenElseAttr) {
		els = p.parseStatement()
	}

	return &ast.IfAttrStmt{Cond: cond, Then: then, Else: els}
}

// ----------------------------------------------------------------------------
// Global Declarations

func (p *parser) parseTopLevelDecl() ast.Decl {
	if p.at(tokenIfAttr) {
		return p.parseIfAttrDecl()
	}

	attrs := p.parseAttributes()
	tok := p.peek()
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
	case tokenFunc:
		return p.parseFuncDecl(attrs)
	case tokenImport:
		return p.parseImportDecl()
	default:
		p.unexpected(tok)
	}
	panic("unreachable")
}

//	attribute* 'diagnostic' diagnostic_control ';'
//
//	diagnostic_control :
//	  '(' severity_control_name ',' diagnostic_rule_name ','? ')'
//
//	severity_control_name : 'error' | 'warning' | 'info' | 'off'
//	diagnostic_rule_name  : ident ( '.' ident )?
//
// parseDiagnosticDirective parses a WGSL diagnostic directive
func (p *parser) parseDiagnosticDirective(attrs []ast.Attribute) *ast.DiagnosticDirective {
	p.expect(tokenDiagnostic)
	ctrl := p.parseDiagnosticControl()
	p.expect(tokenSemicolon)
	return &ast.DiagnosticDirective{Attrs: attrs, Control: ctrl}
}

// diagnostic_rule_name: ident ( '.' ident )?
//
// parseDiagnosticControl
func (p *parser) parseDiagnosticControl() ast.DiagnosticControl {
	p.expect(tokenLParen)

	severity := p.expect(tokenIdent)
	p.expect(tokenComma)

	name := p.expect(tokenIdent)
	ruleName := name.val

	if p.accept(tokenDot) {
		sub := p.expect(tokenIdent)
		ruleName += "." + sub.val
	}

	p.accept(tokenComma)
	p.expect(tokenRParen)

	return ast.DiagnosticControl{Severity: severity.val, RuleName: ruleName}
}

//	attribute* 'enable' enable_extension_list ';'
//	enable_extension_list : ident ( ',' ident )* ','?
//
// parseEnableDirective parses a WGSL enable directive
func (p *parser) parseEnableDirective(attrs []ast.Attribute) *ast.EnableDirective {
	p.expect(tokenEnable)
	extensions := p.parseIdentList()
	p.expect(tokenSemicolon)
	return &ast.EnableDirective{Attrs: attrs, Extensions: extensions}
}

//	attribute* 'requires' software_extension_list ';'
//	software_extension_list : ident ( ',' ident )* ','?
//
// parseRequiresDirective parses a WGSL requires directive
func (p *parser) parseRequiresDirective(attrs []ast.Attribute) *ast.RequiresDirective {
	p.expect(tokenRequires)
	extensions := p.parseIdentList()
	p.expect(tokenSemicolon)
	return &ast.RequiresDirective{Attrs: attrs, Extensions: extensions}
}

// parseIdentList parses a non-empty comma-separated list of identifiers with
// an optional trailing comma.
func (p *parser) parseIdentList() []string {
	var idents []string
	for p.at(tokenIdent) {
		tok := p.expect(tokenIdent)
		idents = append(idents, tok.val)
		p.accept(tokenComma)
	}
	return idents
}

//	attribute* 'struct' ident struct_body_decl
//	struct_body_decl : '{' struct_member* '}'
//	struct_member    : attribute* member_ident ':' type_specifier
//
// parseStructDecl parses a struct declaration
func (p *parser) parseStructDecl(attrs []ast.Attribute) *ast.StructDecl {
	p.expect(tokenStruct)
	tok := p.expect(tokenIdent)

	p.expect(tokenLBrace)
	var members []ast.StructMember
	for !p.at(tokenRBrace) {
		if p.at(tokenIfAttr) {
			members = append(members, p.parseIfAttrStructField())
		} else {
			members = append(members, p.parseStructField())
		}
		p.accept(tokenComma)
	}
	p.expect(tokenRBrace)

	return &ast.StructDecl{Attrs: attrs, Name: tok.val, Members: members}
}

func (p *parser) parseStructField() *ast.StructField {
	attrs := p.parseAttributes()
	tok := p.expect(tokenIdent)
	p.expect(tokenColon)
	typ := p.parseTypeSpecifier()
	return &ast.StructField{Attrs: attrs, Name: tok.val, Type: typ}
}

func (p *parser) parseIfAttrStructField() *ast.IfAttrStructField {
	p.expect(tokenIfAttr)
	cond := p.parseExpression()
	then := p.parseStructField()

	var els ast.StructMember
	if p.accept(tokenComma) {
		if p.accept(tokenElseAttr) {
			els = p.parseStructField()
		}
	}
	return &ast.IfAttrStructField{Cond: cond, Then: then, Else: els}
}

//	attribute* 'alias' ident '=' type_specifier
//
// parseTypeAliasDecl parses an alias type declaration
func (p *parser) parseTypeAliasDecl(attrs []ast.Attribute) *ast.TypeAliasDecl {
	p.expect(tokenAlias)
	tok := p.expect(tokenIdent)
	p.expect(tokenEqual)
	ts := p.parseTypeSpecifier()
	p.expect(tokenSemicolon)
	return &ast.TypeAliasDecl{Attrs: attrs, Name: tok.val, Type: ts}
}

//	variable_decl ( '=' expression )?
//
// parseGlobalVariableDecl parses a global variable declaration
func (p *parser) parseGlobalVariableDecl(attrs []ast.Attribute) *ast.GlobalVariableDecl {
	decl := p.parseVariableDecl(attrs)

	var init ast.Expr
	if p.accept(tokenEqual) {
		init = p.parseExpression()
	}
	p.expect(tokenSemicolon)

	return &ast.GlobalVariableDecl{Decl: &decl, Init: init}
}

//	attribute* 'var' _disambiguate_template template_list? optionally_typed_ident
//
// The attrs argument contains attributes already consumed by the caller.
// The 'var' keyword is consumed here.
//
// parseVariableDecl parses the core variable declaration form
func (p *parser) parseVariableDecl(attrs []ast.Attribute) ast.VariableDecl {
	p.expect(tokenVar)
	var templateArgs []ast.Expr
	if p.at(tokenLAngle) {
		templateArgs = p.parseTemplateList()
	}
	return ast.VariableDecl{Attrs: attrs, TemplateArgs: templateArgs, Ident: p.parseOptionallyTypedIdent()}
}

//	attribute* 'const'    optionally_typed_ident '=' expression
//	attribute* 'override' optionally_typed_ident ( '=' expression )?
//
// parseGlobalValueDecl parses a global value declaration
func (p *parser) parseGlobalValueDecl(attrs []ast.Attribute) *ast.GlobalValueDecl {
	kw := p.expectOneOf(tokenConst, tokenOverride)
	ident := p.parseOptionallyTypedIdent()

	var init ast.Expr

	switch kw.typ {
	case tokenConst:
		p.expect(tokenEqual)
		init = p.parseExpression()

	case tokenOverride:
		if p.accept(tokenEqual) {
			init = p.parseExpression()
		}
	}

	p.expect(tokenSemicolon)
	return &ast.GlobalValueDecl{Attrs: attrs, Keyword: kw.val, Ident: ident, Init: init}
}

// parseGlobalConstAssert parse a global const_assert statement
func (p *parser) parseGlobalConstAssert(attrs []ast.Attribute) *ast.ConstAssertDecl {
	stmt := p.parseConstAssertStatement(attrs)
	p.expect(tokenSemicolon)
	return &ast.ConstAssertDecl{Assert: stmt}
}

// parseFuncDecl parses a function declaration
func (p *parser) parseFuncDecl(attrs []ast.Attribute) *ast.FuncDecl {
	p.expect(tokenFunc)
	name := p.expect(tokenIdent)

	p.expect(tokenLParen)
	var params []ast.Param
	for !p.at(tokenRParen) {
		if p.at(tokenIfAttr) {
			params = append(params, p.parseIfAttrParam())
		} else {
			params = append(params, p.parseParam())
		}
		p.accept(tokenComma)
	}
	p.expect(tokenRParen)

	var retAttrs []ast.Attribute
	var retType *ast.TypeSpecifier
	if p.accept(tokenArrow) {
		retAttrs = p.parseAttributes()
		typ := p.parseTypeSpecifier()
		retType = &typ
	}

	body := p.parseCompoundStatement(nil)

	return &ast.FuncDecl{
		Attrs:       attrs,
		Name:        name.val,
		Params:      params,
		ReturnAttrs: retAttrs,
		ReturnType:  retType,
		Body:        body,
	}
}

func (p *parser) parseIfAttrParam() *ast.IfAttrParam {
	p.expect(tokenIfAttr)
	cond := p.parseExpression()
	then := p.parseParam()

	var els ast.Param
	if p.accept(tokenElseAttr) {
		els = p.parseParam()
	}

	return &ast.IfAttrParam{Cond: cond, Then: then, Else: els}
}

func (p *parser) parseParam() *ast.FuncParam {
	paramAttrs := p.parseAttributes()
	paramName := p.expect(tokenIdent)
	p.expect(tokenColon)
	paramType := p.parseTypeSpecifier()
	return &ast.FuncParam{
		Name:  paramName.val,
		Type:  paramType,
		Attrs: paramAttrs,
	}
}

// | import_statement* global_directive* global_decl*

// import_statement:
// | 'import' import_relative? (import_collection | import_path_or_item) ';'

// import_relative:
// | 'package' '::' | 'super' '::' ('super' '::')*

// import_path_or_item:
// | ident '::' (import_collection | import_path_or_item)
// | ident ('as' ident)?

// import_collection:
// | '{' (import_path_or_item) (',' (import_path_or_item))* ','? '}'
// parseImporDecl parses a WESL import declaration
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
		switch tok := p.next(); tok.typ {
		case tokenLBrace:
			decl.Items = p.parseImportItemList()
			p.expect(tokenRBrace)
			return decl

		case tokenPackage, tokenSuper:
			if !allowSpecial {
				p.unexpected(tok)
			}
			p.expect(tokenColonColon)
			decl.Path = append(decl.Path, tok.val)

		case tokenIdent:
			allowSpecial = false
			if p.accept(tokenColonColon) {
				decl.Path = append(decl.Path, tok.val)
			} else {
				item := ast.ImportItem{Path: []string{tok.val}}
				if p.accept(tokenAs) {
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
		p.accept(tokenComma)
		if p.at(tokenRBrace) {
			return items
		}
	}
}

// parseImportItem parses one item (or a nested brace group) from a brace list,
// prepending prefix to every resulting item's path. Returns one or more items
// because a nested "d::{ e, f }" expands inline.
func (p *parser) parseImportItem(prefix []string) []ast.ImportItem {
	tok := p.expect(tokenIdent)

	if p.accept(tokenColonColon) {
		if p.accept(tokenLBrace) {
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
		return p.parseImportItem(append(append([]string{}, prefix...), tok.val))
	}

	item := ast.ImportItem{Path: append(append([]string{}, prefix...), tok.val)}
	if p.accept(tokenAs) {
		item.Alias = p.expect(tokenIdent).val
	}
	return []ast.ImportItem{item}
}

// ----------------------------------------------------------------------------
// Statement

//	statement :
//	  ';'
//	| return_statement ';'
//	| if_statement
//	| switch_statement
//	| loop_statement
//	| for_statement
//	| while_statement
//	| func_call_statement ';'
//	| variable_or_value_statement ';'
//	| break_statement ';'
//	| continue_statement ';'
//	| discard_statement ';'
//	| variable_updating_statement ';'
//	| compound_statement
//	| const_assert_statement ';'
//
// Attributes consumed before this call are passed in via attrs.
//
// parseStatement parses a single statement.
func (p *parser) parseStatement() ast.Stmt {
	attrs := p.parseAttributes()
	return p.parseStatementBody(attrs)
}

// parseStatementBody dispatches on the next keyword after any leading
// attributes have already been consumed.
func (p *parser) parseStatementBody(attrs []ast.Attribute) ast.Stmt {
	switch tok := p.peek(); tok.typ {
	case tokenIfAttr:
		return p.parseIfAttrStmt()
	case tokenSemicolon:
		p.next()
		return &ast.EmptyStmt{}
	case tokenLBrace:
		return p.parseCompoundStatement(attrs)
	case tokenReturn:
		s := p.parseReturnStatement(attrs)
		p.expect(tokenSemicolon)
		return s
	case tokenBreak:
		s := p.parseBreakOrBreakIf(attrs)
		p.expect(tokenSemicolon)
		return s
	case tokenContinue:
		s := p.parseContinueStatement(attrs)
		p.expect(tokenSemicolon)
		return s
	case tokenContinuing:
		return p.parseContinuingStatement(attrs)
	case tokenDiscard:
		s := p.parseDiscardStatement(attrs)
		p.expect(tokenSemicolon)
		return s
	case tokenIf:
		return p.parseIfStatement(attrs)
	case tokenSwitch:
		return p.parseSwitchStatement(attrs)
	case tokenLoop:
		return p.parseLoopStatement(attrs)
	case tokenFor:
		return p.parseForStatement(attrs)
	case tokenWhile:
		return p.parseWhileStatement(attrs)
	case tokenVar, tokenLet, tokenConst:
		s := p.parseVarOrValueStatement(attrs)
		p.expect(tokenSemicolon)
		return s
	case tokenConstAssert:
		s := p.parseConstAssertStatement(attrs)
		p.expect(tokenSemicolon)
		return s
	case tokenUnderscore:
		s := p.parseBlankAssignment(attrs)
		p.expect(tokenSemicolon)
		return s
	case tokenIdent, tokenStar, tokenAmp, tokenLParen, tokenPackage, tokenSuper:
		s := p.parseExpressionStatement(attrs)
		p.expect(tokenSemicolon)
		return s
	default:
		p.unexpected(tok)
		return nil
	}
}

//	attribute* 'return' expression?
//
// parseReturnStatement parses a return statement
func (p *parser) parseReturnStatement(attrs []ast.Attribute) *ast.ReturnStmt {
	p.expect(tokenReturn)
	var value ast.Expr
	tok := p.peek()
	if tok.typ != tokenSemicolon && tok.typ != tokenRBrace {
		value = p.parseExpression()
	}
	return &ast.ReturnStmt{Attrs: attrs, Value: value}
}

//	break_statement    : attribute* 'break'
//	break_if_statement : attribute* 'break' 'if' expression ';'
//
// parseBreakOrBreakIf disambiguates between break_statement and break_if_statement
// (both start with 'break').
func (p *parser) parseBreakOrBreakIf(attrs []ast.Attribute) ast.Stmt {
	p.expect(tokenBreak)

	// 'break' followed by 'if' → break_if_statement.
	if p.accept(tokenIf) {
		cond := p.parseExpression()
		return &ast.BreakIfStmt{Attrs: attrs, Cond: cond}
	}

	return &ast.BreakStmt{Attrs: attrs}
}

//	attribute* 'continue'
//
// parseContinueStatement parses continue statement
func (p *parser) parseContinueStatement(attrs []ast.Attribute) *ast.ContinueStmt {
	p.expect(tokenContinue)
	return &ast.ContinueStmt{Attrs: attrs}
}

//	attribute* 'continuing' continuing_compound_statement
//
// parseContinuingStatement parses continuing statement
func (p *parser) parseContinuingStatement(attrs []ast.Attribute) *ast.ContinuingStmt {
	p.expect(tokenContinuing)
	body := p.parseCompoundStatement(nil)
	return &ast.ContinuingStmt{Attrs: attrs, Body: body}
}

//	attribute* 'discard'
//
// parseDiscardStatement parses discard statement
func (p *parser) parseDiscardStatement(attrs []ast.Attribute) *ast.DiscardStmt {
	p.expect(tokenDiscard)
	return &ast.DiscardStmt{Attrs: attrs}
}

//	attribute* 'const_assert' expression
//
// parseConstAssertStatement parses const_assert statement;
func (p *parser) parseConstAssertStatement(attrs []ast.Attribute) *ast.ConstAssertStmt {
	p.expect(tokenConstAssert)
	expr := p.parseExpression()
	return &ast.ConstAssertStmt{Attrs: attrs, Expr: expr}
}

//	attribute* '_' '=' expression
//
// parseBlankAssignment parses a blank assignement statement
func (p *parser) parseBlankAssignment(attrs []ast.Attribute) *ast.AssignmentStmt {
	p.expect(tokenUnderscore)
	p.expect(tokenEqual)
	rhs := p.parseExpression()
	return &ast.AssignmentStmt{Attrs: attrs, LHS: nil, Op: "=", RHS: rhs}
}

// parseExpressionStatement handles the ambiguous case where a statement begins
// with an expression (ident / '*' / '&' / '(').  It parses the expression,
// then inspects what follows to decide whether this is:
//
//   - A func_call_statement   (the expression must be a CallExpr)
//   - An assignment_statement ('=' or compound-assignment operator follows)
//   - An increment_statement  ('++' follows)
//   - A decrement_statement   ('--' follows)
func (p *parser) parseExpressionStatement(attrs []ast.Attribute) ast.Stmt {
	expr := p.parsePostfixExpr()

	switch p.peek().typ {
	case tokenPlusPlus:
		p.next()
		return &ast.IncrementStmt{Attrs: attrs, LHS: expr}

	case tokenMinusMinus:
		p.next()
		return &ast.DecrementStmt{Attrs: attrs, LHS: expr}

	case tokenEqual:
		p.next()
		return &ast.AssignmentStmt{Attrs: attrs, LHS: expr, Op: "=", RHS: p.parseExpression()}

	default:
		if op, ok := p.isCompoundAssignOp(); ok {
			p.next()
			return &ast.AssignmentStmt{Attrs: attrs, LHS: expr, Op: op, RHS: p.parseExpression()}
		}

		call, ok := expr.(*ast.CallExpr)
		if !ok {
			p.unexpected(p.peek())
		}
		return &ast.FnCallStmt{Attrs: attrs, Call: *call}
	}
}

func (p *parser) isCompoundAssignOp() (string, bool) {
	switch tok := p.peek(); tok.typ {
	case tokenPlusEq, tokenMinusEq, tokenStarEq, tokenSlashEq, tokenPercentEq,
		tokenAmpEq, tokenPipeEq, tokenCaretEq, tokenLtLtEq, tokenGtGtEq:
		return tok.val, true
	}
	return "", false
}

// parseVarOrValueStatement parses (without the trailing ';'):
//
//	variable_or_value_statement :
//	  variable_decl
//	| variable_decl '=' expression
//	| attribute* 'let'   optionally_typed_ident '=' expression
//	| attribute* 'const' optionally_typed_ident '=' expression
func (p *parser) parseVarOrValueStatement(attrs []ast.Attribute) *ast.VarOrValueStmt {
	switch tok := p.peek(); tok.typ {
	case tokenVar:
		decl := p.parseVariableDecl(attrs)

		var init ast.Expr
		if p.accept(tokenEqual) {
			init = p.parseExpression()
		}

		return &ast.VarOrValueStmt{Attrs: attrs, Keyword: "", Decl: &decl, Init: init}

	case tokenLet:
		p.next()
		ident := p.parseOptionallyTypedIdent()
		p.expect(tokenEqual)
		init := p.parseExpression()
		return &ast.VarOrValueStmt{
			Attrs:   attrs,
			Keyword: "let",
			Ident:   &ident,
			Init:    init,
		}

	case tokenConst:
		p.next()
		ident := p.parseOptionallyTypedIdent()
		p.expect(tokenEqual)
		init := p.parseExpression()
		return &ast.VarOrValueStmt{
			Attrs:   attrs,
			Keyword: "const",
			Ident:   &ident,
			Init:    init,
		}

	default:
		p.unexpected(tok)
		return nil
	}
}

// attribute* '{' statement* '}'
func (p *parser) parseCompoundStatement(attrs []ast.Attribute) *ast.CompoundStmt {
	p.expect(tokenLBrace)
	var stmts []ast.Stmt
	for !p.at(tokenRBrace) {
		stmts = append(stmts, p.parseStatement())
	}
	p.expect(tokenRBrace)
	return &ast.CompoundStmt{Attrs: attrs, Stmts: stmts}
}

// attribute* 'switch' expression switch_body_attributes? '{' switch_clause* '}'
//
// parseSwitchStatement parses a switch statement
func (p *parser) parseSwitchStatement(attrs []ast.Attribute) *ast.SwitchStmt {
	p.expect(tokenSwitch)
	expr := p.parseExpression()

	p.expect(tokenLBrace)
	var clauses []ast.SwitchClause
	for !p.at(tokenRBrace) {
		clauseAttrs := p.parseAttributes()

		switch p.peek().typ {
		case tokenIfAttr:
			clauses = append(clauses, p.parseIfAttrClause())
		default:
			clauses = append(clauses, p.parseSwitchClause(clauseAttrs))
		}
	}
	p.expect(tokenRBrace)
	return &ast.SwitchStmt{Attrs: attrs, Expr: expr, Clauses: clauses}
}

//	attribute* 'case' case_selectors ':'? compound_statement
//
// parseSwitchClause
func (p *parser) parseSwitchClause(attrs []ast.Attribute) ast.SwitchClause {
	switch tok := p.peek(); tok.typ {
	case tokenCase:
		return p.parseCaseClause(attrs)
	case tokenDefault:
		return p.parseDefaultClause(attrs)
	default:
		p.unexpected(tok)
	}
	panic("unreachable")
}

func (p *parser) parseCaseClause(attrs []ast.Attribute) *ast.CaseClause {
	p.expect(tokenCase)
	selectors := p.parseCaseSelectors()
	p.accept(tokenColon)
	body := p.parseCompoundStatement(nil)
	return &ast.CaseClause{Attrs: attrs, Selectors: selectors, Body: body}
}

//	attribute* 'default' ':'? compound_statement
//
// parseDefaultClause parses a default statement within a switch
func (p *parser) parseDefaultClause(attrs []ast.Attribute) *ast.CaseClause {
	p.expect(tokenDefault)
	p.accept(tokenColon)
	body := p.parseCompoundStatement(nil)
	return &ast.CaseClause{Attrs: attrs, Body: body}
}

func (p *parser) parseIfAttrClause() *ast.IfAttrClause {
	p.expect(tokenIfAttr)
	cond := p.parseExpression()
	attrs := p.parseAttributes()
	then := p.parseSwitchClause(attrs)

	var els ast.SwitchClause
	if p.at(tokenElseAttr) {
		p.expect(tokenElseAttr)
		attrs := p.parseAttributes()
		els = p.parseSwitchClause(attrs)
	}
	return &ast.IfAttrClause{Cond: cond, Then: then, Else: els}
}

// parseCaseSelectors parses the comma-separated list of case expressions:
//
//	case_selectors : case_selector ( ',' case_selector )* ','?
//
// In WGSL a case_selector is either 'default' or an expression.
func (p *parser) parseCaseSelectors() []ast.Expr {
	var selectors []ast.Expr
	for {
		if p.accept(tokenDefault) {
			selectors = append(selectors, &ast.Ident{Name: "default"})
		} else {
			selectors = append(selectors, p.parseExpression())
		}
		if !p.at(tokenComma) {
			break
		}
		p.accept(tokenComma)
		next := p.peek()
		if next.typ == tokenColon || next.typ == tokenLBrace {
			break
		}
	}
	return selectors
}

//	attribute* 'if' expression compound_statement
//	  ( 'else' ( if_statement | compound_statement ) )?
//
// parseIfStatement parses an "if" statement
func (p *parser) parseIfStatement(attrs []ast.Attribute) *ast.IfStmt {
	p.expect(tokenIf)
	cond := p.parseExpression()
	then := p.parseCompoundStatement(nil)
	stmt := &ast.IfStmt{Attrs: attrs, Cond: cond, Then: then}

	if p.accept(tokenElse) {
		if p.at(tokenIf) {
			innerAttrs := p.parseAttributes()
			stmt.ElseIf = p.parseIfStatement(innerAttrs)
		} else {
			stmt.Else = p.parseCompoundStatement(nil)
		}
	}
	return stmt
}

//	attribute* 'loop' attribute* '{' statement* ( continuing_statement )? '}'
//
// parseLoopStatement parses a "loop" statement
func (p *parser) parseLoopStatement(attrs []ast.Attribute) *ast.LoopStmt {
	p.expect(tokenLoop)
	bodyAttrs := p.parseAttributes()
	p.expect(tokenLBrace)

	var stmts []ast.Stmt

	for !p.at(tokenRBrace) {
		innerAttrs := p.parseAttributes()
		stmts = append(stmts, p.parseStatementBody(innerAttrs))
	}
	p.expect(tokenRBrace)

	return &ast.LoopStmt{Attrs: attrs, BodyAttrs: bodyAttrs, Body: &ast.CompoundStmt{Stmts: stmts}}
}

//	attribute* 'for' '(' for_init? ';' expression? ';' for_update? ')' compound_statement
//
// parseForStatement parses a "for" statement
func (p *parser) parseForStatement(attrs []ast.Attribute) *ast.ForStmt {
	p.expect(tokenFor)
	p.expect(tokenLParen)

	// for_init: variable_or_value_statement | variable_updating_statement | func_call_statement
	var init ast.Stmt
	if !p.at(tokenSemicolon) {
		initAttrs := p.parseAttributes()
		switch p.peek().typ {
		case tokenVar, tokenLet, tokenConst:
			init = p.parseVarOrValueStatement(initAttrs)
		default:
			init = p.parseExpressionStatement(initAttrs)
		}
	}
	p.expect(tokenSemicolon)

	// for_condition
	var cond ast.Expr
	if !p.at(tokenSemicolon) {
		cond = p.parseExpression()
	}
	p.expect(tokenSemicolon)

	// for_update: variable_updating_statement | func_call_statement
	var update ast.Stmt
	if !p.at(tokenRParen) {
		updateAttrs := p.parseAttributes()
		update = p.parseExpressionStatement(updateAttrs)
	}
	p.expect(tokenRParen)

	body := p.parseCompoundStatement(nil)
	return &ast.ForStmt{Attrs: attrs, Init: init, Cond: cond, Update: update, Body: body}
}

//	attribute* 'while' expression compound_statement
//
// parseWhileStatement parses a "while" statement
func (p *parser) parseWhileStatement(attrs []ast.Attribute) *ast.WhileStmt {
	p.expect(tokenWhile)
	cond := p.parseExpression()
	body := p.parseCompoundStatement(nil)
	return &ast.WhileStmt{Attrs: attrs, Cond: cond, Body: body}
}

// ----------------------------------------------------------------------------
// Expression

// parseExpression parses a full expression.
func (p *parser) parseExpression() ast.Expr {
	return p.parseExprPrec(0)
}

// parseExprPrec is the Pratt precedence-climbing core.
// It consumes the longest expression whose outermost binary operator has
// precedence strictly greater than minPrec.
//
// Left-associativity: when recursing for the RHS we pass the same prec as
// minPrec so that a same-precedence operator on the right breaks the loop
// (the loop condition is  prec <= minPrec, i.e. equal prec stops here).
func (p *parser) parseExprPrec(minPrec int) ast.Expr {
	left := p.parseUnaryExpr()

	for {
		op := p.peek()
		prec := p.infixPrec(op)
		if prec <= minPrec {
			break
		}
		p.next()
		right := p.parseExprPrec(prec)
		left = &ast.BinaryExpr{Op: op.val, Left: left, Right: right}
	}

	return left
}

//	unary_expression :
//	  '!' unary_expression
//	| '~' unary_expression
//	| '-' unary_expression
//	| '*' unary_expression      -- pointer dereference
//	| '&' unary_expression      -- address-of
//	| singular_expression       -- postfix / primary
//
// parseUnaryExpr parses a unary prefix expression.
func (p *parser) parseUnaryExpr() ast.Expr {
	switch tok := p.peek(); tok.typ {
	case tokenBang, tokenTilde, tokenMinus:
		p.next()
		return &ast.UnaryExpr{Op: tok.val, Operand: p.parseUnaryExpr()}
	case tokenStar:
		p.next()
		return &ast.DerefExpr{Operand: p.parseUnaryExpr()}
	case tokenAmp:
		p.next()
		return &ast.AddrOfExpr{Operand: p.parseUnaryExpr()}
	default:
		return p.parsePostfixExpr()
	}
}

//	singular_expression :
//	  primary_expression component_or_swizzle_specifier*
//
//	component_or_swizzle_specifier :
//	'[' expression ']' | '.' ident
//
// parsePostfixExpr applies zero or more postfix operations (indexing and
// member access) to a primary expression.
func (p *parser) parsePostfixExpr() ast.Expr {
	expr := p.parsePrimaryExpr()

	for {
		switch p.peek().typ {
		case tokenLBracket:
			p.expect(tokenLBracket)
			idx := p.parseExpression()
			p.expect(tokenRBracket)
			expr = &ast.IndexExpr{Base: expr, Index: idx}

		case tokenDot:
			p.next()
			member := p.expect(tokenIdent)
			expr = &ast.MemberExpr{Base: expr, Member: member.val}

		default:
			return expr
		}
	}
}

//	primary_expression :
//	int_literal | float_literal | 'true' | 'false'
//	| '(' expression ')'
//	| ident
//	| ident '(' argument_expression_list? ')'                       -- plain call
//	| ident '<' template_list '>' '(' argument_expression_list? ')' -- templated call
//
// parsePrimaryExpr parses literals, parenthesised expressions, identifiers,
// and call expressions (including template-parameterised calls).
func (p *parser) parsePrimaryExpr() ast.Expr {
	switch tok := p.next(); tok.typ {
	case tokenNumber, tokenTrue, tokenFalse:
		return &ast.LitExpr{Val: tok.val}
	case tokenLParen:
		inner := p.parseExpression()
		p.expect(tokenRParen)
		return &ast.ParenExpr{Inner: inner}
	case tokenPackage, tokenSuper:
		ident := tok.val
		p.expect(tokenColonColon)
		for p.at(tokenIdent) {
			tok := p.next()
			ident += "::" + tok.val
			p.accept(tokenColonColon)
		}
		if p.at(tokenLParen) {
			return &ast.CallExpr{Callee: ident, Args: p.parseArgumentExpressionList()}
		}
		return &ast.Ident{Name: ident}

	case tokenIdent:
		ident := tok.val
		for p.at(tokenColonColon) {
			p.next() // consume ::
			seg := p.expect(tokenIdent)
			ident += "::" + seg.val
		}
		if ident != tok.val {
			if p.at(tokenLParen) {
				return &ast.CallExpr{Callee: ident, Args: p.parseArgumentExpressionList()}
			}
			return &ast.Ident{Name: ident}
		}

		if isTemplateableIdent(tok.val) && p.at(tokenLAngle) {
			targs := p.parseTemplateList()
			if p.at(tokenLParen) {
				return &ast.CallExpr{Callee: tok.val, TemplateArgs: targs, Args: p.parseArgumentExpressionList()}
			}
			return &ast.CallExpr{Callee: tok.val, TemplateArgs: targs}
		}
		if p.at(tokenLParen) {
			return &ast.CallExpr{Callee: tok.val, Args: p.parseArgumentExpressionList()}
		}
		return &ast.Ident{Name: tok.val}

	default:
		p.unexpected(tok)
		return nil
	}
}

func (p *parser) parseTemplateList() []ast.Expr {
	p.expect(tokenLAngle)
	p.templateDepth++
	defer func() { p.templateDepth-- }()

	closeTemplate := func() bool {
		if p.pendingClose > 0 {
			p.pendingClose--
			return true
		}
		switch p.peek().typ {
		case tokenRAngle:
			p.next()
			return true
		case tokenGtGt:
			// Consume '>>' but leave one '>' for the enclosing template.
			p.next()
			p.pendingClose++
			return true
		}
		return false
	}

	var args []ast.Expr
	for {
		if closeTemplate() {
			return args
		}
		if len(args) > 0 {
			p.expect(tokenComma)
			if closeTemplate() { // trailing comma
				return args
			}
		}
		args = append(args, p.parseTemplateArg())
	}
}

// parseTemplateArg parses one argument inside a template list.
//
// Template args are either type specifiers (vec3<f32>, ptr<storage,T>) or
// const-expressions (integer literals, named constants, arithmetic on them).
// Identifiers are always parsed as type specifiers because:
//   - A named constant like 'N' becomes TypeSpecifier{Name:"N"}.asExpr() = Ident
//   - A nested type like 'vec3<f32>' is consumed greedily by parseTypeSpecifier,
//     which treats '<' as a template opener unconditionally — avoiding the need
//     for templateDepth to suppress '<' for that nested parse.
func (p *parser) parseTemplateArg() ast.Expr {
	if p.peek().typ == tokenIdent {
		return p.parseTypeSpecifier().AsExpr()
	}
	return p.parseExpression()
}

//	type_specifier : ident template_list?
//
// parseTypeSpecifier parses a type reference:
func (p *parser) parseTypeSpecifier() ast.TypeSpecifier {
	tok := p.expect(tokenIdent)

	var args []ast.Expr
	if p.at(tokenLAngle) { // type context: '<' is always a template opener
		args = p.parseTemplateList()
	}

	return ast.TypeSpecifier{Name: tok.val, TemplateArgs: args}
}

//	optionally_typed_ident : ident ( ':' type_specifier )?
//
// parseOptionallyTypedIdent parses type specifiers that are optional
// FIXME: remove me and use ast.TypeSpecifier instead
func (p *parser) parseOptionallyTypedIdent() ast.OptionallyTypedIdent {
	tok := p.expect(tokenIdent)

	var typ *ast.TypeSpecifier
	if p.accept(tokenColon) {
		ts := p.parseTypeSpecifier()
		typ = &ts
	}

	return ast.OptionallyTypedIdent{Name: tok.val, Type: typ}
}

func (p *parser) parseArgumentExpressionList() []ast.Expr {
	p.expect(tokenLParen)
	var args []ast.Expr
	for !p.at(tokenRParen) {
		if len(args) > 0 {
			p.expect(tokenComma)
			if p.at(tokenRParen) {
				break
			}
		}
		args = append(args, p.parseExpression())
	}
	p.expect(tokenRParen)
	return args
}

// WGSL forbids mixing bitwise (|, ^, &) with logical (||, &&) operators in
// the same expression without explicit parentheses.  That constraint is
// enforced by a later semantic pass; here we assign them the numeric
// precedences from the WGSL specification so that the Pratt parser produces
// the correct tree when they do appear in valid isolation.
//
//   1  ||
//   2  &&
//   3  |
//   4  ^
//   5  &
//   6  == !=
//   7  < > <= >=
//   8  << >>
//   9  + -
//  10  * / %

func (p *parser) infixPrec(tok token) int {
	if p.templateDepth > 0 {
		switch tok.typ {
		case tokenLAngle, tokenRAngle, tokenGtGt, tokenLtLt:
			return 0
		}
	}
	switch tok.typ {
	case tokenPipePipe:
		return 1
	case tokenAmpAmp:
		return 2
	case tokenPipe:
		return 3
	case tokenCaret:
		return 4
	case tokenAmp:
		return 5
	case tokenEqualEqual, tokenBangEqual:
		return 6
	case tokenLAngle, tokenRAngle, tokenLtEqual, tokenGtEqual:
		return 7
	case tokenLtLt, tokenGtGt:
		return 8
	case tokenPlus, tokenMinus:
		return 9
	case tokenStar, tokenSlash, tokenPercent:
		return 10
	}
	return 0
}

func isTemplateableIdent(name string) bool {
	switch name {
	case "array", "atomic", "bool",
		"f16", "f32", "i32", "u32",
		"mat2x2", "mat2x3", "mat2x4",
		"mat3x2", "mat3x3", "mat3x4",
		"mat4x2", "mat4x3", "mat4x4",
		"ptr",
		"sampler", "sampler_comparison",
		"texture_1d", "texture_2d", "texture_2d_array",
		"texture_3d", "texture_cube", "texture_cube_array",
		"texture_depth_2d", "texture_depth_2d_array",
		"texture_depth_cube", "texture_depth_cube_array",
		"texture_depth_multisampled_2d",
		"texture_multisampled_2d",
		"texture_storage_1d", "texture_storage_2d",
		"texture_storage_2d_array", "texture_storage_3d",
		"vec2", "vec3", "vec4",
		"binding_array": // WESL extension
		return true
	}
	return false
}
