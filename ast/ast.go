package ast

import "io"

type Decl interface {
	Emit(w io.Writer)
	//	Node
	//
	// declMarker()
}

type Node interface {
	Emit(w io.Writer)
	//node()
}

type SwitchClause interface {
	Node
}

// type Text struct {
// 	Content string
// }

// func (*Text) node() {}

// type Block struct {
// 	Nodes []Node
// }

// func (*Block) node() {}

type File struct {
	Decls   []Decl
	Imports []ImportDecl
}

func (f *File) Emit(w io.Writer) {
	count := len(f.Decls)
	for i, d := range f.Decls {
		d.Emit(w)
		if i != count-1 {
			w.Write([]byte{' '})
		}
	}
}

type IfAttrDecl struct {
	Cond Expr
	Then Decl
	Else Decl // nil when no @else
}

func (d *IfAttrDecl) Emit(w io.Writer) {
	w.Write([]byte{'@', 'i', 'f', '('})
	d.Cond.Emit(w)
	w.Write([]byte{')'})
	d.Then.Emit(w)
	if d.Else != nil {
		d.Else.Emit(w)
	}
}
