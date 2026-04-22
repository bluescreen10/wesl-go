package ast

import "io"

type Attribute interface {
	Emit(w io.Writer)
}

type GenericAttribute struct {
	Name string
	Args []Expr
}

func (a GenericAttribute) Emit(w io.Writer) {
	w.Write([]byte(a.Name))
	if count := len(a.Args); count > 0 {
		w.Write([]byte{'('})
		for i, arg := range a.Args {
			arg.Emit(w)
			if i != count-1 {
				w.Write([]byte{',', ' '})
			}
		}
		w.Write([]byte{')', ' '})
	}
}
