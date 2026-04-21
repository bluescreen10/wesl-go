package ast

type Decl interface {
	//	Node
	//
	// declMarker()
}

type Node interface {
	//node()
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

type IfAttrDecl struct {
	Cond Expr
	Then Decl
	Else Decl // nil when no @else
}
