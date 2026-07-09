// This file verifies source-plugin public contracts stay aligned with the
// plugin-domain service boundary decisions.

package pluginhost

import (
	"reflect"
	"testing"

	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// TestServicesDoesNotExposeTopLevelTenantTableFilter verifies source-plugin
// table filtering is not exposed as a service-directory method.
func TestServicesDoesNotExposeTopLevelTenantTableFilter(t *testing.T) {
	servicesType := reflect.TypeOf((*capability.Services)(nil)).Elem()
	if _, ok := servicesType.MethodByName("TenantTableFilter"); ok {
		t.Fatal("capability.Services must not expose top-level TenantTableFilter")
	}
}

// TestServicesTenantReturnsOrdinaryTenantService verifies capability.Services
// does not define a tenant-service mirror.
func TestServicesTenantReturnsOrdinaryTenantService(t *testing.T) {
	servicesType := reflect.TypeOf((*capability.Services)(nil)).Elem()
	method, ok := servicesType.MethodByName("Tenant")
	if !ok {
		t.Fatal("capability.Services must expose Tenant")
	}
	want := reflect.TypeOf((*tenantcap.Service)(nil)).Elem()
	if method.Type.NumOut() != 1 || method.Type.Out(0) != want {
		t.Fatalf("capability.Services.Tenant() = %v, want %v", method.Type, want)
	}
}
