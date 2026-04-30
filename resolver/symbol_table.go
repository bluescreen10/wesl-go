package resolver

import (
	"strings"

	"github.com/bluescreen10/wesl-go/ast"
)

// Symbol represents an exported symbol from a file
type Symbol struct {
	File string   // e.g., "./foo.wgsl"
	Name string   // original name, e.g., "foo"
	Decl ast.Decl // the declaration (FuncDecl, StructDecl, etc.)
}

// SymbolTable maps symbol names to their defining symbol
type SymbolTable map[string]Symbol // "foo" → Symbol

// ImplicitImport represents an inline reference like "package::b::foo()"
type ImplicitImport struct {
	Path []string // ["package", "b"]
	Name string   // "foo"
}

// BuildSymbolTable scans a resolved AST and builds a symbol table of exports
func BuildSymbolTable(f *ast.File, filePath string) SymbolTable {
	table := make(SymbolTable)

	for _, d := range f.Decls {
		switch dd := d.(type) {
		case *ast.FuncDecl:
			table[dd.Name] = Symbol{filePath, dd.Name, dd}

		case *ast.StructDecl:
			table[dd.Name] = Symbol{filePath, dd.Name, dd}

		case *ast.GlobalValDecl:
			table[dd.Name] = Symbol{filePath, dd.Name, dd}

		case *ast.GlobalVarDecl:
			table[dd.Name] = Symbol{filePath, dd.Name, dd}

		case *ast.TypeAliasDecl:
			table[dd.Name] = Symbol{filePath, dd.Name, dd}

			// ImportDecl - we track what is needed separately
		}
	}

	return table
}

// BuildFullTable builds symbol tables for ALL files in a map
func BuildFullTable(files map[string]*ast.File) map[string]SymbolTable {
	tables := make(map[string]SymbolTable)

	for path, f := range files {
		tables[path] = BuildSymbolTable(f, path)
	}

	return tables
}

// CollectImplicitImports walks AST and finds all inline references like "package::b::foo()"
func CollectImplicitImports(f *ast.File) []ImplicitImport {
	var imports []ImplicitImport

	// Collect from CallExpr.Callee
	var walk func(e ast.Expr)
	walk = func(e ast.Expr) {
		if e == nil {
			return
		}
		switch ex := e.(type) {
		case *ast.CallExpr:
			if strings.Contains(ex.Callee, "::") {
				parts := strings.Split(ex.Callee, "::")
				if len(parts) >= 2 {
					imports = append(imports, ImplicitImport{
						Path: parts[:len(parts)-1],
						Name: parts[len(parts)-1],
					})
				}
			}
		case *ast.BinaryExpr:
			walk(ex.Left)
			walk(ex.Right)
		case *ast.IndexExpr:
			walk(ex.Base)
			walk(ex.Index)
		case *ast.Ident:
			// Check if this ident is a namespaced reference
		}
	}

	// Walk statements
	var walkStmt func(s ast.Stmt)
	walkStmt = func(s ast.Stmt) {
		if s == nil {
			return
		}
		switch st := s.(type) {
		case *ast.FuncCallStmt:
			walk(&st.Call)
		case *ast.AssignmentStmt:
			walk(st.LHS)
			walk(st.RHS)
		case *ast.ReturnStmt:
			if st.Value != nil {
				walk(st.Value)
			}
		case *ast.WhileStmt:
			walk(st.Cond)
			for _, s := range st.Body.Stmts {
				walkStmt(s)
			}
		case *ast.ForStmt:
			for _, s := range st.Body.Stmts {
				walkStmt(s)
			}
		case *ast.LoopStmt:
			for _, s := range st.Body.Stmts {
				walkStmt(s)
			}
		case *ast.IfStmt:
			walk(st.Cond)
			for _, s := range st.Then.Stmts {
				walkStmt(s)
			}
			if st.Else != nil {
				for _, s := range st.Else.Stmts {
					walkStmt(s)
				}
			}
		}
	}

	// Walk declarations
	for _, d := range f.Decls {
		switch dd := d.(type) {
		case *ast.FuncDecl:
			for _, s := range dd.Body.Stmts {
				walkStmt(s)
			}
		}
	}

	return imports
}
