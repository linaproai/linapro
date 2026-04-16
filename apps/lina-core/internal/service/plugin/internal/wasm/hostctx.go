// This file defines the per-request host call context injected into
// context.Context so that wazero host function callbacks can access
// plugin identity and capability permissions.

package wasm

import (
	"context"
	"strings"

	"lina-core/pkg/pluginbridge"
)

// hostCallContextKey is the private context key for host call state.
type hostCallContextKey struct{}

// hostCallContext carries per-request state into wazero host function callbacks.
type hostCallContext struct {
	// pluginID identifies the calling plugin.
	pluginID string
	// capabilities is the set of granted host capabilities for this plugin.
	capabilities map[string]struct{}
	// hostServices is the structured host service authorization snapshot for this plugin.
	hostServices []*pluginbridge.HostServiceSpec
	// executionSource identifies what triggered this Wasm execution.
	executionSource pluginbridge.ExecutionSource
	// routePath records the matched route path when execution is request-bound.
	routePath string
	// requestID carries the host request identifier for tracing.
	requestID string
	// identity carries the current user identity snapshot when available.
	identity *pluginbridge.IdentitySnapshotV1
}

// withHostCallContext attaches a host call context to the given context.
func withHostCallContext(ctx context.Context, hcc *hostCallContext) context.Context {
	return context.WithValue(ctx, hostCallContextKey{}, hcc)
}

// hostCallContextFrom extracts the host call context from the given context.
func hostCallContextFrom(ctx context.Context) *hostCallContext {
	if hcc, ok := ctx.Value(hostCallContextKey{}).(*hostCallContext); ok {
		return hcc
	}
	return nil
}

// hasCapability checks if the plugin has been granted a specific capability.
func (hcc *hostCallContext) hasCapability(capability string) bool {
	if hcc == nil || hcc.capabilities == nil {
		return false
	}
	_, ok := hcc.capabilities[capability]
	return ok
}

// hasHostServiceAccess checks whether the plugin may invoke one service method and governed target.
func (hcc *hostCallContext) hasHostServiceAccess(service string, method string, resourceRef string, table string) bool {
	if hcc == nil || len(hcc.hostServices) == 0 {
		return false
	}

	var (
		normalizedService     = strings.ToLower(strings.TrimSpace(service))
		normalizedMethod      = strings.ToLower(strings.TrimSpace(method))
		normalizedResourceRef = strings.TrimSpace(resourceRef)
		normalizedTable       = strings.TrimSpace(table)
	)

	// Storage and network authorizations may grant prefixes or URL patterns
	// instead of exact resource IDs, so they must be resolved through the same
	// matcher used by the runtime dispatcher.
	for _, spec := range hcc.hostServices {
		if spec == nil || spec.Service != normalizedService {
			continue
		}
		if !containsString(spec.Methods, normalizedMethod) {
			continue
		}
		if normalizedService == pluginbridge.HostServiceStorage && normalizedResourceRef != "" {
			return matchAuthorizedStoragePath(hcc.hostServices, normalizedResourceRef) != ""
		}
		if normalizedService == pluginbridge.HostServiceNetwork && normalizedResourceRef != "" {
			return hcc.hostServiceResource(normalizedService, normalizedResourceRef) != nil
		}
		if normalizedTable != "" {
			return containsString(spec.Tables, normalizedTable)
		}
		if normalizedResourceRef == "" {
			return len(spec.Resources) == 0 && len(spec.Tables) == 0
		}
		return hcc.hostServiceResource(normalizedService, normalizedResourceRef) != nil
	}
	return false
}

// hostServiceResource returns the authorized governed resource snapshot for one service/ref pair.
func (hcc *hostCallContext) hostServiceResource(service string, resourceRef string) *pluginbridge.HostServiceResourceSpec {
	if hcc == nil || len(hcc.hostServices) == 0 {
		return nil
	}

	normalizedService := strings.ToLower(strings.TrimSpace(service))
	normalizedResourceRef := strings.TrimSpace(resourceRef)
	if normalizedService == "" || normalizedResourceRef == "" {
		return nil
	}

	for _, spec := range hcc.hostServices {
		if spec == nil || spec.Service != normalizedService {
			continue
		}
		if normalizedService == pluginbridge.HostServiceStorage {
			return nil
		}
		if normalizedService == pluginbridge.HostServiceNetwork {
			return matchAuthorizedNetworkResource(hcc.hostServices, normalizedResourceRef)
		}
		for _, resource := range spec.Resources {
			if resource == nil {
				continue
			}
			if strings.TrimSpace(resource.Ref) == normalizedResourceRef {
				return resource
			}
		}
	}
	return nil
}

// hostServiceSpec returns the authorized service snapshot for one logical service.
func (hcc *hostCallContext) hostServiceSpec(service string) *pluginbridge.HostServiceSpec {
	if hcc == nil || len(hcc.hostServices) == 0 {
		return nil
	}
	normalizedService := strings.ToLower(strings.TrimSpace(service))
	if normalizedService == "" {
		return nil
	}
	for _, spec := range hcc.hostServices {
		if spec != nil && spec.Service == normalizedService {
			return spec
		}
	}
	return nil
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
