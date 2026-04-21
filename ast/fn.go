package ast

type FnDecl struct {
	Name        string
	Attrs       []Attribute
	Params      []Param
	ReturnAttrs []Attribute
	ReturnType  TypeSpecifier
	Body        *CompoundStmt
}

type Param struct {
	Name  string
	Type  TypeSpecifier
	Attrs []Attribute
}
