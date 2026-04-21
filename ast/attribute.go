package ast

type Attribute interface {
}

type GenericAttribute struct {
	Name string
	Args []Expr
}
