package wesl

type node interface {
}

type listNode struct {
	children []node
}
