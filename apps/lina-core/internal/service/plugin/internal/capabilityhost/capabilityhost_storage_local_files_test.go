// This file verifies local storage provider routes files/ keys to the host
// file-center namespace without the plugin capability storage root.

package capabilityhost

import (
	"context"
	"io"
	"strings"
	"testing"

	storagesvc "lina-core/internal/service/storage"
	"lina-core/pkg/plugin/capability/storagecap"
)

func TestLocalStorageProviderRoutesFilesPrefixToFilesNamespace(t *testing.T) {
	t.Parallel()
	var (
		filesRoot   = t.TempDir()
		pluginsRoot = t.TempDir()
		provider    = NewLocalStorageProvider(storagesvc.New(storagesvc.Config{
			NamespaceRoots: map[string]string{
				storagesvc.NamespaceFiles:   filesRoot,
				storagesvc.NamespacePlugins: pluginsRoot,
			},
		}))
		ctx = context.Background()
	)

	if _, err := provider.Put(ctx, storagecap.ProviderPutInput{
		Key:       "files/tenant/demo.txt",
		Body:      strings.NewReader("file-body"),
		Size:      9,
		Overwrite: true,
	}); err != nil {
		t.Fatalf("put files key: %v", err)
	}
	if _, err := provider.Put(ctx, storagecap.ProviderPutInput{
		Key:       "plugin-a/demo.txt",
		Body:      strings.NewReader("plugin-body"),
		Size:      11,
		Overwrite: true,
	}); err != nil {
		t.Fatalf("put plugin key: %v", err)
	}

	got, err := provider.Get(ctx, storagecap.ProviderGetInput{Key: "files/tenant/demo.txt"})
	if err != nil || got == nil || !got.Found {
		t.Fatalf("get files key found=%v err=%v", got != nil && got.Found, err)
	}
	raw, _ := io.ReadAll(got.Body)
	_ = got.Body.Close()
	if string(raw) != "file-body" {
		t.Fatalf("files content = %q", raw)
	}
	if got.Object == nil || got.Object.Key != "files/tenant/demo.txt" {
		t.Fatalf("object key = %#v", got.Object)
	}

	pluginGot, err := provider.Get(ctx, storagecap.ProviderGetInput{Key: "plugin-a/demo.txt"})
	if err != nil || pluginGot == nil || !pluginGot.Found {
		t.Fatalf("get plugin key found=%v err=%v", pluginGot != nil && pluginGot.Found, err)
	}
	pluginRaw, _ := io.ReadAll(pluginGot.Body)
	_ = pluginGot.Body.Close()
	if string(pluginRaw) != "plugin-body" {
		t.Fatalf("plugin content = %q", pluginRaw)
	}
}
