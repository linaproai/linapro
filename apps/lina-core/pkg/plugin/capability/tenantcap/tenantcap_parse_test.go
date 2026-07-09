// This file verifies tenant ID parsing stays coupled to the tenant contract.

package tenantcap

import (
	"testing"

	"lina-core/pkg/plugin/capability/capmodel"
)

func TestParseTenantID(t *testing.T) {
	tenantID, err := ParseTenantID(capmodel.DomainID(" 42 "))
	if err != nil {
		t.Fatalf("parse tenant id: %v", err)
	}
	if tenantID != 42 {
		t.Fatalf("unexpected tenant id: %d", tenantID)
	}
}
