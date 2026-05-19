// This file covers source-plugin consumer frontend asset helpers exposed by the
// plugin facade without requiring a full plugin lifecycle database fixture.

package plugin

import (
	"context"
	"errors"
	"strings"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
)

// TestNormalizeSourceConsumerFrontendAssetPath verifies browser-facing paths
// map into the source plugin's frontend/consumer directory and reject escapes.
func TestNormalizeSourceConsumerFrontendAssetPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{name: "root defaults to index", path: "", want: "frontend/consumer/index.html"},
		{name: "plain asset", path: "assets/app.js", want: "frontend/consumer/assets/app.js"},
		{name: "already prefixed", path: "frontend/consumer/index.html", want: "frontend/consumer/index.html"},
		{name: "escape rejected", path: "../plugin.yaml", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeSourceConsumerFrontendAssetPath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for path %q", tt.path)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

// TestSourceConsumerFrontendAssetDeclared verifies asset declaration lookup uses
// exact plugin-relative paths.
func TestSourceConsumerFrontendAssetDeclared(t *testing.T) {
	paths := []string{"frontend/consumer/index.html", "frontend/consumer/assets/app.js"}
	if !sourceConsumerFrontendAssetDeclared(paths, "frontend/consumer/assets/app.js") {
		t.Fatalf("expected declared asset to match")
	}
	if sourceConsumerFrontendAssetDeclared(paths, "frontend/consumer/assets/missing.js") {
		t.Fatalf("expected missing asset not to match")
	}
}

// TestMatchSourceConsumerFrontendMountPath verifies mounted consumer routes map
// to plugin-relative asset paths while unrelated paths are ignored.
func TestMatchSourceConsumerFrontendMountPath(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		mount  string
		want   string
		wantOK bool
	}{
		{name: "mount root", path: "/portal", mount: "/portal", wantOK: true},
		{name: "mount child asset", path: "/portal/assets/app.js", mount: "/portal", want: "assets/app.js", wantOK: true},
		{name: "trailing slash mount", path: "/portal/login", mount: "/portal/", want: "login", wantOK: true},
		{name: "prefix mismatch", path: "/portal-admin", mount: "/portal", wantOK: false},
		{name: "unrelated", path: "/assets/app.js", mount: "/portal", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := matchSourceConsumerFrontendMountPath(tt.path, tt.mount)
			if ok != tt.wantOK {
				t.Fatalf("expected ok=%v, got %v", tt.wantOK, ok)
			}
			if got != tt.want {
				t.Fatalf("expected relative path %q, got %q", tt.want, got)
			}
		})
	}
}

// TestRewriteSourceConsumerHTMLBase injects the stable mount base for clean SPA routes.
func TestRewriteSourceConsumerHTMLBase(t *testing.T) {
	out := rewriteSourceConsumerHTMLBase(
		&SourceConsumerFrontendMountAssetOutput{
			Content:     []byte("<!doctype html><html><HEAD data-app=\"portal\"><title>Portal</title></HEAD><body></body></html>"),
			ContentType: "text/html; charset=utf-8",
		},
		"/portal",
	)
	if out == nil {
		t.Fatalf("expected rewritten output")
	}
	if !strings.Contains(string(out.Content), `<base href="/portal/" />`) {
		t.Fatalf("expected base href to be injected, got %s", string(out.Content))
	}
	if out.CacheControl != frontend.CacheControlRevalidate {
		t.Fatalf("expected HTML mount cache policy %q, got %q", frontend.CacheControlRevalidate, out.CacheControl)
	}
	if strings.TrimSpace(out.ETag) == "" {
		t.Fatalf("expected rewritten HTML to include an ETag")
	}
}

// TestApplySourceConsumerMountAssetPolicyForcesRevalidation verifies stable
// mount entries use validator-based caching regardless of the source asset policy.
func TestApplySourceConsumerMountAssetPolicyForcesRevalidation(t *testing.T) {
	mount := &sourceConsumerFrontendMountEntry{
		pluginID:  "portal",
		version:   "1.0.0",
		mountPath: "/portal",
	}
	out := applySourceConsumerMountAssetPolicy(&SourceConsumerFrontendMountAssetOutput{
		Content:      []byte("body"),
		ContentType:  "text/html; charset=utf-8",
		ETag:         `"source-debug-etag"`,
		CacheControl: frontend.CacheControlStaticRevalidate,
	}, mount, "index.html")
	if out.CacheControl != frontend.CacheControlRevalidate {
		t.Fatalf("expected stable mount cache policy %q, got %q", frontend.CacheControlRevalidate, out.CacheControl)
	}
	if strings.TrimSpace(out.ETag) == "" || out.ETag == `"source-debug-etag"` {
		t.Fatalf("expected stable mount asset to get a mount-specific ETag, got %q", out.ETag)
	}
	otherMount := &sourceConsumerFrontendMountEntry{
		pluginID:  "portal",
		version:   "1.0.0",
		mountPath: "/site",
	}
	otherOut := applySourceConsumerMountAssetPolicy(&SourceConsumerFrontendMountAssetOutput{
		Content:     []byte("body"),
		ContentType: "text/html; charset=utf-8",
	}, otherMount, "index.html")
	if otherOut.ETag == out.ETag {
		t.Fatalf("expected different stable mounts to use different ETags")
	}
}

// TestCloneFrontendAssetOutputProtectsCachedBytes verifies preprocessed HTML
// entries can be returned from cache without exposing mutable state.
func TestCloneFrontendAssetOutputProtectsCachedBytes(t *testing.T) {
	original := &frontend.RuntimeFrontendAssetOutput{
		Content:      []byte("cached"),
		ContentType:  "text/html; charset=utf-8",
		ETag:         `"etag"`,
		CacheControl: frontend.CacheControlRevalidate,
	}
	cloned := cloneFrontendAssetOutput(original)
	if cloned == original {
		t.Fatalf("expected clone to be a distinct pointer")
	}
	cloned.Content[0] = 'C'
	if string(original.Content) != "cached" {
		t.Fatalf("expected clone mutation not to affect original, got %q", string(original.Content))
	}
}

// TestIsSourceConsumerFrontendMountNotFound keeps the exported no-match
// classifier stable for the host frontend catch-all.
func TestIsSourceConsumerFrontendMountNotFound(t *testing.T) {
	if !isSourceConsumerFrontendMountNotFound(
		errSourceConsumerFrontendMountNotFound,
	) {
		t.Fatalf("expected mount-not-found error to be classified")
	}
	if !isSourceConsumerFrontendMountNotFound(errors.Join(errors.New("wrapped"), errSourceConsumerFrontendMountNotFound)) {
		t.Fatalf("expected wrapped mount-not-found error to be classified")
	}
	if isSourceConsumerFrontendMountNotFound(errSourceConsumerFrontendMountDisabled) {
		t.Fatalf("expected non no-match error to stay blocking")
	}
}

// TestIsSourceConsumerFrontendMountAssetNotFound keeps missing asset
// classification stable for mounted static asset requests.
func TestIsSourceConsumerFrontendMountAssetNotFound(t *testing.T) {
	if !isSourceConsumerFrontendMountAssetNotFound(errSourceConsumerFrontendMountAssetNotFound) {
		t.Fatalf("expected sentinel asset-not-found error to be classified")
	}
	if !isSourceConsumerFrontendMountAssetNotFound(errors.Join(errors.New("wrapped"), errSourceConsumerFrontendMountAssetNotFound)) {
		t.Fatalf("expected wrapped asset-not-found error to be classified")
	}
	if isSourceConsumerFrontendMountAssetNotFound(errWithMessage("database is unavailable")) {
		t.Fatalf("expected unrelated error to stay unclassified")
	}
}

// TestLoadSourceConsumerFrontendMountEntriesCachesIndex verifies the mount index can
// be returned without rebuilding after the first successful load.
func TestLoadSourceConsumerFrontendMountEntriesCachesIndex(t *testing.T) {
	service := &serviceImpl{
		sourceConsumerFrontendIndexReady: true,
		sourceConsumerFrontendIndex: &sourceConsumerFrontendResourceIndex{
			mounts: []*sourceConsumerFrontendMountEntry{
				{
					pluginID:    "lina-portal",
					version:     "v0.1.0",
					mountPath:   "/portal",
					index:       "index.html",
					spaFallback: true,
					assets: map[string]struct{}{
						"frontend/consumer/index.html": {},
					},
					indexAsset: &frontend.RuntimeFrontendAssetOutput{
						Content:      []byte("cached index"),
						ContentType:  "text/html; charset=utf-8",
						ETag:         `"etag"`,
						CacheControl: frontend.CacheControlRevalidate,
					},
				},
			},
		},
	}

	mounts, err := service.loadSourceConsumerFrontendMountEntries(context.Background())
	if err != nil {
		t.Fatalf("expected cached mount index, got %v", err)
	}
	if len(mounts) != 1 || mounts[0].mountPath != "/portal" {
		t.Fatalf("unexpected cached mounts: %#v", mounts)
	}
	mounts[0].mountPath = "/changed"
	secondMounts, err := service.loadSourceConsumerFrontendMountEntries(context.Background())
	if err != nil {
		t.Fatalf("expected cached mount index on second read, got %v", err)
	}
	if secondMounts[0].mountPath != "/portal" {
		t.Fatalf("expected cached mount copy to protect service state, got %#v", secondMounts)
	}
	delete(mounts[0].assets, "frontend/consumer/index.html")
	thirdMounts, err := service.loadSourceConsumerFrontendMountEntries(context.Background())
	if err != nil {
		t.Fatalf("expected cached mount index on third read, got %v", err)
	}
	if !sourceConsumerFrontendAssetSet(thirdMounts[0].assets).has("frontend/consumer/index.html") {
		t.Fatalf("expected cached asset set copy to protect service state, got %#v", thirdMounts[0].assets)
	}
	thirdMounts[0].indexAsset.Content[0] = 'C'
	fourthMounts, err := service.loadSourceConsumerFrontendMountEntries(context.Background())
	if err != nil {
		t.Fatalf("expected cached mount index on fourth read, got %v", err)
	}
	if string(fourthMounts[0].indexAsset.Content) != "cached index" {
		t.Fatalf("expected cached index asset copy to protect service state, got %q", string(fourthMounts[0].indexAsset.Content))
	}
}

// TestInvalidateSourceConsumerFrontendMountsClearsIndex verifies lifecycle
// mutations force the next mount request to rebuild the process-local index.
func TestInvalidateSourceConsumerFrontendMountsClearsIndex(t *testing.T) {
	service := &serviceImpl{
		sourceConsumerFrontendIndexReady: true,
		sourceConsumerFrontendIndex: &sourceConsumerFrontendResourceIndex{
			mounts: []*sourceConsumerFrontendMountEntry{
				{pluginID: "lina-portal", version: "v0.1.0", mountPath: "/portal"},
			},
		},
	}

	service.invalidateSourceConsumerFrontendMounts()
	if service.sourceConsumerFrontendIndexReady {
		t.Fatalf("expected mount index ready flag to be cleared")
	}
	if service.sourceConsumerFrontendIndex != nil {
		t.Fatalf("expected mount index entries to be cleared")
	}
}

// TestSourceConsumerFrontendResourceIndexMatchIgnoresSiblingPrefixes verifies
// mount matching does not treat same-prefix sibling paths as nested routes.
func TestSourceConsumerFrontendResourceIndexMatchIgnoresSiblingPrefixes(t *testing.T) {
	index := &sourceConsumerFrontendResourceIndex{
		mounts: []*sourceConsumerFrontendMountEntry{
			{pluginID: "portal", mountPath: "/portal"},
			{pluginID: "portal-admin", mountPath: "/portal-admin"},
		},
	}

	mount, relativePath := index.match("/portal-admin/settings")
	if mount == nil {
		t.Fatalf("expected request path to match resource index")
	}
	if mount.pluginID != "portal-admin" || relativePath != "settings" {
		t.Fatalf("expected sibling mount, got mount=%#v relative=%q", mount, relativePath)
	}
}

// TestFindSourceConsumerFrontendOverlappingMountRejectsNestedMounts verifies
// source plugins cannot nest C-side services inside each other's URL space.
func TestFindSourceConsumerFrontendOverlappingMountRejectsNestedMounts(t *testing.T) {
	seenMounts := map[string]string{"/portal": "portal"}

	conflictMountPath, conflictPluginID, exists := findSourceConsumerFrontendOverlappingMount(seenMounts, "/portal/admin")
	if !exists {
		t.Fatalf("expected nested mount to conflict")
	}
	if conflictMountPath != "/portal" || conflictPluginID != "portal" {
		t.Fatalf("unexpected conflict mount=%q plugin=%q", conflictMountPath, conflictPluginID)
	}

	if _, _, exists := findSourceConsumerFrontendOverlappingMount(seenMounts, "/portal-admin"); exists {
		t.Fatalf("expected sibling prefix mount to be allowed")
	}
}

// TestSourceConsumerSPAFallbackEnabledDefaultsFalse verifies plugins must opt
// in before clean routes fall back to the static consumer frontend entry.
func TestSourceConsumerSPAFallbackEnabledDefaultsFalse(t *testing.T) {
	if sourceConsumerSPAFallbackEnabled(nil) {
		t.Fatalf("expected missing frontend spec to default SPA fallback off")
	}
	enabled := true
	if !sourceConsumerSPAFallbackEnabled(&catalog.ConsumerFrontendSpec{SPAFallback: &enabled}) {
		t.Fatalf("expected explicit enabled SPA fallback to be honored")
	}
	disabled := false
	if sourceConsumerSPAFallbackEnabled(&catalog.ConsumerFrontendSpec{SPAFallback: &disabled}) {
		t.Fatalf("expected explicit disabled SPA fallback to be honored")
	}
}

// TestActiveSourceConsumerFrontendSpecFiltersNonSourceManifests
// verifies only source-plugin consumer frontend declarations become mounts.
func TestActiveSourceConsumerFrontendSpecFiltersNonSourceManifests(t *testing.T) {
	if activeSourceConsumerFrontendSpec(&catalog.Manifest{Type: catalog.TypeDynamic.String()}) != nil {
		t.Fatalf("expected dynamic plugin manifest to be ignored")
	}

	sourceManifest := &catalog.Manifest{
		Type: catalog.TypeSource.String(),
		Consumer: &catalog.ConsumerSpec{Frontend: &catalog.ConsumerFrontendSpec{
			MountPath: "portal",
		}},
	}
	frontendSpec := activeSourceConsumerFrontendSpec(sourceManifest)
	if frontendSpec == nil {
		t.Fatalf("expected source consumer frontend declaration to be active")
	}
	if frontendSpec.MountPath != "/portal" || frontendSpec.Index != "index.html" {
		t.Fatalf("expected normalized frontend spec, got %#v", frontendSpec)
	}
}

// TestLooksLikeSourceConsumerStaticAsset verifies concrete asset misses are not
// treated as clean SPA routes.
func TestLooksLikeSourceConsumerStaticAsset(t *testing.T) {
	if !looksLikeSourceConsumerStaticAsset("assets/app.js") {
		t.Fatalf("expected JavaScript path to be treated as a static asset")
	}
	if looksLikeSourceConsumerStaticAsset("login") {
		t.Fatalf("expected clean route to be eligible for SPA fallback")
	}
}

// errWithMessage is a tiny test error for classifier coverage.
type errWithMessage string

// Error returns the configured test error message.
func (e errWithMessage) Error() string {
	return string(e)
}
