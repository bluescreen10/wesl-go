package ast

import "io"

// type Expr interface {
// 	Emit(w io.Writer)
// }

// type BinaryExpr struct {
// 	Op    string
// 	Left  Expr
// 	Right Expr
// }

func (b BinaryExpr) Emit(w io.Writer) {
	b.Left.Emit(w)
	w.Write([]byte{' '})
	w.Write([]byte(b.Op))
	w.Write([]byte{' '})
	b.Right.Emit(w)
}

// type UnaryExpr struct {
// 	Op      string
// 	Operand Expr
// }

func (u UnaryExpr) Emit(w io.Writer) {
	w.Write([]byte(u.Op))
	w.Write([]byte{' '})
	u.Operand.Emit(w)
}

// type DerefExpr struct {
// 	Operand Expr
// }

func (d DerefExpr) Emit(w io.Writer) {
	w.Write([]byte{'*'})
	d.Operand.Emit(w)
}

// type AddrOfExpr struct {
// 	Operand Expr
// }

func (a AddrOfExpr) Emit(w io.Writer) {
	w.Write([]byte{'&'})
	a.Operand.Emit(w)
}

// type MemberExpr struct {
// 	Base   Expr
// 	Member string
// }

func (m MemberExpr) Emit(w io.Writer) {
	m.Base.Emit(w)
	w.Write([]byte{'.'})
	w.Write([]byte(m.Member))
}

// type IndexExpr struct {
// 	Base  Expr
// 	Index Expr
// }

func (i IndexExpr) Emit(w io.Writer) {
	i.Base.Emit(w)
	w.Write([]byte{'['})
	i.Index.Emit(w)
	w.Write([]byte{']'})
}

// type LitExpr struct {
// 	Val string
// }

func (l LitExpr) Emit(w io.Writer) {
	w.Write([]byte(l.Val))
}

// type ParenExpr struct {
// 	Inner Expr
// }

func (p ParenExpr) Emit(w io.Writer) {
	w.Write([]byte{'('})
	p.Inner.Emit(w)
	w.Write([]byte{')'})
}

// type CallExpr struct {
// 	Callee       string
// 	TemplateArgs []Expr
// 	Args         []Expr
// }

func (c CallExpr) Emit(w io.Writer) {
	w.Write([]byte(c.Callee))
	if c.TemplateArgs != nil {
		w.Write([]byte{'<'})
		count := len(c.TemplateArgs)
		for i, arg := range c.TemplateArgs {
			arg.Emit(w)
			if i != count-1 {
				w.Write([]byte{',', ' '})
			}
		}
		w.Write([]byte{'>'})
	}

	w.Write([]byte{'('})
	count := len(c.Args)
	for i, arg := range c.Args {
		arg.Emit(w)
		if i != count-1 {
			w.Write([]byte{',', ' '})
		}
	}
	w.Write([]byte{')'})
}
