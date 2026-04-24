package wesl

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/bluescreen10/wesl-go/ast"
	"github.com/bluescreen10/wesl-go/parser"
	"github.com/bluescreen10/wesl-go/printer"
)

type Compiler struct {
	files map[string]*ast.File
	mu    sync.Mutex
}

func New() *Compiler {
	w := &Compiler{}
	w.init()
	return w
}

func (c *Compiler) init() {
	c.files = make(map[string]*ast.File)
}

func (c *Compiler) Parse(name, src string) error {
	f, err := parser.Parse(src)
	if err != nil {
		return fmt.Errorf("error parsing %s: %v", name, err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.files[name] = f
	return nil
}

func (c *Compiler) ParseFile(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", path, err)
	}

	return c.Parse(path, string(src))
}

func (c *Compiler) Compile(file string, defines map[string]bool) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	parsedFile, exists := c.files[file]
	if !exists {
		return "", fmt.Errorf("error fetching parsed ast for file %s", file)
	}

	resolved := ResolveFile(parsedFile, defines)

	if c != nil && c.files != nil {
		resolved = ResolveImports(resolved, c.files, defines)
	}
	var buf bytes.Buffer
	printer.Fprint(&buf, resolved)
	return buf.String(), nil
}
