// This file defines the per-request host call context injected into
// context.Context so that wazero host function callbacks can access
// plugin identity and capability permissions.

package wasm

import (
	"context"
	"errors"
	"path"
	"strings"

	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// hostCallContextKey is the context key type for host call state values.
type hostCallContextKey struct{}

// hostCallContext carries per-request plugin identity and authorization state.
type hostCallContext struct {
	// pluginID identifies the calling plugin.
	pluginID string
	// capabilities is the set of granted host capabilities for this plugin.
	capabilities map[string]struct{}
	// hostServices is the structured host service authorization snapshot for this plugin.
	hostServices []*bridgehostservice.HostServiceSpec
	// artifactDefaultConfig is the active-release default config content.
	artifactDefaultConfig []byte
	// artifactManifestResources contains active-release manifest resources
	// keyed relative to manifest/.
	artifactManifestResources map[string][]byte
	// executionSource identifies what triggered this Wasm execution.
	executionSource bridgecontract.ExecutionSource
	// routePath records the matched route path when execution is request-bound.
	routePath string
	// requestID carries the host request identifier for tracing.
	requestID string
	// identity carries the current user identity snapshot when available.
	identity *bridgecontract.IdentitySnapshotV1
	// cronCollector captures dynamic-plugin cron registrations during reserved
	// discovery executions.
	cronCollector CronRegistrationCollector
}

// withHostCallContext attaches the host call context to the execution context.
func withHostCallContext(ctx context.Context, hcc *hostCallContext) context.Context {
	return context.WithValue(ctx, hostCallContextKey{}, hcc)
}

// hostCallContextFrom retrieves the host call context from the execution context.
func hostCallContextFrom(ctx context.Context) *hostCallContext {
	if hcc, ok := ctx.Value(hostCallContextKey{}).(*hostCallContext); ok {
		return hcc
	}
	return nil
}

// hasCapability reports whether the plugin execution was granted the capability.
func (hcc *hostCallContext) hasCapability(capability string) bool {
	if hcc == nil || hcc.capabilities == nil {
		return false
	}
	_, ok := hcc.capabilities[capability]
	return ok
}

// hasHostServiceAccess reports whether the plugin may invoke the governed
// host-service target under the persisted authorization snapshot.
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
		methods := spec.Methods
		if len(methods) == 0 {
			methods = defaultHostServiceMethods(normalizedService)
		}
		if !containsString(methods, normalizedMethod) {
			continue
		}
		if normalizedService == bridgehostservice.HostServiceStorage {
			return normalizedResourceRef != "" && matchAuthorizedStoragePath(hcc.hostServices, normalizedResourceRef) != ""
		}
		if normalizedService == bridgehostservice.HostServiceNetwork {
			return normalizedResourceRef != "" && hcc.hostServiceResource(normalizedService, normalizedResourceRef) != nil
		}
		if normalizedService == bridgehostservice.HostServiceHostConfig {
			return normalizedResourceRef != "" && containsString(spec.Keys, normalizedResourceRef)
		}
		if normalizedService == bridgehostservice.HostServiceManifest {
			return normalizedResourceRef != "" && matchAuthorizedManifestPath(spec.Paths, normalizedResourceRef)
		}
		if normalizedService == bridgehostservice.HostServiceData {
			return normalizedTable != "" && containsString(spec.Tables, normalizedTable)
		}
		if normalizedService == bridgehostservice.HostServiceOrg ||
			normalizedService == bridgehostservice.HostServiceTenant {
			return normalizedResourceRef == "" && normalizedTable == ""
		}
		if normalizedResourceRef == "" {
			return len(spec.Resources) == 0 && len(spec.Tables) == 0
		}
		return hcc.hostServiceResource(normalizedService, normalizedResourceRef) != nil
	}
	return false
}

// hostServiceResource returns the authorized governed resource snapshot for one service/ref pair.
// hostServiceResource returns the authorized resource snapshot for one service/ref pair.
func (hcc *hostCallContext) hostServiceResource(service string, resourceRef string) *bridgehostservice.HostServiceResourceSpec {
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
		if normalizedService == bridgehostservice.HostServiceStorage {
			return nil
		}
		if normalizedService == bridgehostservice.HostServiceNetwork {
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

// defaultHostServiceMethods returns runtime defaults for declarations normalized
// before the get-only read service migration.
func defaultHostServiceMethods(service string) []string {
	switch service {
	case bridgehostservice.HostServiceConfig:
		return []string{bridgehostservice.HostServiceMethodConfigGet}
	case bridgehostservice.HostServiceHostConfig:
		return []string{bridgehostservice.HostServiceMethodHostConfigGet}
	case bridgehostservice.HostServiceManifest:
		return []string{bridgehostservice.HostServiceMethodManifestGet}
	default:
		return nil
	}
}

// hostServiceSpec returns the authorized service snapshot for one logical service.
// hostServiceSpec returns the authorized host-service specification for the service.
func (hcc *hostCallContext) hostServiceSpec(service string) *bridgehostservice.HostServiceSpec {
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

// containsString reports whether target appears in the slice.
func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

// matchAuthorizedManifestPath reports whether target is covered by one exact or glob path.
func matchAuthorizedManifestPath(patterns []string, target string) bool {
	normalizedTarget, err := normalizeManifestAuthorizedPath(target)
	if err != nil {
		return false
	}
	for _, rawPattern := range patterns {
		normalizedPattern, patternErr := normalizeManifestAuthorizedPath(rawPattern)
		if patternErr != nil {
			continue
		}
		if matched, matchErr := path.Match(normalizedPattern, normalizedTarget); matchErr == nil && matched {
			return true
		}
		if normalizedPattern == normalizedTarget {
			return true
		}
	}
	return false
}

// normalizeManifestAuthorizedPath validates the path enough for authorization matching.
func normalizeManifestAuthorizedPath(value string) (string, error) {
	raw := strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	if raw == "" || raw == "." {
		return "", errors.New("invalid manifest host service resource")
	}
	if strings.Contains(raw, "://") || strings.HasPrefix(raw, "/") {
		return "", errors.New("invalid manifest host service resource")
	}
	if len(raw) >= 2 && ((raw[0] >= 'A' && raw[0] <= 'Z') || (raw[0] >= 'a' && raw[0] <= 'z')) && raw[1] == ':' {
		return "", errors.New("invalid manifest host service resource")
	}
	normalized := path.Clean(raw)
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") {
		return "", errors.New("invalid manifest host service resource")
	}
	if normalized == "manifest" || strings.HasPrefix(normalized, "manifest/") {
		return "", errors.New("invalid manifest host service resource")
	}
	return normalized, nil
}
