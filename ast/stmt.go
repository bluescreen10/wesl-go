package ast

type Stmt interface{}

type CompoundStmt struct {
	Attrs []Attribute
	Stmts []Stmt
}

type EmptyStmt struct{}

type ReturnStmt struct {
	Attrs []Attribute
	Value Expr
}

type BreakIfStmt struct {
	Attrs []Attribute
	Cond  Expr
}

type BreakStmt struct {
	Attrs []Attribute
}

type ContinueStmt struct {
	Attrs []Attribute
}

type ContinuingStmt struct {
	Attrs []Attribute
	Body  *CompoundStmt
}

type DiscardStmt struct {
	Attrs []Attribute
}

type ConstAssertStmt struct {
	Attrs []Attribute
	Expr  Expr
}

type AssignmentStmt struct {
	Attrs []Attribute
	LHS   Expr
	RHS   Expr
	Op    string
}

type FuncCallStmt struct {
	Attrs []Attribute
	Call  CallExpr
}

type IncrementStmt struct {
	Attrs []Attribute
	LHS   Expr
}

type DecrementStmt struct {
	Attrs []Attribute
	LHS   Expr
}

type CaseClause struct {
	Attrs     []Attribute
	Selectors Expr
	Body      *CompoundStmt
}

type DefaultAloneClause struct {
	Attrs []Attribute
	Body  *CompoundStmt
}

type IfStmt struct {
	Attrs  []Attribute
	Cond   Expr
	Then   *CompoundStmt
	ElseIf *IfStmt
	Else   *CompoundStmt
}

type IfAttrStmt struct {
	Cond Expr
	Then Stmt
	Else Stmt
}

type SwitchStmt struct {
	Attrs   []Attribute
	Expr    Expr
	Clauses []Node
}

type LoopStmt struct {
	Attrs      []Attribute
	BodyAttrs  []Attribute
	Body       *CompoundStmt
	Continuing *ContinuingStmt
}

type ForStmt struct {
	Attrs  []Attribute
	Init   Stmt
	Cond   Expr
	Update Stmt
	Body   *CompoundStmt
}

type WhileStmt struct {
	Attrs []Attribute
	Cond  Expr
	Body  *CompoundStmt
}
