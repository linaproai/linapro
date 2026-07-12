// Package management owns plugin management read-model projections, caching,
// and request-local manifest snapshots used by the root plugin facade.
package management

import (
	"errors"
	"strconv"
	"strings"
	"sync"

	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/runtime"
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
	// IncludeBuiltin is retained for API compatibility and no longer filters list results.
	// Ordinary management lists always include distribution=builtin plugins.
	IncludeBuiltin bool
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
