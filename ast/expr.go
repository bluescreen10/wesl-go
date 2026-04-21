package ast

type Expr interface {
}

type BinaryExpr struct {
	Op    string
	Left  Expr
	Right Expr
}

type UnaryExpr struct {
	Op      string
	Operand Expr
}

type DerefExpr struct {
	Operand Expr
}

type AddrOfExpr struct {
	Operand Expr
}

type MemberExpr struct {
	Base   Expr
	Member string
}

type IndexExpr struct {
	Base  Expr
	Index Expr
}

type LitExpr struct {
	Val string
}

type ParenExpr struct {
	Inner Expr
}

type CallExpr struct {
	Callee       string
	TemplateArgs Expr
	Args         []Expr
}
