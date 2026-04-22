package wesl

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/bluescreen10/wesl-go/parser"
)

type common struct {
	chunks   map[string]*Translator
	muChunks sync.Mutex
}

type Translator struct {
	name string
	*common
	modulePath string
}

func New(name string) *Translator {
	w := &Translator{name: name}
	w.init()
	return w
}

func (t *Translator) init() {
	if t.common == nil {
		c := new(common)
		c.chunks = make(map[string]*Translator)
		t.common = c
	}
}

func (t *Translator) Parse(src string) (*Translator, error) {
	_, err := parser.Parse(src)
	return t, err
}

func (t *Translator) New(name string) *Translator {
	return &Translator{name: name, common: t.common}
}

func (t *Translator) Translate(src string, defines map[string]bool) (string, error) {
	if src == "@if(false) fn f() -> u32 { return 1; } @else fn f() -> u32 { return 2; } fn main() -> u32 { return f(); }" {
		fmt.Println("here")
	}

	ast, err := parser.Parse(src)
	if err != nil {
		return "", fmt.Errorf("error parsing source file: %v", err)
	}

	resolved := ResolveFile(ast, defines)
	var buf bytes.Buffer
	resolved.Emit(&buf)
	return buf.String(), nil
}
