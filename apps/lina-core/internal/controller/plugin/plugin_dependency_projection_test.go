// This file tests plugin dependency API DTO projections used by management
// endpoints and list/install responses.

package plugin

import (
	"testing"

	pluginsvc "lina-core/internal/service/plugin"
)

// TestBuildPluginDependencyCheckResultProjectsAllSections verifies the
// dependency projection preserves install plan, blockers, auto-installed
// results, soft notices, cycles, reverse dependents, and reverse blockers.
func TestBuildPluginDependencyCheckResultProjectsAllSections(t *testing.T) {
	out := buildPluginDependencyCheckResult(&pluginsvc.DependencyCheckResult{
		TargetID: "plugin-consumer",
		Framework: pluginsvc.DependencyFrameworkCheck{
			RequiredVersion: ">=0.1.0",
			CurrentVersion:  "v0.1.0",
			Status:          "satisfied",
		},
		Dependencies: []*pluginsvc.DependencyPluginCheck{
			{
				OwnerID:         "plugin-consumer",
				DependencyID:    "plugin-base",
				DependencyName:  "Plugin Base",
				RequiredVersion: ">=0.1.0",
				CurrentVersion:  "v0.1.0",
				Required:        true,
				InstallMode:     "auto",
				Installed:       false,
				Discovered:      true,
				Status:          "auto_install_planned",
				Chain:           []string{"plugin-consumer", "plugin-base"},
			},
		},
		AutoInstallPlan: []*pluginsvc.DependencyAutoInstallItem{
			{
				PluginID:   "plugin-base",
				Name:       "Plugin Base",
				Version:    "v0.1.0",
				RequiredBy: "plugin-consumer",
				Chain:      []string{"plugin-consumer", "plugin-base"},
			},
		},
		AutoInstalled: []*pluginsvc.DependencyAutoInstallItem{
			{
				PluginID:   "plugin-base",
				Name:       "Plugin Base",
				Version:    "v0.1.0",
				RequiredBy: "plugin-consumer",
			},
		},
		ManualInstallRequired: []*pluginsvc.DependencyPluginCheck{
			{
				OwnerID:      "plugin-consumer",
				DependencyID: "plugin-manual",
				Required:     true,
				InstallMode:  "manual",
				Status:       "manual_install_required",
			},
		},
		SoftUnsatisfied: []*pluginsvc.DependencyPluginCheck{
			{
				OwnerID:      "plugin-consumer",
				DependencyID: "plugin-optional",
				Required:     false,
				InstallMode:  "manual",
				Status:       "soft_unsatisfied",
			},
		},
		Blockers: []*pluginsvc.DependencyBlocker{
			{
				Code:            "dependency_missing",
				PluginID:        "plugin-consumer",
				DependencyID:    "plugin-missing",
				RequiredVersion: ">=0.1.0",
				CurrentVersion:  "",
				Chain:           []string{"plugin-consumer", "plugin-missing"},
				Detail:          "missing",
			},
		},
		Cycle: []string{"a", "b", "a"},
		ReverseDependents: []*pluginsvc.DependencyReverseDependent{
			{
				PluginID:        "plugin-downstream",
				Name:            "Plugin Downstream",
				Version:         "v0.2.0",
				RequiredVersion: "<0.3.0",
			},
		},
		ReverseBlockers: []*pluginsvc.DependencyBlocker{
			{
				Code:         "dependency_snapshot_unknown",
				PluginID:     "plugin-unknown",
				DependencyID: "",
				Detail:       "unknown snapshot",
			},
		},
	})

	if out == nil {
		t.Fatal("expected projected dependency result")
	}
	if out.TargetId != "plugin-consumer" || out.Framework.Status != "satisfied" {
		t.Fatalf("unexpected top-level projection: %#v", out)
	}
	if len(out.Dependencies) != 1 || out.Dependencies[0].DependencyId != "plugin-base" {
		t.Fatalf("expected dependency edge projection, got %#v", out.Dependencies)
	}
	if len(out.AutoInstallPlan) != 1 || out.AutoInstallPlan[0].PluginId != "plugin-base" {
		t.Fatalf("expected auto-install plan projection, got %#v", out.AutoInstallPlan)
	}
	if len(out.AutoInstalled) != 1 || out.AutoInstalled[0].PluginId != "plugin-base" {
		t.Fatalf("expected auto-installed projection, got %#v", out.AutoInstalled)
	}
	if len(out.ManualInstallRequired) != 1 || out.ManualInstallRequired[0].DependencyId != "plugin-manual" {
		t.Fatalf("expected manual dependency projection, got %#v", out.ManualInstallRequired)
	}
	if len(out.SoftUnsatisfied) != 1 || out.SoftUnsatisfied[0].DependencyId != "plugin-optional" {
		t.Fatalf("expected soft dependency projection, got %#v", out.SoftUnsatisfied)
	}
	if len(out.Blockers) != 1 || out.Blockers[0].Code != "dependency_missing" {
		t.Fatalf("expected blocker projection, got %#v", out.Blockers)
	}
	if len(out.Cycle) != 3 || out.Cycle[2] != "a" {
		t.Fatalf("expected cycle projection, got %#v", out.Cycle)
	}
	if len(out.ReverseDependents) != 1 || out.ReverseDependents[0].PluginId != "plugin-downstream" {
		t.Fatalf("expected reverse dependent projection, got %#v", out.ReverseDependents)
	}
	if len(out.ReverseBlockers) != 1 || out.ReverseBlockers[0].Code != "dependency_snapshot_unknown" {
		t.Fatalf("expected reverse blocker projection, got %#v", out.ReverseBlockers)
	}
}
