// This file covers resource-reference synchronization behaviors owned by integration.

package integration_test

import (
	"context"
	"path/filepath"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
)

func TestSyncPluginResourceReferencesRevivesSoftDeletedRows(t *testing.T) {
	services := testutil.NewServices()
	service := services.Integration
	ctx := context.Background()

	pluginID := "plugin-dynamic-ref-revive"
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		pluginID,
		"Runtime Ref Revive Plugin",
		"v0.9.0",
		[]*catalog.ArtifactSQLAsset{
			{Key: "001-plugin-dynamic-ref-revive.sql", Content: "SELECT 1;"},
		},
		nil,
	)

	manifest := &catalog.Manifest{
		ID:           pluginID,
		Name:         "Runtime Ref Revive Plugin",
		Version:      "v0.9.0",
		Type:         catalog.TypeDynamic.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}
	if err := services.Catalog.ValidateManifest(manifest, manifest.ManifestPath); err != nil {
		t.Fatalf("expected dynamic manifest to be valid, got error: %v", err)
	}

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected plugin manifest sync to succeed, got error: %v", err)
	}

	release, err := services.Catalog.GetRelease(ctx, pluginID, manifest.Version)
	if err != nil {
		t.Fatalf("expected plugin release to exist, got error: %v", err)
	}
	if release == nil {
		t.Fatalf("expected plugin release to be created")
	}

	if _, err = dao.SysPluginResourceRef.Ctx(ctx).
		Where(do.SysPluginResourceRef{
			PluginId:  pluginID,
			ReleaseId: release.Id,
		}).
		Delete(); err != nil {
		t.Fatalf("expected resource refs to be soft-deleted, got error: %v", err)
	}

	if err = service.SyncPluginResourceReferences(ctx, manifest); err != nil {
		t.Fatalf("expected sync to revive soft-deleted rows without duplicate-key errors, got error: %v", err)
	}

	activeRefs, err := service.ListPluginResourceRefs(ctx, pluginID, release.Id)
	if err != nil {
		t.Fatalf("expected resource refs to be queryable, got error: %v", err)
	}
	if len(activeRefs) == 0 {
		t.Fatalf("expected revived resource refs to exist")
	}

	for _, item := range activeRefs {
		if item == nil {
			continue
		}
		if item.DeletedAt != nil {
			t.Fatalf("expected revived resource ref %s/%s to be active, got deleted_at=%v", item.ResourceType, item.ResourceKey, item.DeletedAt)
		}
	}
}
