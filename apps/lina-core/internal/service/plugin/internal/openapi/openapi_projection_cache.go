// This file owns the process-local OpenAPI dynamic-route projection cache. The
// cached value is only the plugin-owned route path subset and is partitioned by
// plugin-runtime revision plus request locale state.

package openapi

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/net/goai"
)

// projectionCacheKey partitions dynamic route docs by runtime and language state.
type projectionCacheKey struct {
	runtimeRevision      int64
	locale               string
	runtimeBundleVersion uint64
}

// projectionCache stores cloned dynamic route path projections.
type projectionCache struct {
	mu     sync.RWMutex
	values map[string]goai.Paths
}

// newProjectionCache creates an empty projection cache.
func newProjectionCache() *projectionCache {
	return &projectionCache{}
}

// openAPIProjectionCacheKey returns the current cache key for dynamic routes.
func (s *serviceImpl) openAPIProjectionCacheKey(ctx context.Context) (projectionCacheKey, error) {
	key := projectionCacheKey{}
	if s != nil && s.revisionReader != nil {
		revision, err := s.revisionReader.CurrentRevision(ctx)
		if err != nil {
			return projectionCacheKey{}, err
		}
		key.runtimeRevision = revision
	}
	key.locale = normalizeProjectionLocale("")
	if s != nil && s.localeReader != nil {
		key.locale = normalizeProjectionLocale(s.localeReader.GetLocale(ctx))
		key.runtimeBundleVersion = s.localeReader.BundleVersion(key.locale)
	}
	return key, nil
}

// String returns the stable map key for one projection cache partition.
func (k projectionCacheKey) String() string {
	return strings.Join([]string{
		strconv.FormatInt(k.runtimeRevision, 10),
		k.locale,
		strconv.FormatUint(k.runtimeBundleVersion, 10),
	}, "@")
}

// get returns a detached cached projection.
func (c *projectionCache) get(key projectionCacheKey) (goai.Paths, bool) {
	if c == nil {
		return nil, false
	}
	c.mu.RLock()
	value := c.values[key.String()]
	c.mu.RUnlock()
	if value == nil {
		return nil, false
	}
	return clonePaths(value), true
}

// store records a detached projection.
func (c *projectionCache) store(key projectionCacheKey, paths goai.Paths) {
	if c == nil {
		return
	}
	cloned := clonePaths(paths)
	c.mu.Lock()
	if c.values == nil {
		c.values = make(map[string]goai.Paths)
	}
	c.values[key.String()] = cloned
	c.mu.Unlock()
}

// invalidate clears all dynamic route projections for this process. Cluster
// peers call the same method after observing a plugin-runtime revision change.
func (c *projectionCache) invalidate() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.values = nil
	c.mu.Unlock()
}

// InvalidateProjectionCache clears cached dynamic route OpenAPI projections.
func (s *serviceImpl) InvalidateProjectionCache(_ context.Context, _ string) {
	if s == nil || s.cache == nil {
		return
	}
	s.cache.invalidate()
}

// normalizeProjectionLocale keeps cache keys stable for startup and tests.
func normalizeProjectionLocale(locale string) string {
	trimmed := strings.TrimSpace(locale)
	if trimmed == "" {
		return DefaultLocaleBundleReader{}.GetLocale(context.Background())
	}
	return trimmed
}

// clonePaths deep-copies the dynamic route OpenAPI path subset owned by this package.
func clonePaths(in goai.Paths) goai.Paths {
	if in == nil {
		return nil
	}
	out := make(goai.Paths, len(in))
	for path, item := range in {
		out[path] = clonePath(item)
	}
	return out
}

// clonePath copies path-level fields and operation pointers so later OpenAPI
// localization cannot mutate cached operations.
func clonePath(in goai.Path) goai.Path {
	out := in
	out.Servers = append(goai.Servers(nil), in.Servers...)
	out.Parameters = append(goai.Parameters(nil), in.Parameters...)
	out.XExtensions = cloneXExtensions(in.XExtensions)
	out.Connect = cloneOperation(in.Connect)
	out.Delete = cloneOperation(in.Delete)
	out.Get = cloneOperation(in.Get)
	out.Head = cloneOperation(in.Head)
	out.Options = cloneOperation(in.Options)
	out.Patch = cloneOperation(in.Patch)
	out.Post = cloneOperation(in.Post)
	out.Put = cloneOperation(in.Put)
	out.Trace = cloneOperation(in.Trace)
	return out
}

// cloneOperation copies the operation fields populated by dynamic route projection.
func cloneOperation(in *goai.Operation) *goai.Operation {
	if in == nil {
		return nil
	}
	out := *in
	out.Tags = append([]string(nil), in.Tags...)
	out.Parameters = append(goai.Parameters(nil), in.Parameters...)
	out.Responses = cloneResponses(in.Responses)
	out.XExtensions = cloneXExtensions(in.XExtensions)
	if in.Security != nil {
		security := append(goai.SecurityRequirements(nil), (*in.Security)...)
		out.Security = &security
	}
	if in.Servers != nil {
		servers := append(goai.Servers(nil), (*in.Servers)...)
		out.Servers = &servers
	}
	return &out
}

// cloneResponses copies response refs used by dynamic route placeholder docs.
func cloneResponses(in goai.Responses) goai.Responses {
	if in == nil {
		return nil
	}
	out := make(goai.Responses, len(in))
	for key, value := range in {
		if value.Value != nil {
			response := *value.Value
			response.XExtensions = cloneXExtensions(value.Value.XExtensions)
			value.Value = &response
		}
		out[key] = value
	}
	return out
}

// cloneXExtensions copies GoFrame extension values.
func cloneXExtensions(in goai.XExtensions) goai.XExtensions {
	if in == nil {
		return nil
	}
	out := make(goai.XExtensions, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
