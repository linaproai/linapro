// This file verifies org-capability i18n projection metadata.

package orgcapadapter

import "testing"

// TestNewUnassignedDeptNodeCarriesLabelKey verifies the synthetic department
// node exposes a stable runtime i18n key for host-side localization.
func TestNewUnassignedDeptNodeCarriesLabelKey(t *testing.T) {
	node := newUnassignedDeptNode(10, 7)
	if node.Id != 0 {
		t.Fatalf("expected synthetic department id 0, got %d", node.Id)
	}
	if node.LabelKey != orgCapUnassignedDeptLabelKey {
		t.Fatalf("expected label key %q, got %q", orgCapUnassignedDeptLabelKey, node.LabelKey)
	}
	if node.Label != "Unassigned (3)" {
		t.Fatalf("expected English fallback label with count, got %q", node.Label)
	}
	if node.UserCount != 3 {
		t.Fatalf("expected unassigned user count 3, got %d", node.UserCount)
	}
}
