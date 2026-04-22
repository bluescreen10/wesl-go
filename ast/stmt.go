package ast

import "io"

type Stmt interface {
	Emit(w io.Writer)
}

type CompoundStmt struct {
	Attrs []Attribute
	Stmts []Stmt
}

func (c CompoundStmt) Emit(w io.Writer) {
	for _, a := range c.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}

	//FIXME: Hack
	w.Write([]byte{'{'})
	//if len(c.Stmts) > 0 {
	w.Write([]byte{' '})
	//}

	var isCompoundStmt bool
	for i, s := range c.Stmts {
		if i > 0 {
			if !isCompoundStmt {
				w.Write([]byte{';'})
			}
			w.Write([]byte{' '})
		}

		switch s.(type) {
		case *CompoundStmt, *IfStmt, *WhileStmt, *ForStmt, *LoopStmt, *SwitchStmt, *ContinuingStmt:
			isCompoundStmt = true
		default:
			isCompoundStmt = false
		}
		s.Emit(w)

	}

	//FIXME: Hack
	if len(c.Stmts) > 0 {
		if !isCompoundStmt {
			w.Write([]byte{';'})
		}
		w.Write([]byte{' '})
	}
	w.Write([]byte{'}'})
}

type EmptyStmt struct{}

func (f EmptyStmt) Emit(w io.Writer) {

}

type ReturnStmt struct {
	Attrs []Attribute
	Value Expr
}

func (r ReturnStmt) Emit(w io.Writer) {
	for _, a := range r.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte{'r', 'e', 't', 'u', 'r', 'n'})
	if r.Value != nil {
		w.Write([]byte{' '})
		r.Value.Emit(w)
	}
}

type BreakIfStmt struct {
	Attrs []Attribute
	Cond  Expr
}

func (b BreakIfStmt) Emit(w io.Writer) {
	for _, a := range b.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte{'b', 'r', 'e', 'a', 'k', ' ', 'i', 'f', ' '})
	b.Cond.Emit(w)
}

type BreakStmt struct {
	Attrs []Attribute
}

func (b BreakStmt) Emit(w io.Writer) {
	for _, a := range b.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte{'b', 'r', 'e', 'a', 'k'})
}

type ContinueStmt struct {
	Attrs []Attribute
}

func (c ContinueStmt) Emit(w io.Writer) {
	for _, a := range c.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte{'c', 'o', 'n', 't', 'i', 'n', 'u', 'e'})
}

type ContinuingStmt struct {
	Attrs []Attribute
	Body  *CompoundStmt
}

func (c ContinuingStmt) Emit(w io.Writer) {
	for _, a := range c.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte{'c', 'o', 'n', 't', 'i', 'n', 'u', 'i', 'n', 'g', ' '})
	c.Body.Emit(w)
}

type DiscardStmt struct {
	Attrs []Attribute
}

func (d DiscardStmt) Emit(w io.Writer) {
	for _, a := range d.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte{'d', 'i', 's', 'c', 'a', 'r', 'd'})
}

type ConstAssertStmt struct {
	Attrs []Attribute
	Expr  Expr
}

func (c ConstAssertStmt) Emit(w io.Writer) {
	for _, a := range c.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte{'c', 'o', 'n', 's', 't', '_', 'a', 's', 's', 'e', 'r', 't', ' '})
	c.Expr.Emit(w)
}

type AssignmentStmt struct {
	Attrs []Attribute
	LHS   Expr
	RHS   Expr
	Op    string
}

func (a AssignmentStmt) Emit(w io.Writer) {
	for _, attr := range a.Attrs {
		attr.Emit(w)
		w.Write([]byte{' '})
	}
	if a.LHS == nil {
		w.Write([]byte{'_'})
	} else {
		a.LHS.Emit(w)
	}
	w.Write([]byte{' '})
	w.Write([]byte(a.Op))
	w.Write([]byte{' '})
	a.RHS.Emit(w)
}

type FuncCallStmt struct {
	Attrs []Attribute
	Call  CallExpr
}

func (f FuncCallStmt) Emit(w io.Writer) {
	for _, a := range f.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	f.Call.Emit(w)
}

type IncrementStmt struct {
	Attrs []Attribute
	LHS   Expr
}

func (s IncrementStmt) Emit(w io.Writer) {
	for _, a := range s.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	s.LHS.Emit(w)
	w.Write([]byte{'+', '+'})
}

type DecrementStmt struct {
	Attrs []Attribute
	LHS   Expr
}

func (s DecrementStmt) Emit(w io.Writer) {
	for _, a := range s.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	s.LHS.Emit(w)
	w.Write([]byte{'-', '-'})
}

type CaseClause struct {
	Attrs     []Attribute
	Selectors []Expr
	Body      *CompoundStmt
}

func (s CaseClause) Emit(w io.Writer) {
	for _, a := range s.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte{'c', 'a', 's', 'e', ' '})
	count := len(s.Selectors)
	for i, sel := range s.Selectors {
		sel.Emit(w)
		if i != count-1 {
			w.Write([]byte{',', ' '})
		}
	}
	w.Write([]byte{':'})
	s.Body.Emit(w)
}

type DefaultAloneClause struct {
	Attrs []Attribute
	Body  *CompoundStmt
}

func (s DefaultAloneClause) Emit(w io.Writer) {
	for _, a := range s.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte{'d', 'e', 'f', 'a', 'u', 'l', 't', ' '})
	s.Body.Emit(w)
}

type IfAttrClause struct {
	Cond Expr
	Then SwitchClause
	Else SwitchClause
}

func (x IfAttrClause) Emit(w io.Writer) {
	w.Write([]byte{'@', 'i', 'f', '('})
	x.Cond.Emit(w)
	w.Write([]byte{')'})
	x.Then.Emit(w)
	if x.Else != nil {
		w.Write([]byte{'@', 'e', 'l', 's', 'e', ' '})
		x.Else.Emit(w)
	}
}

type IfStmt struct {
	Attrs  []Attribute
	Cond   Expr
	Then   *CompoundStmt
	ElseIf *IfStmt
	Else   *CompoundStmt
}

func (s IfStmt) Emit(w io.Writer) {
	for _, a := range s.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte{'i', 'f', ' '})
	s.Cond.Emit(w)
	w.Write([]byte{' '})
	s.Then.Emit(w)
	if s.ElseIf != nil {
		w.Write([]byte{'e', 'l', 's', 'e', ' '})
		s.ElseIf.Emit(w)
	}
	if s.Else != nil {
		w.Write([]byte{'e', 'l', 's', 'e', ' '})
		s.Else.Emit(w)
	}
}

type IfAttrStmt struct {
	Cond Expr
	Then Stmt
	Else Stmt
}

func (s IfAttrStmt) Emit(w io.Writer) {
	w.Write([]byte{'@', 'i', 'f', '('})
	s.Cond.Emit(w)
	w.Write([]byte{')'})
	s.Then.Emit(w)
	if s.Else != nil {
		w.Write([]byte{'@', 'e', 'l', 's', 'e', ' '})
		s.Else.Emit(w)
	}
}

type SwitchStmt struct {
	Attrs   []Attribute
	Expr    Expr
	Clauses []SwitchClause
}

func (s SwitchStmt) Emit(w io.Writer) {
	for _, a := range s.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte{'s', 'w', 'i', 't', 'c', 'h', ' '})
	s.Expr.Emit(w)
	w.Write([]byte{' ', '{', ' '})
	for _, c := range s.Clauses {
		c.Emit(w)
	}
	w.Write([]byte{' ', '}'})
}

type LoopStmt struct {
	Attrs     []Attribute
	BodyAttrs []Attribute
	Body      *CompoundStmt
	//Continuing *ContinuingStmt
}

func (s LoopStmt) Emit(w io.Writer) {
	for _, a := range s.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte{'l', 'o', 'o', 'p', ' '})
	for _, a := range s.BodyAttrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	s.Body.Emit(w)
	// if s.Continuing != nil {
	// 	s.Continuing.Emit(w)
	// }
}

type ForStmt struct {
	Attrs  []Attribute
	Init   Stmt
	Cond   Expr
	Update Stmt
	Body   *CompoundStmt
}

func (s ForStmt) Emit(w io.Writer) {
	for _, a := range s.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte{'f', 'o', 'r', ' ', '('})
	if s.Init != nil {
		s.Init.Emit(w)
	}
	w.Write([]byte{';'})
	if s.Cond != nil {
		w.Write([]byte{' '})
		s.Cond.Emit(w)
	}
	w.Write([]byte{';'})
	if s.Update != nil {
		w.Write([]byte{' '})
		s.Update.Emit(w)
	}
	w.Write([]byte{')', ' '})
	if s.Body != nil {
		s.Body.Emit(w)
	}
}

type WhileStmt struct {
	Attrs []Attribute
	Cond  Expr
	Body  *CompoundStmt
}

func (s WhileStmt) Emit(w io.Writer) {
	for _, a := range s.Attrs {
		a.Emit(w)
		w.Write([]byte{' '})
	}
	w.Write([]byte{'w', 'h', 'i', 'l', 'e', ' '})
	if s.Cond != nil {
		s.Cond.Emit(w)
	}
	w.Write([]byte{' '})
	if s.Body != nil {
		s.Body.Emit(w)
	}
}
