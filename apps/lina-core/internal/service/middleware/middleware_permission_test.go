// This file covers helper logic for declarative static API permission checks.

package middleware

import (
	"reflect"
	"testing"

	filev1 "lina-core/api/file/v1"
	"lina-core/internal/service/role"
)

func TestResolveDeclaredPermissions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		permissionTag string
		aliasTag      string
		expected      []string
	}{
		{
			name:          "prefer primary permission tag",
			permissionTag: "plugin:install, plugin:enable",
			aliasTag:      "plugin:disable",
			expected:      []string{"plugin:install", "plugin:enable"},
		},
		{
			name:          "fallback to perms alias",
			permissionTag: "",
			aliasTag:      "plugin:query",
			expected:      []string{"plugin:query"},
		},
		{
			name:          "trim empty and duplicate values",
			permissionTag: " plugin:query , , plugin:query , plugin:install ",
			expected:      []string{"plugin:query", "plugin:install"},
		},
		{
			name:     "empty tag returns nil",
			expected: nil,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actual := resolveDeclaredPermissions(testCase.permissionTag, testCase.aliasTag)
			if !reflect.DeepEqual(actual, testCase.expected) {
				t.Fatalf("expected permissions %#v, got %#v", testCase.expected, actual)
			}
		})
	}
}

func TestHasRequiredPermissions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		accessContext *role.UserAccessContext
		required      []string
		expected      bool
	}{
		{
			name: "super admin bypasses checks",
			accessContext: &role.UserAccessContext{
				IsSuperAdmin: true,
			},
			required: []string{"plugin:install"},
			expected: true,
		},
		{
			name: "wildcard permission bypasses checks",
			accessContext: &role.UserAccessContext{
				Permissions: []string{staticPermissionWildcard},
			},
			required: []string{"plugin:install"},
			expected: true,
		},
		{
			name: "single exact permission passes",
			accessContext: &role.UserAccessContext{
				Permissions: []string{"plugin:query", "plugin:install"},
			},
			required: []string{"plugin:install"},
			expected: true,
		},
		{
			name: "missing required permission fails",
			accessContext: &role.UserAccessContext{
				Permissions: []string{"plugin:query"},
			},
			required: []string{"plugin:install"},
			expected: false,
		},
		{
			name: "multiple required permissions use all-of semantics",
			accessContext: &role.UserAccessContext{
				Permissions: []string{"plugin:query", "plugin:install"},
			},
			required: []string{"plugin:query", "plugin:install"},
			expected: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actual := hasRequiredPermissions(testCase.accessContext, testCase.required)
			if actual != testCase.expected {
				t.Fatalf("expected permission result %v, got %v", testCase.expected, actual)
			}
		})
	}
}

func TestFileSuffixesReqDeclaresPermission(t *testing.T) {
	t.Parallel()

	metaField, ok := reflect.TypeOf(filev1.FileSuffixesReq{}).FieldByName("Meta")
	if !ok {
		t.Fatal("expected FileSuffixesReq to expose g.Meta field")
	}
	if actual := metaField.Tag.Get("permission"); actual != "system:file:query" {
		t.Fatalf("expected FileSuffixesReq permission tag %q, got %q", "system:file:query", actual)
	}
}
