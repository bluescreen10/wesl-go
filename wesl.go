// Package wesl implements a compiler for WESL (WebGPU Shading Language
// Extended), a superset of WGSL that adds a module system with cross-file
// imports and compile-time @if conditionals.
//
// The typical workflow is:
//
//  1. Create a [Compiler] with [New].
//  2. Register source files using [Compiler.Parse], [Compiler.ParseFile],
//     [Compiler.ParseFS], or [Compiler.ParseGlob].
//  3. Call [Compiler.Compile] with the entry-point filename and any active
//     feature flags to obtain a resolved, flat WGSL string.
//
// Files are stored under a sanitized key: the file extension is stripped and
// any leading "./" is removed (e.g. "./shaders/main.wesl" → "shaders/main").
// If two registrations produce the same key the last one wins.
package wesl

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bluescreen10/wesl-go/ast"
	"github.com/bluescreen10/wesl-go/parser"
	"github.com/bluescreen10/wesl-go/printer"
	"github.com/bluescreen10/wesl-go/resolver"
)

// Compiler parses and compiles WESL source files into WGSL.
// It is safe for concurrent use; multiple goroutines may call Parse* methods
// simultaneously.
type Compiler struct {
	files map[string]*ast.File
	mu    sync.Mutex
}

// New returns a new, empty Compiler ready to accept source files.
func New() *Compiler {
	return &Compiler{files: make(map[string]*ast.File)}
}

// Parse parses src as WESL source and registers it under name. name is
// sanitized (extension stripped, leading "./" removed) before storage. If a
// file with the same sanitized name was registered before, it is replaced.
func (c *Compiler) Parse(name, src string) error {
	f, err := parser.Parse(src)
	if err != nil {
		return fmt.Errorf("error parsing %s: %v", name, err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.files[c.sanitizeName(name)] = f
	return nil
}

// ParseFile reads the file at path and registers it under its base name
// (directory components are stripped). It is intended for single-file use
// where cross-module imports are not required. For projects with a module
// hierarchy use [Compiler.ParseFS] instead.
func (c *Compiler) ParseFile(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", path, err)
	}
	return c.Parse(filepath.Base(path), string(src))
}

// ParseFS walks fsys recursively and registers every file whose path matches
// at least one of the given patterns. Patterns follow the syntax of
// [path.Match]. If no patterns are supplied every file in the FS is
// registered. Files are keyed by their path relative to the root of fsys.
func (c *Compiler) ParseFS(fsys fs.FS, patterns ...string) error {
	// Validate all patterns up-front before touching any files.
	for _, p := range patterns {
		if _, err := filepath.Match(p, ""); err != nil {
			return fmt.Errorf("invalid pattern %q: %v", p, err)
		}
	}

	return fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if len(patterns) > 0 {
			var matched bool
			for _, pat := range patterns {
				// Patterns without a separator are matched against the
				// filename only so that e.g. "*.wgsl" works at any depth.
				target := p
				if !strings.ContainsRune(pat, '/') {
					target = filepath.Base(p)
				}
				if ok, _ := filepath.Match(pat, target); ok {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}
		src, err := fs.ReadFile(fsys, p)
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", p, err)
		}
		if err := c.Parse(p, string(src)); err != nil {
			return fmt.Errorf("error parsing %s: %v", p, err)
		}
		return nil
	})
}

// ParseGlob reads all files under root that match pattern and registers them
// under paths relative to root. It is equivalent to
// ParseFS(os.DirFS(root), pattern). pattern follows the syntax of
// [path.Match].
func (c *Compiler) ParseGlob(root, pattern string) error {
	return c.ParseFS(os.DirFS(root), pattern)
}

// Compile resolves all imports reachable from the named entry-point file and
// returns the merged WGSL source. filename is sanitized the same way as in
// [Compiler.Parse]. defines is the set of active compile-time feature flags
// used to evaluate @if conditionals; a nil map means no flags are set.
// It returns an error if the entry-point has not been registered or if
// resolution fails.
func (c *Compiler) Compile(filename string, defines map[string]bool) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	sanitizedName := c.sanitizeName(filename)
	if _, exists := c.files[sanitizedName]; !exists {
		return "", fmt.Errorf("error fetching parsed ast for file %s", filename)
	}

	ast, err := resolver.ResolveFile(sanitizedName, c.files, defines)
	if err != nil {
		return "", fmt.Errorf("error resolving file: %v", err)
	}

	var buf bytes.Buffer
	printer.Fprint(&buf, ast)
	return buf.String(), nil
}

// sanitizeName strips the file extension and any leading "./" from filename,
// producing the canonical key used to store and look up files.
func (c *Compiler) sanitizeName(filename string) string {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	name = strings.TrimPrefix(name, "./")
	return name
}
