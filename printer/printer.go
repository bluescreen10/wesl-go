package printer

import (
	"io"

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

type printer struct {
	writer io.Writer
}

func (p *printer) Fprint(w io.Writer, n ast.Node) {
	p.writer = w
	p.printNode(n)
}

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

func (p *printer) writeBytes(chars ...byte) {
	p.writer.Write(chars)
}

func (p *printer) writeString(s string) {
	p.writer.Write([]byte(s))
}

func (p *printer) printAttr(attr ast.Attribute) {
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

func (p *printer) printAttrs(attrs []ast.Attribute) {
	for _, a := range attrs {
		p.printAttr(a)
		p.writeBytes(WHITESPACE)
	}
}

func (p *printer) printDecl(d ast.Decl) {
	switch d := d.(type) {
	case *ast.ConstAssertDecl:
		p.printStmt(d.Assert)
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
	case *ast.GlobalValueDecl:
		p.printAttrs(d.Attrs)
		p.writeString(d.Keyword)
		p.writeBytes(WHITESPACE)
		p.writeString(d.Name)
		if d.Type != nil {
			p.writeBytes(COLON, WHITESPACE)
			p.printTypeSpecifier(*d.Type)
		}
		if d.Init != nil {
			p.writeBytes(WHITESPACE, EQUAL, WHITESPACE)
			p.printExpr(d.Init)
		}
		p.writeBytes(SEMICOLON)
	case *ast.GlobalVariableDecl:
		p.printDecl(d.Decl)
		if d.Init != nil {
			p.writeBytes(WHITESPACE, EQUAL, WHITESPACE)
			p.printExpr(d.Init)
		}
		p.writeBytes(SEMICOLON)
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
		p.writeString(d.Name)
		p.writeBytes(WHITESPACE, LBRACE, WHITESPACE)
		for i, m := range d.Members {
			if i > 0 {
				p.writeBytes(COMMA, WHITESPACE)
			}
			p.printStructMember(m)
		}
		p.writeBytes(WHITESPACE, RBRACE)
	case *ast.TypeAliasDecl:
		p.printAttrs(d.Attrs)
		p.writeString(ALIAS)
		p.writeBytes(WHITESPACE)
		p.writeString(d.Name)
		p.writeBytes(WHITESPACE, EQUAL, WHITESPACE)
		p.printTypeSpecifier(d.Type)
		p.writeBytes(SEMICOLON)
	case *ast.VariableDecl:
		p.printAttrs(d.Attrs)
		p.writeString(VAR)
		p.printTemplateArgs(d.TemplateArgs)
		p.writeBytes(WHITESPACE)
		p.writeString(d.Name)
		if d.Type != nil {
			p.writeBytes(COLON, WHITESPACE)
			p.printTypeSpecifier(*d.Type)
		}
	}
}

func (p *printer) printImportDecl(d *ast.ImportDecl) {
	p.writeString(IMPORT)
	p.writeBytes(WHITESPACE)
	for _, seg := range d.Path {
		p.writeString(seg)
	}
	if len(d.Items) == 1 && len(d.Items[0].Path) == 1 && d.Items[0].Alias == "" {
		p.writeString(d.Items[0].Path[0])
	} else {
		p.writeBytes(LBRACE)
		for i, item := range d.Items {
			p.writeString(DCOLON)
			if i > 0 {
				p.writeBytes(COMMA, WHITESPACE)
			}
			for j, seg := range item.Path {
				if j > 0 {

				}
				p.writeString(seg)
			}
			if item.Alias != "" {
				p.writeBytes(WHITESPACE)
				p.writeString(AS)
				p.writeString(DCOLON)
				p.writeBytes(WHITESPACE)
				p.writeString(item.Alias)
			}
		}
		p.writeBytes(RBRACE)
	}
	p.writeBytes(SEMICOLON)
}

func (p *printer) printStructMember(m ast.StructMember) {
	switch m := m.(type) {
	case *ast.IfAttrStructField:
		p.writeString(IF_ATTR)
		p.printExpr(m.Cond)
		p.writeBytes(WHITESPACE)
		p.printStructMember(m.Then)
		if m.Else != nil {
			p.writeString(ELSE_ATTR)
			p.writeBytes(WHITESPACE)
			p.printStructMember(m.Else)
		}
	case *ast.StructField:
		p.printAttrs(m.Attrs)
		p.writeString(m.Name)
		p.writeBytes(COLON, WHITESPACE)
		p.printTypeSpecifier(m.Type)
	}
}

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
		p.writeString(e.Name)
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
	case *ast.BreakIfStmt:
		p.printAttrs(s.Attrs)
		p.writeString(BREAK)
		p.writeBytes(WHITESPACE)
		p.writeString(IF)
		p.writeBytes(WHITESPACE)
		p.printExpr(s.Cond)
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
		p.printCompoundStmt(s.Body)
	case *ast.CompoundStmt:
		p.printCompoundStmt(s)
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
	case *ast.FnCallStmt:
		p.printAttrs(s.Attrs)
		p.printCallExpr(&s.Call)
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
		p.printCompoundStmt(s.Body)
	case *ast.IncDecStmt:
		p.printAttrs(s.Attrs)
		p.printExpr(s.LHS)
		p.writeString(s.Op)
	case *ast.LoopStmt:
		p.printAttrs(s.Attrs)
		p.writeString(LOOP)
		p.writeBytes(WHITESPACE)
		p.printCompoundStmt(s.Body)
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
		p.printSwitchClauses(s.Clauses)
		p.writeBytes(WHITESPACE, RBRACE)
	case *ast.VarOrValueStmt:
		p.printAttrs(s.Attrs)
		if s.Keyword != "" {
			p.writeString(s.Keyword)
			p.writeBytes(WHITESPACE)
		}
		if s.Name != "" {
			p.writeString(s.Name)
			if s.Type != nil {
				p.writeBytes(COLON, WHITESPACE)
				p.printTypeSpecifier(*s.Type)
			}
		}
		if s.Decl != nil {
			p.printDecl(s.Decl)
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
		p.printCompoundStmt(s.Body)
	}
}

func (p *printer) printSwitchClauses(clauses []ast.SwitchClause) {
	for _, c := range clauses {
		p.printSwitchClause(c)
	}
}

func (p *printer) printSwitchClause(c ast.SwitchClause) {
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
		p.printCompoundStmt(c.Body)
	case *ast.IfAttrClause:
		p.writeString(IF_ATTR)
		p.printExpr(c.Cond)
		p.writeBytes(WHITESPACE)
		p.printSwitchClause(c.Then)
		if c.Else != nil {
			p.writeString(ELSE_ATTR)
			p.writeBytes(WHITESPACE)
			p.printSwitchClause(c.Else)
		}
	}
}

func (p *printer) printCallExpr(e *ast.CallExpr) {
	p.writeString(e.Callee)
	p.printTemplateArgs(e.TemplateArgs)
	p.writeBytes(LPAREN)
	p.printExprList(e.Args)
	p.writeBytes(RPAREN)
}

func (p *printer) printIfStmt(s *ast.IfStmt) {
	p.printAttrs(s.Attrs)
	p.writeString(IF)
	p.writeBytes(WHITESPACE)
	p.printExpr(s.Cond)
	p.writeBytes(WHITESPACE)
	p.printCompoundStmt(s.Then)
	if s.ElseIf != nil {
		p.writeString(ELSE)
		p.writeBytes(WHITESPACE)
		p.printIfStmt(s.ElseIf)
	}
	if s.Else != nil {
		p.writeString(ELSE)
		p.writeBytes(WHITESPACE)
		p.printCompoundStmt(s.Else)
	}
}

func (p *printer) printExprList(exprs []ast.Expr) {
	for i, e := range exprs {
		if i > 0 {
			p.writeBytes(COMMA, WHITESPACE)
		}
		p.printExpr(e)
	}
}

func (p *printer) printFile(n *ast.File) {
	for i, d := range n.Decls {
		if i > 0 {
			p.writeBytes(WHITESPACE)
		}
		p.printDecl(d)
	}
}

func (p *printer) printFuncDecl(f *ast.FuncDecl) {
	p.printAttrs(f.Attrs)
	p.writeString(FUNC)
	p.writeBytes(WHITESPACE)
	p.writeString(f.Name)
	p.printParamList(f.Params)
	p.writeBytes(WHITESPACE)
	if f.ReturnType != nil {
		p.writeString(ARROW)
		p.writeBytes(WHITESPACE)
		p.printAttrs(f.ReturnAttrs)
		p.printTypeSpecifier(*f.ReturnType)
		p.writeBytes(WHITESPACE)
	}
	p.printCompoundStmt(f.Body)
}

func (p *printer) printCompoundStmt(s *ast.CompoundStmt) {
	p.printAttrs(s.Attrs)
	p.writeBytes(LBRACE, WHITESPACE)

	var isCompoundStmt bool
	for i, s := range s.Stmts {
		if i > 0 {
			if !isCompoundStmt {
				p.writeBytes(SEMICOLON)
			}
			p.writeBytes(WHITESPACE)
		}

		switch s.(type) {
		case *ast.CompoundStmt, *ast.IfStmt, *ast.WhileStmt, *ast.ForStmt, *ast.LoopStmt, *ast.SwitchStmt, *ast.ContinuingStmt:
			isCompoundStmt = true
		default:
			isCompoundStmt = false
		}
		p.printStmt(s)
	}

	//FIXME: Hack
	if len(s.Stmts) > 0 {
		if !isCompoundStmt {
			p.writeBytes(SEMICOLON)
		}
		p.writeBytes(WHITESPACE)
	}
	p.writeBytes(RBRACE)
}

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

func (p *printer) printTypeSpecifier(t ast.TypeSpecifier) {
	p.writeString(t.Name)
	p.printTemplateArgs(t.TemplateArgs)
}

func (p *printer) printTemplateArgs(args []ast.Expr) {
	if len(args) > 0 {
		p.writeBytes(LANGLE)
		p.printExprList(args)
		p.writeBytes(RANGLE)
	}
}

func Fprint(w io.Writer, n ast.Node) {
	(&printer{}).Fprint(w, n)
}
