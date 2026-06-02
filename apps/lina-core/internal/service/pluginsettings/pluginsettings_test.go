// This file verifies the pure helpers used by the host plugin settings
// service. Database-backed scenarios are covered separately by integration
// tests; the helpers here are intentionally pure so they stay testable in
// CI without standing up sys_config.

package pluginsettings

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"testing"
)

// TestBuildKeyComposesNamespacedKey verifies the canonical "<pluginID>.<key>"
// format used to store one setting in sys_config.
func TestBuildKeyComposesNamespacedKey(t *testing.T) {
	got, err := buildKey("linapro-oidc-google", "clientId")
	if err != nil {
		t.Fatalf("buildKey: %v", err)
	}
	if got != "linapro-oidc-google.clientId" {
		t.Fatalf("buildKey = %q, want linapro-oidc-google.clientId", got)
	}
}

// TestBuildKeyTrimsWhitespace verifies leading/trailing whitespace in the
// inputs is rejected via trimming, not silently accepted into the
// persisted key.
func TestBuildKeyTrimsWhitespace(t *testing.T) {
	got, err := buildKey("  linapro-demo  ", "  clientId  ")
	if err != nil {
		t.Fatalf("buildKey: %v", err)
	}
	if got != "linapro-demo.clientId" {
		t.Fatalf("buildKey = %q, want linapro-demo.clientId", got)
	}
}

// TestBuildKeyRejectsEmptyInputs verifies bad inputs are rejected so
// callers cannot accidentally access another plugin's namespace.
func TestBuildKeyRejectsEmptyInputs(t *testing.T) {
	if _, err := buildKey("", "clientId"); err == nil {
		t.Fatal("expected empty pluginID to be rejected")
	}
	if _, err := buildKey("linapro-demo", ""); err == nil {
		t.Fatal("expected empty key to be rejected")
	}
}

// TestBuildKeyRejectsDotInPluginID verifies the helper refuses pluginIDs
// that contain the namespace separator so plugins cannot escape their
// namespace by crafting a key like "linapro-demo.a" with pluginID
// "linapro-demo.a" and key "extra".
func TestBuildKeyRejectsDotInPluginID(t *testing.T) {
	if _, err := buildKey("linapro-demo.bad", "clientId"); err == nil {
		t.Fatal("expected pluginID containing dot to be rejected")
	}
}

// TestMaskSecretShortValue verifies short secrets fold to a fixed
// placeholder so the projection itself does not leak the length.
func TestMaskSecretShortValue(t *testing.T) {
	if got := MaskSecret(""); got != "" {
		t.Fatalf("expected empty mask for empty value, got %q", got)
	}
	if got := MaskSecret("abcdef"); got != "***" {
		t.Fatalf("expected short mask, got %q", got)
	}
}

// TestMaskSecretLongValue verifies longer secrets preserve the first and
// last three characters so operators can tell whether the secret rotated
// without exposing the middle.
func TestMaskSecretLongValue(t *testing.T) {
	if got := MaskSecret("abcdefghijkl"); got != "abc***jkl" {
		t.Fatalf("expected first-and-last preserving mask, got %q", got)
	}
}

// TestUpsertValueConflictUpdateColumns verifies the atomic upsert path keeps
// framework-owned metadata and built-in protection fields out of conflict
// updates while still resetting deleted_at to recover from any pre-existing
// soft-deleted row. This is a static contract test so it does not depend on
// a live database while still guarding the critical ORM column list.
func TestUpsertValueConflictUpdateColumns(t *testing.T) {
	columns := duplicateColumnsInUpsertValue(t)
	blocked := map[string]struct{}{
		"CreatedAt": {},
		"IsBuiltin": {},
		"UpdatedAt": {},
	}
	for _, column := range columns {
		if _, exists := blocked[column]; exists {
			t.Fatalf("upsertValue OnDuplicate must not update %s", column)
		}
	}
	// Name and Value are the value-bearing columns that must be refreshed on
	// conflict. DeletedAt must also be reset so a previously soft-deleted row
	// becomes visible again after the same (tenant_id, key) is upserted; the
	// regression test TestUpsertValueRecoversSoftDeletedRow exercises the
	// behavior end-to-end against the database.
	for _, required := range []string{"Name", "Value", "DeletedAt"} {
		if !containsString(columns, required) {
			t.Fatalf("upsertValue OnDuplicate columns = %v, want %s", columns, required)
		}
	}
}

// TestDeleteByFullKeyUsesUnscoped verifies the plugin-settings clear path
// physically removes the sys_config row instead of soft-deleting it. Soft
// delete would leave a ghost row under the unique (tenant_id, key) index
// and cause future upserts to update an invisible row in place. This is a
// static contract test so it never requires a live database.
func TestDeleteByFullKeyUsesUnscoped(t *testing.T) {
	if !functionBodyContainsCall(t, "deleteByFullKey", "Unscoped") {
		t.Fatal("deleteByFullKey must call Unscoped() to bypass soft-delete on plugin-settings rows")
	}
}

func duplicateColumnsInUpsertValue(t *testing.T) []string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	filePath := filepath.Join(filepath.Dir(currentFile), "pluginsettings.go")
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, filePath, nil, 0)
	if err != nil {
		t.Fatalf("parse pluginsettings.go: %v", err)
	}
	var columns []string
	ast.Inspect(parsed, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "OnDuplicate" {
			return true
		}
		if !enclosingFunctionContains(parsed, call, "upsertValue") {
			return true
		}
		for _, arg := range call.Args {
			argSelector, ok := arg.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			columns = append(columns, argSelector.Sel.Name)
		}
		return false
	})
	if len(columns) == 0 {
		t.Fatal("upsertValue OnDuplicate columns not found")
	}
	return columns
}

func enclosingFunctionContains(file *ast.File, target ast.Node, functionName string) bool {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != functionName || fn.Body == nil {
			continue
		}
		if fn.Body.Pos() <= target.Pos() && target.End() <= fn.Body.End() {
			return true
		}
	}
	return false
}

// functionBodyContainsCall reports whether the given top-level function
// declared in pluginsettings.go contains any call expression whose selector
// matches the supplied method name. It is intentionally narrow: it only
// inspects direct method invocations like `model.Unscoped()` so static
// contract tests do not have to encode the full chain order.
func functionBodyContainsCall(t *testing.T, functionName string, callName string) bool {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	filePath := filepath.Join(filepath.Dir(currentFile), "pluginsettings.go")
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, filePath, nil, 0)
	if err != nil {
		t.Fatalf("parse pluginsettings.go: %v", err)
	}
	for _, decl := range parsed.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != functionName || fn.Body == nil {
			continue
		}
		found := false
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if selector.Sel.Name == callName {
				found = true
				return false
			}
			return true
		})
		return found
	}
	t.Fatalf("function %s not found in pluginsettings.go", functionName)
	return false
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
