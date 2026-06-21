// This file verifies package-level config service governance and delivery
// configuration defaults that belong to the component contract.

package config

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
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

// TestDatabaseDebugDefaultsOffInDeliveryConfig verifies startup logs do not
// emit ORM SQL details unless operators explicitly enable database debug.
func TestDatabaseDebugDefaultsOffInDeliveryConfig(t *testing.T) {
	templatePath := filepath.Join("..", "..", "..", "manifest", "config", "config.template.yaml")
	assertDatabaseDebugDisabled(t, templatePath)
	packedTemplatePath := filepath.Join("..", "..", "..", "internal", "packed", "manifest", "config", "config.template.yaml")
	assertDatabaseDebugDisabledIfExists(t, packedTemplatePath)

	// Local config.yaml is intentionally git-ignored, but when present it should
	// follow the same default to keep developer startup logs quiet.
	localPath := filepath.Join("..", "..", "..", "manifest", "config", "config.yaml")
	assertDatabaseDebugDisabledIfExists(t, localPath)
}

// assertDatabaseDebugDisabled verifies one config file disables ORM SQL debug
// logging by default.
func assertDatabaseDebugDisabled(t *testing.T, path string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read delivery config %s: %v", path, err)
	}
	if !strings.Contains(string(content), "debug: false") {
		t.Fatalf("expected %s to disable database debug by default", path)
	}
}

// assertDatabaseDebugDisabledIfExists verifies ignored local or packed config
// files only when they are present in the current workspace.
func assertDatabaseDebugDisabledIfExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err == nil {
		assertDatabaseDebugDisabled(t, path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat optional config %s: %v", path, err)
	}
}

// TestDatabaseDebugCanBeEnabledExplicitly verifies SQL diagnostics can still be
// enabled through an explicit config override.
func TestDatabaseDebugCanBeEnabledExplicitly(t *testing.T) {
	setTestServerConfigAdapter(t, `
database:
  default:
    debug: true
`)

	value, err := g.Cfg().Get(context.Background(), "database.default.debug")
	if err != nil {
		t.Fatalf("read explicit database debug config: %v", err)
	}
	if !value.Bool() {
		t.Fatal("expected explicit database.default.debug=true to be readable")
	}
}
