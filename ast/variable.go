package ast

type VariableDecl struct {
	Ident        OptionallyTypedIdent
	Attrs        []Attribute
	TemplateArgs Expr
}

type GlobalVariableDecl struct {
	Decl Expr
	Init Expr
}

type GlobalValueDecl struct {
	Keyword string
	Attrs   []Attribute
	Ident   OptionallyTypedIdent
	Init    Expr
}

type VarOrValueStmt struct {
	Attrs   []Attribute
	Keyword string
	Decl    *VariableDecl
	Ident   *OptionallyTypedIdent
	Init    Expr
}

//: attrs, TemplateArgs: templateArgs, Ident: ident}
