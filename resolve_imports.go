package wesl

import (
	"strings"

	"github.com/bluescreen10/wesl-go/ast"
)

func ResolveImports(f *ast.File, files map[string]*ast.File, defines map[string]bool) *ast.File {
	hasImports := false
	for _, d := range f.Decls {
		if _, ok := d.(*ast.ImportDecl); ok {
			hasImports = true
			break
		}
	}
	if !hasImports {
		return f
	}

	resolver := &importResolver{
		files: files,
	}

	importedDecls := resolver.resolveImports(f)

	var nonImports []ast.Decl
	for _, d := range f.Decls {
		if _, ok := d.(*ast.ImportDecl); !ok {
			nonImports = append(nonImports, d)
		}
	}

	return &ast.File{
		Decls: append(nonImports, importedDecls...),
	}
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

	for _, d := range f.Decls {
		imp, ok := d.(*ast.ImportDecl)
		if !ok {
			continue
		}
		for _, item := range imp.Items {
			symbol := item.Path[len(item.Path)-1]
			alias := item.Alias
			if alias == "" {
				alias = symbol
			}

			// Build the path string for file lookup: prefix + item sub-path (excluding terminal name)
			subPath := item.Path[:len(item.Path)-1]
			fullPrefix := append(append([]string{}, imp.Path...), subPath...)
			pathStr := strings.Join(fullPrefix, "::") + "::"

			targetFile := r.findFile(pathStr)
			if targetFile == nil {
				continue
			}

			extracted := r.extractExport(targetFile, symbol)
			for _, d := range extracted {
				decls = r.addDeclWithRename(decls, d, symbol, alias)
			}
		}
	}

	return decls
}

func (r *importResolver) addDeclWithRename(decls []ast.Decl, d ast.Decl, symbol, alias string) []ast.Decl {
	switch dd := d.(type) {
	case *ast.FuncDecl:
		newDecl := &ast.FuncDecl{
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

func (r *importResolver) findFile(path string) *ast.File {
	cleanImpPath := r.getImportPath(path)

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
	case *ast.FuncDecl:
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
