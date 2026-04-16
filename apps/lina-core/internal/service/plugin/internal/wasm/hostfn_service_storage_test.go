// This file tests storage host service authorization, path isolation, and
// logical path prefix matching.

package wasm

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"lina-core/internal/service/config"
	"lina-core/pkg/pluginbridge"
)

func TestHandleHostServiceInvokeStorageLifecycle(t *testing.T) {
	storageRoot := t.TempDir()
	config.SetPluginDynamicStoragePathOverride(storageRoot)
	t.Cleanup(func() {
		config.SetPluginDynamicStoragePathOverride("")
	})

	authorizedPath := "reports/"
	hcc := newStorageHostCallContext([]string{authorizedPath})

	putResponse := invokeStorageHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodStoragePut,
		"reports/demo.json",
		pluginbridge.MarshalHostServiceStoragePutRequest(&pluginbridge.HostServiceStoragePutRequest{
			Path:        "reports/demo.json",
			Body:        []byte(`{"ok":true}`),
			ContentType: "application/json",
			Overwrite:   false,
		}),
	)
	if putResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("put: expected success, got status=%d payload=%s", putResponse.Status, string(putResponse.Payload))
	}
	putPayload, err := pluginbridge.UnmarshalHostServiceStoragePutResponse(putResponse.Payload)
	if err != nil {
		t.Fatalf("put payload decode failed: %v", err)
	}
	if putPayload.Object == nil || putPayload.Object.Path != "reports/demo.json" {
		t.Fatalf("put object: got %#v", putPayload.Object)
	}

	absolutePath := filepath.Join(
		storageRoot,
		storageHostServiceRootDirName,
		storageHostServiceDirName,
		hcc.pluginID,
		"reports",
		"demo.json",
	)
	content, err := os.ReadFile(absolutePath)
	if err != nil {
		t.Fatalf("expected written file to exist, got error: %v", err)
	}
	if string(content) != `{"ok":true}` {
		t.Fatalf("written content: got %q", content)
	}

	getResponse := invokeStorageHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodStorageGet,
		"reports/demo.json",
		pluginbridge.MarshalHostServiceStorageGetRequest(&pluginbridge.HostServiceStorageGetRequest{Path: "reports/demo.json"}),
	)
	if getResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("get: expected success, got status=%d payload=%s", getResponse.Status, string(getResponse.Payload))
	}
	getPayload, err := pluginbridge.UnmarshalHostServiceStorageGetResponse(getResponse.Payload)
	if err != nil {
		t.Fatalf("get payload decode failed: %v", err)
	}
	if !getPayload.Found || string(getPayload.Body) != `{"ok":true}` {
		t.Fatalf("get payload: got %#v", getPayload)
	}

	listResponse := invokeStorageHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodStorageList,
		"reports",
		pluginbridge.MarshalHostServiceStorageListRequest(&pluginbridge.HostServiceStorageListRequest{
			Prefix: "reports",
			Limit:  10,
		}),
	)
	if listResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("list: expected success, got status=%d payload=%s", listResponse.Status, string(listResponse.Payload))
	}
	listPayload, err := pluginbridge.UnmarshalHostServiceStorageListResponse(listResponse.Payload)
	if err != nil {
		t.Fatalf("list payload decode failed: %v", err)
	}
	if len(listPayload.Objects) != 1 || listPayload.Objects[0].Path != "reports/demo.json" {
		t.Fatalf("list payload: got %#v", listPayload.Objects)
	}

	deleteResponse := invokeStorageHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodStorageDelete,
		"reports/demo.json",
		pluginbridge.MarshalHostServiceStorageDeleteRequest(&pluginbridge.HostServiceStorageDeleteRequest{Path: "reports/demo.json"}),
	)
	if deleteResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("delete: expected success, got status=%d payload=%s", deleteResponse.Status, string(deleteResponse.Payload))
	}

	statResponse := invokeStorageHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodStorageStat,
		"reports/demo.json",
		pluginbridge.MarshalHostServiceStorageStatRequest(&pluginbridge.HostServiceStorageStatRequest{Path: "reports/demo.json"}),
	)
	if statResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("stat: expected success, got status=%d payload=%s", statResponse.Status, string(statResponse.Payload))
	}
	statPayload, err := pluginbridge.UnmarshalHostServiceStorageStatResponse(statResponse.Payload)
	if err != nil {
		t.Fatalf("stat payload decode failed: %v", err)
	}
	if statPayload.Found {
		t.Fatalf("stat: expected object to be deleted, got %#v", statPayload.Object)
	}
}

func TestHandleHostServiceInvokeStorageLifecycleWithRelativeStorageRoot(t *testing.T) {
	storageRoot := t.TempDir()
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("resolve working directory failed: %v", err)
	}
	relativeStorageRoot, err := filepath.Rel(workingDir, storageRoot)
	if err != nil {
		t.Fatalf("build relative storage root failed: %v", err)
	}
	config.SetPluginDynamicStoragePathOverride(relativeStorageRoot)
	t.Cleanup(func() {
		config.SetPluginDynamicStoragePathOverride("")
	})

	hcc := newStorageHostCallContext([]string{"reports/"})
	putResponse := invokeStorageHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodStoragePut,
		"reports/demo.json",
		pluginbridge.MarshalHostServiceStoragePutRequest(&pluginbridge.HostServiceStoragePutRequest{
			Path:        "reports/demo.json",
			Body:        []byte(`{"ok":true}`),
			ContentType: "application/json",
		}),
	)
	if putResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("put with relative storage root: expected success, got status=%d payload=%s", putResponse.Status, string(putResponse.Payload))
	}

	listResponse := invokeStorageHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodStorageList,
		"reports",
		pluginbridge.MarshalHostServiceStorageListRequest(&pluginbridge.HostServiceStorageListRequest{Prefix: "reports", Limit: 10}),
	)
	if listResponse.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("list with relative storage root: expected success, got status=%d payload=%s", listResponse.Status, string(listResponse.Payload))
	}
	listPayload, err := pluginbridge.UnmarshalHostServiceStorageListResponse(listResponse.Payload)
	if err != nil {
		t.Fatalf("list payload decode failed: %v", err)
	}
	if len(listPayload.Objects) != 1 || listPayload.Objects[0].Path != "reports/demo.json" {
		t.Fatalf("list payload: got %#v", listPayload.Objects)
	}
}

func TestHandleHostServiceInvokeStorageRejectsUnauthorizedPath(t *testing.T) {
	config.SetPluginDynamicStoragePathOverride(t.TempDir())
	t.Cleanup(func() {
		config.SetPluginDynamicStoragePathOverride("")
	})

	hcc := newStorageHostCallContext([]string{"reports/"})
	response := invokeStorageHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodStoragePut,
		"private/escape.txt",
		pluginbridge.MarshalHostServiceStoragePutRequest(&pluginbridge.HostServiceStoragePutRequest{
			Path: "private/escape.txt",
			Body: []byte("blocked"),
		}),
	)
	if response.Status != pluginbridge.HostCallStatusCapabilityDenied {
		t.Fatalf("expected capability denied for unauthorized path, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

func TestHandleHostServiceInvokeStorageRejectsTargetMismatch(t *testing.T) {
	config.SetPluginDynamicStoragePathOverride(t.TempDir())
	t.Cleanup(func() {
		config.SetPluginDynamicStoragePathOverride("")
	})

	hcc := newStorageHostCallContext([]string{"reports/"})
	response := invokeStorageHostService(
		t,
		hcc,
		pluginbridge.HostServiceMethodStoragePut,
		"reports/demo.json",
		pluginbridge.MarshalHostServiceStoragePutRequest(&pluginbridge.HostServiceStoragePutRequest{
			Path: "reports/other.json",
			Body: []byte("blocked"),
		}),
	)
	if response.Status != pluginbridge.HostCallStatusInvalidRequest {
		t.Fatalf("expected invalid request for target mismatch, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

func TestMatchAuthorizedStoragePath(t *testing.T) {
	specs := []*pluginbridge.HostServiceSpec{{
		Service: pluginbridge.HostServiceStorage,
		Methods: []string{pluginbridge.HostServiceMethodStorageGet},
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

func newStorageHostCallContext(paths []string) *hostCallContext {
	return &hostCallContext{
		pluginID: "test-plugin-storage",
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityStorage: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{{
			Service: pluginbridge.HostServiceStorage,
			Methods: []string{
				pluginbridge.HostServiceMethodStorageDelete,
				pluginbridge.HostServiceMethodStorageGet,
				pluginbridge.HostServiceMethodStorageList,
				pluginbridge.HostServiceMethodStoragePut,
				pluginbridge.HostServiceMethodStorageStat,
			},
			Paths: paths,
		}},
	}
}

func invokeStorageHostService(
	t *testing.T,
	hcc *hostCallContext,
	method string,
	targetPath string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	t.Helper()

	request := &pluginbridge.HostServiceRequestEnvelope{
		Service:     pluginbridge.HostServiceStorage,
		Method:      method,
		ResourceRef: targetPath,
		Payload:     payload,
	}
	return handleHostServiceInvoke(
		context.Background(),
		hcc,
		pluginbridge.MarshalHostServiceRequestEnvelope(request),
	)
}
