package ast

type Node interface {
	//node()
}

// type SwitchClause interface {
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

	// Enable Directive
	EnableDirective struct {
		Attrs      []Attribute
		Extensions []string
	}

	// Function
	FnDecl struct {
		Name        string
		Attrs       []Attribute
		Params      []Param
		ReturnAttrs []Attribute
		ReturnType  *TypeSpecifier
		Body        *CompoundStmt
	}

	// Global Val
	GlobalValueDecl struct {
		Keyword string
		Attrs   []Attribute
		Ident   OptionallyTypedIdent
		Init    Expr
	}

	// Global Var
	GlobalVariableDecl struct {
		Decl Decl
		Init Expr
	}

	// @if
	IfAttrDecl IfAttr[Decl]

	// Import
	ImportDecl struct {
		Symbol string
		Alias  string
		Path   string
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
		Members []StructMember
	}

	// Type Alias
	TypeAliasDecl struct {
		Name  string
		Attrs []Attribute
		Type  TypeSpecifier
	}

	// Variable
	VariableDecl struct {
		Ident        OptionallyTypedIdent
		Attrs        []Attribute
		TemplateArgs []Expr
	}
)

func (*ConstAssertDecl) declNode()     {}
func (*DiagnosticDirective) declNode() {}
func (*EnableDirective) declNode()     {}
func (*FnDecl) declNode()              {}
func (*GlobalValueDecl) declNode()     {}
func (*GlobalVariableDecl) declNode()  {}
func (*ImportDecl) declNode()          {}
func (ImportsDecl) declNode()          {} //FIXME: remove
func (*IfAttrDecl) declNode()          {}
func (*RequiresDirective) declNode()   {}
func (*StructDecl) declNode()          {}
func (*TypeAliasDecl) declNode()       {}
func (*VariableDecl) declNode()        {}

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

	// Function Call
	FnCallStmt struct {
		Attrs []Attribute
		Call  CallExpr
	}

	// For
	ForStmt struct {
		Attrs  []Attribute
		Init   Stmt
		Cond   Expr
		Update Stmt
		Body   *CompoundStmt
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
	IncrementStmt struct {
		Attrs []Attribute
		LHS   Expr
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
		Clauses []SwitchClause
	}

	// Var or value
	VarOrValueStmt struct {
		Attrs   []Attribute
		Keyword string
		Decl    *VariableDecl
		Ident   *OptionallyTypedIdent
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
func (*DecrementStmt) stmtNode()   {}
func (*DiscardStmt) stmtNode()     {}
func (*EmptyStmt) stmtNode()       {}
func (*FnCallStmt) stmtNode()      {}
func (*ForStmt) stmtNode()         {}
func (*IfStmt) stmtNode()          {}
func (*IfAttrStmt) stmtNode()      {}
func (*IncrementStmt) stmtNode()   {}
func (*LoopStmt) stmtNode()        {}
func (*ReturnStmt) stmtNode()      {}
func (*SwitchStmt) stmtNode()      {}
func (*VarOrValueStmt) stmtNode()  {}
func (*WhileStmt) stmtNode()       {}

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

	// Diagnostic Control
	DiagnosticControl struct {
		Severity string
		RuleName string
	}

	// Optionally Typed
	OptionallyTypedIdent struct {
		Name string
		Type *TypeSpecifier
	}

	// Function Param
	Param interface {
		Node
		paramNode()
	}

	// @if Param
	IfAttrParam IfAttr[Param]

	// Param
	FnParam struct {
		Name  string
		Type  TypeSpecifier
		Attrs []Attribute
	}

	// Type
	TypeSpecifier struct {
		Name         string
		TemplateArgs []Expr
	}

	// Struct Member
	StructMember interface {
		Node
		structMemberNode()
	}

	// Struct Field
	StructField struct {
		Name  string
		Attrs []Attribute
		Type  TypeSpecifier
	}

	// @if Struct Member
	IfAttrStructField IfAttr[StructMember]

	// Switch clauses
	SwitchClause interface {
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
	IfAttrClause IfAttr[SwitchClause]

	// Default
	DefaultAloneClause struct {
		Attrs []Attribute
		Body  *CompoundStmt
	}
)

func (*IfAttrStructField) structMemberNode() {}
func (*StructField) structMemberNode()       {}

func (*IfAttrClause) switchClauseNode()       {}
func (*CaseClause) switchClauseNode()         {}
func (*DefaultAloneClause) switchClauseNode() {}

func (*FnParam) paramNode()     {}
func (*IfAttrParam) paramNode() {}

type File struct {
	Decls   []Decl
	Imports []ImportDecl
}

func (ts TypeSpecifier) AsExpr() Expr {
	if len(ts.TemplateArgs) == 0 {
		return &Ident{Name: ts.Name}
	}
	return &CallExpr{Callee: ts.Name, TemplateArgs: ts.TemplateArgs}
}

type ImportsDecl []ImportDecl
