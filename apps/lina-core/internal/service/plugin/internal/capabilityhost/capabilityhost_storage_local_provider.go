// This file implements the host built-in local storage provider as an adapter
// from the plugin storagecap.Provider contract to the host-internal object
// storage service. The provider receives scoped provider object keys and never
// plugin authorization snapshots. Keys under "files/" map to the host file-center
// namespace without the plugin .capability-storage root so historical upload paths
// stay stable; other keys continue to use NamespacePlugins.

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
	// hostFilesProviderKeyPrefix isolates file-center objects on shared providers.
	hostFilesProviderKeyPrefix = "files/"
)

// localStorageProvider stores objects through the host object storage service.
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
	namespace, storageKey, publicKey, err := resolveLocalStorageTarget(in.Key)
	if err != nil {
		return nil, err
	}
	output, err := p.storage.Put(ctx, storage.PutInput{
		Namespace:   namespace,
		Key:         storageKey,
		Body:        in.Body,
		Size:        in.Size,
		ContentType: in.ContentType,
		Overwrite:   in.Overwrite,
	})
	if err != nil {
		return nil, mapLocalStorageError(err)
	}
	return p.providerObject(publicKey, outputObject(output), in.ContentType), nil
}

// Get reads one scoped object key.
func (p *localStorageProvider) Get(ctx context.Context, in storagecap.ProviderGetInput) (*storagecap.ProviderGetOutput, error) {
	if err := p.ensureAvailable(); err != nil {
		return nil, err
	}
	namespace, storageKey, publicKey, err := resolveLocalStorageTarget(in.Key)
	if err != nil {
		return nil, err
	}
	output, err := p.storage.Get(ctx, storage.GetInput{Namespace: namespace, Key: storageKey})
	if err != nil {
		return nil, mapLocalStorageError(err)
	}
	if output == nil || !output.Found {
		return &storagecap.ProviderGetOutput{Found: false}, nil
	}
	return &storagecap.ProviderGetOutput{
		Object: p.providerObject(publicKey, output.Object, ""),
		Body:   output.Body,
		Found:  true,
	}, nil
}

// Delete removes one scoped object key.
func (p *localStorageProvider) Delete(ctx context.Context, in storagecap.ProviderDeleteInput) error {
	if err := p.ensureAvailable(); err != nil {
		return err
	}
	namespace, storageKey, _, err := resolveLocalStorageTarget(in.Key)
	if err != nil {
		return err
	}
	err = p.storage.Delete(ctx, storage.DeleteInput{Namespace: namespace, Key: storageKey})
	return mapLocalStorageError(err)
}

// DeleteMany removes explicit scoped object keys.
func (p *localStorageProvider) DeleteMany(ctx context.Context, in storagecap.ProviderDeleteManyInput) error {
	if err := p.ensureAvailable(); err != nil {
		return err
	}
	// Group by namespace because host storage DeleteMany is namespace-scoped.
	byNamespace := map[string][]string{}
	for _, key := range in.Keys {
		namespace, storageKey, _, err := resolveLocalStorageTarget(key)
		if err != nil {
			return err
		}
		byNamespace[namespace] = append(byNamespace[namespace], storageKey)
	}
	for namespace, keys := range byNamespace {
		if err := p.storage.DeleteMany(ctx, storage.DeleteManyInput{Namespace: namespace, Keys: keys}); err != nil {
			return mapLocalStorageError(err)
		}
	}
	return nil
}

// List lists scoped object keys under one prefix.
func (p *localStorageProvider) List(ctx context.Context, in storagecap.ProviderListInput) (*storagecap.ProviderListOutput, error) {
	if err := p.ensureAvailable(); err != nil {
		return nil, err
	}
	namespace, storagePrefix, isFiles, err := resolveLocalStoragePrefix(in.Prefix)
	if err != nil {
		return nil, err
	}
	limit := normalizeStorageListLimit(in.Limit)
	output, err := p.storage.List(ctx, storage.ListInput{
		Namespace: namespace,
		Prefix:    storagePrefix,
		Limit:     limit,
	})
	if err != nil {
		return nil, mapLocalStorageError(err)
	}
	objects := make([]*storagecap.ProviderObject, 0)
	if output != nil {
		objects = make([]*storagecap.ProviderObject, 0, len(output.Objects))
		for _, object := range output.Objects {
			providerKey := publicKeyFromLocalObject(object, isFiles)
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
	namespace, storagePrefix, isFiles, err := resolveLocalStoragePrefix(in.Prefix)
	if err != nil {
		return nil, err
	}
	cursor := ""
	if strings.TrimSpace(in.Cursor) != "" {
		_, cursorKey, _, err := resolveLocalStorageTarget(in.Cursor)
		if err != nil {
			return nil, err
		}
		cursor = cursorKey
	}
	limit := normalizeStorageListLimit(in.Limit)
	output, err := p.storage.ListCursor(ctx, storage.ListCursorInput{
		Namespace: namespace,
		Prefix:    storagePrefix,
		Cursor:    cursor,
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
		providerKey := publicKeyFromLocalObject(object, isFiles)
		if providerKey == "" {
			continue
		}
		result.Objects = append(result.Objects, p.providerObject(providerKey, object, ""))
	}
	if next := strings.TrimSpace(output.NextCursor); next != "" {
		if isFiles {
			result.NextCursor = hostFilesProviderKeyPrefix + next
		} else {
			result.NextCursor = providerKeyFromLocalStorageKey(next)
		}
	}
	return result, nil
}

// Stat reads scoped object metadata.
func (p *localStorageProvider) Stat(ctx context.Context, in storagecap.ProviderStatInput) (*storagecap.ProviderStatOutput, error) {
	if err := p.ensureAvailable(); err != nil {
		return nil, err
	}
	namespace, storageKey, publicKey, err := resolveLocalStorageTarget(in.Key)
	if err != nil {
		return nil, err
	}
	output, err := p.storage.Stat(ctx, storage.StatInput{Namespace: namespace, Key: storageKey})
	if err != nil {
		return nil, mapLocalStorageError(err)
	}
	if output == nil || !output.Found {
		return &storagecap.ProviderStatOutput{Found: false}, nil
	}
	return &storagecap.ProviderStatOutput{Object: p.providerObject(publicKey, output.Object, ""), Found: true}, nil
}

// BatchStat reads metadata for explicit scoped object keys.
func (p *localStorageProvider) BatchStat(ctx context.Context, in storagecap.ProviderBatchStatInput) (*storagecap.ProviderBatchStatOutput, error) {
	if err := p.ensureAvailable(); err != nil {
		return nil, err
	}
	result := &storagecap.ProviderBatchStatOutput{Objects: []*storagecap.ProviderObject{}}
	// BatchStat is namespace-scoped; split files vs plugins.
	type item struct {
		namespace  string
		storageKey string
		publicKey  string
	}
	items := make([]item, 0, len(in.Keys))
	for _, key := range in.Keys {
		namespace, storageKey, publicKey, err := resolveLocalStorageTarget(key)
		if err != nil {
			return nil, err
		}
		items = append(items, item{namespace: namespace, storageKey: storageKey, publicKey: publicKey})
	}
	for _, namespace := range []string{storage.NamespaceFiles, storage.NamespacePlugins} {
		var keys []string
		keyByStorage := map[string]string{}
		for _, it := range items {
			if it.namespace != namespace {
				continue
			}
			keys = append(keys, it.storageKey)
			keyByStorage[it.storageKey] = it.publicKey
		}
		if len(keys) == 0 {
			continue
		}
		output, err := p.storage.BatchStat(ctx, storage.BatchStatInput{Namespace: namespace, Keys: keys})
		if err != nil {
			return nil, mapLocalStorageError(err)
		}
		if output == nil {
			continue
		}
		for _, object := range output.Objects {
			publicKey := keyByStorage[object.Key]
			if publicKey == "" {
				continue
			}
			result.Objects = append(result.Objects, p.providerObject(publicKey, object, ""))
		}
		for _, missingKey := range output.MissingKeys {
			if publicKey := keyByStorage[missingKey]; publicKey != "" {
				result.MissingKeys = append(result.MissingKeys, publicKey)
			}
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

// resolveLocalStorageTarget maps a provider key to host namespace + storage key.
// File-center keys use "files/<path>" and land in NamespaceFiles without the
// plugin capability root. Plugin keys keep NamespacePlugins + .capability-storage.
func resolveLocalStorageTarget(providerKey string) (namespace string, storageKey string, publicKey string, err error) {
	objectKey, err := normalizePluginStoragePath(providerKey, false)
	if err != nil {
		return "", "", "", err
	}
	if objectKey == "files" {
		return "", "", "", bizerr.NewCode(storagecap.CodeStoragePathInvalid)
	}
	if strings.HasPrefix(objectKey, hostFilesProviderKeyPrefix) {
		rest := strings.TrimPrefix(objectKey, hostFilesProviderKeyPrefix)
		if rest == "" {
			return "", "", "", bizerr.NewCode(storagecap.CodeStoragePathInvalid)
		}
		return storage.NamespaceFiles, rest, objectKey, nil
	}
	return storage.NamespacePlugins, localStorageKey(objectKey), objectKey, nil
}

func resolveLocalStoragePrefix(prefix string) (namespace string, storagePrefix string, isFiles bool, err error) {
	trimmed := strings.Trim(strings.ReplaceAll(strings.TrimSpace(prefix), "\\", "/"), "/")
	if trimmed == "" {
		// Empty prefix historically lists plugin objects only.
		return storage.NamespacePlugins, localStorageKey(""), false, nil
	}
	objectKey, err := normalizePluginStoragePath(trimmed, false)
	if err != nil {
		return "", "", false, err
	}
	if objectKey == "files" {
		return storage.NamespaceFiles, "", true, nil
	}
	if strings.HasPrefix(objectKey, hostFilesProviderKeyPrefix) {
		return storage.NamespaceFiles, strings.TrimPrefix(objectKey, hostFilesProviderKeyPrefix), true, nil
	}
	return storage.NamespacePlugins, localStorageKey(objectKey), false, nil
}

func publicKeyFromLocalObject(object *storage.Object, isFiles bool) string {
	if object == nil {
		return ""
	}
	key := strings.Trim(strings.ReplaceAll(strings.TrimSpace(object.Key), "\\", "/"), "/")
	if key == "" {
		return ""
	}
	if isFiles {
		return hostFilesProviderKeyPrefix + key
	}
	return providerKeyFromLocalStorageKey(key)
}

func localStorageKey(providerKey string) string {
	providerKey = strings.Trim(strings.ReplaceAll(strings.TrimSpace(providerKey), "\\", "/"), "/")
	if providerKey == "" {
		return localStorageProviderRootDir
	}
	return path.Join(localStorageProviderRootDir, providerKey)
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
