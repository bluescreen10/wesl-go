package parser

type node interface {
}

type listNode struct {
	children []node
}
