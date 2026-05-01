package ast

type Node interface {
	//node()
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

	// Const Assert
	ConstAssertDecl struct {
		Assert *ConstAssertStmt
	}

	// Diagnostic Directive
	DiagnosticDirective struct {
		Attrs   []Attribute
		Control DiagnosticControl
	}

	// Diagnostic Control
	DiagnosticControl struct {
		Severity string
		RuleName string
	}

	// Enable Directive
	EnableDirective struct {
		Attrs      []Attribute
		Extensions []string
	}

	// Function
	FuncDecl struct {
		Name        string
		Attrs       []Attribute
		Params      []Param
		ReturnAttrs []Attribute
		ReturnType  *TypeSpecifier
		Body        *CompoundStmt
	}

	// Function Param
	Param interface {
		Node
		paramNode()
	}

	// Param
	FuncParam struct {
		Name  string
		Type  TypeSpecifier
		Attrs []Attribute
	}

	// @if Param
	IfAttrParam IfAttr[Param]

	// Global Val
	GlobalValDecl struct {
		Keyword string
		Attrs   []Attribute
		Name    string
		Type    *TypeSpecifier
		Init    Expr
	}

	// Global Var
	GlobalVarDecl struct {
		Attrs        []Attribute
		TemplateArgs []Expr
		Name         string
		Type         *TypeSpecifier
		Init         Expr
	}

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
		Attrs      []Attribute
		Extensions []string
	}

	// Struct
	StructDecl struct {
		Name    string
		Attrs   []Attribute
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
		Attrs []Attribute
		Type  TypeSpecifier
	}

	// @if Struct Member
	IfAttrStructMember IfAttr[Member]

	// Type Alias
	TypeAliasDecl struct {
		Name  string
		Attrs []Attribute
		Type  TypeSpecifier
	}
)

func (*ConstAssertDecl) declNode()     {}
func (*DiagnosticDirective) declNode() {}
func (*EnableDirective) declNode()     {}
func (*FuncDecl) declNode()            {}
func (*GlobalValDecl) declNode()       {}
func (*GlobalVarDecl) declNode()       {}
func (*ImportDecl) declNode()          {}
func (*IfAttrDecl) declNode()          {}
func (*RequiresDirective) declNode()   {}
func (*StructDecl) declNode()          {}
func (*TypeAliasDecl) declNode()       {}

func (_ *ConstAssertDecl) GetName() string     { return "" }
func (_ *DiagnosticDirective) GetName() string { return "" }
func (_ *EnableDirective) GetName() string     { return "" }
func (d *FuncDecl) GetName() string            { return d.Name }
func (d *GlobalValDecl) GetName() string       { return d.Name }
func (d *GlobalVarDecl) GetName() string       { return d.Name }
func (_ *ImportDecl) GetName() string          { return "" }
func (_ *IfAttrDecl) GetName() string          { return "" }
func (_ *RequiresDirective) GetName() string   { return "" }
func (d *StructDecl) GetName() string          { return d.Name }
func (d *TypeAliasDecl) GetName() string       { return d.Name }

func (_ *ConstAssertDecl) SetName(string)     {}
func (_ *DiagnosticDirective) SetName(string) {}
func (_ *EnableDirective) SetName(string)     {}
func (d *FuncDecl) SetName(n string)          { d.Name = n }
func (d *GlobalValDecl) SetName(n string)     { d.Name = n }
func (d *GlobalVarDecl) SetName(n string)     { d.Name = n }
func (_ *ImportDecl) SetName(string)          {}
func (_ *IfAttrDecl) SetName(string)          {}
func (_ *RequiresDirective) SetName(string)   {}
func (d *StructDecl) SetName(n string)        { d.Name = n }
func (d *TypeAliasDecl) SetName(n string)     { d.Name = n }

func (*IfAttrStructMember) structMemberNode() {}
func (*StructMember) structMemberNode()       {}

func (*FuncParam) paramNode()   {}
func (*IfAttrParam) paramNode() {}

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
		Attrs []Attribute
		LHS   Expr
		RHS   Expr
		Op    string
	}

	// Break
	BreakStmt struct {
		Attrs []Attribute
	}

	// Break If
	BreakIfStmt struct {
		Attrs []Attribute
		Cond  Expr
	}

	// Const Assert
	ConstAssertStmt struct {
		Attrs []Attribute
		Expr  Expr
	}

	// Continue
	ContinueStmt struct {
		Attrs []Attribute
	}

	// Continuing
	ContinuingStmt struct {
		Attrs []Attribute
		Body  *CompoundStmt
	}

	// Compound
	CompoundStmt struct {
		Attrs []Attribute
		Stmts []Stmt
	}

	// Decrement
	DecrementStmt struct {
		Attrs []Attribute
		LHS   Expr
	}

	// Discard
	DiscardStmt struct {
		Attrs []Attribute
	}

	// Empty
	EmptyStmt struct{}

	// For
	ForStmt struct {
		Attrs  []Attribute
		Init   Stmt
		Cond   Expr
		Update Stmt
		Body   *CompoundStmt
	}

	// Function Call
	FuncCallStmt struct {
		Attrs []Attribute
		Call  CallExpr
	}

	// If
	IfStmt struct {
		Attrs  []Attribute
		Cond   Expr
		Then   *CompoundStmt
		ElseIf *IfStmt
		Else   *CompoundStmt
	}

	// @If
	IfAttrStmt IfAttr[Stmt]

	// Increment
	IncDecStmt struct {
		Attrs []Attribute
		LHS   Expr
		Op    string
	}

	// Loop
	LoopStmt struct {
		Attrs     []Attribute
		BodyAttrs []Attribute
		Body      *CompoundStmt
	}

	// Return
	ReturnStmt struct {
		Attrs []Attribute
		Value Expr
	}

	// Switch
	SwitchStmt struct {
		Attrs   []Attribute
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
		Attrs     []Attribute
		Selectors []Expr
		Body      *CompoundStmt
	}

	// @if
	IfAttrClause IfAttr[Clause]

	// Local var statement
	VarStmt struct {
		Attrs        []Attribute
		TemplateArgs []Expr
		Name         string
		Type         *TypeSpecifier
		Init         Expr
	}

	// Local let/const statement (Keyword is "let" or "const")
	ValStmt struct {
		Attrs   []Attribute
		Keyword string
		Name    string
		Type    *TypeSpecifier
		Init    Expr
	}

	// While
	WhileStmt struct {
		Attrs []Attribute
		Cond  Expr
		Body  *CompoundStmt
	}
)

func (*AssignmentStmt) stmtNode()  {}
func (*BreakStmt) stmtNode()       {}
func (*BreakIfStmt) stmtNode()     {}
func (*CompoundStmt) stmtNode()    {}
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

func (*IfAttrClause) switchClauseNode() {}
func (*CaseClause) switchClauseNode()   {}

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
		Callee       string
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
		Name string
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

// ----------------------------------------------------------------------------
// Type, Attributes, Identifiers, Values, etc.

type (
	// Attribute
	Attribute struct {
		Name string
		Args []Expr
	}

	// Type
	TypeSpecifier struct {
		Name         string
		TemplateArgs []Expr
	}
)

type File struct {
	Decls []Decl
}

func (ts TypeSpecifier) AsExpr() Expr {
	if len(ts.TemplateArgs) == 0 {
		return &Ident{Name: ts.Name}
	}
	return &CallExpr{Callee: ts.Name, TemplateArgs: ts.TemplateArgs}
}
