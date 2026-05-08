// This file implements the governed storage host service backed by one
// plugin-scoped local directory tree with authorized logical path matching.

package wasm

import (
	"context"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/internal/service/config"
	bridgehostcall "lina-core/pkg/pluginbridge/hostcall"
	bridgehostservice "lina-core/pkg/pluginbridge/hostservice"
	"lina-core/pkg/pluginfs"
)

// Local directory layout and pagination limits for the governed storage service.
const (
	storageHostServiceRootDirName = ".host-services"
	storageHostServiceDirName     = "storage"
	defaultStorageListLimit       = 100
	maxStorageListLimit           = 1000
)

// storageConfigSvc provides the runtime storage root configuration.
var storageConfigSvc = config.New()

// storageResourceConfig stores the resolved storage root and visibility for one plugin.
type storageResourceConfig struct {
	rootDir    string
	visibility string
}

// dispatchStorageHostService routes storage host service methods to the local
// governed storage implementation.
func dispatchStorageHostService(
	ctx context.Context,
	hcc *hostCallContext,
	targetPath string,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	if strings.TrimSpace(targetPath) == "" {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			"storage host service requires one authorized target path",
		)
	}

	resourceConfig, err := buildStorageResourceConfig(ctx, hcc)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}

	switch method {
	case bridgehostservice.HostServiceMethodStoragePut:
		return handleStoragePut(resourceConfig, targetPath, payload)
	case bridgehostservice.HostServiceMethodStorageGet:
		return handleStorageGet(resourceConfig, targetPath, payload)
	case bridgehostservice.HostServiceMethodStorageDelete:
		return handleStorageDelete(resourceConfig, targetPath, payload)
	case bridgehostservice.HostServiceMethodStorageList:
		return handleStorageList(resourceConfig, targetPath, payload)
	case bridgehostservice.HostServiceMethodStorageStat:
		return handleStorageStat(resourceConfig, targetPath, payload)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported storage host service method: "+method,
		)
	}
}

// handleStoragePut writes one governed storage object.
func handleStoragePut(
	resourceConfig *storageResourceConfig,
	targetPath string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceStoragePutRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	objectPath, err := normalizeStorageObjectPath(request.Path)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	if err = validateStorageRequestTarget(targetPath, objectPath); err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	if err = resourceConfig.validateWritePolicy(int64(len(request.Body))); err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	absolutePath, err := resourceConfig.resolveObjectPath(objectPath)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	_, exists, err := lookupStorageFileInfo(absolutePath)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	if exists && !request.Overwrite {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, "storage object already exists")
	}

	if err = gfile.Mkdir(filepath.Dir(absolutePath)); err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	if err = os.WriteFile(absolutePath, request.Body, 0o644); err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}

	fileInfo, _, err := lookupStorageFileInfo(absolutePath)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}

	response := &bridgehostservice.HostServiceStoragePutResponse{
		Object: buildStorageObjectSnapshot(
			objectPath,
			fileInfo,
			detectStorageContentType(request.ContentType, request.Body, objectPath),
			resourceConfig.visibility,
		),
	}
	return bridgehostcall.NewHostCallSuccessResponse(
		bridgehostservice.MarshalHostServiceStoragePutResponse(response),
	)
}

// handleStorageGet reads one governed storage object.
func handleStorageGet(
	resourceConfig *storageResourceConfig,
	targetPath string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceStorageGetRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	objectPath, err := normalizeStorageObjectPath(request.Path)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	if err = validateStorageRequestTarget(targetPath, objectPath); err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	absolutePath, err := resourceConfig.resolveObjectPath(objectPath)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	fileInfo, exists, err := lookupStorageFileInfo(absolutePath)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	if !exists {
		return bridgehostcall.NewHostCallSuccessResponse(
			bridgehostservice.MarshalHostServiceStorageGetResponse(&bridgehostservice.HostServiceStorageGetResponse{Found: false}),
		)
	}

	body, err := os.ReadFile(absolutePath)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}

	response := &bridgehostservice.HostServiceStorageGetResponse{
		Found: true,
		Object: buildStorageObjectSnapshot(
			objectPath,
			fileInfo,
			detectStorageContentType("", body, objectPath),
			resourceConfig.visibility,
		),
		Body: body,
	}
	return bridgehostcall.NewHostCallSuccessResponse(
		bridgehostservice.MarshalHostServiceStorageGetResponse(response),
	)
}

// handleStorageDelete deletes one governed storage object.
func handleStorageDelete(
	resourceConfig *storageResourceConfig,
	targetPath string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceStorageDeleteRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	objectPath, err := normalizeStorageObjectPath(request.Path)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	if err = validateStorageRequestTarget(targetPath, objectPath); err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	absolutePath, err := resourceConfig.resolveObjectPath(objectPath)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	if err = os.Remove(absolutePath); err != nil && !os.IsNotExist(err) {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	return bridgehostcall.NewHostCallEmptySuccessResponse()
}

// handleStorageList lists governed storage objects under the authorized prefix.
func handleStorageList(
	resourceConfig *storageResourceConfig,
	targetPath string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceStorageListRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	prefix, err := normalizeStorageListPrefix(request.Prefix)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	if err = validateStorageRequestTarget(targetPath, prefix); err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	limit := int(request.Limit)
	if limit <= 0 {
		limit = defaultStorageListLimit
	}
	if limit > maxStorageListLimit {
		limit = maxStorageListLimit
	}

	objects, err := listStorageObjects(resourceConfig, prefix, limit)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	return bridgehostcall.NewHostCallSuccessResponse(
		bridgehostservice.MarshalHostServiceStorageListResponse(&bridgehostservice.HostServiceStorageListResponse{Objects: objects}),
	)
}

// handleStorageStat returns metadata for one governed storage object.
func handleStorageStat(
	resourceConfig *storageResourceConfig,
	targetPath string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceStorageStatRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	objectPath, err := normalizeStorageObjectPath(request.Path)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	if err = validateStorageRequestTarget(targetPath, objectPath); err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	absolutePath, err := resourceConfig.resolveObjectPath(objectPath)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	fileInfo, exists, err := lookupStorageFileInfo(absolutePath)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	if !exists {
		return bridgehostcall.NewHostCallSuccessResponse(
			bridgehostservice.MarshalHostServiceStorageStatResponse(&bridgehostservice.HostServiceStorageStatResponse{Found: false}),
		)
	}

	return bridgehostcall.NewHostCallSuccessResponse(
		bridgehostservice.MarshalHostServiceStorageStatResponse(&bridgehostservice.HostServiceStorageStatResponse{
			Found: true,
			Object: buildStorageObjectSnapshot(
				objectPath,
				fileInfo,
				detectStorageContentType("", nil, objectPath),
				resourceConfig.visibility,
			),
		}),
	)
}

// buildStorageResourceConfig resolves the plugin-scoped storage root for the current host call.
func buildStorageResourceConfig(
	ctx context.Context,
	hcc *hostCallContext,
) (*storageResourceConfig, error) {
	if hcc == nil {
		return nil, gerror.New("host call context not available")
	}

	rootDir := filepath.Join(
		storageConfigSvc.GetPluginDynamicStoragePath(ctx),
		storageHostServiceRootDirName,
		storageHostServiceDirName,
		hcc.pluginID,
	)
	absoluteRootDir, absErr := filepath.Abs(rootDir)
	if absErr != nil {
		return nil, gerror.Wrap(absErr, "resolve storage resource root directory failed")
	}

	return &storageResourceConfig{
		rootDir:    filepath.Clean(absoluteRootDir),
		visibility: bridgehostservice.HostServiceStorageVisibilityPrivate,
	}, nil
}

// validateWritePolicy enforces basic write constraints for storage uploads.
func (resourceConfig *storageResourceConfig) validateWritePolicy(bodySize int64) error {
	if resourceConfig == nil {
		return gerror.New("storage resource config is nil")
	}
	if bodySize < 0 {
		return gerror.New("storage body size is invalid")
	}
	return nil
}

// resolveObjectPath resolves one logical object path under the plugin storage root.
func (resourceConfig *storageResourceConfig) resolveObjectPath(objectPath string) (string, error) {
	normalizedObjectPath, err := normalizeStorageObjectPath(objectPath)
	if err != nil {
		return "", err
	}

	fullPath := filepath.Clean(filepath.Join(resourceConfig.rootDir, filepath.FromSlash(normalizedObjectPath)))
	rootPath := filepath.Clean(resourceConfig.rootDir)
	if fullPath != rootPath && !strings.HasPrefix(fullPath, rootPath+string(filepath.Separator)) {
		return "", gerror.Newf("storage object path escapes root: %s", objectPath)
	}
	return fullPath, nil
}

// normalizeStorageObjectPath canonicalizes one logical object path.
func normalizeStorageObjectPath(rawPath string) (string, error) {
	return pluginfs.NormalizeRelativePath(rawPath)
}

// normalizeStorageListPrefix canonicalizes one required list prefix.
func normalizeStorageListPrefix(rawPrefix string) (string, error) {
	trimmed := strings.TrimSpace(rawPrefix)
	if trimmed == "" {
		return "", gerror.New("storage list prefix is required")
	}
	return pluginfs.NormalizeRelativePath(trimmed)
}

// normalizeStorageAuthorizedPath canonicalizes one authorized storage target or prefix.
func normalizeStorageAuthorizedPath(rawPath string) (string, error) {
	trimmed := strings.ReplaceAll(strings.TrimSpace(rawPath), "\\", "/")
	if trimmed == "" {
		return "", gerror.New("storage target path is required")
	}
	isPrefix := strings.HasSuffix(trimmed, "/")
	base := strings.TrimSuffix(trimmed, "/")
	if base == "" {
		return "", gerror.New("storage target path is required")
	}
	normalized, err := pluginfs.NormalizeRelativePath(base)
	if err != nil {
		return "", err
	}
	if isPrefix {
		return normalized + "/", nil
	}
	return normalized, nil
}

// matchAuthorizedStoragePath returns the authorized path pattern that matches the target.
func matchAuthorizedStoragePath(specs []*bridgehostservice.HostServiceSpec, targetPath string) string {
	normalizedTarget, err := normalizeStorageAuthorizedPath(targetPath)
	if err != nil {
		return ""
	}
	// Storage authorization supports both exact object paths and directory
	// prefixes ending with `/`, so both the approval snapshot and request path
	// must be normalized before matching.
	for _, spec := range specs {
		if spec == nil || spec.Service != bridgehostservice.HostServiceStorage {
			continue
		}
		for _, authorizedPath := range spec.Paths {
			if matchStoragePathPattern(authorizedPath, normalizedTarget) {
				return authorizedPath
			}
		}
	}
	return ""
}

// matchStoragePathPattern matches exact object paths and directory-prefix patterns.
func matchStoragePathPattern(pattern string, target string) bool {
	normalizedPattern, err := normalizeStorageAuthorizedPath(pattern)
	if err != nil {
		return false
	}
	normalizedTarget, err := normalizeStorageAuthorizedPath(target)
	if err != nil {
		return false
	}
	if strings.HasSuffix(normalizedPattern, "/") {
		base := strings.TrimSuffix(normalizedPattern, "/")
		return normalizedTarget == base || strings.HasPrefix(normalizedTarget, base+"/")
	}
	return normalizedTarget == normalizedPattern
}

// validateStorageRequestTarget ensures the guest request matches the authorized target exactly.
func validateStorageRequestTarget(targetPath string, requestPath string) error {
	normalizedTarget, err := normalizeStorageAuthorizedPath(targetPath)
	if err != nil {
		return err
	}
	normalizedRequest, err := normalizeStorageAuthorizedPath(requestPath)
	if err != nil {
		return err
	}
	if normalizedTarget != normalizedRequest {
		return gerror.Newf("storage request target mismatch: target=%s request=%s", normalizedTarget, normalizedRequest)
	}
	return nil
}

// listStorageObjects scans the plugin storage root and returns matching object snapshots.
func listStorageObjects(
	resourceConfig *storageResourceConfig,
	prefix string,
	limit int,
) ([]*bridgehostservice.HostServiceStorageObject, error) {
	if resourceConfig == nil {
		return []*bridgehostservice.HostServiceStorageObject{}, nil
	}
	if !gfile.Exists(resourceConfig.rootDir) {
		return []*bridgehostservice.HostServiceStorageObject{}, nil
	}

	files, err := gfile.ScanDirFile(resourceConfig.rootDir, "*", true)
	if err != nil {
		return nil, err
	}
	sort.Strings(files)

	objects := make([]*bridgehostservice.HostServiceStorageObject, 0, len(files))
	for _, absolutePath := range files {
		fileInfo, err := os.Stat(absolutePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		relativePath, err := filepath.Rel(resourceConfig.rootDir, absolutePath)
		if err != nil {
			return nil, err
		}
		objectPath := filepath.ToSlash(relativePath)
		if prefix != "" && objectPath != prefix && !strings.HasPrefix(objectPath, prefix+"/") {
			continue
		}

		objects = append(objects, buildStorageObjectSnapshot(
			objectPath,
			fileInfo,
			detectStorageContentType("", nil, objectPath),
			resourceConfig.visibility,
		))
		if limit > 0 && len(objects) >= limit {
			break
		}
	}
	return objects, nil
}

// lookupStorageFileInfo returns file metadata while rejecting directory targets.
func lookupStorageFileInfo(absolutePath string) (os.FileInfo, bool, error) {
	fileInfo, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if fileInfo.IsDir() {
		return nil, false, gerror.Newf("storage object path points to a directory: %s", absolutePath)
	}
	return fileInfo, true, nil
}

// buildStorageObjectSnapshot maps one file info record into the protobuf storage object model.
func buildStorageObjectSnapshot(
	objectPath string,
	fileInfo os.FileInfo,
	contentType string,
	visibility string,
) *bridgehostservice.HostServiceStorageObject {
	if fileInfo == nil {
		return &bridgehostservice.HostServiceStorageObject{
			Path:        objectPath,
			ContentType: contentType,
			Visibility:  visibility,
		}
	}
	return &bridgehostservice.HostServiceStorageObject{
		Path:        objectPath,
		Size:        fileInfo.Size(),
		ContentType: contentType,
		UpdatedAt:   fileInfo.ModTime().UTC().Format(time.RFC3339Nano),
		Visibility:  visibility,
	}
}

// detectStorageContentType derives the best content type from the request, body, or extension.
func detectStorageContentType(rawContentType string, body []byte, objectPath string) string {
	contentType := strings.TrimSpace(rawContentType)
	if contentType != "" {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err == nil && strings.TrimSpace(mediaType) != "" {
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
