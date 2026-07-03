// This file implements the host built-in local storage provider as an adapter
// from the plugin storagecap.Provider contract to the host-internal object
// storage service. The provider receives scoped provider object keys and never
// plugin authorization snapshots.

package capabilityhost

import (
	"context"
	"errors"
	"path"
	"strings"

	"lina-core/internal/service/storage"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
)

const (
	localStorageProviderRootDir = ".capability-storage"
)

// localStorageProvider stores plugin objects through the host object storage service.
type localStorageProvider struct {
	storage storage.Service
}

// NewLocalStorageProvider creates the host built-in local storage provider.
func NewLocalStorageProvider(storageSvc storage.Service) storagecap.Provider {
	return &localStorageProvider{storage: storageSvc}
}

// Put writes one scoped object key.
func (p *localStorageProvider) Put(ctx context.Context, in storagecap.ProviderPutInput) (*storagecap.ProviderObject, error) {
	if err := p.ensureAvailable(); err != nil {
		return nil, err
	}
	objectKey, err := normalizePluginStoragePath(in.Key, false)
	if err != nil {
		return nil, err
	}
	output, err := p.storage.Put(ctx, storage.PutInput{
		Namespace:   storage.NamespacePlugins,
		Key:         localStorageKey(objectKey),
		Body:        in.Body,
		Size:        in.Size,
		ContentType: in.ContentType,
		Overwrite:   in.Overwrite,
	})
	if err != nil {
		return nil, mapLocalStorageError(err)
	}
	return p.providerObject(objectKey, outputObject(output), in.ContentType), nil
}

// Get reads one scoped object key.
func (p *localStorageProvider) Get(ctx context.Context, in storagecap.ProviderGetInput) (*storagecap.ProviderGetOutput, error) {
	if err := p.ensureAvailable(); err != nil {
		return nil, err
	}
	objectKey, err := normalizePluginStoragePath(in.Key, false)
	if err != nil {
		return nil, err
	}
	output, err := p.storage.Get(ctx, storage.GetInput{Namespace: storage.NamespacePlugins, Key: localStorageKey(objectKey)})
	if err != nil {
		return nil, mapLocalStorageError(err)
	}
	if output == nil || !output.Found {
		return &storagecap.ProviderGetOutput{Found: false}, nil
	}
	return &storagecap.ProviderGetOutput{
		Object: p.providerObject(objectKey, output.Object, ""),
		Body:   output.Body,
		Found:  true,
	}, nil
}

// Delete removes one scoped object key.
func (p *localStorageProvider) Delete(ctx context.Context, in storagecap.ProviderDeleteInput) error {
	if err := p.ensureAvailable(); err != nil {
		return err
	}
	objectKey, err := normalizePluginStoragePath(in.Key, false)
	if err != nil {
		return err
	}
	err = p.storage.Delete(ctx, storage.DeleteInput{Namespace: storage.NamespacePlugins, Key: localStorageKey(objectKey)})
	return mapLocalStorageError(err)
}

// DeleteMany removes explicit scoped object keys.
func (p *localStorageProvider) DeleteMany(ctx context.Context, in storagecap.ProviderDeleteManyInput) error {
	if err := p.ensureAvailable(); err != nil {
		return err
	}
	keys := make([]string, 0, len(in.Keys))
	for _, key := range in.Keys {
		objectKey, err := normalizePluginStoragePath(key, false)
		if err != nil {
			return err
		}
		keys = append(keys, localStorageKey(objectKey))
	}
	err := p.storage.DeleteMany(ctx, storage.DeleteManyInput{Namespace: storage.NamespacePlugins, Keys: keys})
	return mapLocalStorageError(err)
}

// List lists scoped object keys under one prefix.
func (p *localStorageProvider) List(ctx context.Context, in storagecap.ProviderListInput) (*storagecap.ProviderListOutput, error) {
	if err := p.ensureAvailable(); err != nil {
		return nil, err
	}
	prefix, err := normalizePluginStoragePath(in.Prefix, false)
	if err != nil {
		return nil, err
	}
	limit := normalizeStorageListLimit(in.Limit)
	output, err := p.storage.List(ctx, storage.ListInput{
		Namespace: storage.NamespacePlugins,
		Prefix:    localStorageKey(prefix),
		Limit:     limit,
	})
	if err != nil {
		return nil, mapLocalStorageError(err)
	}
	objects := make([]*storagecap.ProviderObject, 0)
	if output != nil {
		objects = make([]*storagecap.ProviderObject, 0, len(output.Objects))
		for _, object := range output.Objects {
			providerKey := providerKeyFromLocalStorageObject(object)
			if providerKey == "" {
				continue
			}
			objects = append(objects, p.providerObject(providerKey, object, ""))
		}
	}
	return &storagecap.ProviderListOutput{Objects: objects}, nil
}

// ListCursor lists scoped object keys under one prefix using deterministic key order.
func (p *localStorageProvider) ListCursor(ctx context.Context, in storagecap.ProviderListCursorInput) (*storagecap.ProviderListCursorOutput, error) {
	if err := p.ensureAvailable(); err != nil {
		return nil, err
	}
	prefix, err := normalizePluginStoragePath(in.Prefix, false)
	if err != nil {
		return nil, err
	}
	cursor := ""
	if strings.TrimSpace(in.Cursor) != "" {
		cursor, err = normalizePluginStoragePath(in.Cursor, false)
		if err != nil {
			return nil, err
		}
	}
	limit := normalizeStorageListLimit(in.Limit)
	output, err := p.storage.ListCursor(ctx, storage.ListCursorInput{
		Namespace: storage.NamespacePlugins,
		Prefix:    localStorageKey(prefix),
		Cursor:    localStorageKey(cursor),
		Limit:     limit,
	})
	if err != nil {
		return nil, mapLocalStorageError(err)
	}
	result := &storagecap.ProviderListCursorOutput{Objects: []*storagecap.ProviderObject{}}
	if output == nil {
		return result, nil
	}
	for _, object := range output.Objects {
		providerKey := providerKeyFromLocalStorageObject(object)
		if providerKey == "" {
			continue
		}
		result.Objects = append(result.Objects, p.providerObject(providerKey, object, ""))
	}
	result.NextCursor = providerKeyFromLocalStorageKey(output.NextCursor)
	return result, nil
}

// Stat reads scoped object metadata.
func (p *localStorageProvider) Stat(ctx context.Context, in storagecap.ProviderStatInput) (*storagecap.ProviderStatOutput, error) {
	if err := p.ensureAvailable(); err != nil {
		return nil, err
	}
	objectKey, err := normalizePluginStoragePath(in.Key, false)
	if err != nil {
		return nil, err
	}
	output, err := p.storage.Stat(ctx, storage.StatInput{Namespace: storage.NamespacePlugins, Key: localStorageKey(objectKey)})
	if err != nil {
		return nil, mapLocalStorageError(err)
	}
	if output == nil || !output.Found {
		return &storagecap.ProviderStatOutput{Found: false}, nil
	}
	return &storagecap.ProviderStatOutput{Object: p.providerObject(objectKey, output.Object, ""), Found: true}, nil
}

// BatchStat reads metadata for explicit scoped object keys.
func (p *localStorageProvider) BatchStat(ctx context.Context, in storagecap.ProviderBatchStatInput) (*storagecap.ProviderBatchStatOutput, error) {
	if err := p.ensureAvailable(); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(in.Keys))
	keyByStorageKey := make(map[string]string, len(in.Keys))
	for _, key := range in.Keys {
		objectKey, err := normalizePluginStoragePath(key, false)
		if err != nil {
			return nil, err
		}
		storageKey := localStorageKey(objectKey)
		keys = append(keys, storageKey)
		keyByStorageKey[storageKey] = objectKey
	}
	output, err := p.storage.BatchStat(ctx, storage.BatchStatInput{Namespace: storage.NamespacePlugins, Keys: keys})
	if err != nil {
		return nil, mapLocalStorageError(err)
	}
	result := &storagecap.ProviderBatchStatOutput{Objects: []*storagecap.ProviderObject{}}
	if output == nil {
		return result, nil
	}
	for _, object := range output.Objects {
		providerKey := providerKeyFromLocalStorageObject(object)
		if providerKey == "" {
			continue
		}
		result.Objects = append(result.Objects, p.providerObject(providerKey, object, ""))
	}
	for _, missingKey := range output.MissingKeys {
		if providerKey := keyByStorageKey[missingKey]; providerKey != "" {
			result.MissingKeys = append(result.MissingKeys, providerKey)
		}
	}
	return result, nil
}

// ensureAvailable verifies the local provider has a storage backend.
func (p *localStorageProvider) ensureAvailable() error {
	if p == nil || p.storage == nil {
		return bizerr.NewCode(storagecap.CodeStorageProviderUnavailable)
	}
	return nil
}

// providerObject maps host storage metadata into provider metadata.
func (p *localStorageProvider) providerObject(key string, object *storage.Object, contentType string) *storagecap.ProviderObject {
	if strings.TrimSpace(contentType) == "" && object != nil {
		contentType = object.ContentType
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = detectObjectContentType("", nil, key)
	}
	if object == nil {
		return &storagecap.ProviderObject{Key: key, ContentType: strings.TrimSpace(contentType), Visibility: storagecap.VisibilityPrivate}
	}
	return &storagecap.ProviderObject{
		Key:         key,
		Size:        object.Size,
		ContentType: strings.TrimSpace(contentType),
		ETag:        object.ETag,
		UpdatedAt:   object.UpdatedAt,
		Visibility:  storagecap.VisibilityPrivate,
	}
}

func outputObject(output *storage.PutOutput) *storage.Object {
	if output == nil {
		return nil
	}
	return output.Object
}

func localStorageKey(providerKey string) string {
	providerKey = strings.Trim(strings.ReplaceAll(strings.TrimSpace(providerKey), "\\", "/"), "/")
	if providerKey == "" {
		return localStorageProviderRootDir
	}
	return path.Join(localStorageProviderRootDir, providerKey)
}

func providerKeyFromLocalStorageObject(object *storage.Object) string {
	if object == nil {
		return ""
	}
	return providerKeyFromLocalStorageKey(object.Key)
}

func providerKeyFromLocalStorageKey(key string) string {
	key = strings.Trim(strings.ReplaceAll(strings.TrimSpace(key), "\\", "/"), "/")
	if key == "" {
		return ""
	}
	prefix := localStorageProviderRootDir + "/"
	if key == localStorageProviderRootDir {
		return ""
	}
	if !strings.HasPrefix(key, prefix) {
		return ""
	}
	return strings.TrimPrefix(key, prefix)
}

func mapLocalStorageError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, storage.ErrUnavailable) {
		return bizerr.NewCode(storagecap.CodeStorageProviderUnavailable)
	}
	if errors.Is(err, storage.ErrPathInvalid) {
		return bizerr.NewCode(storagecap.CodeStoragePathInvalid)
	}
	if errors.Is(err, storage.ErrObjectExist) {
		return bizerr.NewCode(storagecap.CodeStorageObjectExists)
	}
	return err
}
