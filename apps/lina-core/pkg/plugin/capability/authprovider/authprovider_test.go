// This file verifies the authprovider registry and ListViews filtering
// behavior. Tests use ResetForTest to keep the global registry isolated.

package authprovider

import (
	"context"
	"errors"
	"testing"
)

// fakeProvider is a test-only Provider implementation.
type fakeProvider struct {
	pluginID   string
	providerID string
	entry      LoginEntry
	loadErr    error
}

func (f *fakeProvider) ProviderID() string { return f.providerID }
func (f *fakeProvider) PluginID() string   { return f.pluginID }
func (f *fakeProvider) Kind() Kind         { return f.entry.Kind }
func (f *fakeProvider) LoginEntry(context.Context) (*LoginEntry, error) {
	if f.loadErr != nil {
		return nil, f.loadErr
	}
	out := f.entry
	out.ProviderID = f.providerID
	out.PluginID = f.pluginID
	return &out, nil
}

// TestRegisterProviderAndListViews verifies registration plus ListViews
// returns sorted snapshots scoped to enabled plugins.
func TestRegisterProviderAndListViews(t *testing.T) {
	ResetForTest()
	defer ResetForTest()

	RegisterProvider(&fakeProvider{
		pluginID:   "plugin-a",
		providerID: "a-provider",
		entry: LoginEntry{
			Name:         "Provider A",
			Kind:         KindOAuth2,
			DisplayOrder: 20,
		},
	})
	RegisterProvider(&fakeProvider{
		pluginID:   "plugin-b",
		providerID: "b-provider",
		entry: LoginEntry{
			Name:         "Provider B",
			Kind:         KindOIDC,
			DisplayOrder: 10,
		},
	})

	views, err := ListViews(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListViews: %v", err)
	}
	if len(views) != 2 {
		t.Fatalf("expected 2 views, got %d", len(views))
	}
	// Sort order should be by DisplayOrder then ProviderID.
	if views[0].ProviderID != "b-provider" {
		t.Fatalf("expected b-provider first by display order, got %s", views[0].ProviderID)
	}
	if views[1].ProviderID != "a-provider" {
		t.Fatalf("expected a-provider second, got %s", views[1].ProviderID)
	}
	if !views[0].Enabled || !views[1].Enabled {
		t.Fatalf("expected both views Enabled=true when no checker is supplied")
	}
}

// TestListViewsFiltersDisabledPlugins verifies the checker callback hides
// providers whose owning plugin is disabled.
func TestListViewsFiltersDisabledPlugins(t *testing.T) {
	ResetForTest()
	defer ResetForTest()

	RegisterProvider(&fakeProvider{
		pluginID:   "plugin-active",
		providerID: "alpha",
		entry:      LoginEntry{Name: "Alpha", Kind: KindOIDC, DisplayOrder: 1},
	})
	RegisterProvider(&fakeProvider{
		pluginID:   "plugin-disabled",
		providerID: "beta",
		entry:      LoginEntry{Name: "Beta", Kind: KindOIDC, DisplayOrder: 2},
	})

	views, err := ListViews(context.Background(), func(_ context.Context, pluginID string) bool {
		return pluginID == "plugin-active"
	})
	if err != nil {
		t.Fatalf("ListViews: %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("expected 1 enabled view, got %d", len(views))
	}
	if views[0].ProviderID != "alpha" {
		t.Fatalf("expected alpha view, got %s", views[0].ProviderID)
	}
}

// TestListViewsPropagatesLoaderError verifies that an error from
// Provider.LoginEntry is surfaced to the caller.
func TestListViewsPropagatesLoaderError(t *testing.T) {
	ResetForTest()
	defer ResetForTest()

	RegisterProvider(&fakeProvider{
		pluginID:   "plugin-bad",
		providerID: "bad",
		loadErr:    errors.New("simulated"),
	})
	if _, err := ListViews(context.Background(), nil); err == nil {
		t.Fatal("expected provider load error to be surfaced")
	}
}

// TestRegisterProviderRejectsEmptyID verifies bad inputs are ignored so
// the registry stays free of empty-id entries.
func TestRegisterProviderRejectsEmptyID(t *testing.T) {
	ResetForTest()
	defer ResetForTest()

	RegisterProvider(nil)
	RegisterProvider(&fakeProvider{providerID: ""})
	if got := len(Providers()); got != 0 {
		t.Fatalf("expected 0 providers, got %d", got)
	}
}
