package ast

import "slices"

func (n *ConstAssertDecl) Clone() *ConstAssertDecl {
	if n == nil {
		return nil
	}
	return &ConstAssertDecl{Assert: n.Assert.Clone()}
}

func (n *DiagnosticDirective) Clone() *DiagnosticDirective {
	if n == nil {
		return nil
	}
	return &DiagnosticDirective{
		Attrs:   CloneList(n.Attrs),
		Control: n.Control,
	}
}

func (n *EnableDirective) Clone() *EnableDirective {
	if n == nil {
		return nil
	}
	return &EnableDirective{
		Attrs:      CloneList(n.Attrs),
		Extensions: slices.Clone(n.Extensions),
	}
}

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

func (n *GlobalValDecl) Clone() *GlobalValDecl {
	if n == nil {
		return nil
	}
	return &GlobalValDecl{
		Attrs:   CloneList(n.Attrs),
		Name:    n.Name.Clone(),
		Keyword: n.Keyword,
		Type:    n.Type.Clone(),
		Init:    CloneExpr(n.Init),
	}
}

func (n *GlobalVarDecl) Clone() *GlobalVarDecl {
	if n == nil {
		return nil
	}
	return &GlobalVarDecl{
		Attrs:        CloneList(n.Attrs),
		Name:         n.Name.Clone(),
		TemplateArgs: CloneListFunc(n.TemplateArgs, CloneExpr),
		Type:         n.Type.Clone(),
		Init:         CloneExpr(n.Init),
	}
}

func (n *ImportDecl) Clone() *ImportDecl {
	if n == nil {
		return nil
	}
	return &ImportDecl{
		Imports: slices.Clone(n.Imports),
	}
}

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

func (n *RequiresDirective) Clone() *RequiresDirective {
	if n == nil {
		return nil
	}
	return &RequiresDirective{
		Attrs:      CloneList(n.Attrs),
		Extensions: slices.Clone(n.Extensions),
	}
}

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

func (n *BreakStmt) Clone() *BreakStmt {
	if n == nil {
		return nil
	}
	return &BreakStmt{
		Attrs: CloneList(n.Attrs),
	}
}

func (n *BreakIfStmt) Clone() *BreakIfStmt {
	if n == nil {
		return nil
	}
	return &BreakIfStmt{
		Attrs: CloneList(n.Attrs),
		Cond:  CloneExpr(n.Cond),
	}
}

func (n *BlockStmt) Clone() *BlockStmt {
	if n == nil {
		return nil
	}
	return &BlockStmt{
		Attrs: CloneList(n.Attrs),
		Stmts: CloneListFunc(n.Stmts, CloneStmt),
	}
}

func (n *ConstAssertStmt) Clone() *ConstAssertStmt {
	if n == nil {
		return nil
	}
	return &ConstAssertStmt{
		Attrs: CloneList(n.Attrs),
		Expr:  CloneExpr(n.Expr),
	}
}

func (n *ContinueStmt) Clone() *ContinueStmt {
	if n == nil {
		return nil
	}
	return &ContinueStmt{
		Attrs: CloneList(n.Attrs),
	}
}

func (n *ContinuingStmt) Clone() *ContinuingStmt {
	if n == nil {
		return nil
	}
	return &ContinuingStmt{
		Attrs: CloneList(n.Attrs),
		Body:  n.Body.Clone(),
	}
}

func (n *DiscardStmt) Clone() *DiscardStmt {
	if n == nil {
		return nil
	}
	return &DiscardStmt{
		Attrs: CloneList(n.Attrs),
	}
}

func (n *EmptyStmt) Clone() *EmptyStmt {
	if n == nil {
		return nil
	}
	return &EmptyStmt{}
}

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

func (n *FuncCallStmt) Clone() *FuncCallStmt {
	if n == nil {
		return nil
	}
	return &FuncCallStmt{
		Attrs: CloneList(n.Attrs),
		Call:  n.Call.Clone(),
	}
}

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

func (n *ReturnStmt) Clone() *ReturnStmt {
	if n == nil {
		return nil
	}
	return &ReturnStmt{
		Attrs: CloneList(n.Attrs),
		Value: CloneExpr(n.Value),
	}
}

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

func (n *AddrOfExpr) Clone() *AddrOfExpr {
	if n == nil {
		return nil
	}
	return &AddrOfExpr{
		Operand: CloneExpr(n.Operand),
	}
}

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

func (n *DerefExpr) Clone() *DerefExpr {
	if n == nil {
		return nil
	}
	return &DerefExpr{
		Operand: CloneExpr(n.Operand),
	}
}

func (n *Ident) Clone() *Ident {
	if n == nil {
		return nil
	}
	return &Ident{Val: n.Val}
}

func (n *IndexExpr) Clone() *IndexExpr {
	if n == nil {
		return nil
	}
	return &IndexExpr{
		Base:  CloneExpr(n.Base),
		Index: CloneExpr(n.Index),
	}
}

func (n *LitExpr) Clone() *LitExpr {
	if n == nil {
		return nil
	}
	return &LitExpr{Val: n.Val}
}

func (n *MemberExpr) Clone() *MemberExpr {
	if n == nil {
		return nil
	}
	return &MemberExpr{
		Base:   CloneExpr(n.Base),
		Member: n.Member,
	}
}

func (n *ParenExpr) Clone() *ParenExpr {
	if n == nil {
		return nil
	}
	return &ParenExpr{
		Inner: CloneExpr(n.Inner),
	}
}

func (n *UnaryExpr) Clone() *UnaryExpr {
	if n == nil {
		return nil
	}
	return &UnaryExpr{
		Op:      n.Op,
		Operand: CloneExpr(n.Operand),
	}
}

func (n *Attribute) Clone() *Attribute {
	if n == nil {
		return nil
	}
	return &Attribute{
		Name: n.Name,
		Args: CloneListFunc(n.Args, CloneExpr),
	}
}

func (n *TypeSpecifier) Clone() *TypeSpecifier {
	if n == nil {
		return nil
	}
	return &TypeSpecifier{
		Name:         n.Name,
		TemplateArgs: CloneListFunc(n.TemplateArgs, CloneExpr),
	}
}

func (n *File) Clone() *File {
	if n == nil {
		return nil
	}
	return &File{
		Decls: CloneListFunc(n.Decls, CloneDecl),
	}
}

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

func (n *CaseClause) Clone() *CaseClause {
	if n == nil {
		return nil
	}
	return &CaseClause{
		Attrs:     CloneList(n.Attrs),
		Selectors: slices.Clone(n.Selectors),
		Body:      n.Body.Clone(),
	}
}

type Cloner[T any] interface {
	Clone() T
}

func CloneDecl(item Decl) Decl {
	if item == nil {
		return nil
	}
	switch item := item.(type) {
	case *ConstAssertDecl:
		return item.Clone()
	case *DiagnosticDirective:
		return item.Clone()
	case *EnableDirective:
		return item.Clone()
	case *FuncDecl:
		return item.Clone()
	case *GlobalValDecl:
		return item.Clone()
	case *GlobalVarDecl:
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
	default:
		return item
	}
}

func CloneStmt(item Stmt) Stmt {
	if item == nil {
		return nil
	}
	switch item := item.(type) {
	case *AssignmentStmt:
		return item.Clone()
	case *BreakStmt:
		return item.Clone()
	case *BreakIfStmt:
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

func CloneList[T Cloner[T]](items []T) []T {
	out := make([]T, 0, len(items))
	for _, item := range items {
		out = append(out, item.Clone())
	}
	return out
}

func CloneListFunc[T any](items []T, fn func(T) T) []T {
	out := make([]T, 0, len(items))
	for _, item := range items {
		out = append(out, fn(item))
	}
	return out
}
