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
	i18nsvc "lina-core/internal/service/i18n"
)

const (
	// runtimeMessagesETagHeader is the response header carrying the per-locale
	// bundle revision validator used for HTTP cache revalidation.
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
// locale. The handler emits an ETag derived from the cache-owned bundle
// revision and short-circuits to 304 before cloning and nesting the bundle when
// the client's If-None-Match header matches a warm cache revision.
func (c *ControllerV1) RuntimeMessages(ctx context.Context, req *v1.RuntimeMessagesReq) (res *v1.RuntimeMessagesRes, err error) {
	locale := c.localeResolver.ResolveLocale(ctx, req.Lang)
	if err = c.bundleProvider.EnsureRuntimeBundleCacheFresh(ctx); err != nil {
		return nil, err
	}

	r := g.RequestFromCtx(ctx)
	if r != nil {
		r.Response.Header().Set(runtimeMessagesCacheControlHeader, runtimeMessagesCacheControlValue)
		if etag, ok := buildRuntimeMessagesETag(locale, c.bundleProvider.BundleRevision(locale)); ok {
			r.Response.Header().Set(runtimeMessagesETagHeader, etag)
			if matchesIfNoneMatch(r.Header.Get(runtimeMessagesIfNoneMatchHeader), etag) {
				r.Response.Status = http.StatusNotModified
				return nil, nil
			}
		}
	}

	messages := c.bundleProvider.BuildRuntimeMessages(ctx, locale)
	if r != nil {
		if etag, ok := buildRuntimeMessagesETag(locale, c.bundleProvider.BundleRevision(locale)); ok {
			r.Response.Header().Set(runtimeMessagesETagHeader, etag)
			if matchesIfNoneMatch(r.Header.Get(runtimeMessagesIfNoneMatchHeader), etag) {
				r.Response.Status = http.StatusNotModified
				return nil, nil
			}
		}
	}
	return &v1.RuntimeMessagesRes{
		Locale:   locale,
		Messages: messages,
	}, nil
}

// buildRuntimeMessagesETag formats a strong ETag value such as
// `"en-US-42-bb8c546a83d93ed4bb8c546a83d93ed4"`. The content fingerprint is
// intentionally part of the validator because in-process bundle versions
// restart from zero after a process restart while embedded locale resources may
// have changed between builds.
func buildRuntimeMessagesETag(locale string, revision i18nsvc.RuntimeBundleRevision) (string, bool) {
	fingerprint := strings.TrimSpace(revision.Fingerprint)
	if fingerprint == "" {
		return "", false
	}
	return `"` + locale + "-" + strconv.FormatUint(revision.Version, 10) + "-" + fingerprint + `"`, true
}

// matchesIfNoneMatch reports whether one If-None-Match header value matches
// the server-side ETag. It accepts both the exact ETag string and the special
// `*` wildcard used by some clients to mean "any current representation".
func matchesIfNoneMatch(headerValue string, etag string) bool {
	if etag == "" {
		return false
	}
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
