// This file verifies plugin storage provider registration and zero-configuration
// provider selection semantics.

package storagecap

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"lina-core/pkg/bizerr"
)

func TestResolveProviderFallsBackToLocalWhenNoProviderIsAvailable(t *testing.T) {
	localProvider := &testStorageProvider{}
	providerID, provider, err := ResolveProvider(context.Background(), testProviderRuntime{}, localProvider)
	if err != nil {
		t.Fatalf("resolve provider: %v", err)
	}
	if providerID != LocalProviderID || provider != localProvider {
		t.Fatalf("expected local provider fallback, got id=%q provider=%T", providerID, provider)
	}
}

func TestResolveProviderSelectsUniqueEnabledPluginProvider(t *testing.T) {
	ctx := context.Background()
	providerID := fmt.Sprintf("storage-provider-unique-%d", nextStorageProviderTestID())
	pluginProvider := &testStorageProvider{}
	registerStorageProviderFactoryForTest(t, providerID, pluginProvider)

	resolvedID, provider, err := ResolveProvider(ctx, testProviderRuntime{
		available: map[string]bool{providerID: true},
	}, &testStorageProvider{})
	if err != nil {
		t.Fatalf("resolve provider: %v", err)
	}
	if resolvedID != providerID || provider != pluginProvider {
		t.Fatalf("expected plugin provider %q, got id=%q provider=%T", providerID, resolvedID, provider)
	}
}

func TestResolveProviderFallsBackToLocalWhenRegisteredProviderIsDisabled(t *testing.T) {
	ctx := context.Background()
	providerID := fmt.Sprintf("storage-provider-disabled-%d", nextStorageProviderTestID())
	registerStorageProviderFactoryForTest(t, providerID, &testStorageProvider{})
	localProvider := &testStorageProvider{}

	resolvedID, provider, err := ResolveProvider(ctx, testProviderRuntime{
		available: map[string]bool{providerID: false},
	}, localProvider)
	if err != nil {
		t.Fatalf("resolve provider: %v", err)
	}
	if resolvedID != LocalProviderID || provider != localProvider {
		t.Fatalf("expected local provider fallback, got id=%q provider=%T", resolvedID, provider)
	}
}

func TestResolveProviderRejectsMultipleEnabledPluginProviders(t *testing.T) {
	ctx := context.Background()
	providerAID := fmt.Sprintf("storage-provider-conflict-a-%d", nextStorageProviderTestID())
	providerBID := fmt.Sprintf("storage-provider-conflict-b-%d", nextStorageProviderTestID())
	registerStorageProviderFactoryForTest(t, providerAID, &testStorageProvider{})
	registerStorageProviderFactoryForTest(t, providerBID, &testStorageProvider{})

	_, _, err := ResolveProvider(ctx, testProviderRuntime{
		available: map[string]bool{
			providerAID: true,
			providerBID: true,
		},
	}, &testStorageProvider{})
	if !bizerr.Is(err, CodeStorageProviderConflict) {
		t.Fatalf("expected multiple provider conflict, got %v", err)
	}
	bizErr, ok := bizerr.As(err)
	providerIDs, hasProviderIDs := bizErr.Params()["providerIds"].(string)
	if !ok || !hasProviderIDs || providerIDs == "" {
		t.Fatalf("expected conflict provider IDs, got %v", err)
	}

	statuses := ProviderStatuses(ctx, testProviderRuntime{
		available: map[string]bool{
			providerAID: true,
			providerBID: true,
		},
	}, &testStorageProvider{})
	for _, providerID := range []string{providerAID, providerBID} {
		status := testProviderStatusByID(statuses, providerID)
		if status == nil || status.Active || !status.Available || status.Message == "" {
			t.Fatalf("expected conflicting provider status for %s, got %#v in %#v", providerID, status, statuses)
		}
	}
}

type testProviderRuntime struct {
	available map[string]bool
}

func (r testProviderRuntime) ProviderPluginAvailable(_ context.Context, pluginID string) bool {
	return r.available[strings.TrimSpace(pluginID)]
}

type testStorageProvider struct{}

func (*testStorageProvider) Put(context.Context, ProviderPutInput) (*ProviderObject, error) {
	return nil, nil
}

func (*testStorageProvider) Get(context.Context, ProviderGetInput) (*ProviderGetOutput, error) {
	return nil, nil
}

func (*testStorageProvider) Delete(context.Context, ProviderDeleteInput) error {
	return nil
}

func (*testStorageProvider) DeleteMany(context.Context, ProviderDeleteManyInput) error {
	return nil
}

func (*testStorageProvider) List(context.Context, ProviderListInput) (*ProviderListOutput, error) {
	return nil, nil
}

func (*testStorageProvider) ListCursor(context.Context, ProviderListCursorInput) (*ProviderListCursorOutput, error) {
	return nil, nil
}

func (*testStorageProvider) Stat(context.Context, ProviderStatInput) (*ProviderStatOutput, error) {
	return nil, nil
}

func (*testStorageProvider) BatchStat(context.Context, ProviderBatchStatInput) (*ProviderBatchStatOutput, error) {
	return nil, nil
}

func registerStorageProviderFactoryForTest(t *testing.T, providerID string, provider Provider) {
	t.Helper()
	if err := Provide(providerID, func(context.Context, ProviderEnv) (Provider, error) {
		return provider, nil
	}); err != nil {
		t.Fatalf("register storage provider %s: %v", providerID, err)
	}
}

func testProviderStatusByID(statuses []*ProviderStatus, providerID string) *ProviderStatus {
	for _, status := range statuses {
		if status != nil && status.ProviderID == providerID {
			return status
		}
	}
	return nil
}

var storageProviderTestSequence struct {
	sync.Mutex
	value int
}

func nextStorageProviderTestID() int {
	storageProviderTestSequence.Lock()
	defer storageProviderTestSequence.Unlock()
	storageProviderTestSequence.value++
	return storageProviderTestSequence.value
}
