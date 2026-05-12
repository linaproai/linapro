// This file verifies platform API permission tags and DTO validation metadata.

package v1

import (
	"reflect"
	"strings"
	"testing"
)

// TestPlatformPermissionTagsUseSystemPrefix verifies platform APIs use unified
// system permission tags and keep platform boundaries in route/context checks.
func TestPlatformPermissionTagsUseSystemPrefix(t *testing.T) {
	requestTypes := []reflect.Type{
		reflect.TypeOf(TenantListReq{}),
		reflect.TypeOf(TenantGetReq{}),
		reflect.TypeOf(TenantCreateReq{}),
		reflect.TypeOf(TenantUpdateReq{}),
		reflect.TypeOf(TenantStatusReq{}),
		reflect.TypeOf(TenantDeleteReq{}),
		reflect.TypeOf(TenantImpersonateReq{}),
		reflect.TypeOf(TenantEndImpersonateReq{}),
	}

	for _, requestType := range requestTypes {
		field, ok := requestType.FieldByName("Meta")
		if !ok {
			t.Fatalf("%s is missing g.Meta field", requestType.Name())
		}
		permission := field.Tag.Get("permission")
		if !strings.HasPrefix(permission, "system:") {
			t.Fatalf("%s permission must use system:* prefix, got %q", requestType.Name(), permission)
		}
	}
}

// TestTenantCreateCodeValidationMatchesServiceContract verifies the DTO exposes the 2-32 code contract.
func TestTenantCreateCodeValidationMatchesServiceContract(t *testing.T) {
	field, ok := reflect.TypeOf(TenantCreateReq{}).FieldByName("Code")
	if !ok {
		t.Fatal("TenantCreateReq.Code field missing")
	}
	validation := field.Tag.Get("v")
	if !strings.Contains(validation, "min-length:2") || !strings.Contains(validation, "max-length:32") {
		t.Fatalf("expected tenant code validation to include 2-32 length contract, got %q", validation)
	}
}
