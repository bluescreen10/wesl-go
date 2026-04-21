package ast

type DiagnosticDirective struct {
	Attrs   []Attribute
	Control DiagnosticControl
}

type DiagnosticControl struct {
	Severity string
	RuleName string
}

type EnableDirective struct {
	Attrs      []Attribute
	Extensions []string
}

type RequiresDirective struct {
	Attrs      []Attribute
	Extensions []string
}
