// This file implements the governed storage host service dispatcher. Dynamic
// plugin transport authorization stays here; object storage semantics are
// delegated to the plugin-scoped storagecap.Service.

package wasm

import (
	"bytes"
	"context"
	"io"
	"path"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/storagecap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchStorageHostService routes storage host service methods to the
// plugin-scoped storage domain service after path authorization has passed.
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
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return domainServiceNotScoped("storage")
	}
	service := services.Storage()
	if service == nil {
		return domainServiceNotScoped("storage")
	}

	switch method {
	case bridgehostservice.HostServiceMethodStoragePut:
		return handleStoragePut(ctx, service, targetPath, payload)
	case bridgehostservice.HostServiceMethodStoragePutInit:
		return handleStoragePutInit(hcc, targetPath, payload)
	case bridgehostservice.HostServiceMethodStoragePutChunk:
		return handleStoragePutChunk(hcc, targetPath, payload)
	case bridgehostservice.HostServiceMethodStoragePutCommit:
		return handleStoragePutCommit(ctx, hcc, service, targetPath, payload)
	case bridgehostservice.HostServiceMethodStoragePutAbort:
		return handleStoragePutAbort(hcc, targetPath, payload)
	case bridgehostservice.HostServiceMethodStorageGet:
		return handleStorageGet(ctx, service, targetPath, payload)
	case bridgehostservice.HostServiceMethodStorageDelete:
		return handleStorageDelete(ctx, service, targetPath, payload)
	case bridgehostservice.HostServiceMethodStorageList:
		return handleStorageList(ctx, service, targetPath, payload)
	case bridgehostservice.HostServiceMethodStorageStat:
		return handleStorageStat(ctx, service, targetPath, payload)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported storage host service method: "+method,
		)
	}
}

// handleStoragePut writes one governed storage object through storagecap.
func handleStoragePut(
	ctx context.Context,
	service storagecap.Service,
	targetPath string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceStoragePutRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	objectPath, err := normalizeStorageObjectPath(request.Path)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if err = validateStorageRequestTarget(targetPath, objectPath); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	output, err := service.Put(ctx, storagecap.PutInput{
		Path:        objectPath,
		Body:        bytes.NewReader(request.Body),
		Size:        int64(len(request.Body)),
		ContentType: request.ContentType,
		Overwrite:   request.Overwrite,
	})
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	return bridgehostcall.NewHostCallSuccessResponse(
		bridgehostservice.MarshalHostServiceStoragePutResponse(&bridgehostservice.HostServiceStoragePutResponse{
			Object: storageObjectResponse(outputObject(output)),
		}),
	)
}

// handleStorageGet reads one governed storage object through storagecap.
func handleStorageGet(
	ctx context.Context,
	service storagecap.Service,
	targetPath string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceStorageGetRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	objectPath, err := normalizeStorageObjectPath(request.Path)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if err = validateStorageRequestTarget(targetPath, objectPath); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	output, err := service.Get(ctx, storagecap.GetInput{Path: objectPath})
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if output == nil || !output.Found {
		return bridgehostcall.NewHostCallSuccessResponse(
			bridgehostservice.MarshalHostServiceStorageGetResponse(&bridgehostservice.HostServiceStorageGetResponse{Found: false}),
		)
	}
	body, err := readAndCloseStorageBody(output.Body)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	return bridgehostcall.NewHostCallSuccessResponse(
		bridgehostservice.MarshalHostServiceStorageGetResponse(&bridgehostservice.HostServiceStorageGetResponse{
			Found:  true,
			Object: storageObjectResponse(output.Object),
			Body:   body,
		}),
	)
}

// handleStorageDelete deletes one governed storage object through storagecap.
func handleStorageDelete(
	ctx context.Context,
	service storagecap.Service,
	targetPath string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceStorageDeleteRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	objectPath, err := normalizeStorageObjectPath(request.Path)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if err = validateStorageRequestTarget(targetPath, objectPath); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if err = service.Delete(ctx, storagecap.DeleteInput{Path: objectPath}); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	return bridgehostcall.NewHostCallEmptySuccessResponse()
}

// handleStorageList lists governed storage objects through storagecap.
func handleStorageList(
	ctx context.Context,
	service storagecap.Service,
	targetPath string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceStorageListRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	prefix, err := normalizeStorageListPrefix(request.Prefix)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if err = validateStorageRequestTarget(targetPath, prefix); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	output, err := service.List(ctx, storagecap.ListInput{Prefix: prefix, Limit: int(request.Limit)})
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	objects := make([]*bridgehostservice.HostServiceStorageObject, 0)
	if output != nil {
		objects = make([]*bridgehostservice.HostServiceStorageObject, 0, len(output.Objects))
		for _, object := range output.Objects {
			objects = append(objects, storageObjectResponse(object))
		}
	}
	return bridgehostcall.NewHostCallSuccessResponse(
		bridgehostservice.MarshalHostServiceStorageListResponse(&bridgehostservice.HostServiceStorageListResponse{Objects: objects}),
	)
}

// handleStorageStat reads governed storage metadata through storagecap.
func handleStorageStat(
	ctx context.Context,
	service storagecap.Service,
	targetPath string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceStorageStatRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	objectPath, err := normalizeStorageObjectPath(request.Path)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if err = validateStorageRequestTarget(targetPath, objectPath); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	output, err := service.Stat(ctx, storagecap.StatInput{Path: objectPath})
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if output == nil || !output.Found {
		return bridgehostcall.NewHostCallSuccessResponse(
			bridgehostservice.MarshalHostServiceStorageStatResponse(&bridgehostservice.HostServiceStorageStatResponse{Found: false}),
		)
	}
	return bridgehostcall.NewHostCallSuccessResponse(
		bridgehostservice.MarshalHostServiceStorageStatResponse(&bridgehostservice.HostServiceStorageStatResponse{
			Found:  true,
			Object: storageObjectResponse(output.Object),
		}),
	)
}

// normalizeStorageObjectPath canonicalizes one logical object path.
func normalizeStorageObjectPath(rawPath string) (string, error) {
	normalized, err := normalizeStorageAuthorizedPath(rawPath)
	if err != nil {
		return "", err
	}
	if strings.HasSuffix(normalized, "/") {
		return "", gerror.New("storage object path cannot be a prefix")
	}
	return normalized, nil
}

// normalizeStorageListPrefix canonicalizes one required list prefix.
func normalizeStorageListPrefix(rawPrefix string) (string, error) {
	trimmed := strings.TrimSpace(rawPrefix)
	if trimmed == "" {
		return "", gerror.New("storage list prefix is required")
	}
	return normalizeStorageAuthorizedPath(trimmed)
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
	if strings.Contains(base, "://") || strings.HasPrefix(base, "/") {
		return "", gerror.New("invalid storage target path")
	}
	if len(base) >= 2 && ((base[0] >= 'A' && base[0] <= 'Z') || (base[0] >= 'a' && base[0] <= 'z')) && base[1] == ':' {
		return "", gerror.New("invalid storage target path")
	}
	normalized := path.Clean(base)
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") {
		return "", gerror.New("invalid storage target path")
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

// validateStorageRequestTarget ensures the guest request is covered by the authorized target.
func validateStorageRequestTarget(targetPath string, requestPath string) error {
	normalizedTarget, err := normalizeStorageAuthorizedPath(targetPath)
	if err != nil {
		return err
	}
	normalizedRequest, err := normalizeStorageAuthorizedPath(requestPath)
	if err != nil {
		return err
	}
	if strings.HasSuffix(normalizedTarget, "/") {
		base := strings.TrimSuffix(normalizedTarget, "/")
		if normalizedRequest == base || strings.HasPrefix(normalizedRequest, base+"/") {
			return nil
		}
		return gerror.Newf("storage request target mismatch: target=%s request=%s", normalizedTarget, normalizedRequest)
	}
	if normalizedTarget != normalizedRequest {
		return gerror.Newf("storage request target mismatch: target=%s request=%s", normalizedTarget, normalizedRequest)
	}
	return nil
}

// storageObjectResponse maps storagecap object metadata into the bridge payload.
func storageObjectResponse(object *storagecap.Object) *bridgehostservice.HostServiceStorageObject {
	if object == nil {
		return nil
	}
	response := &bridgehostservice.HostServiceStorageObject{
		Path:        object.Path,
		Size:        object.Size,
		ContentType: object.ContentType,
		Visibility:  object.Visibility,
	}
	if object.UpdatedAt != nil {
		response.UpdatedAt = object.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}
	return response
}

// outputObject returns the object metadata from a Put response.
func outputObject(output *storagecap.PutOutput) *storagecap.Object {
	if output == nil {
		return nil
	}
	return output.Object
}

// readAndCloseStorageBody drains and closes a storage body.
func readAndCloseStorageBody(body io.ReadCloser) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	content, readErr := io.ReadAll(body)
	closeErr := body.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return content, nil
}
