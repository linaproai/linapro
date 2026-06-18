// This file adapts plugin-visible object storage operations to the active
// storage provider while enforcing logical path, plugin scope, tenant scope,
// streaming writes, and bounded listing rules.

package capabilityhost

import (
	"bytes"
	"context"
	"io"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/capability/tenantcap"
)

const (
	// objectContentTypeProbeBytes is the maximum prefix read for MIME sniffing
	// before the original object stream is passed through to the provider.
	objectContentTypeProbeBytes = 512
)

// storageAdapter binds provider-backed object storage to one plugin.
type storageAdapter struct {
	runtime       storagecap.ProviderRuntime
	localProvider storagecap.Provider
	bizCtx        bizctxcap.Service
	pluginID      string
}

// newStorageAdapter creates one plugin-scoped object storage adapter.
func newStorageAdapter(
	runtime storagecap.ProviderRuntime,
	localProvider storagecap.Provider,
	bizCtx bizctxcap.Service,
	pluginID string,
) storagecap.Service {
	return &storageAdapter{
		runtime:       runtime,
		localProvider: localProvider,
		bizCtx:        bizCtx,
		pluginID:      strings.TrimSpace(pluginID),
	}
}

// Put writes one plugin object and returns plugin-visible metadata.
func (s *storageAdapter) Put(ctx context.Context, in storagecap.PutInput) (*storagecap.PutOutput, error) {
	objectPath, err := s.normalizeObjectPath(in.Path)
	if err != nil {
		return nil, err
	}
	body, probeBytes, err := storageBody(in.Body)
	if err != nil {
		return nil, err
	}
	providerID, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, err
	}
	object, err := provider.Put(ctx, storagecap.ProviderPutInput{
		Key:         s.objectKey(ctx, objectPath),
		Body:        body,
		Size:        in.Size,
		ContentType: detectObjectContentType(in.ContentType, probeBytes, objectPath),
		Overwrite:   in.Overwrite,
	})
	if err != nil {
		return nil, err
	}
	return &storagecap.PutOutput{Object: s.providerObject(objectPath, providerID, object)}, nil
}

// Get reads one plugin object.
func (s *storageAdapter) Get(ctx context.Context, in storagecap.GetInput) (*storagecap.GetOutput, error) {
	objectPath, err := s.normalizeObjectPath(in.Path)
	if err != nil {
		return nil, err
	}
	providerID, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, err
	}
	output, err := provider.Get(ctx, storagecap.ProviderGetInput{Key: s.objectKey(ctx, objectPath)})
	if err != nil {
		return nil, err
	}
	if output == nil || !output.Found {
		return &storagecap.GetOutput{Found: false}, nil
	}
	return &storagecap.GetOutput{
		Object: s.providerObject(objectPath, providerID, output.Object),
		Body:   output.Body,
		Found:  true,
	}, nil
}

// Delete removes one plugin object.
func (s *storageAdapter) Delete(ctx context.Context, in storagecap.DeleteInput) error {
	objectPath, err := s.normalizeObjectPath(in.Path)
	if err != nil {
		return err
	}
	_, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return err
	}
	return provider.Delete(ctx, storagecap.ProviderDeleteInput{Key: s.objectKey(ctx, objectPath)})
}

// DeleteMany removes an explicit bounded set of plugin objects.
func (s *storageAdapter) DeleteMany(ctx context.Context, in storagecap.DeleteManyInput) error {
	paths, err := s.normalizeBatchObjectPaths(in.Paths)
	if err != nil {
		return err
	}
	_, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(paths))
	for _, objectPath := range paths {
		keys = append(keys, s.objectKey(ctx, objectPath))
	}
	return provider.DeleteMany(ctx, storagecap.ProviderDeleteManyInput{Keys: keys})
}

// List returns bounded plugin object metadata under one logical prefix.
func (s *storageAdapter) List(ctx context.Context, in storagecap.ListInput) (*storagecap.ListOutput, error) {
	prefix, err := s.normalizeListPrefix(in.Prefix)
	if err != nil {
		return nil, err
	}
	limit := normalizeStorageListLimit(in.Limit)
	providerID, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, err
	}
	providerOutput, err := provider.List(ctx, storagecap.ProviderListInput{
		Prefix: s.objectKey(ctx, prefix),
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}
	objects := make([]*storagecap.Object, 0)
	if providerOutput != nil {
		objects = make([]*storagecap.Object, 0, len(providerOutput.Objects))
		for _, object := range providerOutput.Objects {
			logicalPath := s.logicalPathFromKey(ctx, object)
			if logicalPath == "" {
				continue
			}
			objects = append(objects, s.providerObject(logicalPath, providerID, object))
		}
	}
	return &storagecap.ListOutput{Objects: objects, Limit: limit}, nil
}

// ListCursor returns bounded plugin object metadata under one logical prefix.
func (s *storageAdapter) ListCursor(ctx context.Context, in storagecap.ListCursorInput) (*storagecap.ListCursorOutput, error) {
	prefix, err := s.normalizeListPrefix(in.Prefix)
	if err != nil {
		return nil, err
	}
	cursor := ""
	if strings.TrimSpace(in.Cursor) != "" {
		cursor, err = s.normalizeObjectPath(in.Cursor)
		if err != nil {
			return nil, err
		}
		if cursor != prefix && !strings.HasPrefix(cursor, strings.TrimSuffix(prefix, "/")+"/") {
			return nil, bizerr.NewCode(storagecap.CodeStoragePathInvalid)
		}
	}
	limit := normalizeStorageListLimit(in.Limit)
	providerID, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, err
	}
	providerCursor := ""
	if cursor != "" {
		providerCursor = s.objectKey(ctx, cursor)
	}
	providerOutput, err := provider.ListCursor(ctx, storagecap.ProviderListCursorInput{
		Prefix: s.objectKey(ctx, prefix),
		Cursor: providerCursor,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}
	output := &storagecap.ListCursorOutput{Objects: []*storagecap.Object{}, Limit: limit}
	if providerOutput == nil {
		return output, nil
	}
	for _, object := range providerOutput.Objects {
		logicalPath := s.logicalPathFromKey(ctx, object)
		if logicalPath == "" {
			continue
		}
		output.Objects = append(output.Objects, s.providerObject(logicalPath, providerID, object))
	}
	output.NextCursor = s.logicalPathFromKey(ctx, &storagecap.ProviderObject{Key: providerOutput.NextCursor})
	return output, nil
}

// Stat reads plugin object metadata.
func (s *storageAdapter) Stat(ctx context.Context, in storagecap.StatInput) (*storagecap.StatOutput, error) {
	objectPath, err := s.normalizeObjectPath(in.Path)
	if err != nil {
		return nil, err
	}
	providerID, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, err
	}
	output, err := provider.Stat(ctx, storagecap.ProviderStatInput{Key: s.objectKey(ctx, objectPath)})
	if err != nil {
		return nil, err
	}
	if output == nil || !output.Found {
		return &storagecap.StatOutput{Found: false}, nil
	}
	return &storagecap.StatOutput{Object: s.providerObject(objectPath, providerID, output.Object), Found: true}, nil
}

// BatchStat reads plugin object metadata for an explicit bounded path set.
func (s *storageAdapter) BatchStat(ctx context.Context, in storagecap.BatchStatInput) (*storagecap.BatchStatOutput, error) {
	paths, err := s.normalizeBatchObjectPaths(in.Paths)
	if err != nil {
		return nil, err
	}
	providerID, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, err
	}
	pathByKey := make(map[string]string, len(paths))
	keys := make([]string, 0, len(paths))
	for _, objectPath := range paths {
		key := s.objectKey(ctx, objectPath)
		pathByKey[key] = objectPath
		keys = append(keys, key)
	}
	providerOutput, err := provider.BatchStat(ctx, storagecap.ProviderBatchStatInput{Keys: keys})
	if err != nil {
		return nil, err
	}
	found := make(map[string]struct{}, len(paths))
	output := &storagecap.BatchStatOutput{Objects: []*storagecap.Object{}}
	if providerOutput != nil {
		for _, object := range providerOutput.Objects {
			logicalPath := pathByKey[strings.TrimSpace(object.Key)]
			if logicalPath == "" {
				continue
			}
			found[logicalPath] = struct{}{}
			output.Objects = append(output.Objects, s.providerObject(logicalPath, providerID, object))
		}
	}
	for _, objectPath := range paths {
		if _, ok := found[objectPath]; !ok {
			output.MissingPaths = append(output.MissingPaths, objectPath)
		}
	}
	return output, nil
}

// ProviderStatuses returns registered provider status snapshots.
func (s *storageAdapter) ProviderStatuses(ctx context.Context) ([]*storagecap.ProviderStatus, error) {
	return storagecap.ProviderStatuses(ctx, s.runtime, s.localProvider), nil
}

// normalizeObjectPath validates one plugin logical object path.
func (s *storageAdapter) normalizeObjectPath(rawPath string) (string, error) {
	if err := s.validateServiceScope(); err != nil {
		return "", err
	}
	normalized, err := normalizePluginStoragePath(rawPath, false)
	if err != nil {
		return "", err
	}
	return normalized, nil
}

// normalizeListPrefix validates one plugin logical list prefix.
func (s *storageAdapter) normalizeListPrefix(rawPrefix string) (string, error) {
	if err := s.validateServiceScope(); err != nil {
		return "", err
	}
	return normalizePluginStoragePath(rawPrefix, false)
}

// normalizeBatchObjectPaths validates, deduplicates, and bounds explicit object paths.
func (s *storageAdapter) normalizeBatchObjectPaths(rawPaths []string) ([]string, error) {
	if err := s.validateServiceScope(); err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(rawPaths))
	paths := make([]string, 0, len(rawPaths))
	totalBytes := 0
	for _, rawPath := range rawPaths {
		normalized, err := normalizePluginStoragePath(rawPath, false)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		paths = append(paths, normalized)
		totalBytes += len([]byte(normalized))
		if len(paths) > storagecap.MaxBatchPathCount || totalBytes > storagecap.MaxBatchPathBytes {
			return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", storagecap.MaxBatchPathCount))
		}
	}
	return paths, nil
}

// validateServiceScope verifies storage calls are plugin-bound and provider-backed.
func (s *storageAdapter) validateServiceScope() error {
	if s == nil || s.localProvider == nil {
		return bizerr.NewCode(storagecap.CodeStorageProviderUnavailable)
	}
	if strings.TrimSpace(s.pluginID) == "" {
		return bizerr.NewCode(storagecap.CodeStoragePluginIDRequired)
	}
	return nil
}

// objectKey maps a plugin logical path into a provider object key.
func (s *storageAdapter) objectKey(ctx context.Context, logicalPath string) string {
	tenantID := s.currentTenantID(ctx)
	if tenantID > 0 {
		return path.Join(
			"plugins",
			s.pluginID,
			"tenant",
			strconv.FormatInt(tenantID, 10),
			logicalPath,
		)
	}
	return path.Join("plugins", s.pluginID, "platform", logicalPath)
}

// logicalPathFromKey converts provider metadata back into the plugin logical path.
func (s *storageAdapter) logicalPathFromKey(ctx context.Context, object *storagecap.ProviderObject) string {
	if object == nil {
		return ""
	}
	keyPrefix := path.Dir(s.objectKey(ctx, "__placeholder__"))
	key := strings.TrimPrefix(strings.TrimSpace(object.Key), keyPrefix+"/")
	if key == strings.TrimSpace(object.Key) || key == "" {
		return ""
	}
	return key
}

// providerObject converts provider metadata into plugin-visible metadata.
func (s *storageAdapter) providerObject(
	logicalPath string,
	_ string,
	object *storagecap.ProviderObject,
) *storagecap.Object {
	if object == nil {
		return nil
	}
	visibility := strings.TrimSpace(object.Visibility)
	if visibility == "" {
		visibility = storagecap.VisibilityPrivate
	}
	return &storagecap.Object{
		Path:        logicalPath,
		Size:        object.Size,
		ContentType: object.ContentType,
		ETag:        object.ETag,
		UpdatedAt:   object.UpdatedAt,
		Visibility:  visibility,
	}
}

// currentTenantID returns the current plugin storage tenant scope.
func (s *storageAdapter) currentTenantID(ctx context.Context) int64 {
	current := bizctxcap.CurrentFromContext(ctx)
	if current.TenantID > 0 {
		return int64(current.TenantID)
	}
	if current.PlatformBypass {
		return int64(tenantcap.PLATFORM)
	}
	if s != nil && s.bizCtx != nil {
		if tenantID := s.bizCtx.Current(ctx).TenantID; tenantID > 0 {
			return int64(tenantID)
		}
	}
	return int64(tenantcap.PLATFORM)
}

// normalizePluginStoragePath canonicalizes a logical object path or prefix.
func normalizePluginStoragePath(rawPath string, allowEmpty bool) (string, error) {
	trimmed := strings.ReplaceAll(strings.TrimSpace(rawPath), "\\", "/")
	if trimmed == "" || trimmed == "." {
		if allowEmpty {
			return "", nil
		}
		return "", bizerr.NewCode(storagecap.CodeStoragePathRequired)
	}
	if strings.Contains(trimmed, "://") || strings.HasPrefix(trimmed, "/") {
		return "", bizerr.NewCode(storagecap.CodeStoragePathInvalid)
	}
	if len(trimmed) >= 2 && ((trimmed[0] >= 'A' && trimmed[0] <= 'Z') || (trimmed[0] >= 'a' && trimmed[0] <= 'z')) && trimmed[1] == ':' {
		return "", bizerr.NewCode(storagecap.CodeStoragePathInvalid)
	}
	normalized := path.Clean(trimmed)
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") {
		return "", bizerr.NewCode(storagecap.CodeStoragePathInvalid)
	}
	if len([]byte(normalized)) > storagecap.MaxLogicalPathBytes {
		return "", bizerr.NewCode(storagecap.CodeStoragePathTooLong, bizerr.P("maxBytes", storagecap.MaxLogicalPathBytes))
	}
	return normalized, nil
}

// normalizeStorageListLimit enforces bounded list calls.
func normalizeStorageListLimit(limit int) int {
	if limit <= 0 {
		return storagecap.DefaultListLimit
	}
	if limit > storagecap.MaxListLimit {
		return storagecap.MaxListLimit
	}
	return limit
}

// storageBody prepares a streaming body for provider operations while reading
// only a small prefix for content-type detection.
func storageBody(reader io.Reader) (io.Reader, []byte, error) {
	if reader == nil {
		reader = bytes.NewReader(nil)
	}
	probe, err := io.ReadAll(io.LimitReader(reader, objectContentTypeProbeBytes))
	if err != nil {
		return nil, nil, err
	}
	return io.MultiReader(bytes.NewReader(probe), reader), probe, nil
}

// detectObjectContentType derives the best content type from request, body, or extension.
func detectObjectContentType(rawContentType string, body []byte, objectPath string) string {
	contentType := strings.TrimSpace(rawContentType)
	if contentType != "" {
		if mediaType, _, err := mime.ParseMediaType(contentType); err == nil {
			contentType = mediaType
		}
		contentType = strings.ToLower(strings.TrimSpace(contentType))
	}
	if contentType != "" {
		return contentType
	}
	if len(body) > 0 {
		return strings.ToLower(strings.TrimSpace(strings.Split(http.DetectContentType(body), ";")[0]))
	}
	extension := strings.ToLower(path.Ext(objectPath))
	if extension != "" {
		if detected := mime.TypeByExtension(extension); strings.TrimSpace(detected) != "" {
			return strings.ToLower(strings.TrimSpace(strings.Split(detected, ";")[0]))
		}
	}
	return "application/octet-stream"
}
