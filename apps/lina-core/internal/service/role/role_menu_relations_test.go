// This file verifies role-menu relation normalization used by batch inserts.

package role

import "testing"

// TestBuildRoleMenuRelationsNormalizesInput verifies invalid and duplicate menu
// IDs do not reach the batch insert payload.
func TestBuildRoleMenuRelationsNormalizesInput(t *testing.T) {
	relations := buildRoleMenuRelations(7, []int{3, 0, 3, -1, 9}, 42)
	if len(relations) != 2 {
		t.Fatalf("expected 2 normalized relations, got %d", len(relations))
	}
	if relations[0].RoleId != 7 || relations[0].MenuId != 3 || relations[0].TenantId != 42 {
		t.Fatalf("unexpected first relation: %#v", relations[0])
	}
	if relations[1].RoleId != 7 || relations[1].MenuId != 9 || relations[1].TenantId != 42 {
		t.Fatalf("unexpected second relation: %#v", relations[1])
	}
}

// TestBuildRoleMenuRelationsRejectsInvalidRole verifies an invalid role ID
// produces no insert payload.
func TestBuildRoleMenuRelationsRejectsInvalidRole(t *testing.T) {
	relations := buildRoleMenuRelations(0, []int{1, 2}, 42)
	if len(relations) != 0 {
		t.Fatalf("expected no relations for invalid role, got %#v", relations)
	}
}
