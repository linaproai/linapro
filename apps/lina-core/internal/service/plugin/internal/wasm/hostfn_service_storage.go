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
	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/pluginfs"
)

const (
	storageHostServiceRootDirName = ".host-services"
	storageHostServiceDirName     = "storage"
	defaultStorageListLimit       = 100
	maxStorageListLimit           = 1000
)

var storageConfigSvc = config.New()

type storageResourceConfig struct {
	rootDir    string
	visibility string
}

func dispatchStorageHostService(
	ctx context.Context,
	hcc *hostCallContext,
	targetPath string,
	method string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	if strings.TrimSpace(targetPath) == "" {
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusCapabilityDenied,
			"storage host service requires one authorized target path",
		)
	}

	resourceConfig, err := buildStorageResourceConfig(ctx, hcc)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}

	switch method {
	case pluginbridge.HostServiceMethodStoragePut:
		return handleStoragePut(resourceConfig, targetPath, payload)
	case pluginbridge.HostServiceMethodStorageGet:
		return handleStorageGet(resourceConfig, targetPath, payload)
	case pluginbridge.HostServiceMethodStorageDelete:
		return handleStorageDelete(resourceConfig, targetPath, payload)
	case pluginbridge.HostServiceMethodStorageList:
		return handleStorageList(resourceConfig, targetPath, payload)
	case pluginbridge.HostServiceMethodStorageStat:
		return handleStorageStat(resourceConfig, targetPath, payload)
	default:
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusNotFound,
			"unsupported storage host service method: "+method,
		)
	}
}

func handleStoragePut(
	resourceConfig *storageResourceConfig,
	targetPath string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	request, err := pluginbridge.UnmarshalHostServiceStoragePutRequest(payload)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	objectPath, err := normalizeStorageObjectPath(request.Path)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}
	if err = validateStorageRequestTarget(targetPath, objectPath); err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}
	if err = resourceConfig.validateWritePolicy(int64(len(request.Body))); err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	absolutePath, err := resourceConfig.resolveObjectPath(objectPath)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	_, exists, err := lookupStorageFileInfo(absolutePath)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	if exists && !request.Overwrite {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, "storage object already exists")
	}

	if err = gfile.Mkdir(filepath.Dir(absolutePath)); err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	if err = os.WriteFile(absolutePath, request.Body, 0o644); err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}

	fileInfo, _, err := lookupStorageFileInfo(absolutePath)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}

	response := &pluginbridge.HostServiceStoragePutResponse{
		Object: buildStorageObjectSnapshot(
			objectPath,
			fileInfo,
			detectStorageContentType(request.ContentType, request.Body, objectPath),
			resourceConfig.visibility,
		),
	}
	return pluginbridge.NewHostCallSuccessResponse(
		pluginbridge.MarshalHostServiceStoragePutResponse(response),
	)
}

func handleStorageGet(
	resourceConfig *storageResourceConfig,
	targetPath string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	request, err := pluginbridge.UnmarshalHostServiceStorageGetRequest(payload)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	objectPath, err := normalizeStorageObjectPath(request.Path)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}
	if err = validateStorageRequestTarget(targetPath, objectPath); err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	absolutePath, err := resourceConfig.resolveObjectPath(objectPath)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	fileInfo, exists, err := lookupStorageFileInfo(absolutePath)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	if !exists {
		return pluginbridge.NewHostCallSuccessResponse(
			pluginbridge.MarshalHostServiceStorageGetResponse(&pluginbridge.HostServiceStorageGetResponse{Found: false}),
		)
	}

	body, err := os.ReadFile(absolutePath)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}

	response := &pluginbridge.HostServiceStorageGetResponse{
		Found: true,
		Object: buildStorageObjectSnapshot(
			objectPath,
			fileInfo,
			detectStorageContentType("", body, objectPath),
			resourceConfig.visibility,
		),
		Body: body,
	}
	return pluginbridge.NewHostCallSuccessResponse(
		pluginbridge.MarshalHostServiceStorageGetResponse(response),
	)
}

func handleStorageDelete(
	resourceConfig *storageResourceConfig,
	targetPath string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	request, err := pluginbridge.UnmarshalHostServiceStorageDeleteRequest(payload)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	objectPath, err := normalizeStorageObjectPath(request.Path)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}
	if err = validateStorageRequestTarget(targetPath, objectPath); err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	absolutePath, err := resourceConfig.resolveObjectPath(objectPath)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	if err = os.Remove(absolutePath); err != nil && !os.IsNotExist(err) {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	return pluginbridge.NewHostCallEmptySuccessResponse()
}

func handleStorageList(
	resourceConfig *storageResourceConfig,
	targetPath string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	request, err := pluginbridge.UnmarshalHostServiceStorageListRequest(payload)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	prefix, err := normalizeStorageListPrefix(request.Prefix)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}
	if err = validateStorageRequestTarget(targetPath, prefix); err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
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
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	return pluginbridge.NewHostCallSuccessResponse(
		pluginbridge.MarshalHostServiceStorageListResponse(&pluginbridge.HostServiceStorageListResponse{Objects: objects}),
	)
}

func handleStorageStat(
	resourceConfig *storageResourceConfig,
	targetPath string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	request, err := pluginbridge.UnmarshalHostServiceStorageStatRequest(payload)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	objectPath, err := normalizeStorageObjectPath(request.Path)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}
	if err = validateStorageRequestTarget(targetPath, objectPath); err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	absolutePath, err := resourceConfig.resolveObjectPath(objectPath)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	fileInfo, exists, err := lookupStorageFileInfo(absolutePath)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	if !exists {
		return pluginbridge.NewHostCallSuccessResponse(
			pluginbridge.MarshalHostServiceStorageStatResponse(&pluginbridge.HostServiceStorageStatResponse{Found: false}),
		)
	}

	return pluginbridge.NewHostCallSuccessResponse(
		pluginbridge.MarshalHostServiceStorageStatResponse(&pluginbridge.HostServiceStorageStatResponse{
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
		return nil, gerror.Wrap(absErr, "storage resource 根目录解析失败")
	}

	return &storageResourceConfig{
		rootDir:    filepath.Clean(absoluteRootDir),
		visibility: pluginbridge.HostServiceStorageVisibilityPrivate,
	}, nil
}

func (resourceConfig *storageResourceConfig) validateWritePolicy(bodySize int64) error {
	if resourceConfig == nil {
		return gerror.New("storage resource config is nil")
	}
	if bodySize < 0 {
		return gerror.New("storage body size is invalid")
	}
	return nil
}

func (resourceConfig *storageResourceConfig) resolveObjectPath(objectPath string) (string, error) {
	normalizedObjectPath, err := normalizeStorageObjectPath(objectPath)
	if err != nil {
		return "", err
	}

	fullPath := filepath.Clean(filepath.Join(resourceConfig.rootDir, filepath.FromSlash(normalizedObjectPath)))
	rootPath := filepath.Clean(resourceConfig.rootDir)
	if fullPath != rootPath && !strings.HasPrefix(fullPath, rootPath+string(filepath.Separator)) {
		return "", gerror.Newf("storage object path 越界: %s", objectPath)
	}
	return fullPath, nil
}

func normalizeStorageObjectPath(rawPath string) (string, error) {
	return pluginfs.NormalizeRelativePath(rawPath)
}

func normalizeStorageListPrefix(rawPrefix string) (string, error) {
	trimmed := strings.TrimSpace(rawPrefix)
	if trimmed == "" {
		return "", gerror.New("storage list prefix is required")
	}
	return pluginfs.NormalizeRelativePath(trimmed)
}

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

func matchAuthorizedStoragePath(specs []*pluginbridge.HostServiceSpec, targetPath string) string {
	normalizedTarget, err := normalizeStorageAuthorizedPath(targetPath)
	if err != nil {
		return ""
	}
	// Storage authorization supports both exact object paths and directory
	// prefixes ending with `/`, so both the approval snapshot and request path
	// must be normalized before matching.
	for _, spec := range specs {
		if spec == nil || spec.Service != pluginbridge.HostServiceStorage {
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

func listStorageObjects(
	resourceConfig *storageResourceConfig,
	prefix string,
	limit int,
) ([]*pluginbridge.HostServiceStorageObject, error) {
	if resourceConfig == nil {
		return []*pluginbridge.HostServiceStorageObject{}, nil
	}
	if !gfile.Exists(resourceConfig.rootDir) {
		return []*pluginbridge.HostServiceStorageObject{}, nil
	}

	files, err := gfile.ScanDirFile(resourceConfig.rootDir, "*", true)
	if err != nil {
		return nil, err
	}
	sort.Strings(files)

	objects := make([]*pluginbridge.HostServiceStorageObject, 0, len(files))
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

func lookupStorageFileInfo(absolutePath string) (os.FileInfo, bool, error) {
	fileInfo, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if fileInfo.IsDir() {
		return nil, false, gerror.Newf("storage object path 指向目录: %s", absolutePath)
	}
	return fileInfo, true, nil
}

func buildStorageObjectSnapshot(
	objectPath string,
	fileInfo os.FileInfo,
	contentType string,
	visibility string,
) *pluginbridge.HostServiceStorageObject {
	if fileInfo == nil {
		return &pluginbridge.HostServiceStorageObject{
			Path:        objectPath,
			ContentType: contentType,
			Visibility:  visibility,
		}
	}
	return &pluginbridge.HostServiceStorageObject{
		Path:        objectPath,
		Size:        fileInfo.Size(),
		ContentType: contentType,
		UpdatedAt:   fileInfo.ModTime().UTC().Format(time.RFC3339Nano),
		Visibility:  visibility,
	}
}

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
