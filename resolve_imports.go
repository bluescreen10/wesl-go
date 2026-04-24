package wesl

import (
	"strings"

	"github.com/bluescreen10/wesl-go/ast"
)

func ResolveImports(f *ast.File, files map[string]*ast.File, defines map[string]bool) *ast.File {
	if len(f.Imports) == 0 {
		return f
	}

	resolver := &importResolver{
		files: files,
	}

	importedDecls := resolver.resolveImports(f)

	out := &ast.File{
		Decls:   append(f.Decls, importedDecls...),
		Imports: nil,
	}

	return out
}

type importResolver struct {
	files         map[string]*ast.File
	seenDecls     map[string]bool
	importedNames map[string]string
}

func (r *importResolver) resolveImports(f *ast.File) []ast.Decl {
	r.seenDecls = make(map[string]bool)
	r.importedNames = make(map[string]string)

	var decls []ast.Decl

	for _, imp := range f.Imports {
		targetFile := r.findFile(imp)
		if targetFile == nil {
			continue
		}

		symbol := imp.Symbol
		alias := imp.Alias

		extracted := r.extractExport(targetFile, symbol)
		for _, d := range extracted {
			decls = r.addDeclWithRename(decls, d, symbol, alias)
		}
	}

	return decls
}

func (r *importResolver) addDeclWithRename(decls []ast.Decl, d ast.Decl, symbol, alias string) []ast.Decl {
	switch dd := d.(type) {
	case *ast.FnDecl:
		newDecl := &ast.FnDecl{
			Name:        dd.Name,
			Attrs:       dd.Attrs,
			Params:      dd.Params,
			ReturnAttrs: dd.ReturnAttrs,
			ReturnType:  dd.ReturnType,
			Body:        dd.Body,
		}
		if alias != "" && dd.Name == symbol {
			newDecl.Name = alias
			r.importedNames[symbol] = alias
		}
		decls = append(decls, newDecl)
	case *ast.StructDecl:
		newDecl := &ast.StructDecl{
			Name:    dd.Name,
			Attrs:   dd.Attrs,
			Members: dd.Members,
		}
		if alias != "" && dd.Name == symbol {
			newDecl.Name = alias
		}
		decls = append(decls, newDecl)
	default:
		decls = append(decls, d)
	}
	return decls
}

func (r *importResolver) findFile(imp ast.ImportDecl) *ast.File {
	cleanImpPath := r.getImportPath(imp.Path)

	for filePath, f := range r.files {
		cleanFilePath := strings.TrimPrefix(filePath, "./")
		cleanFilePath = strings.TrimSuffix(cleanFilePath, ".wgsl")

		if cleanFilePath == cleanImpPath {
			return f
		}

		if filePath == "./"+cleanImpPath+".wgsl" {
			return f
		}
	}
	return nil
}

func (r *importResolver) getImportPath(path string) string {
	parts := strings.Split(path, "::")
	if len(parts) >= 2 {
		return parts[1]
	}
	return path
}

func (r *importResolver) extractExport(f *ast.File, symbol string) []ast.Decl {
	var decls []ast.Decl
	for _, d := range f.Decls {
		if r.matchesExport(d, symbol) {
			decls = append(decls, d)
		}
	}
	return decls
}

func (r *importResolver) matchesExport(d ast.Decl, symbol string) bool {
	switch d := d.(type) {
	case *ast.FnDecl:
		return d.Name == symbol
	case *ast.StructDecl:
		return d.Name == symbol
	case *ast.GlobalValueDecl:
		return d.Ident.Name == symbol
	case *ast.GlobalVariableDecl:
		if vd, ok := d.Decl.(*ast.VariableDecl); ok {
			return vd.Ident.Name == symbol
		}
	case *ast.TypeAliasDecl:
		return d.Name == symbol
	}
	return false
}
