package wesl

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

func (c *Compiler) ParseFS(fsys fs.FS, patterns ...string) error {
	var paths []string
	for _, pattern := range patterns {
		matches, err := fs.Glob(fsys, pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern %q: %v", pattern, err)
		}
		paths = append(paths, matches...)
	}

	errs := make(chan error, len(paths))
	for _, path := range paths {
		go func(p string) {
			src, err := fs.ReadFile(fsys, p)
			if err != nil {
				errs <- fmt.Errorf("error reading file %s: %v", p, err)
				return
			}
			f, err := parser.Parse(string(src))
			if err != nil {
				errs <- fmt.Errorf("error parsing %s: %v", p, err)
				return
			}
			c.mu.Lock()
			c.files[p] = f
			c.mu.Unlock()
			errs <- nil
		}(path)
	}

	var errsSlice []error
	for range paths {
		errsSlice = append(errsSlice, <-errs)
	}
	return errors.Join(errsSlice...)
}

func (c *Compiler) ParseGlob(pattern string) error {
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern %q: %v", pattern, err)
	}

	errs := make(chan error, len(paths))
	for _, path := range paths {
		go func(p string) {
			src, err := os.ReadFile(p)
			if err != nil {
				errs <- fmt.Errorf("error reading file %s: %v", p, err)
				return
			}
			f, err := parser.Parse(string(src))
			if err != nil {
				errs <- fmt.Errorf("error parsing %s: %v", p, err)
				return
			}
			c.mu.Lock()
			c.files[p] = f
			c.mu.Unlock()
			errs <- nil
		}(path)
	}

	var errsSlice []error
	for range paths {
		errsSlice = append(errsSlice, <-errs)
	}
	return errors.Join(errsSlice...)
}

func (c *Compiler) Compile(file string, defines map[string]bool) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.files[file]; !exists {
		return "", fmt.Errorf("error fetching parsed ast for file %s", file)
	}

	// Use the new resolver for full resolution
	resolver := NewResolver(c.files, defines)
	resolved := resolver.Resolve(file)

	var buf bytes.Buffer
	printer.Fprint(&buf, resolved)
	return buf.String(), nil
}
