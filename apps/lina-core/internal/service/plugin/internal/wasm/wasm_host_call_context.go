// This file defines the per-request host call context injected into
// context.Context so that wazero host function callbacks can access
// plugin identity and capability permissions.

package wasm

import (
	"context"
	"errors"
	"path"
	"strings"
	"sync"

	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// hostCallContextKey is the context key type for host call state values.
type hostCallContextKey struct{}

// hostCallContext carries per-request plugin identity and authorization state.
type hostCallContext struct {
	// runtime stores host-service dependencies for the current WASM runtime instance.
	runtime *hostServiceRuntime
	// pluginID identifies the calling plugin.
	pluginID string
	// capabilities is the set of granted host capabilities for this plugin.
	capabilities map[string]struct{}
	// hostServices is the structured host service authorization snapshot for this plugin.
	hostServices []*bridgehostservice.HostServiceSpec
	// hostServiceSnapshot indexes hostServices once for the current guest execution.
	hostServiceSnapshot *hostServiceAccessSnapshot
	// hostServiceSnapshotOnce guards lazy snapshot construction for tests and
	// direct host-call contexts that do not prebuild the snapshot.
	hostServiceSnapshotOnce sync.Once
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
	// jobCollector captures Jobs declarations during dynamic-plugin discovery.
	jobCollector JobRegistrationCollector
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
	snapshot := hcc.accessSnapshot()
	if snapshot == nil {
		return false
	}

	var (
		normalizedService     = strings.ToLower(strings.TrimSpace(service))
		normalizedMethod      = strings.ToLower(strings.TrimSpace(method))
		normalizedResourceRef = strings.TrimSpace(resourceRef)
		normalizedTable       = strings.TrimSpace(table)
	)

	if !snapshot.hasMethod(normalizedService, normalizedMethod) {
		return false
	}
	spec := snapshot.spec(normalizedService)
	if spec == nil {
		return false
	}

	// Storage and network authorizations may grant prefixes or URL patterns
	// instead of exact resource IDs, so they must be resolved through the same
	// matcher used by the runtime dispatcher.
	if normalizedService == bridgehostservice.HostServiceStorage {
		return normalizedResourceRef != "" && matchAuthorizedStoragePath(snapshot.hostServices, normalizedResourceRef) != ""
	}
	if normalizedService == bridgehostservice.HostServiceNetwork {
		return normalizedResourceRef != "" && hcc.hostServiceResource(normalizedService, normalizedResourceRef) != nil
	}
	if normalizedService == bridgehostservice.HostServiceHostConfig {
		return normalizedResourceRef != "" && snapshot.hasKey(normalizedService, normalizedResourceRef)
	}
	if normalizedService == bridgehostservice.HostServiceManifest {
		return normalizedResourceRef != "" && matchAuthorizedManifestPath(spec.Paths, normalizedResourceRef)
	}
	if normalizedService == bridgehostservice.HostServiceData {
		return normalizedTable != "" && snapshot.hasTable(normalizedService, normalizedTable)
	}
	if normalizedService == bridgehostservice.HostServiceNotifications &&
		normalizedMethod == bridgehostservice.HostServiceMethodNotificationsSend {
		return normalizedResourceRef != "" && hcc.hostServiceResource(normalizedService, normalizedResourceRef) != nil
	}
	if normalizedService == bridgehostservice.HostServiceAPIDoc ||
		normalizedService == bridgehostservice.HostServiceAuth ||
		normalizedService == bridgehostservice.HostServiceAuthz ||
		normalizedService == bridgehostservice.HostServiceUsers ||
		normalizedService == bridgehostservice.HostServiceBizCtx ||
		normalizedService == bridgehostservice.HostServiceDict ||
		normalizedService == bridgehostservice.HostServiceFiles ||
		normalizedService == bridgehostservice.HostServiceInfra ||
		normalizedService == bridgehostservice.HostServiceJobs ||
		normalizedService == bridgehostservice.HostServiceNotifications ||
		normalizedService == bridgehostservice.HostServicePlugins ||
		normalizedService == bridgehostservice.HostServiceRoute ||
		normalizedService == bridgehostservice.HostServiceSessions ||
		normalizedService == bridgehostservice.HostServiceOrg ||
		normalizedService == bridgehostservice.HostServiceTenant {
		return normalizedResourceRef == "" && normalizedTable == ""
	}
	if normalizedResourceRef == "" {
		return len(spec.Resources) == 0 && len(spec.Tables) == 0
	}
	return hcc.hostServiceResource(normalizedService, normalizedResourceRef) != nil
}

// hostServiceResource returns the authorized resource snapshot for one service/ref pair.
func (hcc *hostCallContext) hostServiceResource(service string, resourceRef string) *bridgehostservice.HostServiceResourceSpec {
	snapshot := hcc.accessSnapshot()
	if snapshot == nil {
		return nil
	}

	normalizedService := strings.ToLower(strings.TrimSpace(service))
	normalizedResourceRef := strings.TrimSpace(resourceRef)
	if normalizedService == "" || normalizedResourceRef == "" {
		return nil
	}

	if normalizedService == bridgehostservice.HostServiceStorage {
		return nil
	}
	if normalizedService == bridgehostservice.HostServiceNetwork {
		return matchAuthorizedNetworkResource(snapshot.hostServices, normalizedResourceRef)
	}
	return snapshot.resource(normalizedService, normalizedResourceRef)
}

// hostServiceSpec returns the authorized host-service specification for the service.
func (hcc *hostCallContext) hostServiceSpec(service string) *bridgehostservice.HostServiceSpec {
	snapshot := hcc.accessSnapshot()
	if snapshot == nil {
		return nil
	}
	normalizedService := strings.ToLower(strings.TrimSpace(service))
	if normalizedService == "" {
		return nil
	}
	return snapshot.spec(normalizedService)
}

// accessSnapshot returns the prebuilt request snapshot, building it lazily for
// unit tests that construct hostCallContext directly.
func (hcc *hostCallContext) accessSnapshot() *hostServiceAccessSnapshot {
	if hcc == nil {
		return nil
	}
	hcc.hostServiceSnapshotOnce.Do(func() {
		if hcc.hostServiceSnapshot == nil {
			hcc.hostServiceSnapshot = newHostServiceAccessSnapshot(hcc.hostServices)
		}
	})
	return hcc.hostServiceSnapshot
}

// hostServiceAccessSnapshot indexes one request's host service authorization
// snapshot without changing storage/network matcher semantics.
type hostServiceAccessSnapshot struct {
	hostServices []*bridgehostservice.HostServiceSpec
	specs        map[string]*bridgehostservice.HostServiceSpec
	methods      map[string]map[string]struct{}
	resources    map[string]map[string]*bridgehostservice.HostServiceResourceSpec
	keys         map[string]map[string]struct{}
	tables       map[string]map[string]struct{}
}

// newHostServiceAccessSnapshot builds one immutable-enough request-local index
// from the host-confirmed active-release authorization snapshot.
func newHostServiceAccessSnapshot(specs []*bridgehostservice.HostServiceSpec) *hostServiceAccessSnapshot {
	if len(specs) == 0 {
		return nil
	}
	snapshot := &hostServiceAccessSnapshot{
		specs:     make(map[string]*bridgehostservice.HostServiceSpec, len(specs)),
		methods:   make(map[string]map[string]struct{}, len(specs)),
		resources: make(map[string]map[string]*bridgehostservice.HostServiceResourceSpec, len(specs)),
		keys:      make(map[string]map[string]struct{}, len(specs)),
		tables:    make(map[string]map[string]struct{}, len(specs)),
	}
	for _, spec := range specs {
		cloned := cloneHostServiceSpec(spec)
		if cloned == nil || cloned.Service == "" {
			continue
		}
		snapshot.hostServices = append(snapshot.hostServices, cloned)
		snapshot.specs[cloned.Service] = cloned
		snapshot.methods[cloned.Service] = normalizedStringSet(cloned.Methods)
		snapshot.keys[cloned.Service] = normalizedStringSet(cloned.Keys)
		snapshot.tables[cloned.Service] = normalizedStringSet(cloned.Tables)
		if len(cloned.Resources) > 0 {
			byRef := make(map[string]*bridgehostservice.HostServiceResourceSpec, len(cloned.Resources))
			for _, resource := range cloned.Resources {
				if resource == nil || strings.TrimSpace(resource.Ref) == "" {
					continue
				}
				byRef[strings.TrimSpace(resource.Ref)] = resource
			}
			snapshot.resources[cloned.Service] = byRef
		}
	}
	if len(snapshot.hostServices) == 0 {
		return nil
	}
	return snapshot
}

func (snapshot *hostServiceAccessSnapshot) spec(service string) *bridgehostservice.HostServiceSpec {
	if snapshot == nil {
		return nil
	}
	return snapshot.specs[strings.ToLower(strings.TrimSpace(service))]
}

func (snapshot *hostServiceAccessSnapshot) hasMethod(service string, method string) bool {
	if snapshot == nil {
		return false
	}
	methods := snapshot.methods[strings.ToLower(strings.TrimSpace(service))]
	_, ok := methods[strings.ToLower(strings.TrimSpace(method))]
	return ok
}

func (snapshot *hostServiceAccessSnapshot) hasKey(service string, key string) bool {
	if snapshot == nil {
		return false
	}
	keys := snapshot.keys[strings.ToLower(strings.TrimSpace(service))]
	_, ok := keys[strings.TrimSpace(key)]
	return ok
}

func (snapshot *hostServiceAccessSnapshot) hasTable(service string, table string) bool {
	if snapshot == nil {
		return false
	}
	tables := snapshot.tables[strings.ToLower(strings.TrimSpace(service))]
	_, ok := tables[strings.TrimSpace(table)]
	return ok
}

func (snapshot *hostServiceAccessSnapshot) resource(
	service string,
	resourceRef string,
) *bridgehostservice.HostServiceResourceSpec {
	if snapshot == nil {
		return nil
	}
	resources := snapshot.resources[strings.ToLower(strings.TrimSpace(service))]
	return resources[strings.TrimSpace(resourceRef)]
}

func cloneHostServiceSpec(spec *bridgehostservice.HostServiceSpec) *bridgehostservice.HostServiceSpec {
	if spec == nil {
		return nil
	}
	cloned := &bridgehostservice.HostServiceSpec{
		Service:   strings.ToLower(strings.TrimSpace(spec.Service)),
		Methods:   normalizeLowerStringSlice(spec.Methods),
		Paths:     trimStringSlice(spec.Paths),
		Tables:    trimStringSlice(spec.Tables),
		Keys:      trimStringSlice(spec.Keys),
		Resources: cloneHostServiceResources(spec.Resources),
	}
	return cloned
}

func cloneHostServiceResources(
	resources []*bridgehostservice.HostServiceResourceSpec,
) []*bridgehostservice.HostServiceResourceSpec {
	if len(resources) == 0 {
		return nil
	}
	cloned := make([]*bridgehostservice.HostServiceResourceSpec, 0, len(resources))
	for _, resource := range resources {
		if resource == nil {
			continue
		}
		cloned = append(cloned, &bridgehostservice.HostServiceResourceSpec{
			Ref:             strings.TrimSpace(resource.Ref),
			AllowMethods:    trimStringSlice(resource.AllowMethods),
			HeaderAllowList: trimStringSlice(resource.HeaderAllowList),
			TimeoutMs:       resource.TimeoutMs,
			MaxBodyBytes:    resource.MaxBodyBytes,
			Attributes:      cloneStringMap(resource.Attributes),
		})
	}
	return cloned
}

func normalizeLowerStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.ToLower(strings.TrimSpace(value))
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	return normalized
}

func trimStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	trimmedValues := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			trimmedValues = append(trimmedValues, trimmed)
		}
	}
	return trimmedValues
}

func normalizedStringSet(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized != "" {
			set[normalized] = struct{}{}
		}
	}
	return set
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	clone := make(map[string]string, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
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
		trimmedPattern := strings.ReplaceAll(strings.TrimSpace(rawPattern), "\\", "/")
		normalizedPattern, patternErr := normalizeManifestAuthorizedPath(rawPattern)
		if patternErr != nil {
			continue
		}
		if strings.HasSuffix(trimmedPattern, "/") {
			base := strings.TrimSuffix(normalizedPattern, "/")
			if normalizedTarget == base || strings.HasPrefix(normalizedTarget, base+"/") {
				return true
			}
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
