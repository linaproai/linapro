// This file implements a host Storage Service facade that keeps the wide
// storage.Service contract while routing NamespaceFiles content operations
// through storagecap.ResolveProvider (same rules as plugin Storage()). Other
// namespaces continue to use the local Service implementation. File-center
// keys are stored as "files/<key>" on the active provider so they do not
// collide with plugin-scoped keys.

package storage

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
)

const (
	// filesProviderKeyPrefix isolates file-center objects on shared providers.
	filesProviderKeyPrefix = "files/"
)

// resolvingService implements Service by resolving the active object backend for
// the files namespace and delegating plugin/other namespaces to the local disk
// service used as ResolveProvider's local fallback.
type resolvingService struct {
	local         Service
	runtime       storagecap.ProviderRuntime
	localProvider storagecap.Provider
}

// NewResolvingService creates a Storage Service that uses unified provider
// selection for NamespaceFiles and the provided local Service for other
// namespaces. local must be the same disk-backed Service used by
// NewLocalStorageProvider to avoid recursive resolution.
func NewResolvingService(
	local Service,
	runtime storagecap.ProviderRuntime,
	localProvider storagecap.Provider,
) Service {
	return &resolvingService{
		local:         local,
		runtime:       runtime,
		localProvider: localProvider,
	}
}

// Put writes one object. NamespaceFiles uses the active resolved provider.
func (s *resolvingService) Put(ctx context.Context, in PutInput) (*PutOutput, error) {
	if !isFilesNamespace(in.Namespace) {
		return s.local.Put(ctx, in)
	}
	providerKey, err := filesProviderKey(in.Key)
	if err != nil {
		return nil, err
	}
	_, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, mapResolveError(err)
	}
	object, err := provider.Put(ctx, storagecap.ProviderPutInput{
		Key:         providerKey,
		Body:        in.Body,
		Size:        in.Size,
		ContentType: in.ContentType,
		Overwrite:   in.Overwrite,
	})
	if err != nil {
		return nil, mapResolveError(err)
	}
	return &PutOutput{Object: objectFromProvider(object, in.Key)}, nil
}

// Get reads one object. For NamespaceFiles, falls back to the local provider
// when the active backend is cloud and the key is missing (legacy local files).
func (s *resolvingService) Get(ctx context.Context, in GetInput) (*GetOutput, error) {
	if !isFilesNamespace(in.Namespace) {
		return s.local.Get(ctx, in)
	}
	providerKey, err := filesProviderKey(in.Key)
	if err != nil {
		return nil, err
	}
	providerID, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, mapResolveError(err)
	}
	output, err := provider.Get(ctx, storagecap.ProviderGetInput{Key: providerKey})
	if err != nil {
		return nil, mapResolveError(err)
	}
	if output != nil && output.Found && output.Body != nil {
		return &GetOutput{
			Object: objectFromProvider(output.Object, in.Key),
			Body:   output.Body,
			Found:  true,
		}, nil
	}
	if providerID == storagecap.LocalProviderID || s.localProvider == nil {
		return &GetOutput{Found: false}, nil
	}
	localOut, err := s.localProvider.Get(ctx, storagecap.ProviderGetInput{Key: providerKey})
	if err != nil {
		return nil, mapResolveError(err)
	}
	if localOut == nil || !localOut.Found || localOut.Body == nil {
		return &GetOutput{Found: false}, nil
	}
	return &GetOutput{
		Object: objectFromProvider(localOut.Object, in.Key),
		Body:   localOut.Body,
		Found:  true,
	}, nil
}

// Delete removes one object from the active backend and best-effort cleans a
// local legacy copy when the active backend is not local.
func (s *resolvingService) Delete(ctx context.Context, in DeleteInput) error {
	if !isFilesNamespace(in.Namespace) {
		return s.local.Delete(ctx, in)
	}
	providerKey, err := filesProviderKey(in.Key)
	if err != nil {
		return err
	}
	providerID, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return mapResolveError(err)
	}
	if err = provider.Delete(ctx, storagecap.ProviderDeleteInput{Key: providerKey}); err != nil {
		return mapResolveError(err)
	}
	if providerID != storagecap.LocalProviderID && s.localProvider != nil {
		_ = s.localProvider.Delete(ctx, storagecap.ProviderDeleteInput{Key: providerKey})
	}
	return nil
}

// DeleteMany removes explicit keys; files namespace is resolved per call group.
func (s *resolvingService) DeleteMany(ctx context.Context, in DeleteManyInput) error {
	if !isFilesNamespace(in.Namespace) {
		return s.local.DeleteMany(ctx, in)
	}
	for _, key := range in.Keys {
		if err := s.Delete(ctx, DeleteInput{Namespace: NamespaceFiles, Key: key}); err != nil {
			return err
		}
	}
	return nil
}

// List lists objects. NamespaceFiles lists via the active provider under files/.
func (s *resolvingService) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	if !isFilesNamespace(in.Namespace) {
		return s.local.List(ctx, in)
	}
	providerKey, err := filesProviderPrefix(in.Prefix)
	if err != nil {
		return nil, err
	}
	_, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, mapResolveError(err)
	}
	output, err := provider.List(ctx, storagecap.ProviderListInput{Prefix: providerKey, Limit: in.Limit})
	if err != nil {
		return nil, mapResolveError(err)
	}
	return listOutputFromProvider(output, in.Limit), nil
}

// ListCursor lists objects with a cursor under NamespaceFiles via the active provider.
func (s *resolvingService) ListCursor(ctx context.Context, in ListCursorInput) (*ListCursorOutput, error) {
	if !isFilesNamespace(in.Namespace) {
		return s.local.ListCursor(ctx, in)
	}
	providerKey, err := filesProviderPrefix(in.Prefix)
	if err != nil {
		return nil, err
	}
	cursor := ""
	if strings.TrimSpace(in.Cursor) != "" {
		cursor, err = filesProviderKey(in.Cursor)
		if err != nil {
			return nil, err
		}
	}
	_, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, mapResolveError(err)
	}
	output, err := provider.ListCursor(ctx, storagecap.ProviderListCursorInput{
		Prefix: providerKey,
		Cursor: cursor,
		Limit:  in.Limit,
	})
	if err != nil {
		return nil, mapResolveError(err)
	}
	return listCursorOutputFromProvider(output, in.Limit), nil
}

// Stat returns metadata for one NamespaceFiles object via the active provider.
func (s *resolvingService) Stat(ctx context.Context, in StatInput) (*StatOutput, error) {
	if !isFilesNamespace(in.Namespace) {
		return s.local.Stat(ctx, in)
	}
	providerKey, err := filesProviderKey(in.Key)
	if err != nil {
		return nil, err
	}
	providerID, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, mapResolveError(err)
	}
	output, err := provider.Stat(ctx, storagecap.ProviderStatInput{Key: providerKey})
	if err != nil {
		return nil, mapResolveError(err)
	}
	if output != nil && output.Found {
		return &StatOutput{Object: objectFromProvider(output.Object, in.Key), Found: true}, nil
	}
	if providerID == storagecap.LocalProviderID || s.localProvider == nil {
		return &StatOutput{Found: false}, nil
	}
	localOut, err := s.localProvider.Stat(ctx, storagecap.ProviderStatInput{Key: providerKey})
	if err != nil {
		return nil, mapResolveError(err)
	}
	if localOut == nil || !localOut.Found {
		return &StatOutput{Found: false}, nil
	}
	return &StatOutput{Object: objectFromProvider(localOut.Object, in.Key), Found: true}, nil
}

// BatchStat returns metadata for explicit NamespaceFiles keys.
func (s *resolvingService) BatchStat(ctx context.Context, in BatchStatInput) (*BatchStatOutput, error) {
	if !isFilesNamespace(in.Namespace) {
		return s.local.BatchStat(ctx, in)
	}
	keys := make([]string, 0, len(in.Keys))
	keyByProvider := map[string]string{}
	for _, key := range in.Keys {
		providerKey, err := filesProviderKey(key)
		if err != nil {
			return nil, err
		}
		keys = append(keys, providerKey)
		keyByProvider[providerKey] = key
	}
	_, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, mapResolveError(err)
	}
	output, err := provider.BatchStat(ctx, storagecap.ProviderBatchStatInput{Keys: keys})
	if err != nil {
		return nil, mapResolveError(err)
	}
	result := &BatchStatOutput{}
	if output == nil {
		return result, nil
	}
	for _, object := range output.Objects {
		hostKey := keyByProvider[strings.TrimSpace(object.Key)]
		if hostKey == "" {
			hostKey = stripFilesPrefix(object.Key)
		}
		result.Objects = append(result.Objects, objectFromProvider(object, hostKey))
	}
	for _, missing := range output.MissingKeys {
		if hostKey := keyByProvider[missing]; hostKey != "" {
			result.MissingKeys = append(result.MissingKeys, hostKey)
		} else {
			result.MissingKeys = append(result.MissingKeys, stripFilesPrefix(missing))
		}
	}
	return result, nil
}

// filesProviderID returns the currently resolved provider identifier for
// NamespaceFiles (local or the unique enabled cloud plugin id).
func (s *resolvingService) filesProviderID(ctx context.Context) (string, error) {
	providerID, _, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return "", mapResolveError(err)
	}
	return providerID, nil
}

// ResolveFilesProviderID returns the active object-backend id for NamespaceFiles
// when svc is a resolving Storage Service. Provider selection errors such as
// multi-plugin conflicts are returned so callers can fail closed before any
// metadata write or hash-dedup shortcut. Non-resolving services return local
// with a nil error.
func ResolveFilesProviderID(ctx context.Context, svc Service) (string, error) {
	type providerAware interface {
		filesProviderID(context.Context) (string, error)
	}
	if aware, ok := svc.(providerAware); ok {
		id, err := aware.filesProviderID(ctx)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(id) != "" {
			return id, nil
		}
	}
	return storagecap.LocalProviderID, nil
}

// FilesProviderID returns the active object-backend id for NamespaceFiles when
// svc is a resolving Storage Service; otherwise it returns "local". Selection
// failures fall back to "local" for best-effort metadata labeling only; writers
// must use ResolveFilesProviderID to fail closed on conflict.
func FilesProviderID(ctx context.Context, svc Service) string {
	id, err := ResolveFilesProviderID(ctx, svc)
	if err != nil || strings.TrimSpace(id) == "" {
		return storagecap.LocalProviderID
	}
	return id
}

func isFilesNamespace(namespace string) bool {
	return strings.TrimSpace(namespace) == NamespaceFiles
}

func filesProviderKey(rawKey string) (string, error) {
	trimmed := strings.ReplaceAll(strings.TrimSpace(rawKey), "\\", "/")
	if trimmed == "" || strings.Contains(trimmed, "://") || strings.HasPrefix(trimmed, "/") {
		return "", ErrPathInvalid
	}
	key := strings.Trim(trimmed, "/")
	if key == "" || key == ".." || strings.HasPrefix(key, "../") {
		return "", ErrPathInvalid
	}
	return filesProviderKeyPrefix + key, nil
}

func filesProviderPrefix(rawPrefix string) (string, error) {
	trimmed := strings.Trim(strings.ReplaceAll(strings.TrimSpace(rawPrefix), "\\", "/"), "/")
	if trimmed == "" {
		return filesProviderKeyPrefix, nil
	}
	return filesProviderKey(trimmed)
}

func stripFilesPrefix(key string) string {
	key = strings.Trim(strings.ReplaceAll(strings.TrimSpace(key), "\\", "/"), "/")
	return strings.TrimPrefix(key, filesProviderKeyPrefix)
}

func objectFromProvider(object *storagecap.ProviderObject, hostKey string) *Object {
	if object == nil {
		return &Object{Key: hostKey}
	}
	return &Object{
		Key:         hostKey,
		Size:        object.Size,
		ContentType: object.ContentType,
		ETag:        object.ETag,
		UpdatedAt:   object.UpdatedAt,
	}
}

func listOutputFromProvider(output *storagecap.ProviderListOutput, limit int) *ListOutput {
	result := &ListOutput{Limit: limit, Objects: []*Object{}}
	if output == nil {
		return result
	}
	for _, object := range output.Objects {
		result.Objects = append(result.Objects, objectFromProvider(object, stripFilesPrefix(object.Key)))
	}
	return result
}

func listCursorOutputFromProvider(output *storagecap.ProviderListCursorOutput, limit int) *ListCursorOutput {
	result := &ListCursorOutput{Limit: limit, Objects: []*Object{}}
	if output == nil {
		return result
	}
	for _, object := range output.Objects {
		result.Objects = append(result.Objects, objectFromProvider(object, stripFilesPrefix(object.Key)))
	}
	result.NextCursor = stripFilesPrefix(output.NextCursor)
	return result
}

func mapResolveError(err error) error {
	if err == nil {
		return nil
	}
	if bizerr.Is(err, storagecap.CodeStorageProviderConflict) {
		return err
	}
	if bizerr.Is(err, storagecap.CodeStoragePathInvalid) || bizerr.Is(err, storagecap.CodeStoragePathRequired) {
		return ErrPathInvalid
	}
	if bizerr.Is(err, storagecap.CodeStorageObjectExists) {
		return ErrObjectExist
	}
	if bizerr.Is(err, storagecap.CodeStorageProviderUnavailable) {
		return ErrUnavailable
	}
	return err
}
