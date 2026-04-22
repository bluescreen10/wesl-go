package ast

import "io"

type FnDecl struct {
	Name        string
	Attrs       []Attribute
	Params      []Param
	ReturnAttrs []Attribute
	ReturnType  *TypeSpecifier
	Body        *CompoundStmt
}

func (f FnDecl) Emit(w io.Writer) {
	for i, a := range f.Attrs {
		if i > 0 {
			w.Write([]byte{' '})
		}
		a.Emit(w)

	}
	w.Write([]byte{'f', 'n', ' '})
	w.Write([]byte(f.Name))
	w.Write([]byte{'('})
	count := len(f.Params)
	for i, p := range f.Params {
		p.Emit(w)
		if i != count-1 {
			w.Write([]byte{',', ' '})
		}
	}
	w.Write([]byte{')', ' '})
	if f.ReturnType != nil {
		w.Write([]byte{'-', '>', ' '})
		for _, a := range f.ReturnAttrs {
			a.Emit(w)
			w.Write([]byte{' '})
		}
		f.ReturnType.Emit(w)
		w.Write([]byte{' '})
	}
	f.Body.Emit(w)
}

type Param interface {
	Emit(w io.Writer)
}

type IfAttrParam struct {
	Cond Expr
	Then FnParam
	Else *FnParam
}

func (p IfAttrParam) Emit(w io.Writer) {
	w.Write([]byte{'@', 'i', 'f'})
	p.Cond.Emit(w)
	w.Write([]byte{' '})
	p.Then.Emit(w)
	if p.Else != nil {
		p.Else.Emit(w)
	}
}

type FnParam struct {
	Name  string
	Type  TypeSpecifier
	Attrs []Attribute
}

func (p FnParam) Emit(w io.Writer) {
	for _, a := range p.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte(p.Name))
	w.Write([]byte{':', ' '})
	p.Type.Emit(w)
}
