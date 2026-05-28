// Package authprovider defines the published contract for source-plugin
// authentication providers: the provider interface, login-entry shape, and
// the process-level registry plugins use to publish themselves.
//
// Source plugins register an auth provider during route registration so the
// host /auth/providers endpoint can enumerate enabled third-party login
// entries (Google, Discord, GitHub, custom OIDC, …) without depending on
// any specific plugin package. The host filters out providers whose owning
// plugin is currently disabled before returning the list.
package authprovider

import (
	"context"
	"sort"
	"strings"
	"sync"
)

// OAuthHandoffPath is the frontend route that consumes the host login
// outcome produced by source-plugin OAuth callbacks. Source plugins MUST
// redirect to this path (with the login result encoded in the URL query)
// instead of hard-coding their own value, so the workbench owns one stable
// handoff URL across deployments and the AuthService.OAuthHandoffURL host
// helper can compose the right full URL for hash vs history routing.
const OAuthHandoffPath = "/oauth-handoff"

// Kind classifies one authentication provider's protocol family.
type Kind string

// String returns the canonical lowercase string form of the kind.
func (k Kind) String() string {
	return string(k)
}

// Provider kind constants understood by the host workbench.
const (
	// KindOAuth2 represents a generic OAuth 2.0 authorization-code provider.
	KindOAuth2 Kind = "oauth2"
	// KindOIDC represents an OpenID Connect provider (OAuth 2.0 + ID token).
	KindOIDC Kind = "oidc"
	// KindLDAP represents an LDAP/AD directory authentication provider.
	KindLDAP Kind = "ldap"
	// KindSAML represents a SAML2 federated identity provider.
	KindSAML Kind = "saml"
	// KindCAS represents a Central Authentication Service provider.
	KindCAS Kind = "cas"
)

// Provider is the interface every source-plugin auth provider must
// implement. Implementations are registered through RegisterProvider and
// must be goroutine-safe because LoginEntry is invoked on each request.
type Provider interface {
	// ProviderID returns the stable provider identifier (e.g. "google",
	// "discord"). Two providers cannot register with the same identifier.
	ProviderID() string
	// PluginID returns the owning source-plugin identifier. The host uses
	// this to filter out providers whose owning plugin is disabled.
	PluginID() string
	// Kind returns the provider's protocol family.
	Kind() Kind
	// LoginEntry returns the runtime login-entry metadata. Implementations
	// are expected to read live plugin settings each call so admin toggles
	// take effect without restarting the plugin.
	LoginEntry(ctx context.Context) (*LoginEntry, error)
}

// LoginEntry is the runtime metadata returned by Provider.LoginEntry. The
// host maps this verbatim into ProviderView and /auth/providers responses
// after filtering disabled plugins.
type LoginEntry struct {
	// ProviderID matches Provider.ProviderID() for traceability.
	ProviderID string
	// PluginID matches Provider.PluginID() for traceability.
	PluginID string
	// Kind is the provider kind for the workbench UI router.
	Kind Kind
	// Name is the display name shown on the login button.
	Name string
	// Description is the short tagline shown next to the entry.
	Description string
	// Icon is the icon identifier rendered by the workbench (Iconify name).
	Icon string
	// EntryURL is the URL the browser visits to start the login flow.
	EntryURL string
	// BackendRedirectEnabled reports whether the provider should append a
	// state token so the callback can route the user to a state-specific
	// redirect URL after login.
	BackendRedirectEnabled bool
	// BackendRedirectDefault is the fallback redirect URL when no state
	// matches any configured rule.
	BackendRedirectDefault string
	// BackendRedirectRules is the JSON object mapping state to redirect URL.
	BackendRedirectRules string
	// DisplayOrder controls login entry sort order; smaller values first.
	DisplayOrder int
}

// ProviderView is the host-published projection of one auth provider used
// by /auth/providers responses. It mirrors LoginEntry but also carries the
// post-filtering Enabled flag.
type ProviderView struct {
	// ProviderID is the stable provider identifier.
	ProviderID string
	// PluginID is the owning source-plugin identifier.
	PluginID string
	// Kind is the provider kind.
	Kind Kind
	// Name is the display name.
	Name string
	// Description is the short tagline.
	Description string
	// Icon is the icon identifier.
	Icon string
	// EntryURL is the login entry URL.
	EntryURL string
	// BackendRedirectEnabled mirrors LoginEntry.
	BackendRedirectEnabled bool
	// BackendRedirectDefault mirrors LoginEntry.
	BackendRedirectDefault string
	// BackendRedirectRules mirrors LoginEntry.
	BackendRedirectRules string
	// DisplayOrder controls login entry sort order.
	DisplayOrder int
	// Enabled is true when the owning plugin is enabled.
	Enabled bool
}

// registry holds the process-level auth provider registry.
var registry = struct {
	mu        sync.RWMutex
	providers map[string]Provider
}{
	providers: make(map[string]Provider),
}

// RegisterProvider registers one auth provider in the process-level
// registry. It returns no error so plugin init() can call it directly;
// duplicate registrations replace the prior entry to allow plugin reload.
func RegisterProvider(p Provider) {
	if p == nil {
		return
	}
	id := strings.TrimSpace(p.ProviderID())
	if id == "" {
		return
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.providers[id] = p
}

// UnregisterProvider removes one auth provider from the registry. It is
// safe to call on missing IDs; the call is a no-op in that case.
func UnregisterProvider(providerID string) {
	id := strings.TrimSpace(providerID)
	if id == "" {
		return
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	delete(registry.providers, id)
}

// ResetForTest clears the registry. Tests use this to keep the global
// registry isolated between cases.
func ResetForTest() {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.providers = make(map[string]Provider)
}

// Providers returns a snapshot of every registered provider sorted by
// provider id. The caller may iterate the slice without holding the
// registry lock.
func Providers() []Provider {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	out := make([]Provider, 0, len(registry.providers))
	ids := make([]string, 0, len(registry.providers))
	for id := range registry.providers {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		out = append(out, registry.providers[id])
	}
	return out
}

// PluginEnabledChecker returns whether one plugin id is currently enabled.
// The host implements this against its plugin lifecycle service; the
// registry uses it to filter out disabled plugins before returning views.
type PluginEnabledChecker func(ctx context.Context, pluginID string) bool

// ListViews returns the host-facing snapshot of every registered provider
// whose owning plugin is enabled, sorted by DisplayOrder then ProviderID.
// The supplied checker MUST be safe to call from request goroutines.
// Errors encountered while loading individual provider entries are
// returned to the caller so the host can surface a structured failure.
func ListViews(ctx context.Context, isEnabled PluginEnabledChecker) ([]ProviderView, error) {
	providers := Providers()
	if len(providers) == 0 {
		return nil, nil
	}
	views := make([]ProviderView, 0, len(providers))
	for _, p := range providers {
		if p == nil {
			continue
		}
		enabled := true
		if isEnabled != nil {
			enabled = isEnabled(ctx, p.PluginID())
		}
		if !enabled {
			continue
		}
		entry, err := p.LoginEntry(ctx)
		if err != nil {
			return nil, err
		}
		if entry == nil {
			continue
		}
		views = append(views, ProviderView{
			ProviderID:             entry.ProviderID,
			PluginID:               entry.PluginID,
			Kind:                   entry.Kind,
			Name:                   entry.Name,
			Description:            entry.Description,
			Icon:                   entry.Icon,
			EntryURL:               entry.EntryURL,
			BackendRedirectEnabled: entry.BackendRedirectEnabled,
			BackendRedirectDefault: entry.BackendRedirectDefault,
			BackendRedirectRules:   entry.BackendRedirectRules,
			DisplayOrder:           entry.DisplayOrder,
			Enabled:                true,
		})
	}
	sort.SliceStable(views, func(i, j int) bool {
		if views[i].DisplayOrder != views[j].DisplayOrder {
			return views[i].DisplayOrder < views[j].DisplayOrder
		}
		return views[i].ProviderID < views[j].ProviderID
	})
	return views, nil
}
