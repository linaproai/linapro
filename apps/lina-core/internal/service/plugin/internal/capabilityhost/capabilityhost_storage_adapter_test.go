// This file verifies plugin storage capability adapters and providers keep
// durable object storage behind the unified storagecap contract.

package capabilityhost

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"testing"

	"lina-core/internal/service/hostlock"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/locker"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/storagecap"
)

// TestLocalStorageProviderListsFromPrefixWithLimit verifies local provider
// listing is bounded and starts from the requested prefix instead of scanning
// unrelated plugin object roots.
func TestLocalStorageProviderListsFromPrefixWithLimit(t *testing.T) {
	ctx := context.Background()
	provider := NewLocalStorageProvider(t.TempDir())

	writeProviderObject(t, ctx, provider, "plugins/reporting/platform/reports/a.json", "a")
	writeProviderObject(t, ctx, provider, "plugins/reporting/platform/reports/b.json", "b")
	writeProviderObject(t, ctx, provider, "plugins/reporting/platform/reports/c.json", "c")
	writeProviderObject(t, ctx, provider, "plugins/reporting/platform/private/hidden.json", "hidden")
	writeProviderObject(t, ctx, provider, "plugins/other/platform/reports/other.json", "other")

	output, err := provider.List(ctx, storagecap.ProviderListInput{
		Prefix: "plugins/reporting/platform/reports",
		Limit:  2,
	})
	if err != nil {
		t.Fatalf("list local provider: %v", err)
	}
	got := providerObjectKeys(output.Objects)
	if len(got) != 2 {
		t.Fatalf("expected bounded list of 2 objects, got %#v", got)
	}
	for _, key := range got {
		if !strings.HasPrefix(key, "plugins/reporting/platform/reports/") {
			t.Fatalf("expected only reports prefix objects, got %#v", got)
		}
	}
}

// TestStorageAdapterSelectsActivePluginProvider verifies scoped Storage()
// delegates to the unique enabled plugin provider.
func TestStorageAdapterSelectsActivePluginProvider(t *testing.T) {
	ctx := context.Background()
	providerID := fmt.Sprintf("storage-provider-test-%d", storageProviderTestSequence())
	provider := &storageProviderTestProvider{}
	registerStorageProviderForTest(t, providerID, provider)

	services := newStorageAdapterTestDirectory(t, &storageProviderTestRuntime{
		available: map[string]bool{providerID: true},
	}, NewLocalStorageProvider(t.TempDir()))
	storageSvc := capability.ServicesForPlugin(services, "reporting").Storage()

	_, err := storageSvc.Put(ctx, storagecap.PutInput{
		Path:        "reports/a.json",
		Body:        strings.NewReader("a"),
		ContentType: "application/json",
	})
	if err != nil {
		t.Fatalf("put through active plugin provider: %v", err)
	}
	if provider.putKey != "plugins/reporting/platform/reports/a.json" {
		t.Fatalf("expected plugin-scoped provider key, got %q", provider.putKey)
	}

	statuses, err := storageSvc.ProviderStatuses(ctx)
	if err != nil {
		t.Fatalf("provider statuses: %v", err)
	}
	status := storageProviderStatusByID(statuses, providerID)
	if status == nil || !status.Active || !status.Available {
		t.Fatalf("expected active plugin provider status, got %#v in %#v", status, statuses)
	}
}

// TestStorageAdapterFallsBackToLocalWhenProviderUnavailable verifies disabled
// provider plugins do not override the built-in local fallback.
func TestStorageAdapterFallsBackToLocalWhenProviderUnavailable(t *testing.T) {
	ctx := context.Background()
	providerID := fmt.Sprintf("storage-provider-unavailable-test-%d", storageProviderTestSequence())
	registerStorageProviderForTest(t, providerID, &storageProviderTestProvider{})
	localProvider := &storageProviderTestProvider{}

	services := newStorageAdapterTestDirectory(t, &storageProviderTestRuntime{
		available: map[string]bool{providerID: false},
	}, localProvider)
	storageSvc := capability.ServicesForPlugin(services, "reporting").Storage()

	_, err := storageSvc.Put(ctx, storagecap.PutInput{
		Path: "reports/a.json",
		Body: strings.NewReader("a"),
	})
	if err != nil {
		t.Fatalf("put through local provider fallback: %v", err)
	}
	if localProvider.putKey != "plugins/reporting/platform/reports/a.json" {
		t.Fatalf("expected local provider key, got %q", localProvider.putKey)
	}
}

// TestStorageAdapterUsesLocalProviderByDefault verifies empty active provider
// configuration selects the built-in local provider.
func TestStorageAdapterUsesLocalProviderByDefault(t *testing.T) {
	ctx := context.Background()
	localProvider := &storageProviderTestProvider{}
	services := newStorageAdapterTestDirectory(t, &storageProviderTestRuntime{}, localProvider)
	storageSvc := capability.ServicesForPlugin(services, "reporting").Storage()

	output, err := storageSvc.Put(ctx, storagecap.PutInput{
		Path: "reports/a.json",
		Body: strings.NewReader("a"),
	})
	if err != nil {
		t.Fatalf("put through local provider: %v", err)
	}
	if output == nil || output.Object == nil || output.Object.Path != "reports/a.json" {
		t.Fatalf("unexpected local provider output: %#v", output)
	}
	if localProvider.putKey != "plugins/reporting/platform/reports/a.json" {
		t.Fatalf("expected local provider key, got %q", localProvider.putKey)
	}
}

// TestStorageAdapterRejectsMultipleEnabledProviders verifies provider selection
// fails when more than one storage provider plugin is serviceable.
func TestStorageAdapterRejectsMultipleEnabledProviders(t *testing.T) {
	ctx := context.Background()
	providerAID := fmt.Sprintf("storage-provider-conflict-a-%d", storageProviderTestSequence())
	providerBID := fmt.Sprintf("storage-provider-conflict-b-%d", storageProviderTestSequence())
	registerStorageProviderForTest(t, providerAID, &storageProviderTestProvider{})
	registerStorageProviderForTest(t, providerBID, &storageProviderTestProvider{})

	services := newStorageAdapterTestDirectory(t, &storageProviderTestRuntime{
		available: map[string]bool{
			providerAID: true,
			providerBID: true,
		},
	}, &storageProviderTestProvider{})
	storageSvc := capability.ServicesForPlugin(services, "reporting").Storage()

	_, err := storageSvc.Put(ctx, storagecap.PutInput{
		Path: "reports/a.json",
		Body: strings.NewReader("a"),
	})
	if !bizerr.Is(err, storagecap.CodeStorageProviderConflict) {
		t.Fatalf("expected multiple provider conflict error, got %v", err)
	}

	statuses, err := storageSvc.ProviderStatuses(ctx)
	if err != nil {
		t.Fatalf("provider statuses: %v", err)
	}
	for _, providerID := range []string{providerAID, providerBID} {
		status := storageProviderStatusByID(statuses, providerID)
		if status == nil || status.Active || !status.Available || status.Message == "" {
			t.Fatalf("expected conflicting provider status for %s, got %#v in %#v", providerID, status, statuses)
		}
	}
}

// TestStorageAdapterStreamsPutWithoutFixedObjectLimit verifies Storage() no
// longer rejects writes at the adapter layer based on a fixed object-size cap.
func TestStorageAdapterStreamsPutWithoutFixedObjectLimit(t *testing.T) {
	ctx := context.Background()
	localProvider := &storageProviderTestProvider{}
	storageSvc := newStorageAdapter(nil, localProvider, nil, "reporting")
	declaredSize := int64(64 * 1024 * 1024)

	output, err := storageSvc.Put(ctx, storagecap.PutInput{
		Path: "reports/large.bin",
		Body: strings.NewReader("large body"),
		Size: declaredSize,
	})
	if err != nil {
		t.Fatalf("put with large declared size: %v", err)
	}
	if output == nil || output.Object == nil || output.Object.Size != int64(len("large body")) {
		t.Fatalf("unexpected object output: %#v", output)
	}
	if localProvider.putSize != declaredSize {
		t.Fatalf("expected provider size %d, got %d", declaredSize, localProvider.putSize)
	}
}

// TestStorageAdapterContentTypeProbePreservesBody verifies MIME sniffing reads
// only a prefix and passes the full object stream to the provider.
func TestStorageAdapterContentTypeProbePreservesBody(t *testing.T) {
	ctx := context.Background()
	localProvider := &storageProviderTestProvider{}
	storageSvc := newStorageAdapter(nil, localProvider, nil, "reporting")
	body := strings.Repeat("a", objectContentTypeProbeBytes+32)

	_, err := storageSvc.Put(ctx, storagecap.PutInput{
		Path: "reports/plain",
		Body: strings.NewReader(body),
	})
	if err != nil {
		t.Fatalf("put with sniffed body: %v", err)
	}
	if got := string(localProvider.objects["plugins/reporting/platform/reports/plain"]); got != body {
		t.Fatalf("expected preserved body length %d, got length %d", len(body), len(got))
	}
	if localProvider.putContentType != "text/plain" {
		t.Fatalf("expected sniffed text/plain content type, got %q", localProvider.putContentType)
	}
}

// TestStorageAdapterBatchMethodsUseScopedProviderKeys verifies batch storage
// operations keep provider keys inside the plugin and tenant scope.
func TestStorageAdapterBatchMethodsUseScopedProviderKeys(t *testing.T) {
	ctx := bizctxcap.WithCurrentContext(context.Background(), bizctxcap.CurrentContext{TenantID: 42})
	localProvider := &storageProviderTestProvider{}
	storageSvc := newStorageAdapter(nil, localProvider, nil, "reporting")

	_, err := storageSvc.Put(ctx, storagecap.PutInput{Path: "reports/a.json", Body: strings.NewReader("a")})
	if err != nil {
		t.Fatalf("put a: %v", err)
	}
	_, err = storageSvc.Put(ctx, storagecap.PutInput{Path: "reports/b.json", Body: strings.NewReader("b")})
	if err != nil {
		t.Fatalf("put b: %v", err)
	}

	statOutput, err := storageSvc.BatchStat(ctx, storagecap.BatchStatInput{
		Paths: []string{"reports/a.json", "reports/missing.json"},
	})
	if err != nil {
		t.Fatalf("batch stat: %v", err)
	}
	if len(statOutput.Objects) != 1 || statOutput.Objects[0].Path != "reports/a.json" {
		t.Fatalf("unexpected batch stat objects: %#v", statOutput.Objects)
	}
	if strings.Join(statOutput.MissingPaths, ",") != "reports/missing.json" {
		t.Fatalf("unexpected missing paths: %#v", statOutput.MissingPaths)
	}

	if err = storageSvc.DeleteMany(ctx, storagecap.DeleteManyInput{Paths: []string{"reports/a.json", "reports/b.json"}}); err != nil {
		t.Fatalf("delete many: %v", err)
	}
	if len(localProvider.objects) != 0 {
		t.Fatalf("expected scoped objects deleted, got %#v", localProvider.objects)
	}
}

// TestStorageAdapterPrefersExplicitPluginTenantContext verifies lifecycle and
// plugin cleanup code can bind the target tenant even when the original host
// context still carries another tenant.
func TestStorageAdapterPrefersExplicitPluginTenantContext(t *testing.T) {
	ctx := bizctxcap.WithCurrentContext(context.Background(), bizctxcap.CurrentContext{TenantID: 2002})
	localProvider := &storageProviderTestProvider{}
	storageSvc := newStorageAdapter(
		nil,
		localProvider,
		storageAdapterTestBizCtx{current: bizctxcap.CurrentContext{TenantID: 1001}},
		"reporting",
	)

	_, err := storageSvc.Put(ctx, storagecap.PutInput{
		Path: "reports/a.json",
		Body: strings.NewReader("a"),
	})
	if err != nil {
		t.Fatalf("put with explicit plugin tenant context: %v", err)
	}
	if localProvider.putKey != "plugins/reporting/tenant/2002/reports/a.json" {
		t.Fatalf("expected explicit tenant provider key, got %q", localProvider.putKey)
	}
}

func writeProviderObject(
	t *testing.T,
	ctx context.Context,
	provider storagecap.Provider,
	key string,
	body string,
) {
	t.Helper()
	if _, err := provider.Put(ctx, storagecap.ProviderPutInput{
		Key:       key,
		Body:      strings.NewReader(body),
		Overwrite: true,
	}); err != nil {
		t.Fatalf("write provider object %s: %v", key, err)
	}
}

func providerObjectKeys(objects []*storagecap.ProviderObject) []string {
	keys := make([]string, 0, len(objects))
	for _, object := range objects {
		if object != nil {
			keys = append(keys, object.Key)
		}
	}
	sort.Strings(keys)
	return keys
}

func newStorageAdapterTestDirectory(
	t *testing.T,
	runtime storagecap.ProviderRuntime,
	localProvider storagecap.Provider,
) capability.Services {
	t.Helper()
	services, err := New(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		kvcache.New(),
		newStorageAdapterTestLockService(t),
		runtime,
		localProvider,
	)
	if err != nil {
		t.Fatalf("create storage adapter test services: %v", err)
	}
	return services
}

func newStorageAdapterTestLockService(t *testing.T) hostlock.Service {
	t.Helper()
	service, err := hostlock.New(locker.New())
	if err != nil {
		t.Fatalf("create storage adapter test lock service: %v", err)
	}
	return service
}

type storageProviderTestRuntime struct {
	available map[string]bool
}

func (r *storageProviderTestRuntime) ProviderPluginAvailable(_ context.Context, pluginID string) bool {
	if r == nil {
		return false
	}
	return r.available[strings.TrimSpace(pluginID)]
}

type storageAdapterTestBizCtx struct {
	current bizctxcap.CurrentContext
}

func (s storageAdapterTestBizCtx) Current(context.Context) bizctxcap.CurrentContext {
	return s.current
}

type storageProviderTestProvider struct {
	mu             sync.Mutex
	objects        map[string][]byte
	putKey         string
	putSize        int64
	putContentType string
	listKeys       []string
}

func (p *storageProviderTestProvider) Put(_ context.Context, in storagecap.ProviderPutInput) (*storagecap.ProviderObject, error) {
	body, err := io.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ensureObjects()
	p.putKey = in.Key
	p.putSize = in.Size
	p.putContentType = strings.TrimSpace(in.ContentType)
	p.objects[in.Key] = append([]byte(nil), body...)
	return &storagecap.ProviderObject{
		Key:         in.Key,
		Size:        int64(len(body)),
		ContentType: strings.TrimSpace(in.ContentType),
		Visibility:  storagecap.VisibilityPrivate,
	}, nil
}

func (p *storageProviderTestProvider) Get(_ context.Context, in storagecap.ProviderGetInput) (*storagecap.ProviderGetOutput, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ensureObjects()
	body, ok := p.objects[in.Key]
	if !ok {
		return &storagecap.ProviderGetOutput{Found: false}, nil
	}
	return &storagecap.ProviderGetOutput{
		Object: &storagecap.ProviderObject{
			Key:        in.Key,
			Size:       int64(len(body)),
			Visibility: storagecap.VisibilityPrivate,
		},
		Body:  io.NopCloser(bytes.NewReader(append([]byte(nil), body...))),
		Found: true,
	}, nil
}

func (p *storageProviderTestProvider) Delete(_ context.Context, in storagecap.ProviderDeleteInput) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ensureObjects()
	delete(p.objects, in.Key)
	return nil
}

func (p *storageProviderTestProvider) DeleteMany(_ context.Context, in storagecap.ProviderDeleteManyInput) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ensureObjects()
	for _, key := range in.Keys {
		delete(p.objects, key)
	}
	return nil
}

func (p *storageProviderTestProvider) List(_ context.Context, in storagecap.ProviderListInput) (*storagecap.ProviderListOutput, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ensureObjects()
	keys := make([]string, 0, len(p.objects))
	prefix := strings.TrimSuffix(in.Prefix, "/")
	for key := range p.objects {
		if key == prefix || strings.HasPrefix(key, prefix+"/") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	limit := in.Limit
	if limit <= 0 || limit > len(keys) {
		limit = len(keys)
	}
	keys = keys[:limit]
	p.listKeys = append([]string(nil), keys...)
	objects := make([]*storagecap.ProviderObject, 0, len(keys))
	for _, key := range keys {
		objects = append(objects, &storagecap.ProviderObject{
			Key:        key,
			Size:       int64(len(p.objects[key])),
			Visibility: storagecap.VisibilityPrivate,
		})
	}
	return &storagecap.ProviderListOutput{Objects: objects}, nil
}

func (p *storageProviderTestProvider) ListCursor(_ context.Context, in storagecap.ProviderListCursorInput) (*storagecap.ProviderListCursorOutput, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ensureObjects()
	keys := make([]string, 0, len(p.objects))
	prefix := strings.TrimSuffix(in.Prefix, "/")
	for key := range p.objects {
		if key == prefix || strings.HasPrefix(key, prefix+"/") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	cursor := strings.TrimSpace(in.Cursor)
	filtered := keys[:0]
	for _, key := range keys {
		if cursor == "" || key > cursor {
			filtered = append(filtered, key)
		}
	}
	limit := in.Limit
	if limit <= 0 || limit > len(filtered) {
		limit = len(filtered)
	}
	page := filtered[:limit]
	nextCursor := ""
	if limit < len(filtered) && len(page) > 0 {
		nextCursor = page[len(page)-1]
	}
	objects := make([]*storagecap.ProviderObject, 0, len(page))
	for _, key := range page {
		objects = append(objects, &storagecap.ProviderObject{
			Key:        key,
			Size:       int64(len(p.objects[key])),
			Visibility: storagecap.VisibilityPrivate,
		})
	}
	return &storagecap.ProviderListCursorOutput{Objects: objects, NextCursor: nextCursor}, nil
}

func (p *storageProviderTestProvider) Stat(_ context.Context, in storagecap.ProviderStatInput) (*storagecap.ProviderStatOutput, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ensureObjects()
	body, ok := p.objects[in.Key]
	if !ok {
		return &storagecap.ProviderStatOutput{Found: false}, nil
	}
	return &storagecap.ProviderStatOutput{
		Object: &storagecap.ProviderObject{
			Key:        in.Key,
			Size:       int64(len(body)),
			Visibility: storagecap.VisibilityPrivate,
		},
		Found: true,
	}, nil
}

func (p *storageProviderTestProvider) BatchStat(_ context.Context, in storagecap.ProviderBatchStatInput) (*storagecap.ProviderBatchStatOutput, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ensureObjects()
	output := &storagecap.ProviderBatchStatOutput{Objects: []*storagecap.ProviderObject{}}
	for _, key := range in.Keys {
		body, ok := p.objects[key]
		if !ok {
			output.MissingKeys = append(output.MissingKeys, key)
			continue
		}
		output.Objects = append(output.Objects, &storagecap.ProviderObject{
			Key:        key,
			Size:       int64(len(body)),
			Visibility: storagecap.VisibilityPrivate,
		})
	}
	return output, nil
}

func (p *storageProviderTestProvider) ensureObjects() {
	if p.objects == nil {
		p.objects = make(map[string][]byte)
	}
}

func storageProviderStatusByID(statuses []*storagecap.ProviderStatus, providerID string) *storagecap.ProviderStatus {
	for _, status := range statuses {
		if status != nil && status.ProviderID == providerID {
			return status
		}
	}
	return nil
}

var storageProviderTestCounter struct {
	sync.Mutex
	value int
}

func storageProviderTestSequence() int {
	storageProviderTestCounter.Lock()
	defer storageProviderTestCounter.Unlock()
	storageProviderTestCounter.value++
	return storageProviderTestCounter.value
}

func registerStorageProviderForTest(t *testing.T, providerID string, provider storagecap.Provider) {
	t.Helper()
	if err := storagecap.Provide(providerID, func(context.Context, storagecap.ProviderEnv) (storagecap.Provider, error) {
		return provider, nil
	}); err != nil {
		t.Fatalf("register storage provider %s: %v", providerID, err)
	}
}
