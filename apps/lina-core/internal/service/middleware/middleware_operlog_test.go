// This file verifies operation-log request-parameter sanitization for
// passwords and shell environment variables.

package middleware

import (
	"encoding/json"
	"testing"
)

// TestSanitizeOperLogParamMasksNestedSensitiveFields verifies password fields
// and shell environment payloads are recursively sanitized before logging.
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
		t.Fatalf("expected nested env object, got %#v", nested["env"])
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
