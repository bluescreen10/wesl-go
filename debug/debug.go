package main

import (
	
	"github.com/bluescreen10/wesl-go/parser"
)

func main() {
	src := `const foo = 10; const bar = 10; fn func() { @if(true) let foo = 20; let x = foo; /* foo is shadowed. */ @if(false) let bar = 20; let y = bar; /* bar is not shadowed. */ }`
	parser.DebugParse(src)
}
