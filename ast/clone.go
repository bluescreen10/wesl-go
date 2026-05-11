package ast

import "slices"

// Clone returns a deep copy of the DiagnosticDirective, or nil if n is nil.
func (n *DiagnosticDirective) Clone() *DiagnosticDirective {
	if n == nil {
		return nil
	}
	return &DiagnosticDirective{
		Attrs:   CloneList(n.Attrs),
		Control: n.Control,
	}
}

// Clone returns a deep copy of the EnableDirective, or nil if n is nil.
func (n *EnableDirective) Clone() *EnableDirective {
	if n == nil {
		return nil
	}
	return &EnableDirective{
		Attrs:      CloneList(n.Attrs),
		Extensions: slices.Clone(n.Extensions),
	}
}

// Clone returns a deep copy of the FuncDecl, or nil if n is nil.
func (n *FuncDecl) Clone() *FuncDecl {
	if n == nil {
		return nil
	}
	return &FuncDecl{
		Attrs:       CloneList(n.Attrs),
		Name:        n.Name.Clone(),
		Params:      CloneListFunc(n.Params, CloneParam),
		ReturnAttrs: CloneList(n.ReturnAttrs),
		ReturnType:  n.ReturnType.Clone(),
		Body:        n.Body.Clone(),
	}
}

// Clone returns a deep copy of the ImportDecl, or nil if n is nil.
func (n *ImportDecl) Clone() *ImportDecl {
	if n == nil {
		return nil
	}
	return &ImportDecl{
		Imports: slices.Clone(n.Imports),
	}
}

// Clone returns a deep copy of the IfAttrDecl, or nil if n is nil.
func (n *IfAttrDecl) Clone() *IfAttrDecl {
	if n == nil {
		return nil
	}
	return &IfAttrDecl{
		Cond: CloneExpr(n.Cond),
		Then: CloneDecl(n.Then),
		Else: CloneDecl(n.Else),
	}
}

// Clone returns a deep copy of the RequiresDirective, or nil if n is nil.
func (n *RequiresDirective) Clone() *RequiresDirective {
	if n == nil {
		return nil
	}
	return &RequiresDirective{
		Attrs:      CloneList(n.Attrs),
		Extensions: slices.Clone(n.Extensions),
	}
}

// Clone returns a deep copy of the StructDecl, or nil if n is nil.
func (n *StructDecl) Clone() *StructDecl {
	if n == nil {
		return nil
	}
	return &StructDecl{
		Attrs:   CloneList(n.Attrs),
		Name:    n.Name.Clone(),
		Members: CloneListFunc(n.Members, CloneMember),
	}
}

// Clone returns a deep copy of the TypeAliasDecl, or nil if n is nil.
func (n *TypeAliasDecl) Clone() *TypeAliasDecl {
	if n == nil {
		return nil
	}
	return &TypeAliasDecl{
		Attrs: CloneList(n.Attrs),
		Name:  n.Name.Clone(),
		Type:  n.Type.Clone(),
	}
}

// Clone returns a deep copy of the AssignmentStmt, or nil if n is nil.
func (n *AssignmentStmt) Clone() *AssignmentStmt {
	if n == nil {
		return nil
	}
	return &AssignmentStmt{
		Attrs: CloneList(n.Attrs),
		LHS:   CloneExpr(n.LHS),
		RHS:   CloneExpr(n.RHS),
		Op:    n.Op,
	}
}

// Clone returns a deep copy of the BreakStmt, or nil if n is nil.
func (n *BreakStmt) Clone() *BreakStmt {
	if n == nil {
		return nil
	}
	return &BreakStmt{
		Attrs: CloneList(n.Attrs),
		Cond:  CloneExpr(n.Cond),
	}
}

// Clone returns a deep copy of the BlockStmt, or nil if n is nil.
func (n *BlockStmt) Clone() *BlockStmt {
	if n == nil {
		return nil
	}
	return &BlockStmt{
		Attrs: CloneList(n.Attrs),
		Stmts: CloneListFunc(n.Stmts, CloneStmt),
	}
}

// Clone returns a deep copy of the ConstAssertStmt, or nil if n is nil.
func (n *ConstAssertStmt) Clone() *ConstAssertStmt {
	if n == nil {
		return nil
	}
	return &ConstAssertStmt{
		Attrs: CloneList(n.Attrs),
		Expr:  CloneExpr(n.Expr),
	}
}

// Clone returns a deep copy of the ContinueStmt, or nil if n is nil.
func (n *ContinueStmt) Clone() *ContinueStmt {
	if n == nil {
		return nil
	}
	return &ContinueStmt{
		Attrs: CloneList(n.Attrs),
	}
}

// Clone returns a deep copy of the ContinuingStmt, or nil if n is nil.
func (n *ContinuingStmt) Clone() *ContinuingStmt {
	if n == nil {
		return nil
	}
	return &ContinuingStmt{
		Attrs: CloneList(n.Attrs),
		Body:  n.Body.Clone(),
	}
}

// Clone returns a deep copy of the DiscardStmt, or nil if n is nil.
func (n *DiscardStmt) Clone() *DiscardStmt {
	if n == nil {
		return nil
	}
	return &DiscardStmt{
		Attrs: CloneList(n.Attrs),
	}
}

// Clone returns a deep copy of the EmptyStmt, or nil if n is nil.
func (n *EmptyStmt) Clone() *EmptyStmt {
	if n == nil {
		return nil
	}
	return &EmptyStmt{}
}

// Clone returns a deep copy of the ForStmt, or nil if n is nil.
func (n *ForStmt) Clone() *ForStmt {
	if n == nil {
		return nil
	}
	return &ForStmt{
		Attrs:  CloneList(n.Attrs),
		Init:   CloneStmt(n.Init),
		Cond:   CloneExpr(n.Cond),
		Update: CloneStmt(n.Update),
		Body:   n.Body.Clone(),
	}
}

// Clone returns a deep copy of the FuncCallStmt, or nil if n is nil.
func (n *FuncCallStmt) Clone() *FuncCallStmt {
	if n == nil {
		return nil
	}
	return &FuncCallStmt{
		Attrs: CloneList(n.Attrs),
		Call:  n.Call.Clone(),
	}
}

// Clone returns a deep copy of the IfStmt, or nil if n is nil.
func (n *IfStmt) Clone() *IfStmt {
	if n == nil {
		return nil
	}
	return &IfStmt{
		Attrs:  CloneList(n.Attrs),
		Cond:   CloneExpr(n.Cond),
		Then:   n.Then.Clone(),
		ElseIf: n.ElseIf.Clone(),
		Else:   n.Else.Clone(),
	}
}

// Clone returns a deep copy of the IfAttrStmt, or nil if n is nil.
func (n *IfAttrStmt) Clone() *IfAttrStmt {
	if n == nil {
		return nil
	}
	return &IfAttrStmt{
		Cond: CloneExpr(n.Cond),
		Then: CloneStmt(n.Then),
		Else: CloneStmt(n.Else),
	}
}

// Clone returns a deep copy of the IncDecStmt, or nil if n is nil.
func (n *IncDecStmt) Clone() *IncDecStmt {
	if n == nil {
		return nil
	}
	return &IncDecStmt{
		Attrs: CloneList(n.Attrs),
		LHS:   CloneExpr(n.LHS),
		Op:    n.Op,
	}
}

// Clone returns a deep copy of the LoopStmt, or nil if n is nil.
func (n *LoopStmt) Clone() *LoopStmt {
	if n == nil {
		return nil
	}
	return &LoopStmt{
		Attrs:     CloneList(n.Attrs),
		BodyAttrs: CloneList(n.BodyAttrs),
		Body:      n.Body.Clone(),
	}
}

// Clone returns a deep copy of the ReturnStmt, or nil if n is nil.
func (n *ReturnStmt) Clone() *ReturnStmt {
	if n == nil {
		return nil
	}
	return &ReturnStmt{
		Attrs: CloneList(n.Attrs),
		Value: CloneExpr(n.Value),
	}
}

// Clone returns a deep copy of the SwitchStmt, or nil if n is nil.
func (n *SwitchStmt) Clone() *SwitchStmt {
	if n == nil {
		return nil
	}
	return &SwitchStmt{
		Attrs:   CloneList(n.Attrs),
		Expr:    CloneExpr(n.Expr),
		Clauses: CloneListFunc(n.Clauses, CloneClause),
	}
}

// Clone returns a deep copy of the VarStmt, or nil if n is nil.
func (n *VarStmt) Clone() *VarStmt {
	if n == nil {
		return nil
	}
	return &VarStmt{
		Attrs:        CloneList(n.Attrs),
		TemplateArgs: CloneListFunc(n.TemplateArgs, CloneExpr),
		Name:         n.Name.Clone(),
		Type:         n.Type.Clone(),
		Init:         CloneExpr(n.Init),
	}
}

// Clone returns a deep copy of the ValStmt, or nil if n is nil.
func (n *ValStmt) Clone() *ValStmt {
	if n == nil {
		return nil
	}
	return &ValStmt{
		Attrs:   CloneList(n.Attrs),
		Keyword: n.Keyword,
		Name:    n.Name.Clone(),
		Type:    n.Type.Clone(),
		Init:    CloneExpr(n.Init),
	}
}

// Clone returns a deep copy of the WhileStmt, or nil if n is nil.
func (n *WhileStmt) Clone() *WhileStmt {
	if n == nil {
		return nil
	}
	return &WhileStmt{
		Attrs: CloneList(n.Attrs),
		Cond:  CloneExpr(n.Cond),
		Body:  n.Body.Clone(),
	}
}

// Clone returns a deep copy of the AddrOfExpr, or nil if n is nil.
func (n *AddrOfExpr) Clone() *AddrOfExpr {
	if n == nil {
		return nil
	}
	return &AddrOfExpr{
		Operand: CloneExpr(n.Operand),
	}
}

// Clone returns a deep copy of the BinaryExpr, or nil if n is nil.
func (n *BinaryExpr) Clone() *BinaryExpr {
	if n == nil {
		return nil
	}
	return &BinaryExpr{
		Op:    n.Op,
		Left:  CloneExpr(n.Left),
		Right: CloneExpr(n.Right),
	}
}

// Clone returns a deep copy of the CallExpr, or nil if n is nil.
func (n *CallExpr) Clone() *CallExpr {
	if n == nil {
		return nil
	}
	return &CallExpr{
		Callee:       n.Callee.Clone(),
		TemplateArgs: CloneListFunc(n.TemplateArgs, CloneExpr),
		Args:         CloneListFunc(n.Args, CloneExpr),
	}
}

// Clone returns a deep copy of the DerefExpr, or nil if n is nil.
func (n *DerefExpr) Clone() *DerefExpr {
	if n == nil {
		return nil
	}
	return &DerefExpr{
		Operand: CloneExpr(n.Operand),
	}
}

// Clone returns a deep copy of the Ident, or nil if n is nil.
func (n *Ident) Clone() *Ident {
	if n == nil {
		return nil
	}
	return &Ident{Path: slices.Clone(n.Path), Val: n.Val}
}

// Clone returns a deep copy of the IndexExpr, or nil if n is nil.
func (n *IndexExpr) Clone() *IndexExpr {
	if n == nil {
		return nil
	}
	return &IndexExpr{
		Base:  CloneExpr(n.Base),
		Index: CloneExpr(n.Index),
	}
}

// Clone returns a deep copy of the LitExpr, or nil if n is nil.
func (n *LitExpr) Clone() *LitExpr {
	if n == nil {
		return nil
	}
	return &LitExpr{Val: n.Val}
}

// Clone returns a deep copy of the MemberExpr, or nil if n is nil.
func (n *MemberExpr) Clone() *MemberExpr {
	if n == nil {
		return nil
	}
	return &MemberExpr{
		Base:   CloneExpr(n.Base),
		Member: n.Member,
	}
}

// Clone returns a deep copy of the ParenExpr, or nil if n is nil.
func (n *ParenExpr) Clone() *ParenExpr {
	if n == nil {
		return nil
	}
	return &ParenExpr{
		Inner: CloneExpr(n.Inner),
	}
}

// Clone returns a deep copy of the UnaryExpr, or nil if n is nil.
func (n *UnaryExpr) Clone() *UnaryExpr {
	if n == nil {
		return nil
	}
	return &UnaryExpr{
		Op:      n.Op,
		Operand: CloneExpr(n.Operand),
	}
}

// Clone returns a deep copy of the Attribute, or nil if n is nil.
func (n *Attribute) Clone() *Attribute {
	if n == nil {
		return nil
	}
	return &Attribute{
		Name: n.Name,
		Args: CloneListFunc(n.Args, CloneExpr),
	}
}

// Clone returns a deep copy of the TypeSpecifier, or nil if n is nil.
func (n *TypeSpecifier) Clone() *TypeSpecifier {
	if n == nil {
		return nil
	}
	return &TypeSpecifier{
		Name:         n.Name.Clone(),
		TemplateArgs: CloneListFunc(n.TemplateArgs, CloneExpr),
	}
}

// Clone returns a deep copy of the File, or nil if n is nil.
func (n *File) Clone() *File {
	if n == nil {
		return nil
	}
	return &File{
		Decls: CloneListFunc(n.Decls, CloneDecl),
	}
}

// Clone returns a deep copy of the IfAttrParam, or nil if n is nil.
func (n *IfAttrParam) Clone() *IfAttrParam {
	if n == nil {
		return nil
	}
	return &IfAttrParam{
		Cond: CloneExpr(n.Cond),
		Then: CloneParam(n.Then),
		Else: CloneParam(n.Else),
	}
}

// Clone returns a deep copy of the FuncParam, or nil if n is nil.
func (n *FuncParam) Clone() *FuncParam {
	if n == nil {
		return nil
	}
	return &FuncParam{
		Attrs: CloneList(n.Attrs),
		Name:  n.Name,
		Type:  n.Type.Clone(),
	}
}

// Clone returns a deep copy of the IfAttrStructMember, or nil if n is nil.
func (n *IfAttrStructMember) Clone() *IfAttrStructMember {
	if n == nil {
		return nil
	}
	return &IfAttrStructMember{
		Cond: CloneExpr(n.Cond),
		Then: CloneMember(n.Then),
		Else: CloneMember(n.Else),
	}
}

// Clone returns a deep copy of the StructMember, or nil if n is nil.
func (n *StructMember) Clone() *StructMember {
	if n == nil {
		return nil
	}
	return &StructMember{
		Attrs: CloneList(n.Attrs),
		Name:  n.Name,
		Type:  n.Type.Clone(),
	}
}

// Clone returns a deep copy of the IfAttrClause, or nil if n is nil.
func (n *IfAttrClause) Clone() *IfAttrClause {
	if n == nil {
		return nil
	}
	return &IfAttrClause{
		Cond: CloneExpr(n.Cond),
		Then: CloneClause(n.Then),
		Else: CloneClause(n.Else),
	}
}

// Clone returns a deep copy of the CaseClause, or nil if n is nil.
func (n *CaseClause) Clone() *CaseClause {
	if n == nil {
		return nil
	}
	return &CaseClause{
		Attrs:     CloneList(n.Attrs),
		Selectors: CloneListFunc(n.Selectors, CloneExpr),
		Body:      n.Body.Clone(),
	}
}

// Cloner is a type constraint satisfied by any type T that provides a Clone
// method returning a fresh T.
type Cloner[T any] interface {
	Clone() T
}

// CloneDecl returns a deep copy of the given Decl by dispatching to the
// concrete Clone method for each known declaration type. Unknown types are
// returned as-is.
func CloneDecl(item Decl) Decl {
	if item == nil {
		return nil
	}
	switch item := item.(type) {
	case *DiagnosticDirective:
		return item.Clone()
	case *EnableDirective:
		return item.Clone()
	case *FuncDecl:
		return item.Clone()
	case *ImportDecl:
		return item.Clone()
	case *IfAttrDecl:
		return item.Clone()
	case *RequiresDirective:
		return item.Clone()
	case *StructDecl:
		return item.Clone()
	case *TypeAliasDecl:
		return item.Clone()
	case *ConstAssertStmt:
		return item.Clone()
	case *VarStmt:
		return item.Clone()
	case *ValStmt:
		return item.Clone()
	default:
		return item
	}
}

// CloneStmt returns a deep copy of the given Stmt by dispatching to the
// concrete Clone method for each known statement type. Unknown types are
// returned as-is.
func CloneStmt(item Stmt) Stmt {
	if item == nil {
		return nil
	}
	switch item := item.(type) {
	case *AssignmentStmt:
		return item.Clone()
	case *BreakStmt:
		return item.Clone()
	case *BlockStmt:
		return item.Clone()
	case *ConstAssertStmt:
		return item.Clone()
	case *ContinueStmt:
		return item.Clone()
	case *ContinuingStmt:
		return item.Clone()
	case *DiscardStmt:
		return item.Clone()
	case *EmptyStmt:
		return item.Clone()
	case *ForStmt:
		return item.Clone()
	case *FuncCallStmt:
		return item.Clone()
	case *IfStmt:
		return item.Clone()
	case *IfAttrStmt:
		return item.Clone()
	case *IncDecStmt:
		return item.Clone()
	case *LoopStmt:
		return item.Clone()
	case *ReturnStmt:
		return item.Clone()
	case *SwitchStmt:
		return item.Clone()
	case *VarStmt:
		return item.Clone()
	case *ValStmt:
		return item.Clone()
	case *WhileStmt:
		return item.Clone()
	default:
		return item
	}
}

// CloneExpr returns a deep copy of the given Expr by dispatching to the
// concrete Clone method for each known expression type. Unknown types are
// returned as-is.
func CloneExpr(item Expr) Expr {
	if item == nil {
		return nil
	}
	switch item := item.(type) {
	case *AddrOfExpr:
		return item.Clone()
	case *BinaryExpr:
		return item.Clone()
	case *CallExpr:
		return item.Clone()
	case *DerefExpr:
		return item.Clone()
	case *Ident:
		return item.Clone()
	case *IndexExpr:
		return item.Clone()
	case *LitExpr:
		return item.Clone()
	case *MemberExpr:
		return item.Clone()
	case *ParenExpr:
		return item.Clone()
	case *UnaryExpr:
		return item.Clone()
	default:
		return item
	}
}

// CloneParam returns a deep copy of the given Param by dispatching to the
// concrete Clone method for each known parameter type. Unknown types are
// returned as-is.
func CloneParam(item Param) Param {
	if item == nil {
		return nil
	}
	switch item := item.(type) {
	case *IfAttrParam:
		return item.Clone()
	case *FuncParam:
		return item.Clone()
	default:
		return item
	}
}

// CloneMember returns a deep copy of the given Member by dispatching to the
// concrete Clone method for each known member type. Unknown types are returned
// as-is.
func CloneMember(item Member) Member {
	if item == nil {
		return nil
	}
	switch item := item.(type) {
	case *IfAttrStructMember:
		return item.Clone()
	case *StructMember:
		return item.Clone()
	default:
		return item
	}
}

// CloneClause returns a deep copy of the given Clause by dispatching to the
// concrete Clone method for each known clause type. Unknown types are returned
// as-is.
func CloneClause(item Clause) Clause {
	if item == nil {
		return nil
	}
	switch item := item.(type) {
	case *IfAttrClause:
		return item.Clone()
	case *CaseClause:
		return item.Clone()
	default:
		return item
	}
}

// CloneList returns a new slice containing deep copies of every element in
// items. Each element must satisfy the Cloner constraint so its Clone method
// can be called directly.
func CloneList[T Cloner[T]](items []T) []T {
	out := make([]T, 0, len(items))
	for _, item := range items {
		out = append(out, item.Clone())
	}
	return out
}

// CloneListFunc returns a new slice built by applying fn to every element of
// items. fn is typically one of the CloneDecl/CloneStmt/CloneExpr helpers or
// an inline closure. A nil items slice is returned as nil, preserving the
// nil-vs-empty distinction used by some AST nodes (e.g. CaseClause.Selectors
// where nil signals the default clause).
func CloneListFunc[T any](items []T, fn func(T) T) []T {
	if items == nil {
		return nil
	}
	out := make([]T, 0, len(items))
	for _, item := range items {
		out = append(out, fn(item))
	}
	return out
}
