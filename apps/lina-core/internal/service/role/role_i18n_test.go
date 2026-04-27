// This file verifies role-owned localization projection rules.

package role

import (
	"context"
	"testing"

	"lina-core/internal/model/entity"
)

// roleTestTranslator stubs the narrow role translation dependency.
type roleTestTranslator map[string]string

// Translate returns a configured translation or the caller fallback.
func (t roleTestTranslator) Translate(_ context.Context, key string, fallback string) string {
	if value, ok := t[key]; ok {
		return value
	}
	return fallback
}

// TestLocalizeListRoleNameTranslatesBuiltinAdmin verifies only the built-in admin role is projected.
func TestLocalizeListRoleNameTranslatesBuiltinAdmin(t *testing.T) {
	svc := &serviceImpl{
		i18nSvc: roleTestTranslator{
			"role.builtin.admin.name": "Administrator",
		},
	}

	name := svc.localizeListRoleName(context.Background(), &entity.SysRole{
		Key:  "admin",
		Name: "超级管理员",
	})

	if name != "Administrator" {
		t.Fatalf("expected built-in admin role name to be localized, got %q", name)
	}
}

// TestLocalizeListRoleNameKeepsCustomRole verifies custom role names remain stored values.
func TestLocalizeListRoleNameKeepsCustomRole(t *testing.T) {
	svc := &serviceImpl{
		i18nSvc: roleTestTranslator{
			"role.builtin.admin.name": "Administrator",
		},
	}

	name := svc.localizeListRoleName(context.Background(), &entity.SysRole{
		Key:  "operator",
		Name: "Operator",
	})

	if name != "Operator" {
		t.Fatalf("expected custom role name to remain unchanged, got %q", name)
	}
}
