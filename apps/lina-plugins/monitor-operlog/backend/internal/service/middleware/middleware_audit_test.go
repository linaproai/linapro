// This file verifies operation-log audit middleware helper behavior.

package middleware

import (
	"encoding/json"
	"net/http"
	"testing"

	operlogsvc "lina-plugin-monitor-operlog/backend/internal/service/operlog"
)

// TestSanitizeOperLogParamMasksNestedSensitiveFields verifies password fields
// and shell-environment payloads are recursively sanitized before logging.
func TestSanitizeOperLogParamMasksNestedSensitiveFields(t *testing.T) {
	input := `{
		"password":"secret",
		"nested":{"newPassword":"next","env":{"TOKEN":"abc","SECRET":"def"}},
		"items":[
			{"oldPassword":"prev"},
			{"env":[{"key":"API_KEY","value":"123"},{"name":"TOKEN","value":"456"}]}
		]
	}`

	sanitized := sanitizeOperLogParam(input)

	var payload map[string]any
	if err := json.Unmarshal([]byte(sanitized), &payload); err != nil {
		t.Fatalf("unmarshal sanitized oper log param: %v", err)
	}

	if payload["password"] != operLogMaskedPassword {
		t.Fatalf("expected password masked, got %#v", payload["password"])
	}

	nested, ok := payload["nested"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested object, got %#v", payload["nested"])
	}
	if nested["newPassword"] != operLogMaskedPassword {
		t.Fatalf("expected nested newPassword masked, got %#v", nested["newPassword"])
	}

	env, ok := nested["env"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested env object, got %#v", payload["nested"])
	}
	if env["TOKEN"] != operLogRedactedValue || env["SECRET"] != operLogRedactedValue {
		t.Fatalf("expected nested env values redacted, got %#v", env)
	}

	items, ok := payload["items"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("expected two sanitized items, got %#v", payload["items"])
	}

	firstItem, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first item object, got %#v", items[0])
	}
	if firstItem["oldPassword"] != operLogMaskedPassword {
		t.Fatalf("expected oldPassword masked, got %#v", firstItem["oldPassword"])
	}

	secondItem, ok := items[1].(map[string]any)
	if !ok {
		t.Fatalf("expected second item object, got %#v", items[1])
	}
	envList, ok := secondItem["env"].([]any)
	if !ok || len(envList) != 2 {
		t.Fatalf("expected env list with two items, got %#v", secondItem["env"])
	}

	firstEnv, ok := envList[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first env item object, got %#v", envList[0])
	}
	if firstEnv["key"] != "API_KEY" || firstEnv["value"] != operLogRedactedValue {
		t.Fatalf("expected env key preserved and value redacted, got %#v", firstEnv)
	}

	secondEnv, ok := envList[1].(map[string]any)
	if !ok {
		t.Fatalf("expected second env item object, got %#v", envList[1])
	}
	if secondEnv["name"] != "TOKEN" || secondEnv["value"] != operLogRedactedValue {
		t.Fatalf("expected env name preserved and value redacted, got %#v", secondEnv)
	}
}

// TestSanitizeOperLogParamLeavesInvalidJSONUntouched verifies malformed JSON
// payloads are preserved verbatim instead of producing broken audit content.
func TestSanitizeOperLogParamLeavesInvalidJSONUntouched(t *testing.T) {
	input := `{"password":"secret"`
	if sanitized := sanitizeOperLogParam(input); sanitized != input {
		t.Fatalf("expected invalid JSON to stay unchanged, got %q", sanitized)
	}
}

// TestShouldRecordAuditRequest verifies audit capture rules stay aligned with the HTTP semantics.
func TestShouldRecordAuditRequest(t *testing.T) {
	testCases := []struct {
		name       string
		method     string
		operLogTag string
		expected   bool
	}{
		{name: "post always records", method: http.MethodPost, expected: true},
		{name: "put always records", method: http.MethodPut, expected: true},
		{name: "delete always records", method: http.MethodDelete, expected: true},
		{name: "get requires operlog tag", method: http.MethodGet, expected: false},
		{name: "get with operlog tag records", method: http.MethodGet, operLogTag: "export", expected: true},
		{name: "patch never records", method: http.MethodPatch, expected: false},
	}

	for _, testCase := range testCases {
		actual := shouldRecordAuditRequest(testCase.method, testCase.operLogTag)
		if actual != testCase.expected {
			t.Fatalf("%s: expected %v, got %v", testCase.name, testCase.expected, actual)
		}
	}
}

// TestInferOperType verifies the middleware reuses the shared operlog service constants.
func TestInferOperType(t *testing.T) {
	testCases := []struct {
		name       string
		method     string
		path       string
		operLogTag string
		expected   int
	}{
		{name: "operlog tag wins", method: http.MethodGet, path: "/api/v1/export", operLogTag: "export", expected: operlogsvc.OperTypeExport},
		{name: "unknown operlog tag falls back to other", method: http.MethodGet, path: "/api/v1/query", operLogTag: "custom", expected: operlogsvc.OperTypeOther},
		{name: "post import path maps to import", method: http.MethodPost, path: "/api/v1/file/import", expected: operlogsvc.OperTypeImport},
		{name: "post create defaults to create", method: http.MethodPost, path: "/api/v1/file", expected: operlogsvc.OperTypeCreate},
		{name: "put maps to update", method: http.MethodPut, path: "/api/v1/file", expected: operlogsvc.OperTypeUpdate},
		{name: "delete maps to delete", method: http.MethodDelete, path: "/api/v1/file", expected: operlogsvc.OperTypeDelete},
	}

	for _, testCase := range testCases {
		actual := inferOperType(testCase.method, testCase.path, testCase.operLogTag)
		if actual != testCase.expected {
			t.Fatalf("%s: expected %d, got %d", testCase.name, testCase.expected, actual)
		}
	}
}
