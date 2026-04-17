package ast

type ImportDecl struct {
	Symbol string
	Alias  string
	Path   string
}

func (i *ImportDecl) String() string {
	return i.Path + i.Symbol + " as " + i.Alias
}
