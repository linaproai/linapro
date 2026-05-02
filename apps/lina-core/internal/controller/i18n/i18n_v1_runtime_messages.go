// This file handles the runtime i18n message-bundle endpoint and the per-locale
// ETag protocol that lets the workbench skip transferring the whole catalog
// when its persisted bundle is still current.

package i18n

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/i18n/v1"
)

const (
	// runtimeMessagesETagHeader is the response header carrying the per-locale
	// bundle version identifier used for HTTP cache revalidation.
	runtimeMessagesETagHeader = "ETag"
	// runtimeMessagesIfNoneMatchHeader is the request header the workbench sends
	// to revalidate a previously persisted bundle.
	runtimeMessagesIfNoneMatchHeader = "If-None-Match"
	// runtimeMessagesCacheControlHeader sets the recommended caching policy for
	// runtime translation packages: per-user privacy with mandatory revalidation.
	runtimeMessagesCacheControlHeader = "Cache-Control"
	// runtimeMessagesCacheControlValue declares per-user, must-revalidate caching.
	runtimeMessagesCacheControlValue = "private, must-revalidate"
)

// RuntimeMessages returns the aggregated runtime translation bundle for one
// locale. The handler emits an ETag derived from the per-locale bundle version
// and short-circuits to 304 when the client's If-None-Match header matches,
// skipping the bundle clone and nesting work entirely on the revalidation path.
func (c *ControllerV1) RuntimeMessages(ctx context.Context, req *v1.RuntimeMessagesReq) (res *v1.RuntimeMessagesRes, err error) {
	locale := c.localeResolver.ResolveLocale(ctx, req.Lang)
	if err = c.bundleProvider.EnsureRuntimeBundleCacheFresh(ctx); err != nil {
		return nil, err
	}

	r := g.RequestFromCtx(ctx)
	// BundleVersion is 0 only when the locale entry has never been built.
	// In that case every client must receive a fresh 200 to populate its
	// persisted cache; we therefore skip 304 short-circuiting until the
	// merged catalog has been initialized at least once.
	if r != nil {
		preBuildVersion := c.bundleProvider.BundleVersion(locale)
		if preBuildVersion > 0 {
			etag := buildRuntimeMessagesETag(locale, preBuildVersion)
			if matchesIfNoneMatch(r.Header.Get(runtimeMessagesIfNoneMatchHeader), etag) {
				r.Response.Header().Set(runtimeMessagesCacheControlHeader, runtimeMessagesCacheControlValue)
				r.Response.Header().Set(runtimeMessagesETagHeader, etag)
				r.Response.Status = http.StatusNotModified
				return nil, nil
			}
		}
	}

	// BuildRuntimeMessages primes the merged catalog so the post-build version
	// always reflects every sector that participated in this response body.
	messages := c.bundleProvider.BuildRuntimeMessages(ctx, locale)
	etag := buildRuntimeMessagesETag(locale, c.bundleProvider.BundleVersion(locale))
	if r != nil {
		r.Response.Header().Set(runtimeMessagesCacheControlHeader, runtimeMessagesCacheControlValue)
		r.Response.Header().Set(runtimeMessagesETagHeader, etag)
	}
	return &v1.RuntimeMessagesRes{
		Locale:   locale,
		Messages: messages,
	}, nil
}

// buildRuntimeMessagesETag formats a strong ETag value such as `"en-US-42"`.
// Strong validators are required because the bundle bytes are byte-stable for
// a given (locale, version) pair: the merged catalog is rebuilt deterministically
// and re-serialized in sorted order.
func buildRuntimeMessagesETag(locale string, version uint64) string {
	return `"` + locale + "-" + strconv.FormatUint(version, 10) + `"`
}

// matchesIfNoneMatch reports whether one If-None-Match header value matches
// the server-side ETag. It accepts both the exact ETag string and the special
// `*` wildcard used by some clients to mean "any current representation".
func matchesIfNoneMatch(headerValue string, etag string) bool {
	trimmedHeader := strings.TrimSpace(headerValue)
	if trimmedHeader == "" {
		return false
	}
	for _, candidate := range strings.Split(trimmedHeader, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if candidate == "*" || candidate == etag {
			return true
		}
	}
	return false
}
