// Package management owns plugin management read-model projections, caching,
// and request-local manifest snapshots used by the root plugin facade.
package management

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// PluginItem is the display-ready management projection of one plugin entry.
type PluginItem struct {
	runtime.PluginItem
	// DependencyCheck carries server-side dependency status for management UIs.
	DependencyCheck *plugindep.CheckProjection
}

// ListOutput defines output for plugin list query.
type ListOutput struct {
	// List contains the filtered plugin list.
	List []*PluginItem
	// Total is the number of returned plugins.
	Total int
}

// ListInput defines input for plugin list query.
type ListInput struct {
	// PageNum is the one-based page number for the summary list.
	PageNum int
	// PageSize is the bounded page size for the summary list.
	PageSize int
	// ID filters by plugin identifier.
	ID string
	// Name filters by plugin display name.
	Name string
	// Type filters by normalized plugin type.
	Type string
	// Status filters by enabled flag.
	Status *int
	// Installed filters by installed flag.
	Installed *int
}

const (
	// DefaultListPageNum is used when callers omit the page number.
	DefaultListPageNum = 1
	// DefaultListPageSize is used when callers omit the page size.
	DefaultListPageSize = 20
	// MaxListPageSize bounds the plugin management summary list response size.
	MaxListPageSize = 100
)

// NormalizeListPage applies default and maximum pagination bounds.
func NormalizeListPage(pageNum int, pageSize int) (int, int) {
	if pageNum < DefaultListPageNum {
		pageNum = DefaultListPageNum
	}
	if pageSize <= 0 {
		pageSize = DefaultListPageSize
	}
	if pageSize > MaxListPageSize {
		pageSize = MaxListPageSize
	}
	return pageNum, pageSize
}

// PaginatePluginItems returns the requested page and the total item count.
func PaginatePluginItems(items []*PluginItem, pageNum int, pageSize int) ([]*PluginItem, int) {
	pageNum, pageSize = NormalizeListPage(pageNum, pageSize)
	total := len(items)
	start := (pageNum - 1) * pageSize
	if start >= total {
		return []*PluginItem{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return items[start:end], total
}

// ListCache stores one complete unfiltered plugin management summary read model.
// Filtered API requests derive page data from this lightweight projection; detail
// and action modals use the separate plugin detail path.
type ListCache struct {
	mu         sync.RWMutex
	values     map[string]*ListOutput
	builds     map[string]*listCacheBuildCall
	generation uint64
}

// listCacheBuildCall records one in-flight cache build for a single cache key.
type listCacheBuildCall struct {
	wg         sync.WaitGroup
	generation uint64
	value      *ListOutput
	err        error
}

// NewListCache creates an empty process-local management list cache.
func NewListCache() *ListCache {
	return &ListCache{}
}

var errListCacheBuildInvalidated = errors.New("plugin management list cache build invalidated")

// Get returns a defensive copy of the cached list, if available.
func (c *ListCache) Get(key ListCacheKey) (*ListOutput, bool) {
	if c == nil {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.values == nil {
		return nil, false
	}
	value := c.values[key.String()]
	if value == nil {
		return nil, false
	}
	return CloneListOutput(value), true
}

// LoadOrBuild returns a cached list or runs one in-flight build per cache key.
func (c *ListCache) LoadOrBuild(key ListCacheKey, build func() (*ListOutput, error)) (*ListOutput, error) {
	if c == nil {
		return build()
	}
	for {
		value, err, retry := c.loadOrBuildOnce(key, build)
		if retry {
			continue
		}
		return value, err
	}
}

// loadOrBuildOnce performs one cached lookup/build cycle. If the cache is
// invalidated while a build is running, callers retry against the new generation.
func (c *ListCache) loadOrBuildOnce(key ListCacheKey, build func() (*ListOutput, error)) (*ListOutput, error, bool) {
	keyString := key.String()
	c.mu.Lock()
	if c.values != nil {
		if value := c.values[keyString]; value != nil {
			out := CloneListOutput(value)
			c.mu.Unlock()
			return out, nil, false
		}
	}
	if c.builds != nil {
		if call := c.builds[keyString]; call != nil {
			c.mu.Unlock()
			call.wg.Wait()
			if errors.Is(call.err, errListCacheBuildInvalidated) {
				return nil, nil, true
			}
			if call.err != nil {
				return nil, call.err, false
			}
			return CloneListOutput(call.value), nil, false
		}
	}
	if c.builds == nil {
		c.builds = make(map[string]*listCacheBuildCall)
	}
	call := &listCacheBuildCall{generation: c.generation}
	call.wg.Add(1)
	c.builds[keyString] = call
	c.mu.Unlock()

	value, err := build()
	built := CloneListOutput(value)

	c.mu.Lock()
	defer c.mu.Unlock()
	defer call.wg.Done()
	if err == nil && built != nil && c.generation != call.generation {
		call.err = errListCacheBuildInvalidated
		delete(c.builds, keyString)
		return nil, nil, true
	}
	if err == nil && built != nil {
		c.storeLocked(key, built)
	}
	call.value = built
	call.err = err
	delete(c.builds, keyString)
	if len(c.builds) == 0 {
		c.builds = nil
	}
	if err != nil {
		return nil, err, false
	}
	return CloneListOutput(built), nil, false
}

// Store replaces the cached list with a defensive copy and drops stale entries
// for the same locale but older runtime bundle versions.
func (c *ListCache) Store(key ListCacheKey, value *ListOutput) {
	if c == nil || value == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.storeLocked(key, value)
}

// storeLocked writes one defensive cache entry while the caller holds c.mu.
func (c *ListCache) storeLocked(key ListCacheKey, value *ListOutput) {
	if c.values == nil {
		c.values = make(map[string]*ListOutput)
	}
	for existingKey := range c.values {
		if listCacheKeyLocale(existingKey) == key.Locale && existingKey != key.String() {
			delete(c.values, existingKey)
		}
	}
	c.values[key.String()] = CloneListOutput(value)
}

// Invalidate clears all cached management list projections.
func (c *ListCache) Invalidate() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values = nil
	c.generation++
}

// ListCacheKey identifies one localized management list read model.
type ListCacheKey struct {
	// Locale is the request locale used while building display metadata.
	Locale string
	// RuntimeBundleVersion is the i18n runtime bundle version for Locale.
	RuntimeBundleVersion uint64
	// RuntimeRevision is the shared plugin-runtime cache revision.
	RuntimeRevision int64
}

// String returns a stable map key for the localized read-model cache.
func (k ListCacheKey) String() string {
	return k.Locale + "@" +
		strconv.FormatUint(k.RuntimeBundleVersion, 10) + "@" +
		strconv.FormatInt(k.RuntimeRevision, 10)
}

// listCacheKeyLocale extracts the locale prefix from one cache key.
func listCacheKeyLocale(key string) string {
	locale, _, ok := strings.Cut(key, "@")
	if !ok {
		return key
	}
	return locale
}

// WithManifestSnapshot stores one already-scanned manifest list in context so
// dependency checks inside the same list build do not rescan source plugins and
// dynamic artifacts for every plugin row.
func WithManifestSnapshot(ctx context.Context, manifests []*catalog.Manifest) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if ManifestSnapshotFromContext(ctx) != nil {
		return ctx
	}
	return context.WithValue(ctx, manifestSnapshotContextKey{}, CloneManifestSlice(manifests))
}

// ManifestSnapshotFromContext returns the request-local manifest list, if set.
func ManifestSnapshotFromContext(ctx context.Context) []*catalog.Manifest {
	if ctx == nil {
		return nil
	}
	manifests, ok := ctx.Value(manifestSnapshotContextKey{}).([]*catalog.Manifest)
	if !ok || manifests == nil {
		return nil
	}
	return CloneManifestSlice(manifests)
}

// ManifestByIDFromContext returns a manifest from the request-local discovery
// snapshot without triggering another scan.
func ManifestByIDFromContext(ctx context.Context, pluginID string) *catalog.Manifest {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return nil
	}
	for _, manifest := range ManifestSnapshotFromContext(ctx) {
		if manifest != nil && strings.TrimSpace(manifest.ID) == normalizedPluginID {
			return manifest
		}
	}
	return nil
}

// CloneManifestSlice copies the manifest slice header so callers cannot mutate
// the request-local list ordering.
func CloneManifestSlice(in []*catalog.Manifest) []*catalog.Manifest {
	if in == nil {
		return nil
	}
	out := make([]*catalog.Manifest, len(in))
	copy(out, in)
	return out
}

// RegistryByPluginID indexes registry rows by plugin ID for read-only list projection.
func RegistryByPluginID(registries []*entity.SysPlugin) map[string]*entity.SysPlugin {
	result := make(map[string]*entity.SysPlugin, len(registries))
	for _, registry := range registries {
		if registry == nil || strings.TrimSpace(registry.PluginId) == "" {
			continue
		}
		result[registry.PluginId] = registry
	}
	return result
}

// SortPluginItems sorts facade plugin projections by plugin ID.
func SortPluginItems(items []*PluginItem) {
	sort.Slice(items, func(i int, j int) bool {
		if items[i] == nil {
			return false
		}
		if items[j] == nil {
			return true
		}
		return items[i].Id < items[j].Id
	})
}

// MatchesPluginType compares normalized plugin types so list filtering accepts
// user input that differs only by case or alias formatting.
func MatchesPluginType(actual string, expected string) bool {
	actualType := catalog.NormalizeType(actual)
	expectedType := catalog.NormalizeType(expected)
	if expectedType == "" {
		return true
	}
	return actualType == expectedType
}

// CloneListOutput copies one list output and the plugin rows it owns.
func CloneListOutput(in *ListOutput) *ListOutput {
	if in == nil {
		return nil
	}
	out := &ListOutput{
		List:  make([]*PluginItem, 0, len(in.List)),
		Total: in.Total,
	}
	for _, item := range in.List {
		out.List = append(out.List, ClonePluginItem(item))
	}
	return out
}

// ClonePluginItem copies one plugin item while preserving immutable nested
// projections by value where callers may otherwise mutate list rows.
func ClonePluginItem(in *PluginItem) *PluginItem {
	if in == nil {
		return nil
	}
	out := *in
	if in.LastUpgradeFailure != nil {
		lastUpgradeFailure := *in.LastUpgradeFailure
		out.LastUpgradeFailure = &lastUpgradeFailure
	}
	out.RequestedHostServices = cloneHostServiceSpecs(in.RequestedHostServices)
	out.AuthorizedHostServices = cloneHostServiceSpecs(in.AuthorizedHostServices)
	out.DeclaredRoutes = cloneRouteContracts(in.DeclaredRoutes)
	out.DependencyCheck = plugindep.CloneCheckProjection(in.DependencyCheck)
	return &out
}

// cloneHostServiceSpecs deep-copies host-service declarations because list
// consumers may reuse rows while action modals are open.
func cloneHostServiceSpecs(in []*protocol.HostServiceSpec) []*protocol.HostServiceSpec {
	if len(in) == 0 {
		return nil
	}
	out := make([]*protocol.HostServiceSpec, 0, len(in))
	for _, item := range in {
		if item == nil {
			continue
		}
		out = append(out, &protocol.HostServiceSpec{
			Service:   item.Service,
			Methods:   append([]string(nil), item.Methods...),
			Paths:     append([]string(nil), item.Paths...),
			Tables:    append([]string(nil), item.Tables...),
			Keys:      append([]string(nil), item.Keys...),
			Resources: cloneHostServiceResources(item.Resources),
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// cloneHostServiceResources deep-copies governed host-service resource specs.
func cloneHostServiceResources(in []*protocol.HostServiceResourceSpec) []*protocol.HostServiceResourceSpec {
	if len(in) == 0 {
		return nil
	}
	out := make([]*protocol.HostServiceResourceSpec, 0, len(in))
	for _, item := range in {
		if item == nil {
			continue
		}
		out = append(out, &protocol.HostServiceResourceSpec{
			Ref:             item.Ref,
			AllowMethods:    append([]string(nil), item.AllowMethods...),
			HeaderAllowList: append([]string(nil), item.HeaderAllowList...),
			TimeoutMs:       item.TimeoutMs,
			MaxBodyBytes:    item.MaxBodyBytes,
			Attributes:      cloneStringMap(item.Attributes),
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// cloneRouteContracts deep-copies route review declarations.
func cloneRouteContracts(in []*protocol.RouteContract) []*protocol.RouteContract {
	if len(in) == 0 {
		return nil
	}
	out := make([]*protocol.RouteContract, 0, len(in))
	for _, item := range in {
		if item == nil {
			continue
		}
		out = append(out, &protocol.RouteContract{
			Path:        item.Path,
			Method:      item.Method,
			Tags:        append([]string(nil), item.Tags...),
			Summary:     item.Summary,
			Description: item.Description,
			Access:      item.Access,
			Permission:  item.Permission,
			Meta:        cloneStringMap(item.Meta),
			RequestType: item.RequestType,
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// cloneStringMap copies string maps used by cached list projections.
func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

// manifestSnapshotContextKey stores one request-local manifest discovery result.
type manifestSnapshotContextKey struct{}
