// This file contains API contract guard tests that prevent entity leakage.
package api_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	configv1 "lina-core/api/config/v1"
	dictv1 "lina-core/api/dict/v1"
	filev1 "lina-core/api/file/v1"
	jobv1 "lina-core/api/job/v1"
	jobgroupv1 "lina-core/api/jobgroup/v1"
	joblogv1 "lina-core/api/joblog/v1"
	userv1 "lina-core/api/user/v1"
)

// TestAPIPackagesDoNotImportEntities verifies that public API contracts do not depend on database entities.
func TestAPIPackagesDoNotImportEntities(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fileSet := token.NewFileSet()
		file, parseErr := parser.ParseFile(fileSet, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}
		entityImport := strings.Join([]string{"internal", "model", "entity"}, "/")
		for _, imported := range file.Imports {
			if strings.Contains(strings.Trim(imported.Path.Value, `"`), entityImport) {
				t.Errorf("%s must not import internal entity types", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk api files: %v", err)
	}
}

// TestResponseDTOsDoNotExposeInternalJSONFields verifies that shared response DTOs omit sensitive entity fields.
func TestResponseDTOsDoNotExposeInternalJSONFields(t *testing.T) {
	dtoTypes := []any{
		configv1.ByKeyRes{},
		configv1.ConfigItem{},
		configv1.GetRes{},
		configv1.ListRes{},
		dictv1.DataByTypeRes{},
		dictv1.DataGetRes{},
		dictv1.DataListRes{},
		dictv1.DictDataItem{},
		dictv1.DictTypeItem{},
		dictv1.DictTypeOptionItem{},
		dictv1.TypeGetRes{},
		dictv1.TypeListRes{},
		dictv1.TypeOptionsRes{},
		filev1.DetailRes{},
		filev1.FileItem{},
		filev1.InfoByIdsRes{},
		filev1.ListItem{},
		filev1.ListRes{},
		jobv1.DetailRes{},
		jobv1.JobItem{},
		jobv1.ListItem{},
		jobv1.ListRes{},
		jobgroupv1.JobGroupItem{},
		jobgroupv1.ListItem{},
		jobgroupv1.ListRes{},
		joblogv1.DetailRes{},
		joblogv1.JobLogItem{},
		joblogv1.ListItem{},
		joblogv1.ListRes{},
		userv1.UserItem{},
		userv1.GetRes{},
		userv1.GetProfileRes{},
		userv1.ListItem{},
	}
	for _, dto := range dtoTypes {
		assertDTOOmitsJSONFields(t, reflect.TypeOf(dto), map[string]struct{}{
			"password":  {},
			"deletedAt": {},
			"path":      {},
			"engine":    {},
			"hash":      {},
		})
	}
}

// assertDTOOmitsJSONFields walks embedded DTO fields and fails when a forbidden JSON field is exposed.
func assertDTOOmitsJSONFields(t *testing.T, typ reflect.Type, forbidden map[string]struct{}) {
	t.Helper()
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		t.Fatalf("%s is not a struct DTO", typ)
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Anonymous {
			assertDTOOmitsJSONFields(t, field.Type, forbidden)
			continue
		}
		jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonName == "" || jsonName == "-" {
			continue
		}
		if _, blocked := forbidden[jsonName]; blocked {
			t.Errorf("%s exposes forbidden JSON field %q", typ.Name(), jsonName)
		}
		assertDTOFieldTypeDoesNotEmbedEntity(t, typ, field.Type)
	}
}

// assertDTOFieldTypeDoesNotEmbedEntity fails when a DTO field directly embeds an internal entity type.
func assertDTOFieldTypeDoesNotEmbedEntity(t *testing.T, parent reflect.Type, fieldType reflect.Type) {
	t.Helper()
	for fieldType.Kind() == reflect.Pointer || fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
		fieldType = fieldType.Elem()
	}
	if fieldType.Kind() != reflect.Struct {
		return
	}
	if strings.Contains(fieldType.PkgPath(), strings.Join([]string{"internal", "model", "entity"}, "/")) {
		t.Errorf("%s embeds internal entity type %s", parent.Name(), fieldType)
	}
}
