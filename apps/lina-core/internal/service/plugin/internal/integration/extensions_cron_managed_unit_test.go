// This file tests managed-cron discovery helpers for dynamic plugin manifests.

package integration

import (
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
