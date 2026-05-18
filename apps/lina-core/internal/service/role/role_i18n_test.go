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

// TestDisplayNameTranslatesBuiltinAdmin verifies the built-in admin role is projected.
func TestDisplayNameTranslatesBuiltinAdmin(t *testing.T) {
	svc := &serviceImpl{
		i18nSvc: roleTestTranslator{
			"role.builtin.admin.name": "Administrator",
		},
	}

	name := svc.DisplayName(context.Background(), &entity.SysRole{
		Key:  "admin",
		Name: "超级管理员",
	})

	if name != "Administrator" {
		t.Fatalf("expected built-in admin role name to be localized, got %q", name)
	}
}

// TestDisplayNameTranslatesBuiltinUser verifies the built-in standard user role is projected.
func TestDisplayNameTranslatesBuiltinUser(t *testing.T) {
	svc := &serviceImpl{
		i18nSvc: roleTestTranslator{
			"role.builtin.user.name": "User",
		},
	}

	name := svc.DisplayName(context.Background(), &entity.SysRole{
		Key:  "user",
		Name: "普通用户",
	})

	if name != "User" {
		t.Fatalf("expected built-in user role name to be localized, got %q", name)
	}
}

// TestDisplayNameKeepsCustomRole verifies custom role names remain stored values.
func TestDisplayNameKeepsCustomRole(t *testing.T) {
	svc := &serviceImpl{
		i18nSvc: roleTestTranslator{
			"role.builtin.admin.name": "Administrator",
		},
	}

	name := svc.DisplayName(context.Background(), &entity.SysRole{
		Key:  "operator",
		Name: "Operator",
	})

	if name != "Operator" {
		t.Fatalf("expected custom role name to remain unchanged, got %q", name)
	}
}
