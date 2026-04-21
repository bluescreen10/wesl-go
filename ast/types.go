package ast

type Ident struct {
	Name string
}

type TypeSpecifier struct {
	Name         string
	TemplateArgs []Expr
}

func (ts TypeSpecifier) AsExpr() Expr {
	if len(ts.TemplateArgs) == 0 {
		return &Ident{Name: ts.Name}
	}
	return &CallExpr{Callee: ts.Name, TemplateArgs: ts.TemplateArgs}
}

type OptionallyTypedIdent struct {
	Name string
	Type *TypeSpecifier
}

type Arg struct {
}

type StructMember interface {
	//structMemberMarker()
}

type StructDecl struct {
	Name    string
	Attrs   []Attribute
	Members []StructMember
}

type StructField struct {
	Name  string
	Attrs []Attribute
	Type  TypeSpecifier
}

type IfAttrStructField struct {
	Cond Expr
	Then StructField
	Else *StructField
}

type TypeAliasDecl struct {
	Name  string
	Attrs []Attribute
	Type  TypeSpecifier
}
