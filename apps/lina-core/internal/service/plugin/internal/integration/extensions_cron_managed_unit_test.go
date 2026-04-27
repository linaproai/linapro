// This file tests managed-cron discovery helpers for dynamic plugin manifests.

package integration

import (
	"context"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginbridge"
)

// TestManifestDeclaresCronHostService verifies dynamic cron discovery only
// runs for manifests that explicitly declare the cron host service.
func TestManifestDeclaresCronHostService(t *testing.T) {
	t.Run("missing cron service", func(t *testing.T) {
		manifest := &catalog.Manifest{
			HostServices: []*pluginbridge.HostServiceSpec{
				{Service: pluginbridge.HostServiceRuntime},
				{Service: pluginbridge.HostServiceStorage},
			},
		}
		if manifestDeclaresCronHostService(manifest) {
			t.Fatal("expected manifest without cron service to skip cron discovery")
		}
	})

	t.Run("with cron service", func(t *testing.T) {
		manifest := &catalog.Manifest{
			HostServices: []*pluginbridge.HostServiceSpec{
				{Service: pluginbridge.HostServiceCron},
			},
		}
		if !manifestDeclaresCronHostService(manifest) {
			t.Fatal("expected manifest with cron service to enable cron discovery")
		}
	})
}

// TestManagedCronCollectorAddWithMetadataPreservesSourceText verifies source
// plugins can register English display metadata without relying on en-US JSON.
func TestManagedCronCollectorAddWithMetadataPreservesSourceText(t *testing.T) {
	collector := &managedCronCollector{
		pluginID: "source-plugin",
		items:    make([]ManagedCronJob, 0),
	}

	err := collector.AddWithMetadata(
		context.Background(),
		"# */15 * * * *",
		"source-plugin-echo-inspection",
		"Source Plugin Echo Inspection",
		"Runs a lightweight source-plugin inspection task for scheduler integration validation.",
		func(ctx context.Context) error {
			return nil
		},
	)
	if err != nil {
		t.Fatalf("expected metadata cron registration to succeed, got error: %v", err)
	}
	if len(collector.items) != 1 {
		t.Fatalf("expected 1 collected cron job, got %d", len(collector.items))
	}
	item := collector.items[0]
	if item.DisplayName != "Source Plugin Echo Inspection" {
		t.Fatalf("expected display name to come from source metadata, got %q", item.DisplayName)
	}
	if item.Description != "Runs a lightweight source-plugin inspection task for scheduler integration validation." {
		t.Fatalf("expected description to come from source metadata, got %q", item.Description)
	}
}
