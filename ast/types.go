package ast

import "io"

type Ident struct {
	Name string
}

func (i Ident) Emit(w io.Writer) {
	w.Write([]byte(i.Name))
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

func (t TypeSpecifier) Emit(w io.Writer) {
	w.Write([]byte(t.Name))
	if t.TemplateArgs != nil {
		w.Write([]byte{'<'})
		count := len(t.TemplateArgs)
		for i, arg := range t.TemplateArgs {
			arg.Emit(w)
			if i != count-1 {
				w.Write([]byte{',', ' '})
			}
		}
		w.Write([]byte{'>'})
	}
}

type OptionallyTypedIdent struct {
	Name string
	Type *TypeSpecifier
}

func (o OptionallyTypedIdent) Emit(w io.Writer) {
	w.Write([]byte(o.Name))
	if o.Type != nil {
		w.Write([]byte{':', ' '})
		o.Type.Emit(w)
	}
}

type Arg struct {
}

type StructMember interface {
	Emit(w io.Writer)
	//structMemberMarker()
}

type StructDecl struct {
	Name    string
	Attrs   []Attribute
	Members []StructMember
}

func (s StructDecl) Emit(w io.Writer) {
	for _, a := range s.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}

	w.Write([]byte{'s', 't', 'r', 'u', 'c', 't', ' '})
	w.Write([]byte(s.Name))
	w.Write([]byte{' ', '{'})
	//FIXME: hack
	count := len(s.Members)
	if count > 0 {
		w.Write([]byte{' '})
	}
	for i, m := range s.Members {
		m.Emit(w)
		if i != count-1 {
			w.Write([]byte{',', ' '})
		}
	}
	//FIXME: hack
	if count > 0 {
		w.Write([]byte{' '})
	}
	w.Write([]byte{'}'})
}

type StructField struct {
	Name  string
	Attrs []Attribute
	Type  TypeSpecifier
}

func (s StructField) Emit(w io.Writer) {
	for _, a := range s.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte(s.Name))
	w.Write([]byte{':', ' '})
	s.Type.Emit(w)
}

type IfAttrStructField struct {
	Cond Expr
	Then StructField
	Else *StructField
}

func (i IfAttrStructField) Emit(w io.Writer) {
	w.Write([]byte{'@', 'i', 'f'})
	i.Cond.Emit(w)
	w.Write([]byte{' '})
	i.Then.Emit(w)
	if i.Else != nil {
		w.Write([]byte{'@', 'e', 'l', 's', 'e', ' '})
		i.Else.Emit(w)
	}
}

type TypeAliasDecl struct {
	Name  string
	Attrs []Attribute
	Type  TypeSpecifier
}

func (t TypeAliasDecl) Emit(w io.Writer) {
	for _, a := range t.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte{'a', 'l', 'i', 'a', 's', ' '})
	w.Write([]byte(t.Name))
	w.Write([]byte{' ', '=', ' '})
	t.Type.Emit(w)
	w.Write([]byte{';'})
}
