package ast

type Node interface {
	node()
}

// type Clause interface {
// 	Node
// }

type IfAttr[T any] struct {
	Cond Expr
	Then T
	Else T
}

// ----------------------------------------------------------------------------
// Decls
type (
	// Interface
	Decl interface {
		Node
		GetName() string
		SetName(string)
		declNode()
	}

	// Diagnostic Directive
	DiagnosticDirective struct {
		Attrs   []*Attribute
		Control DiagnosticControl
	}

	// Diagnostic Control
	DiagnosticControl struct {
		Severity string
		RuleName string
	}

	// Enable Directive
	EnableDirective struct {
		Attrs      []*Attribute
		Extensions []string
	}

	// Function
	FuncDecl struct {
		Name        *Ident
		Attrs       []*Attribute
		Params      []Param
		ReturnAttrs []*Attribute
		ReturnType  *TypeSpecifier
		Body        *BlockStmt
	}

	// Function Param
	Param interface {
		Node
		paramNode()
	}

	// Param
	FuncParam struct {
		Name  string
		Type  *TypeSpecifier
		Attrs []*Attribute
	}

	// @if Param
	IfAttrParam IfAttr[Param]

	// @if
	IfAttrDecl IfAttr[Decl]

	// Import declaration
	ImportDecl struct {
		Imports []ImportedItem
	}

	// ImportedItem is a single fully-qualified import with an optional alias.
	// Path holds every segment including the leading anchor (package/super) and
	// the final symbol name.  Example: import package::foo::bar as b
	// → ImportedItem{Path: ["package","foo","bar"], Alias: "b"}
	ImportedItem struct {
		Path  []string
		Alias string
	}

	// Reqiures Directive
	RequiresDirective struct {
		Attrs      []*Attribute
		Extensions []string
	}

	// Struct
	StructDecl struct {
		Name    *Ident
		Attrs   []*Attribute
		Members []Member
	}

	// Struct Member
	Member interface {
		Node
		structMemberNode()
	}

	// Struct Field
	StructMember struct {
		Name  string
		Attrs []*Attribute
		Type  *TypeSpecifier
	}

	// @if Struct Member
	IfAttrStructMember IfAttr[Member]

	// Type Alias
	TypeAliasDecl struct {
		Name  *Ident
		Attrs []*Attribute
		Type  *TypeSpecifier
	}
)

func (*DiagnosticDirective) declNode() {}
func (*EnableDirective) declNode()     {}
func (*FuncDecl) declNode()            {}
func (*ImportDecl) declNode()          {}
func (*IfAttrDecl) declNode()          {}
func (*RequiresDirective) declNode()   {}
func (*StructDecl) declNode()          {}
func (*TypeAliasDecl) declNode()       {}
func (*VarStmt) declNode()             {}
func (*ValStmt) declNode()             {}
func (*ConstAssertStmt) declNode()     {}

func (*DiagnosticDirective) node() {}
func (*EnableDirective) node()     {}
func (*FuncDecl) node()            {}
func (*ImportDecl) node()          {}
func (*IfAttrDecl) node()          {}
func (*RequiresDirective) node()   {}
func (*StructDecl) node()          {}
func (*TypeAliasDecl) node()       {}

func (_ *DiagnosticDirective) GetName() string  { return "" }
func (_ *EnableDirective) GetName() string      { return "" }
func (d *FuncDecl) GetName() string             { return d.Name.Val }
func (_ *ImportDecl) GetName() string           { return "" }
func (_ *IfAttrDecl) GetName() string           { return "" }
func (_ *RequiresDirective) GetName() string    { return "" }
func (d *StructDecl) GetName() string           { return d.Name.Val }
func (d *TypeAliasDecl) GetName() string        { return d.Name.Val }
func (d *VarStmt) GetName() string              { return d.Name.Val }
func (d *ValStmt) GetName() string              { return d.Name.Val }
func (_ *ConstAssertStmt) GetName() string      { return "" }

func (_ *DiagnosticDirective) SetName(string)  {}
func (_ *EnableDirective) SetName(string)      {}
func (d *FuncDecl) SetName(n string)           { d.Name.Val = n }
func (_ *ImportDecl) SetName(string)           {}
func (_ *IfAttrDecl) SetName(string)           {}
func (_ *RequiresDirective) SetName(string)    {}
func (d *StructDecl) SetName(n string)         { d.Name.Val = n }
func (d *TypeAliasDecl) SetName(n string)      { d.Name.Val = n }
func (d *VarStmt) SetName(n string)            { d.Name.Val = n }
func (d *ValStmt) SetName(n string)            { d.Name.Val = n }
func (_ *ConstAssertStmt) SetName(string)      {}

func (*IfAttrStructMember) structMemberNode() {}
func (*StructMember) structMemberNode()       {}

func (*IfAttrStructMember) node() {}
func (*StructMember) node()       {}

func (*FuncParam) paramNode()   {}
func (*IfAttrParam) paramNode() {}

func (*FuncParam) node()   {}
func (*IfAttrParam) node() {}

// ----------------------------------------------------------------------------
// Stmt
type (
	// Interface
	Stmt interface {
		Node
		stmtNode()
	}

	// Assignment
	AssignmentStmt struct {
		Attrs []*Attribute
		LHS   Expr
		RHS   Expr
		Op    string
	}

	// Break
	BreakStmt struct {
		Attrs []*Attribute
	}

	// Break If
	BreakIfStmt struct {
		Attrs []*Attribute
		Cond  Expr
	}

	// Const Assert
	ConstAssertStmt struct {
		Attrs []*Attribute
		Expr  Expr
	}

	// Continue
	ContinueStmt struct {
		Attrs []*Attribute
	}

	// Continuing
	ContinuingStmt struct {
		Attrs []*Attribute
		Body  *BlockStmt
	}

	// Compound
	BlockStmt struct {
		Attrs []*Attribute
		Stmts []Stmt
	}

	// Decrement
	DecrementStmt struct {
		Attrs []*Attribute
		LHS   Expr
	}

	// Discard
	DiscardStmt struct {
		Attrs []*Attribute
	}

	// Empty
	EmptyStmt struct{}

	// For
	ForStmt struct {
		Attrs  []*Attribute
		Init   Stmt
		Cond   Expr
		Update Stmt
		Body   *BlockStmt
	}

	// Function Call
	FuncCallStmt struct {
		Attrs []*Attribute
		Call  *CallExpr
	}

	// If
	IfStmt struct {
		Attrs  []*Attribute
		Cond   Expr
		Then   *BlockStmt
		ElseIf *IfStmt
		Else   *BlockStmt
	}

	// @If
	IfAttrStmt IfAttr[Stmt]

	// Increment
	IncDecStmt struct {
		Attrs []*Attribute
		LHS   Expr
		Op    string
	}

	// Loop
	LoopStmt struct {
		Attrs     []*Attribute
		BodyAttrs []*Attribute
		Body      *BlockStmt
	}

	// Return
	ReturnStmt struct {
		Attrs []*Attribute
		Value Expr
	}

	// Switch
	SwitchStmt struct {
		Attrs   []*Attribute
		Expr    Expr
		Clauses []Clause
	}

	// Switch clauses
	Clause interface {
		Node
		switchClauseNode()
	}

	// Case
	CaseClause struct {
		Attrs     []*Attribute
		Selectors []Expr
		Body      *BlockStmt
	}

	// @if
	IfAttrClause IfAttr[Clause]

	// Local var statement
	VarStmt struct {
		Attrs        []*Attribute
		TemplateArgs []Expr
		Name         *Ident
		Type         *TypeSpecifier
		Init         Expr
	}

	// Local let/const statement (Keyword is "let" or "const")
	ValStmt struct {
		Attrs   []*Attribute
		Keyword string
		Name    *Ident
		Type    *TypeSpecifier
		Init    Expr
	}

	// While
	WhileStmt struct {
		Attrs []*Attribute
		Cond  Expr
		Body  *BlockStmt
	}
)

func (*AssignmentStmt) stmtNode()  {}
func (*BreakStmt) stmtNode()       {}
func (*BreakIfStmt) stmtNode()     {}
func (*BlockStmt) stmtNode()       {}
func (*ConstAssertStmt) stmtNode() {}
func (*ContinueStmt) stmtNode()    {}
func (*ContinuingStmt) stmtNode()  {}
func (*DiscardStmt) stmtNode()     {}
func (*EmptyStmt) stmtNode()       {}
func (*ForStmt) stmtNode()         {}
func (*FuncCallStmt) stmtNode()    {}
func (*IfStmt) stmtNode()          {}
func (*IfAttrStmt) stmtNode()      {}
func (*IncDecStmt) stmtNode()      {}
func (*LoopStmt) stmtNode()        {}
func (*ReturnStmt) stmtNode()      {}
func (*SwitchStmt) stmtNode()      {}
func (*VarStmt) stmtNode()         {}
func (*ValStmt) stmtNode()         {}
func (*WhileStmt) stmtNode()       {}

func (*AssignmentStmt) node()  {}
func (*BreakStmt) node()       {}
func (*BreakIfStmt) node()     {}
func (*BlockStmt) node()       {}
func (*ConstAssertStmt) node() {}
func (*ContinueStmt) node()    {}
func (*ContinuingStmt) node()  {}
func (*DiscardStmt) node()     {}
func (*EmptyStmt) node()       {}
func (*ForStmt) node()         {}
func (*FuncCallStmt) node()    {}
func (*IfStmt) node()          {}
func (*IfAttrStmt) node()      {}
func (*IncDecStmt) node()      {}
func (*LoopStmt) node()        {}
func (*ReturnStmt) node()      {}
func (*SwitchStmt) node()      {}
func (*VarStmt) node()         {}
func (*ValStmt) node()         {}
func (*WhileStmt) node()       {}

func (*IfAttrClause) switchClauseNode() {}
func (*CaseClause) switchClauseNode()   {}

func (*IfAttrClause) node() {}
func (*CaseClause) node()   {}

// ----------------------------------------------------------------------------
// Expr
type (
	// Interface
	Expr interface {
		Node
		exprNode()
	}

	// Ref
	AddrOfExpr struct {
		Operand Expr
	}

	// Binary
	BinaryExpr struct {
		Op    string
		Left  Expr
		Right Expr
	}

	// Function Call
	CallExpr struct {
		Callee       *Ident
		TemplateArgs []Expr
		Args         []Expr
	}

	// Deref
	DerefExpr struct {
		Operand Expr
	}

	// @if
	//IfAttrExpr IfAttr[Expr]

	// Ident
	Ident struct {
		Val string
	}

	// Index
	IndexExpr struct {
		Base  Expr
		Index Expr
	}

	// Literal
	LitExpr struct {
		Val string
	}

	// Member
	MemberExpr struct {
		Base   Expr
		Member string
	}

	// Parenthesis
	ParenExpr struct {
		Inner Expr
	}

	// Unary
	UnaryExpr struct {
		Op      string
		Operand Expr
	}
)

func (*AddrOfExpr) exprNode() {}
func (*BinaryExpr) exprNode() {}
func (*CallExpr) exprNode()   {}
func (*DerefExpr) exprNode()  {}
func (*Ident) exprNode()      {}
func (*IndexExpr) exprNode()  {}
func (*LitExpr) exprNode()    {}
func (*MemberExpr) exprNode() {}
func (*ParenExpr) exprNode()  {}
func (*UnaryExpr) exprNode()  {}

func (*AddrOfExpr) node() {}
func (*BinaryExpr) node() {}
func (*CallExpr) node()   {}
func (*DerefExpr) node()  {}
func (*Ident) node()      {}
func (*IndexExpr) node()  {}
func (*LitExpr) node()    {}
func (*MemberExpr) node() {}
func (*ParenExpr) node()  {}
func (*UnaryExpr) node()  {}

// ----------------------------------------------------------------------------
// Type, Attributes, Identifiers, Values, etc.

type (
	// Attribute
	Attribute struct {
		Node
		Name string
		Args []Expr
	}

	// Type
	TypeSpecifier struct {
		Node
		Name         string
		TemplateArgs []Expr
	}
)

type File struct {
	Node
	Decls []Decl
}

func (*Attribute) node()     {}
func (*TypeSpecifier) node() {}
func (*File) node()          {}

func (ts TypeSpecifier) AsExpr() Expr {
	if len(ts.TemplateArgs) == 0 {
		return &Ident{Val: ts.Name}
	}
	return &CallExpr{Callee: &Ident{Val: ts.Name}, TemplateArgs: ts.TemplateArgs}
}
