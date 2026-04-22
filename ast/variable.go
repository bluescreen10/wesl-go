package ast

import "io"

// type ConstAssertDecl struct {
// 	Assert *ConstAssertStmt
// }

func (x ConstAssertDecl) Emit(w io.Writer) {
	x.Assert.Emit(w)
	w.Write([]byte{';'})
}

// type VariableDecl struct {
// 	Ident        OptionallyTypedIdent
// 	Attrs        []Attribute
// 	TemplateArgs []Expr
// }

func (v VariableDecl) Emit(w io.Writer) {
	for _, a := range v.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte{'v', 'a', 'r'})
	if v.TemplateArgs != nil {
		w.Write([]byte{'<'})
		count := len(v.TemplateArgs)
		for i, arg := range v.TemplateArgs {
			arg.Emit(w)
			if i != count-1 {
				w.Write([]byte{',', ' '})
			}
		}
		w.Write([]byte{'>'})
	}
	w.Write([]byte{' '})
	v.Ident.Emit(w)
}

// type GlobalVariableDecl struct {
// 	Decl Expr
// 	Init Expr
// }

func (g GlobalVariableDecl) Emit(w io.Writer) {
	g.Decl.Emit(w)
	if g.Init != nil {
		w.Write([]byte{' ', '=', ' '})
		g.Init.Emit(w)
	}
	w.Write([]byte{';'})
}

// type GlobalValueDecl struct {
// 	Keyword string
// 	Attrs   []Attribute
// 	Ident   OptionallyTypedIdent
// 	Init    Expr
// }

func (g GlobalValueDecl) Emit(w io.Writer) {
	for _, a := range g.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte(g.Keyword))
	w.Write([]byte{' '})
	g.Ident.Emit(w)
	if g.Init != nil {
		w.Write([]byte{' ', '=', ' '})
		g.Init.Emit(w)
	}
	w.Write([]byte{';'})
}

// type VarOrValueStmt struct {
// 	Attrs   []Attribute
// 	Keyword string
// 	Decl    *VariableDecl
// 	Ident   *OptionallyTypedIdent
// 	Init    Expr
// }

func (s VarOrValueStmt) Emit(w io.Writer) {
	for _, a := range s.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	if s.Keyword != "" {
		w.Write([]byte(s.Keyword))
		w.Write([]byte{' '})
	}
	if s.Ident != nil {
		s.Ident.Emit(w)
	}
	if s.Decl != nil {
		s.Decl.Emit(w)
	}
	if s.Init != nil {
		w.Write([]byte{' ', '=', ' '})
		s.Init.Emit(w)
	}
}

//: attrs, TemplateArgs: templateArgs, Ident: ident}
