package ast

import "io"

type DiagnosticDirective struct {
	Attrs   []Attribute
	Control DiagnosticControl
}

func (d DiagnosticDirective) Emit(w io.Writer) {
	for _, a := range d.Attrs {
		a.Emit(w)
	}
	w.Write([]byte{'d', 'i', 'a', 'g', 'n', 'o', 's', 't', 'i', 'c', '('})
	d.Control.Emit(w)
	w.Write([]byte{')', ';'})
}

type DiagnosticControl struct {
	Severity string
	RuleName string
}

func (dc DiagnosticControl) Emit(w io.Writer) {
	w.Write([]byte(dc.Severity))
	w.Write([]byte{',', ' '})
	w.Write([]byte(dc.RuleName))
}

type EnableDirective struct {
	Attrs      []Attribute
	Extensions []string
}

func (e EnableDirective) Emit(w io.Writer) {
	for _, a := range e.Attrs {
		a.Emit(w)
	}
	w.Write([]byte{'e', 'n', 'a', 'b', 'l', 'e', ' '})
	count := len(e.Extensions)
	for i, ext := range e.Extensions {
		w.Write([]byte(ext))
		if i != count-1 {
			w.Write([]byte{',', ' '})
		}
	}
	w.Write([]byte{';'})
}

type RequiresDirective struct {
	Attrs      []Attribute
	Extensions []string
}

func (r RequiresDirective) Emit(w io.Writer) {
	for _, a := range r.Attrs {
		a.Emit(w)
	}
	w.Write([]byte{'r', 'e', 'q', 'u', 'i', 'r', 'e', 's', ' '})
	count := len(r.Extensions)
	for i, ext := range r.Extensions {
		w.Write([]byte(ext))
		if i != count-1 {
			w.Write([]byte{',', ' '})
		}
	}
	w.Write([]byte{';'})
}
