// This file adapts file host-service calls to the shared file capability
// service.

package wasm

import (
	"bytes"
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/filecap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchFilesHostService routes file-domain host-service calls.
func dispatchFilesHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := filesServiceForHostCall(hcc)
	if service == nil {
		return domainServiceNotScoped("files")
	}
	switch method {
	case bridgehostservice.HostServiceMethodFilesBatchGet:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.BatchGet(ctx, fileIDs(request.IDs))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodFilesList:
		var request filesListRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.List(ctx, filecap.ListInput{
			BusinessScene: strings.TrimSpace(request.BusinessScene),
			Keyword:       strings.TrimSpace(request.Keyword),
			MimeType:      strings.TrimSpace(request.MimeType),
			Page: capmodel.PageRequest{
				PageNum:  request.PageNum,
				PageSize: request.PageSize,
			},
		})
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodFilesEnsureVisible:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.EnsureVisible(ctx, fileIDs(request.IDs))
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodFilesUpload:
		var request filesUploadRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		if int64(len(request.Body)) > filecap.MaxDirectUploadBytes {
			return domainCapabilityResult(nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", filecap.MaxDirectUploadBytes)))
		}
		sizeBytes := request.SizeBytes
		if sizeBytes <= 0 {
			sizeBytes = int64(len(request.Body))
		}
		result, err := service.Upload(ctx, filecap.UploadInput{
			Filename:      strings.TrimSpace(request.Filename),
			BusinessScene: strings.TrimSpace(request.BusinessScene),
			Reader:        bytes.NewReader(request.Body),
			SizeBytes:     sizeBytes,
		})
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodFilesCreateFromStorage:
		var request filesCreateFromStorageRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		storagePath := strings.TrimSpace(request.StoragePath)
		if !hcc.hasCapability(bridgehostservice.CapabilityStorage) ||
			!hcc.hasHostServiceAccess(bridgehostservice.HostServiceStorage, bridgehostservice.HostServiceMethodStorageGet, storagePath, "") {
			return bridgehostcall.NewHostCallErrorResponse(
				bridgehostcall.HostCallStatusCapabilityDenied,
				"files.create_from_storage requires storage.get authorization for the source path",
			)
		}
		result, err := service.CreateFromStorage(ctx, filecap.CreateFromStorageInput{
			StoragePath:   storagePath,
			Filename:      strings.TrimSpace(request.Filename),
			BusinessScene: strings.TrimSpace(request.BusinessScene),
			SizeBytes:     request.SizeBytes,
		})
		return domainCapabilityResult(result, err)
	default:
		return domainMethodNotFound("files", method)
	}
}

// filesServiceForHostCall resolves the file service for one host call.
func filesServiceForHostCall(hcc *hostCallContext) filecap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.Files()
}

// fileIDs converts transport string identifiers into typed file IDs.
func fileIDs(ids []string) []filecap.FileID {
	out := make([]filecap.FileID, 0, len(ids))
	for _, id := range ids {
		if value := strings.TrimSpace(id); value != "" {
			out = append(out, filecap.FileID(value))
		}
	}
	return out
}

// filesListRequest carries governed file search parameters.
type filesListRequest struct {
	BusinessScene string `json:"businessScene,omitempty"`
	Keyword       string `json:"keyword,omitempty"`
	MimeType      string `json:"mimeType,omitempty"`
	PageNum       int    `json:"pageNum,omitempty"`
	PageSize      int    `json:"pageSize,omitempty"`
}

// filesUploadRequest carries direct dynamic file upload content.
type filesUploadRequest struct {
	Filename      string `json:"filename,omitempty"`
	BusinessScene string `json:"businessScene,omitempty"`
	Body          []byte `json:"body,omitempty"`
	SizeBytes     int64  `json:"sizeBytes,omitempty"`
}

// filesCreateFromStorageRequest carries one plugin storage promotion request.
type filesCreateFromStorageRequest struct {
	StoragePath   string `json:"storagePath,omitempty"`
	Filename      string `json:"filename,omitempty"`
	BusinessScene string `json:"businessScene,omitempty"`
	SizeBytes     int64  `json:"sizeBytes,omitempty"`
}
