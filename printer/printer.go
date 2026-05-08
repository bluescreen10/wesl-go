package printer

import (
	"io"
	"strings"

	"github.com/bluescreen10/wesl-go/ast"
)

const (
	WHITESPACE = ' '

	LANGLE   = '<'
	RANGLE   = '>'
	LPAREN   = '('
	RPAREN   = ')'
	LBRACKET = '['
	RBRACKET = ']'
	LBRACE   = '{'
	RBRACE   = '}'

	DOT       = '.'
	COMMA     = ','
	COLON     = ':'
	SEMICOLON = ';'

	STAR       = '*'
	AMP        = '&'
	UNDERSCORE = '_'
	EQUAL      = '='

	ARROW = "->"

	// Keywords
	DIAGNOSTIC   = "diagnostic"
	DISCARD      = "discard"
	ENABLE       = "enable"
	FUNC         = "fn"
	WHILE        = "while"
	LOOP         = "loop"
	FOR          = "for"
	IF           = "if"
	ELSE         = "else"
	CONTINUE     = "continue"
	CONTINUING   = "continuing"
	BREAK        = "break"
	CONST_ASSERT = "const_assert"
	RETURN       = "return"
	SWITCH       = "switch"
	CASE         = "case"
	DEFAULT      = "default"
	STRUCT       = "struct"
	ALIAS        = "alias"
	VAR          = "var"
	REQUIRES     = "requires"

	// WESL
	IF_ATTR   = "@if"
	ELSE_ATTR = "@else"
	IMPORT    = "import"
	AS        = "as"
	DCOLON    = "::"
)

// printer serializes AST nodes to a Writer. It is an internal type; use Fprint
// to access printing functionality.
type printer struct {
	writer io.Writer // destination for all output written by the printer
}

// Fprint sets the destination writer and prints node n to it. Calling Fprint
// multiple times on the same printer reuses the instance with the new writer.
func (p *printer) Fprint(w io.Writer, n ast.Node) {
	p.writer = w
	p.printNode(n)
}

// printNode dispatches to the appropriate typed print method based on the
// runtime type of n.
func (p *printer) printNode(n ast.Node) {
	switch n := n.(type) {
	case *ast.File:
		p.printFile(n)
	case ast.Decl:
		p.printDecl(n)
	case ast.Stmt:
		p.printStmt(n)
	case ast.Expr:
		p.printExpr(n)
	}
}

// writeBytes writes one or more raw bytes directly to the writer.
func (p *printer) writeBytes(chars ...byte) {
	p.writer.Write(chars)
}

// writeString writes s as a byte slice to the writer.
func (p *printer) writeString(s string) {
	p.writer.Write([]byte(s))
}

// printAttr prints a single attribute in the form @name or @name(args...).
func (p *printer) printAttr(attr *ast.Attribute) {
	p.writeString(attr.Name)
	if len(attr.Args) > 0 {
		p.writeBytes(LPAREN)
		for i, a := range attr.Args {
			if i > 0 {
				p.writeBytes(COMMA, WHITESPACE)
			}
			p.printExpr(a)
		}
		p.writeBytes(RPAREN)
	}
}

// printAttrs prints each attribute in attrs followed by a space separator.
func (p *printer) printAttrs(attrs []*ast.Attribute) {
	for _, a := range attrs {
		p.printAttr(a)
		p.writeBytes(WHITESPACE)
	}
}

// printDecl prints a top-level declaration. Statements that are also valid as
// declarations (VarStmt, ValStmt, ConstAssertStmt) are printed as statements
// with a trailing semicolon.
func (p *printer) printDecl(d ast.Decl) {
	switch d := d.(type) {
	case *ast.ConstAssertStmt, *ast.VarStmt, *ast.ValStmt:
		p.printStmt(d.(ast.Stmt))
		p.writeBytes(SEMICOLON)
	case *ast.DiagnosticDirective:
		p.printAttrs(d.Attrs)
		p.writeString(DIAGNOSTIC)
		p.writeBytes(LPAREN)
		p.writeString(d.Control.Severity)
		p.writeBytes(COMMA, WHITESPACE)
		p.writeString(d.Control.RuleName)
		p.writeBytes(RPAREN, SEMICOLON)
	case *ast.EnableDirective:
		p.printAttrs(d.Attrs)
		p.writeString(ENABLE)
		p.writeBytes(WHITESPACE)
		for i, e := range d.Extensions {
			if i > 0 {
				p.writeBytes(COMMA, WHITESPACE)
			}
			p.writeString(e)
		}
		p.writeBytes(SEMICOLON)
	case *ast.FuncDecl:
		p.printFuncDecl(d)
	case *ast.IfAttrDecl:
		p.writeString(IF_ATTR)
		p.printExpr(d.Cond)
		p.writeBytes(WHITESPACE)
		p.printDecl(d.Then)
		if d.Else != nil {
			p.writeString(ELSE_ATTR)
			p.writeBytes(WHITESPACE)
			p.printDecl(d.Else)
		}
	case *ast.ImportDecl:
		p.printImportDecl(d)
	case *ast.RequiresDirective:
		p.printAttrs(d.Attrs)
		p.writeString(REQUIRES)
		p.writeBytes(WHITESPACE)
		for i, e := range d.Extensions {
			if i > 0 {
				p.writeBytes(COMMA, WHITESPACE)
			}
			p.writeString(e)
		}
		p.writeBytes(SEMICOLON)
	case *ast.StructDecl:
		p.printAttrs(d.Attrs)
		p.writeString(STRUCT)
		p.writeBytes(WHITESPACE)
		p.printIdent(d.Name)
		p.writeBytes(WHITESPACE, LBRACE, WHITESPACE)
		for i, m := range d.Members {
			if i > 0 {
				p.writeBytes(COMMA, WHITESPACE)
			}
			p.printMember(m)
		}
		p.writeBytes(WHITESPACE, RBRACE)
	case *ast.TypeAliasDecl:
		p.printAttrs(d.Attrs)
		p.writeString(ALIAS)
		p.writeBytes(WHITESPACE)
		p.printIdent(d.Name)
		p.writeBytes(WHITESPACE, EQUAL, WHITESPACE)
		p.printTypeSpecifier(d.Type)
		p.writeBytes(SEMICOLON)
	}
}

// printImportDecl prints each imported item in d as a separate import
// statement, including its alias when one is present.
func (p *printer) printImportDecl(d *ast.ImportDecl) {
	for _, imp := range d.Imports {
		p.writeString(IMPORT)
		p.writeBytes(WHITESPACE)
		p.writeString(strings.Join(imp.Path, DCOLON))
		if imp.Alias != "" {
			p.writeBytes(WHITESPACE)
			p.writeString(AS)
			p.writeBytes(WHITESPACE)
			p.writeString(imp.Alias)
		}
		p.writeBytes(SEMICOLON)
	}
}

// printMember prints a struct member, handling both plain StructMember fields
// and conditional @if IfAttrStructMember nodes.
func (p *printer) printMember(m ast.Member) {
	switch m := m.(type) {
	case *ast.IfAttrStructMember:
		p.writeString(IF_ATTR)
		p.printExpr(m.Cond)
		p.writeBytes(WHITESPACE)
		p.printMember(m.Then)
		if m.Else != nil {
			p.writeString(ELSE_ATTR)
			p.writeBytes(WHITESPACE)
			p.printMember(m.Else)
		}
	case *ast.StructMember:
		p.printAttrs(m.Attrs)
		p.writeString(m.Name)
		p.writeBytes(COLON, WHITESPACE)
		p.printTypeSpecifier(m.Type)
	}
}

// printExpr prints an expression node, dispatching on its concrete type.
func (p *printer) printExpr(e ast.Expr) {
	switch e := e.(type) {
	case *ast.AddrOfExpr:
		p.writeBytes(AMP)
		p.printExpr(e.Operand)
	case *ast.BinaryExpr:
		p.printExpr(e.Left)
		p.writeBytes(WHITESPACE)
		p.writeString(e.Op)
		p.writeBytes(WHITESPACE)
		p.printExpr(e.Right)
	case *ast.CallExpr:
		p.printCallExpr(e)
	case *ast.DerefExpr:
		p.writeBytes(STAR)
		p.printExpr(e.Operand)
	case *ast.Ident:
		p.printIdent(e)
	case *ast.IndexExpr:
		p.printExpr(e.Base)
		p.writeBytes(LBRACKET)
		p.printExpr(e.Index)
		p.writeBytes(RBRACKET)
	case *ast.LitExpr:
		p.writeString(e.Val)
	case *ast.MemberExpr:
		p.printExpr(e.Base)
		p.writeBytes(DOT)
		p.writeString(e.Member)
	case *ast.ParenExpr:
		p.writeBytes(LPAREN)
		p.printExpr(e.Inner)
		p.writeBytes(RPAREN)
	case *ast.UnaryExpr:
		p.writeString(e.Op)
		p.writeBytes(WHITESPACE)
		p.printExpr(e.Operand)
	}
}

// printStmt prints a statement node, dispatching on its concrete type.
// Block-structured statements (if, for, while, etc.) do not print a trailing
// semicolon; simple statements do.
func (p *printer) printStmt(s ast.Stmt) {
	switch s := s.(type) {
	case *ast.AssignmentStmt:
		p.printAttrs(s.Attrs)
		if s.LHS != nil {
			p.printExpr(s.LHS)
		} else {
			p.writeBytes(UNDERSCORE)
		}
		p.writeBytes(WHITESPACE)
		p.writeString(s.Op)
		p.writeBytes(WHITESPACE)
		p.printExpr(s.RHS)
	case *ast.BreakStmt:
		p.printAttrs(s.Attrs)
		p.writeString(BREAK)
		if s.Cond != nil {
			p.writeBytes(WHITESPACE)
			p.writeString(IF)
			p.writeBytes(WHITESPACE)
			p.printExpr(s.Cond)
		}
	case *ast.ConstAssertStmt:
		p.printAttrs(s.Attrs)
		p.writeString(CONST_ASSERT)
		p.writeBytes(WHITESPACE)
		p.printExpr(s.Expr)
	case *ast.ContinueStmt:
		p.printAttrs(s.Attrs)
		p.writeString(CONTINUE)
	case *ast.ContinuingStmt:
		p.printAttrs(s.Attrs)
		p.writeString(CONTINUING)
		p.writeBytes(WHITESPACE)
		p.printBlockStmt(s.Body)
	case *ast.BlockStmt:
		p.printBlockStmt(s)
	case *ast.DiscardStmt:
		p.printAttrs(s.Attrs)
		p.writeString(DISCARD)
	case *ast.EmptyStmt:
		// nothing to do
	case *ast.IfAttrStmt:
		p.writeString(IF_ATTR)
		p.printExpr(s.Cond)
		p.writeBytes(WHITESPACE)
		p.printStmt(s.Then)
		if s.Else != nil {
			p.writeString(ELSE_ATTR)
			p.writeBytes(WHITESPACE)
			p.printStmt(s.Else)
		}
	case *ast.IfStmt:
		p.printIfStmt(s)
	case *ast.FuncCallStmt:
		p.printAttrs(s.Attrs)
		p.printCallExpr(s.Call)
	case *ast.ForStmt:
		p.printAttrs(s.Attrs)
		p.writeString(FOR)
		p.writeBytes(WHITESPACE, LPAREN)
		p.printStmt(s.Init)
		p.writeBytes(SEMICOLON, WHITESPACE)
		p.printExpr(s.Cond)
		p.writeBytes(SEMICOLON, WHITESPACE)
		p.printStmt(s.Update)
		p.writeBytes(RPAREN, WHITESPACE)
		p.printBlockStmt(s.Body)
	case *ast.IncDecStmt:
		p.printAttrs(s.Attrs)
		p.printExpr(s.LHS)
		p.writeString(s.Op)
	case *ast.LoopStmt:
		p.printAttrs(s.Attrs)
		p.writeString(LOOP)
		p.writeBytes(WHITESPACE)
		p.printBlockStmt(s.Body)
	case *ast.ReturnStmt:
		p.printAttrs(s.Attrs)
		p.writeString(RETURN)
		if s.Value != nil {
			p.writeBytes(WHITESPACE)
			p.printExpr(s.Value)
		}
	case *ast.SwitchStmt:
		p.printAttrs(s.Attrs)
		p.writeString(SWITCH)
		p.writeBytes(WHITESPACE)
		p.printExpr(s.Expr)
		p.writeBytes(WHITESPACE, LBRACE, WHITESPACE)
		p.printClauses(s.Clauses)
		p.writeBytes(WHITESPACE, RBRACE)
	case *ast.VarStmt:
		p.printAttrs(s.Attrs)
		p.writeString(VAR)
		p.printTemplateArgs(s.TemplateArgs)
		p.writeBytes(WHITESPACE)
		p.printIdent(s.Name)
		if s.Type != nil {
			p.writeBytes(COLON, WHITESPACE)
			p.printTypeSpecifier(s.Type)
		}
		if s.Init != nil {
			p.writeBytes(WHITESPACE, EQUAL, WHITESPACE)
			p.printExpr(s.Init)
		}
	case *ast.ValStmt:
		p.printAttrs(s.Attrs)
		p.writeString(s.Keyword)
		p.writeBytes(WHITESPACE)
		p.printIdent(s.Name)
		if s.Type != nil {
			p.writeBytes(COLON, WHITESPACE)
			p.printTypeSpecifier(s.Type)
		}
		if s.Init != nil {
			p.writeBytes(WHITESPACE, EQUAL, WHITESPACE)
			p.printExpr(s.Init)
		}
	case *ast.WhileStmt:
		p.printAttrs(s.Attrs)
		p.writeString(WHILE)
		p.writeBytes(WHITESPACE)
		if s.Cond != nil {
			p.printExpr(s.Cond)
			p.writeBytes(WHITESPACE)
		}
		p.printBlockStmt(s.Body)
	}
}

// printIdent prints an identifier, prefixing qualified names with their
// path segments separated by "::".
func (p *printer) printIdent(i *ast.Ident) {
	if len(i.Path) > 0 {
		p.writeString(strings.Join(i.Path, DCOLON))
		p.writeString(DCOLON)
	}
	p.writeString(i.Val)
}

// printClauses prints all switch clauses in the order they appear.
func (p *printer) printClauses(clauses []ast.Clause) {
	for _, c := range clauses {
		p.printClause(c)
	}
}

// printClause prints a single switch clause. A CaseClause with nil Selectors
// is printed as the default clause.
func (p *printer) printClause(c ast.Clause) {
	switch c := c.(type) {
	case *ast.CaseClause:
		p.printAttrs(c.Attrs)
		if c.Selectors == nil {
			p.writeString(DEFAULT)
			p.writeBytes(WHITESPACE)
		} else {
			p.writeString(CASE)
			p.writeBytes(WHITESPACE)
			for i, s := range c.Selectors {
				if i > 0 {
					p.writeBytes(COMMA, WHITESPACE)
				}
				p.printExpr(s)
			}
		}
		p.printBlockStmt(c.Body)
	case *ast.IfAttrClause:
		p.writeString(IF_ATTR)
		p.printExpr(c.Cond)
		p.writeBytes(WHITESPACE)
		p.printClause(c.Then)
		if c.Else != nil {
			p.writeString(ELSE_ATTR)
			p.writeBytes(WHITESPACE)
			p.printClause(c.Else)
		}
	}
}

// printCallExpr prints a function call expression, including optional template
// arguments and the parenthesized argument list.
func (p *printer) printCallExpr(e *ast.CallExpr) {
	p.printIdent(e.Callee)
	p.printTemplateArgs(e.TemplateArgs)
	p.writeBytes(LPAREN)
	p.printExprList(e.Args)
	p.writeBytes(RPAREN)
}

// printIfStmt prints an if statement including any chained else-if and else
// branches.
func (p *printer) printIfStmt(s *ast.IfStmt) {
	p.printAttrs(s.Attrs)
	p.writeString(IF)
	p.writeBytes(WHITESPACE)
	p.printExpr(s.Cond)
	p.writeBytes(WHITESPACE)
	p.printBlockStmt(s.Then)
	if s.ElseIf != nil {
		p.writeString(ELSE)
		p.writeBytes(WHITESPACE)
		p.printIfStmt(s.ElseIf)
	}
	if s.Else != nil {
		p.writeString(ELSE)
		p.writeBytes(WHITESPACE)
		p.printBlockStmt(s.Else)
	}
}

// printExprList prints a comma-separated list of expressions.
func (p *printer) printExprList(exprs []ast.Expr) {
	for i, e := range exprs {
		if i > 0 {
			p.writeBytes(COMMA, WHITESPACE)
		}
		p.printExpr(e)
	}
}

// printFile prints all top-level declarations in n, separated by spaces.
func (p *printer) printFile(n *ast.File) {
	for i, d := range n.Decls {
		if i > 0 {
			p.writeBytes(WHITESPACE)
		}
		p.printDecl(d)
	}
}

// printFuncDecl prints a function declaration including its attributes, name,
// parameter list, optional return type, and body.
func (p *printer) printFuncDecl(f *ast.FuncDecl) {
	p.printAttrs(f.Attrs)
	p.writeString(FUNC)
	p.writeBytes(WHITESPACE)
	p.printIdent(f.Name)
	p.printParamList(f.Params)
	p.writeBytes(WHITESPACE)
	if f.ReturnType != nil {
		p.writeString(ARROW)
		p.writeBytes(WHITESPACE)
		p.printAttrs(f.ReturnAttrs)
		p.printTypeSpecifier(f.ReturnType)
		p.writeBytes(WHITESPACE)
	}
	p.printBlockStmt(f.Body)
}

// printBlockStmt prints a brace-enclosed block, inserting semicolons between
// simple statements and omitting them after block-structured statements.
func (p *printer) printBlockStmt(s *ast.BlockStmt) {
	p.printAttrs(s.Attrs)
	p.writeBytes(LBRACE, WHITESPACE)

	var isBlockStmt bool
	for i, s := range s.Stmts {
		if i > 0 {
			if !isBlockStmt {
				p.writeBytes(SEMICOLON)
			}
			p.writeBytes(WHITESPACE)
		}

		switch s.(type) {
		case *ast.BlockStmt, *ast.IfStmt, *ast.WhileStmt, *ast.ForStmt, *ast.LoopStmt, *ast.SwitchStmt, *ast.ContinuingStmt:
			isBlockStmt = true
		default:
			isBlockStmt = false
		}
		p.printStmt(s)
	}

	//FIXME: Hack
	if len(s.Stmts) > 0 {
		if !isBlockStmt {
			p.writeBytes(SEMICOLON)
		}
		p.writeBytes(WHITESPACE)
	}
	p.writeBytes(RBRACE)
}

// printParamList prints the parenthesized, comma-separated parameter list for
// a function declaration.
func (p *printer) printParamList(params []ast.Param) {
	p.writeBytes(LPAREN)
	for i, param := range params {
		if i > 0 {
			p.writeBytes(COMMA, WHITESPACE)
		}
		p.printParam(param)
	}
	p.writeBytes(RPAREN)
}

// printParam prints a single function parameter, handling both plain FuncParam
// nodes and conditional @if IfAttrParam nodes.
func (p *printer) printParam(param ast.Param) {
	switch param := param.(type) {
	case *ast.IfAttrParam:
		p.writeString(IF_ATTR)
		p.printExpr(param.Cond)
		p.writeBytes(WHITESPACE)
		p.printParam(param.Then)
		if param.Else != nil {
			p.writeString(ELSE_ATTR)
			p.writeBytes(WHITESPACE)
			p.printParam(param.Else)
		}
	case *ast.FuncParam:
		p.printAttrs(param.Attrs)
		p.writeString(param.Name)
		p.writeBytes(COLON, WHITESPACE)
		p.printTypeSpecifier(param.Type)
	}
}

// printTypeSpecifier prints a type reference followed by any template
// arguments enclosed in angle brackets.
func (p *printer) printTypeSpecifier(t *ast.TypeSpecifier) {
	p.printIdent(t.Name)
	p.printTemplateArgs(t.TemplateArgs)
}

// printTemplateArgs prints angle-bracket enclosed template arguments when
// args is non-empty; otherwise it is a no-op.
func (p *printer) printTemplateArgs(args []ast.Expr) {
	if len(args) > 0 {
		p.writeBytes(LANGLE)
		p.printExprList(args)
		p.writeBytes(RANGLE)
	}
}

// Fprint serializes the AST node n as WESL/WGSL source text and writes the
// result to w.
func Fprint(w io.Writer, n ast.Node) {
	(&printer{}).Fprint(w, n)
}
