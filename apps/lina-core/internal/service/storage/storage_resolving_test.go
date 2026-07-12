// This file verifies the resolving Storage Service routes NamespaceFiles
// through provider selection and keeps other namespaces on the local service.

package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
)

type resolveRecordingProvider struct {
	putKeys []string
	bodies  map[string]string
}

func (p *resolveRecordingProvider) Put(_ context.Context, in storagecap.ProviderPutInput) (*storagecap.ProviderObject, error) {
	p.putKeys = append(p.putKeys, in.Key)
	if p.bodies == nil {
		p.bodies = map[string]string{}
	}
	content, _ := io.ReadAll(in.Body)
	p.bodies[in.Key] = string(content)
	return &storagecap.ProviderObject{Key: in.Key, Size: int64(len(content))}, nil
}

func (p *resolveRecordingProvider) Get(_ context.Context, in storagecap.ProviderGetInput) (*storagecap.ProviderGetOutput, error) {
	content, ok := p.bodies[in.Key]
	if !ok {
		return &storagecap.ProviderGetOutput{Found: false}, nil
	}
	return &storagecap.ProviderGetOutput{
		Found:  true,
		Object: &storagecap.ProviderObject{Key: in.Key, Size: int64(len(content))},
		Body:   io.NopCloser(strings.NewReader(content)),
	}, nil
}

func (p *resolveRecordingProvider) Delete(_ context.Context, in storagecap.ProviderDeleteInput) error {
	delete(p.bodies, in.Key)
	return nil
}
func (p *resolveRecordingProvider) DeleteMany(context.Context, storagecap.ProviderDeleteManyInput) error {
	return nil
}
func (p *resolveRecordingProvider) List(context.Context, storagecap.ProviderListInput) (*storagecap.ProviderListOutput, error) {
	return &storagecap.ProviderListOutput{}, nil
}
func (p *resolveRecordingProvider) ListCursor(context.Context, storagecap.ProviderListCursorInput) (*storagecap.ProviderListCursorOutput, error) {
	return &storagecap.ProviderListCursorOutput{}, nil
}
func (p *resolveRecordingProvider) Stat(context.Context, storagecap.ProviderStatInput) (*storagecap.ProviderStatOutput, error) {
	return &storagecap.ProviderStatOutput{Found: false}, nil
}
func (p *resolveRecordingProvider) BatchStat(context.Context, storagecap.ProviderBatchStatInput) (*storagecap.ProviderBatchStatOutput, error) {
	return &storagecap.ProviderBatchStatOutput{}, nil
}

type emptyRuntime struct{}

func (emptyRuntime) ProviderPluginAvailable(context.Context, string) bool { return false }

func TestResolvingServiceUsesFilesPrefixOnLocalProvider(t *testing.T) {
	t.Parallel()
	provider := &resolveRecordingProvider{}
	// Use recording provider as local provider so we observe key mapping without disk.
	svc := NewResolvingService(New(Config{}), emptyRuntime{}, provider)
	ctx := context.Background()

	_, err := svc.Put(ctx, PutInput{
		Namespace: NamespaceFiles,
		Key:       "1/demo.txt",
		Body:      strings.NewReader("hello"),
		Overwrite: true,
	})
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	if len(provider.putKeys) != 1 || provider.putKeys[0] != "files/1/demo.txt" {
		t.Fatalf("put keys = %#v", provider.putKeys)
	}
	if id := FilesProviderID(ctx, svc); id != storagecap.LocalProviderID {
		t.Fatalf("provider id = %q", id)
	}
}

func TestFilesProviderKeyRejectsUnsafe(t *testing.T) {
	t.Parallel()
	if _, err := filesProviderKey("../x"); err == nil {
		t.Fatal("expected rejection")
	}
	if _, err := filesProviderKey("/abs"); err == nil {
		t.Fatal("expected absolute rejection")
	}
}

type multiAvailableRuntime struct {
	ids map[string]bool
}

func (r multiAvailableRuntime) ProviderPluginAvailable(_ context.Context, pluginID string) bool {
	return r.ids[pluginID]
}

func TestResolveFilesProviderIDFailsOnMultipleCloudProviders(t *testing.T) {
	// Not parallel: process-global storagecap.Provide registry.
	aID := fmt.Sprintf("test-files-resolve-a-%d", time.Now().UnixNano())
	bID := fmt.Sprintf("test-files-resolve-b-%d", time.Now().UnixNano())
	for _, id := range []string{aID, bID} {
		if err := storagecap.Provide(id, func(context.Context, storagecap.ProviderEnv) (storagecap.Provider, error) {
			return &resolveRecordingProvider{}, nil
		}); err != nil {
			t.Fatalf("provide %s: %v", id, err)
		}
	}

	svc := NewResolvingService(New(Config{}), multiAvailableRuntime{
		ids: map[string]bool{aID: true, bID: true},
	}, &resolveRecordingProvider{})
	_, err := ResolveFilesProviderID(context.Background(), svc)
	if err == nil {
		t.Fatal("expected multi-provider conflict")
	}
	if !bizerr.Is(err, storagecap.CodeStorageProviderConflict) {
		t.Fatalf("expected CodeStorageProviderConflict, got %v", err)
	}
	if id := FilesProviderID(context.Background(), svc); id != storagecap.LocalProviderID {
		t.Fatalf("FilesProviderID on conflict should fall back to local, got %q", id)
	}
}
