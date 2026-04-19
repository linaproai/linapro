// This file guards the repository rule that every Service interface method in
// internal/service must carry an explanatory comment for long-term maintenance.

package config

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestAllServiceInterfaceMethodsHaveComments verifies every Service interface
// method under internal/service keeps an adjacent explanatory comment.
func TestAllServiceInterfaceMethodsHaveComments(t *testing.T) {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	serviceRoot := filepath.Clean(filepath.Join(workingDir, ".."))
	fileSet := token.NewFileSet()
	var missing []string

	err = filepath.WalkDir(serviceRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fileNode, parseErr := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
		if parseErr != nil {
			return parseErr
		}

		for _, decl := range fileNode.Decls {
			generalDecl, ok := decl.(*ast.GenDecl)
			if !ok || generalDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range generalDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name == nil || typeSpec.Name.Name != "Service" {
					continue
				}
				interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}
				for _, field := range interfaceType.Methods.List {
					if len(field.Names) == 0 {
						continue
					}
					if field.Doc != nil && len(field.Doc.List) > 0 {
						continue
					}
					position := fileSet.Position(field.Pos())
					for _, name := range field.Names {
						missing = append(
							missing,
							position.Filename+":"+strconv.Itoa(position.Line)+":"+name.Name,
						)
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan service interfaces: %v", err)
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("service interface methods missing comments:\n%s", strings.Join(missing, "\n"))
	}
}
