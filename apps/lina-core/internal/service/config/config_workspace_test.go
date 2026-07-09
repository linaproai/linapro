// This file verifies startup-scoped admin workspace base-path configuration.

package config

import (
	"context"
	"testing"
)

// TestGetWorkspaceBasePathUsesDefault verifies the workspace defaults to the
// built-in admin entry when the config section is absent.
func TestGetWorkspaceBasePathUsesDefault(t *testing.T) {
	setTestServerConfigAdapter(t, `
server:
  address: ":9120"
`)

	cfg := New().GetWorkspace(context.Background())
	if cfg.BasePath != defaultWorkspaceBasePath {
		t.Fatalf("expected default workspace base path %q, got %q", defaultWorkspaceBasePath, cfg.BasePath)
	}
}

// TestGetWorkspaceBasePathNormalizesConfiguredPath verifies configured values
// are normalized once at static config load.
func TestGetWorkspaceBasePathNormalizesConfiguredPath(t *testing.T) {
	setTestServerConfigAdapter(t, `
workspace:
  basePath: "/console/"
`)

	basePath := New().GetWorkspaceBasePath(context.Background())
	if basePath != "/console" {
		t.Fatalf("expected normalized workspace base path /console, got %q", basePath)
	}
}

// TestGetWorkspaceBasePathAllowsRoot verifies dedicated admin-domain
// deployments can mount the workspace at the public root.
func TestGetWorkspaceBasePathAllowsRoot(t *testing.T) {
	setTestServerConfigAdapter(t, `
workspace:
  basePath: "/"
`)

	basePath := New().GetWorkspaceBasePath(context.Background())
	if basePath != "/" {
		t.Fatalf("expected root workspace base path, got %q", basePath)
	}
}

// TestGetWorkspaceBasePathRejectsInvalidValues verifies reserved and ambiguous
// workspace entry paths fail before route binding continues.
func TestGetWorkspaceBasePathRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "relative", content: "workspace:\n  basePath: \"admin\"\n"},
		{name: "wildcard", content: "workspace:\n  basePath: \"/admin/*\"\n"},
		{name: "host api", content: "workspace:\n  basePath: \"/api\"\n"},
		{name: "plugin api", content: "workspace:\n  basePath: \"/x\"\n"},
		{name: "plugin assets", content: "workspace:\n  basePath: \"/x-assets\"\n"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			setTestServerConfigAdapter(t, testCase.content)
			defer func() {
				if recovered := recover(); recovered == nil {
					t.Fatal("expected workspace base path validation to panic")
				}
			}()
			_ = New().GetWorkspaceBasePath(context.Background())
		})
	}
}
