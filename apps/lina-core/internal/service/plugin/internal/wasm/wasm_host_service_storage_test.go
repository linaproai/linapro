// This file tests storage host service authorization, path isolation, and
// logical path prefix matching through the unified storage capability service.

package wasm

import (
	"bytes"
	"context"
	"io"
	"sort"
	"strings"
	"sync"
	"testing"

	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// storageDomainTestService is an in-memory storagecap service used by WASM
// dispatcher tests. Provider behavior is tested separately in capabilityhost.
type storageDomainTestService struct {
	mu          sync.Mutex
	objects     map[string]*storageDomainTestObject
	putCalls    int
	getCalls    int
	deleteCalls int
	listCalls   int
	statCalls   int
	lastPath    string
	lastPrefix  string
	lastLimit   int
}

type storageDomainTestObject struct {
	body        []byte
	contentType string
}

// Put stores one plugin-visible object.
func (s *storageDomainTestService) Put(_ context.Context, in storagecap.PutInput) (*storagecap.PutOutput, error) {
	body, err := io.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureObjects()
	s.putCalls++
	s.lastPath = in.Path
	s.objects[in.Path] = &storageDomainTestObject{
		body:        append([]byte(nil), body...),
		contentType: strings.TrimSpace(in.ContentType),
	}
	return &storagecap.PutOutput{Object: s.objectMetadataLocked(in.Path)}, nil
}

// Get reads one plugin-visible object.
func (s *storageDomainTestService) Get(_ context.Context, in storagecap.GetInput) (*storagecap.GetOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureObjects()
	s.getCalls++
	s.lastPath = in.Path
	object, ok := s.objects[in.Path]
	if !ok {
		return &storagecap.GetOutput{Found: false}, nil
	}
	return &storagecap.GetOutput{
		Object: s.objectMetadataLocked(in.Path),
		Body:   io.NopCloser(bytes.NewReader(append([]byte(nil), object.body...))),
		Found:  true,
	}, nil
}

// Delete removes one plugin-visible object.
func (s *storageDomainTestService) Delete(_ context.Context, in storagecap.DeleteInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureObjects()
	s.deleteCalls++
	s.lastPath = in.Path
	delete(s.objects, in.Path)
	return nil
}

// List lists plugin-visible objects under a bounded prefix.
func (s *storageDomainTestService) List(_ context.Context, in storagecap.ListInput) (*storagecap.ListOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureObjects()
	s.listCalls++
	s.lastPrefix = in.Prefix
	limit := in.Limit
	if limit <= 0 {
		limit = storagecap.DefaultListLimit
	}
	if limit > storagecap.MaxListLimit {
		limit = storagecap.MaxListLimit
	}
	s.lastLimit = limit

	prefix := strings.TrimSuffix(strings.TrimSpace(in.Prefix), "/")
	keys := make([]string, 0, len(s.objects))
	for key := range s.objects {
		if key == prefix || strings.HasPrefix(key, prefix+"/") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if len(keys) > limit {
		keys = keys[:limit]
	}
	objects := make([]*storagecap.Object, 0, len(keys))
	for _, key := range keys {
		objects = append(objects, s.objectMetadataLocked(key))
	}
	return &storagecap.ListOutput{Objects: objects, Limit: limit}, nil
}

// Stat reads plugin-visible object metadata.
func (s *storageDomainTestService) Stat(_ context.Context, in storagecap.StatInput) (*storagecap.StatOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureObjects()
	s.statCalls++
	s.lastPath = in.Path
	if _, ok := s.objects[in.Path]; !ok {
		return &storagecap.StatOutput{Found: false}, nil
	}
	return &storagecap.StatOutput{Object: s.objectMetadataLocked(in.Path), Found: true}, nil
}

// ProviderStatuses returns a deterministic local-provider status snapshot.
func (*storageDomainTestService) ProviderStatuses(context.Context) ([]*storagecap.ProviderStatus, error) {
	return []*storagecap.ProviderStatus{{
		ProviderID: storagecap.LocalProviderID,
		Active:     true,
		Available:  true,
	}}, nil
}

func (s *storageDomainTestService) ensureObjects() {
	if s.objects == nil {
		s.objects = make(map[string]*storageDomainTestObject)
	}
}

func (s *storageDomainTestService) objectMetadataLocked(path string) *storagecap.Object {
	object := s.objects[path]
	if object == nil {
		return nil
	}
	return &storagecap.Object{
		Path:        path,
		Size:        int64(len(object.body)),
		ContentType: object.contentType,
		Visibility:  storagecap.VisibilityPrivate,
	}
}

// TestHandleHostServiceInvokeStorageLifecycle verifies storage put/get/list/
// delete/stat behavior through storagecap.Service.
func TestHandleHostServiceInvokeStorageLifecycle(t *testing.T) {
	storageSvc := &storageDomainTestService{}
	configureStorageDomainServiceForTest(t, storageSvc)

	authorizedPath := "reports/"
	hcc := newStorageHostCallContext([]string{authorizedPath})

	putResponse := invokeStorageHostService(
		t,
		hcc,
		protocol.HostServiceMethodStoragePut,
		"reports/demo.json",
		protocol.MarshalHostServiceStoragePutRequest(&protocol.HostServiceStoragePutRequest{
			Path:        "reports/demo.json",
			Body:        []byte(`{"ok":true}`),
			ContentType: "application/json",
			Overwrite:   false,
		}),
	)
	if putResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("put: expected success, got status=%d payload=%s", putResponse.Status, string(putResponse.Payload))
	}
	putPayload, err := protocol.UnmarshalHostServiceStoragePutResponse(putResponse.Payload)
	if err != nil {
		t.Fatalf("put payload decode failed: %v", err)
	}
	if putPayload.Object == nil || putPayload.Object.Path != "reports/demo.json" {
		t.Fatalf("put object: got %#v", putPayload.Object)
	}
	if storageSvc.putCalls != 1 || storageSvc.lastPath != "reports/demo.json" {
		t.Fatalf("expected put to use storagecap service, calls=%d path=%q", storageSvc.putCalls, storageSvc.lastPath)
	}

	getResponse := invokeStorageHostService(
		t,
		hcc,
		protocol.HostServiceMethodStorageGet,
		"reports/demo.json",
		protocol.MarshalHostServiceStorageGetRequest(&protocol.HostServiceStorageGetRequest{Path: "reports/demo.json"}),
	)
	if getResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("get: expected success, got status=%d payload=%s", getResponse.Status, string(getResponse.Payload))
	}
	getPayload, err := protocol.UnmarshalHostServiceStorageGetResponse(getResponse.Payload)
	if err != nil {
		t.Fatalf("get payload decode failed: %v", err)
	}
	if !getPayload.Found || string(getPayload.Body) != `{"ok":true}` {
		t.Fatalf("get payload: got %#v", getPayload)
	}

	listResponse := invokeStorageHostService(
		t,
		hcc,
		protocol.HostServiceMethodStorageList,
		"reports",
		protocol.MarshalHostServiceStorageListRequest(&protocol.HostServiceStorageListRequest{
			Prefix: "reports",
			Limit:  10,
		}),
	)
	if listResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("list: expected success, got status=%d payload=%s", listResponse.Status, string(listResponse.Payload))
	}
	listPayload, err := protocol.UnmarshalHostServiceStorageListResponse(listResponse.Payload)
	if err != nil {
		t.Fatalf("list payload decode failed: %v", err)
	}
	if len(listPayload.Objects) != 1 || listPayload.Objects[0].Path != "reports/demo.json" {
		t.Fatalf("list payload: got %#v", listPayload.Objects)
	}
	if storageSvc.listCalls != 1 || storageSvc.lastPrefix != "reports" || storageSvc.lastLimit != 10 {
		t.Fatalf("expected bounded list through storagecap, calls=%d prefix=%q limit=%d", storageSvc.listCalls, storageSvc.lastPrefix, storageSvc.lastLimit)
	}

	deleteResponse := invokeStorageHostService(
		t,
		hcc,
		protocol.HostServiceMethodStorageDelete,
		"reports/demo.json",
		protocol.MarshalHostServiceStorageDeleteRequest(&protocol.HostServiceStorageDeleteRequest{Path: "reports/demo.json"}),
	)
	if deleteResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("delete: expected success, got status=%d payload=%s", deleteResponse.Status, string(deleteResponse.Payload))
	}

	statResponse := invokeStorageHostService(
		t,
		hcc,
		protocol.HostServiceMethodStorageStat,
		"reports/demo.json",
		protocol.MarshalHostServiceStorageStatRequest(&protocol.HostServiceStorageStatRequest{Path: "reports/demo.json"}),
	)
	if statResponse.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("stat: expected success, got status=%d payload=%s", statResponse.Status, string(statResponse.Payload))
	}
	statPayload, err := protocol.UnmarshalHostServiceStorageStatResponse(statResponse.Payload)
	if err != nil {
		t.Fatalf("stat payload decode failed: %v", err)
	}
	if statPayload.Found {
		t.Fatalf("stat: expected object to be deleted, got %#v", statPayload.Object)
	}
}

// TestHandleHostServiceInvokeStorageRejectsUnauthorizedPath verifies requests
// outside the authorized logical path set are denied before storagecap is called.
func TestHandleHostServiceInvokeStorageRejectsUnauthorizedPath(t *testing.T) {
	hcc := newStorageHostCallContext([]string{"reports/"})
	response := invokeStorageHostService(
		t,
		hcc,
		protocol.HostServiceMethodStoragePut,
		"private/escape.txt",
		protocol.MarshalHostServiceStoragePutRequest(&protocol.HostServiceStoragePutRequest{
			Path: "private/escape.txt",
			Body: []byte("blocked"),
		}),
	)
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected capability denied for unauthorized path, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// TestHandleHostServiceInvokeStorageRejectsTargetMismatch verifies the request
// payload path must match the declared target resource reference.
func TestHandleHostServiceInvokeStorageRejectsTargetMismatch(t *testing.T) {
	configureStorageDomainServiceForTest(t, &storageDomainTestService{})

	hcc := newStorageHostCallContext([]string{"reports/"})
	response := invokeStorageHostService(
		t,
		hcc,
		protocol.HostServiceMethodStoragePut,
		"reports/demo.json",
		protocol.MarshalHostServiceStoragePutRequest(&protocol.HostServiceStoragePutRequest{
			Path: "reports/other.json",
			Body: []byte("blocked"),
		}),
	)
	if response.Status != protocol.HostCallStatusInvalidRequest {
		t.Fatalf("expected invalid request for target mismatch, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// TestHandleHostServiceInvokeStorageRequiresConfiguredService verifies missing
// storage domain wiring fails explicitly instead of using a package default.
func TestHandleHostServiceInvokeStorageRequiresConfiguredService(t *testing.T) {
	configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{})

	hcc := newStorageHostCallContext([]string{"reports/"})
	response := invokeStorageHostService(
		t,
		hcc,
		protocol.HostServiceMethodStoragePut,
		"reports/demo.json",
		protocol.MarshalHostServiceStoragePutRequest(&protocol.HostServiceStoragePutRequest{
			Path: "reports/demo.json",
			Body: []byte("blocked"),
		}),
	)
	if response.Status != protocol.HostCallStatusInternalError {
		t.Fatalf("expected internal error for unconfigured storage service, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// TestHandleHostServiceInvokeStorageUsesConfiguredSharedService verifies
// storage host service dispatch reuses the explicitly configured domain service.
func TestHandleHostServiceInvokeStorageUsesConfiguredSharedService(t *testing.T) {
	storageSvc := &storageDomainTestService{}
	configureStorageDomainServiceForTest(t, storageSvc)

	hcc := newStorageHostCallContext([]string{"reports/"})
	response := invokeStorageHostService(
		t,
		hcc,
		protocol.HostServiceMethodStoragePut,
		"reports/demo.json",
		protocol.MarshalHostServiceStoragePutRequest(&protocol.HostServiceStoragePutRequest{
			Path: "reports/demo.json",
			Body: []byte("shared"),
		}),
	)
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("put through shared storage service: expected success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if storageSvc.putCalls != 1 || storageSvc.lastPath != "reports/demo.json" {
		t.Fatalf("expected shared storage service to receive one put, calls=%d path=%q", storageSvc.putCalls, storageSvc.lastPath)
	}
}

// TestHandleHostServiceInvokeStorageConcurrentDispatchIsRaceSafe verifies
// concurrent dispatch can reuse the same storagecap service instance.
func TestHandleHostServiceInvokeStorageConcurrentDispatchIsRaceSafe(t *testing.T) {
	storageSvc := &storageDomainTestService{}
	storageSvc.objects = map[string]*storageDomainTestObject{
		"reports/demo.json": {body: []byte("ready"), contentType: "text/plain"},
	}
	configureStorageDomainServiceForTest(t, storageSvc)

	hcc := newStorageHostCallContext([]string{"reports/"})
	const (
		workers    = 8
		iterations = 50
	)
	errCh := make(chan string, workers*iterations)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
				for j := 0; j < iterations; j++ {
					response := dispatchStorageHostServiceRequest(
						t,
						hcc,
						protocol.HostServiceMethodStorageGet,
					"reports/demo.json",
					protocol.MarshalHostServiceStorageGetRequest(&protocol.HostServiceStorageGetRequest{Path: "reports/demo.json"}),
				)
				if response.Status != protocol.HostCallStatusSuccess {
					errCh <- string(response.Payload)
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)
	for msg := range errCh {
		t.Fatalf("concurrent storage host service dispatch failed: %s", msg)
	}
}

// TestMatchAuthorizedStoragePath verifies logical prefix and exact-file path
// matching for authorized storage resources.
func TestMatchAuthorizedStoragePath(t *testing.T) {
	specs := []*protocol.HostServiceSpec{{
		Service: protocol.HostServiceStorage,
		Methods: []string{protocol.HostServiceMethodStorageGet},
		Paths:   []string{"reports/", "exports/daily.json"},
	}}

	if matched := matchAuthorizedStoragePath(specs, "reports/2026/summary.json"); matched != "reports/" {
		t.Fatalf("expected reports/ prefix to match, got %q", matched)
	}
	if matched := matchAuthorizedStoragePath(specs, "exports/daily.json"); matched != "exports/daily.json" {
		t.Fatalf("expected exact file path to match, got %q", matched)
	}
	if matched := matchAuthorizedStoragePath(specs, "reports-v2/demo.json"); matched != "" {
		t.Fatalf("expected sibling prefix to be rejected, got %q", matched)
	}
}

// configureStorageDomainServiceForTest installs one storagecap service directory.
func configureStorageDomainServiceForTest(t *testing.T, service storagecap.Service) {
	t.Helper()
	configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{storage: service})
}

// newStorageHostCallContext constructs a storage-capable host call context for
// the provided authorized logical paths.
func newStorageHostCallContext(paths []string) *hostCallContext {
	return &hostCallContext{
		pluginID: "test-plugin-storage",
		capabilities: map[string]struct{}{
			protocol.CapabilityStorage: {},
		},
		hostServices: []*protocol.HostServiceSpec{{
			Service: protocol.HostServiceStorage,
			Methods: []string{
				protocol.HostServiceMethodStorageDelete,
				protocol.HostServiceMethodStorageGet,
				protocol.HostServiceMethodStorageList,
				protocol.HostServiceMethodStoragePut,
				protocol.HostServiceMethodStorageStat,
			},
			Paths: paths,
		}},
	}
}

// invokeStorageHostService dispatches a storage host-service request through
// the shared handler and returns the raw response envelope.
func invokeStorageHostService(
	t *testing.T,
	hcc *hostCallContext,
	method string,
	targetPath string,
	payload []byte,
) *protocol.HostCallResponseEnvelope {
	t.Helper()
	return dispatchStorageHostServiceRequest(t, hcc, method, targetPath, payload)
}

// dispatchStorageHostServiceRequest dispatches one storage host-service request.
func dispatchStorageHostServiceRequest(
	t *testing.T,
	hcc *hostCallContext,
	method string,
	targetPath string,
	payload []byte,
) *protocol.HostCallResponseEnvelope {
	t.Helper()
	request := &protocol.HostServiceRequestEnvelope{
		Service:     protocol.HostServiceStorage,
		Method:      method,
		ResourceRef: targetPath,
		Payload:     payload,
	}
	return handleHostServiceInvoke(
		context.Background(),
		withTestHostCallRuntime(t, hcc),
		protocol.MarshalHostServiceRequestEnvelope(request),
	)
}
