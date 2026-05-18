// This file verifies menu-owned localization projection rules.

package menu

import (
	"context"
	"testing"

	"lina-core/internal/model/entity"
)

// menuTestTranslator stubs the narrow menu translation dependency.
type menuTestTranslator map[string]string

// Translate returns a configured translation or the caller fallback.
func (t menuTestTranslator) Translate(_ context.Context, key string, fallback string) string {
	if value, ok := t[key]; ok {
		return value
	}
	return fallback
}

// TestLocalizeMenuEntityUsesMenuKey verifies menu_key remains the preferred translation anchor.
func TestLocalizeMenuEntityUsesMenuKey(t *testing.T) {
	svc := &serviceImpl{
		i18nSvc: menuTestTranslator{
			"menu.dashboard.title": "Dashboard",
		},
	}
	menu := &entity.SysMenu{
		MenuKey: "dashboard",
		Name:    "仪表盘",
	}

	svc.localizeMenuEntity(context.Background(), menu)

	if menu.Name != "Dashboard" {
		t.Fatalf("expected menu name to be localized by menu key, got %q", menu.Name)
	}
}

// TestLocalizeMenuEntityKeepsLiteralName verifies literal non-key names stay unchanged.
func TestLocalizeMenuEntityKeepsLiteralName(t *testing.T) {
	svc := &serviceImpl{
		i18nSvc: menuTestTranslator{
			"menu.dashboard.title": "Dashboard",
		},
	}
	menu := &entity.SysMenu{
		Name: "Custom Menu",
	}

	svc.localizeMenuEntity(context.Background(), menu)

	if menu.Name != "Custom Menu" {
		t.Fatalf("expected literal menu name to remain unchanged, got %q", menu.Name)
	}
}
