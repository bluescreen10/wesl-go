package ast

import "io"

type ImportDecl struct {
	Symbol string
	Alias  string
	Path   string
}

func (i ImportDecl) Emit(w io.Writer) {
	w.Write([]byte(i.Path))
	w.Write([]byte(i.Symbol))
	w.Write([]byte{' ', 'a', 's', ' '})
	w.Write([]byte(i.Alias))
}

func (i *ImportDecl) String() string {
	return i.Path + i.Symbol + " as " + i.Alias
}

// FIXME: hack
type ImportsDecl []ImportDecl

func (d ImportsDecl) Emit(w io.Writer) {
	for _, decl := range d {
		decl.Emit(w)
	}
}
