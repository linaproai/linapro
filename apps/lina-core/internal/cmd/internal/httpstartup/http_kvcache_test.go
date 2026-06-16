// This file verifies HTTP startup selects the shared KV cache backend without
// mutating process-wide kvcache defaults.

package httpstartup

import (
	"testing"

	"lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
	"lina-core/internal/service/kvcache"
)

// TestNewHTTPKVCacheProviderSelectsSQLTableForSingleNode verifies single-node
// startup explicitly wires the SQL table provider.
func TestNewHTTPKVCacheProviderSelectsSQLTableForSingleNode(t *testing.T) {
	provider, err := newHTTPKVCacheProvider(&config.ClusterConfig{Enabled: false}, nil)
	if err != nil {
		t.Fatalf("select single-node kvcache provider: %v", err)
	}
	service := kvcache.New(kvcache.WithProvider(provider))
	if service.BackendName() != kvcache.BackendSQLTable {
		t.Fatalf("expected sql table backend, got %q", service.BackendName())
	}
	if !service.RequiresExpiredCleanup() {
		t.Fatal("expected sql table backend to require expired cleanup")
	}
}

// TestNewHTTPKVCacheProviderSelectsCoordinationForCluster verifies cluster
// startup explicitly wires the coordination KV provider.
func TestNewHTTPKVCacheProviderSelectsCoordinationForCluster(t *testing.T) {
	provider, err := newHTTPKVCacheProvider(&config.ClusterConfig{Enabled: true}, coordination.NewMemory(nil))
	if err != nil {
		t.Fatalf("select cluster kvcache provider: %v", err)
	}
	service := kvcache.New(kvcache.WithProvider(provider))
	if service.BackendName() != kvcache.BackendCoordinationKV {
		t.Fatalf("expected coordination KV backend, got %q", service.BackendName())
	}
	if service.RequiresExpiredCleanup() {
		t.Fatal("expected coordination KV backend to skip expired cleanup")
	}
}

// TestNewHTTPKVCacheProviderRejectsClusterWithoutCoordination verifies cluster
// startup does not silently fall back to SQL or process defaults.
func TestNewHTTPKVCacheProviderRejectsClusterWithoutCoordination(t *testing.T) {
	if _, err := newHTTPKVCacheProvider(&config.ClusterConfig{Enabled: true}, nil); err == nil {
		t.Fatal("expected cluster kvcache provider selection to require coordination service")
	}
}

// TestNewHTTPKVCacheProviderDoesNotReadProcessDefault verifies HTTP startup
// uses the explicit provider returned by its topology helper.
func TestNewHTTPKVCacheProviderDoesNotReadProcessDefault(t *testing.T) {
	original := kvcache.DefaultProvider()
	kvcache.SetDefaultProvider(kvcache.NewCoordinationKVProvider(coordination.NewMemory(nil)))
	t.Cleanup(func() {
		kvcache.SetDefaultProvider(original)
	})

	provider, err := newHTTPKVCacheProvider(&config.ClusterConfig{Enabled: false}, nil)
	if err != nil {
		t.Fatalf("select single-node kvcache provider: %v", err)
	}
	service := kvcache.New(kvcache.WithProvider(provider))
	if service.BackendName() != kvcache.BackendSQLTable {
		t.Fatalf("expected explicit sql table backend despite process default, got %q", service.BackendName())
	}
}
