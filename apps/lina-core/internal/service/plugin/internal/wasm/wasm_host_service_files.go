// This file adapts file host-service calls to the shared file capability
// service.

package wasm

import (
	"context"
	"strings"

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
	capCtx := capabilityContextForHostCall(hcc, bridgehostservice.HostServiceFiles, method)
	switch method {
	case bridgehostservice.HostServiceMethodFilesBatchGet:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.BatchGet(ctx, capCtx, fileIDs(request.IDs))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodFilesSearch:
		var request filesSearchRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.Search(ctx, capCtx, filecap.SearchInput{
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
		err := service.EnsureVisible(ctx, capCtx, fileIDs(request.IDs))
		return domainCapabilityResult(true, err)
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

// filesSearchRequest carries governed file search parameters.
type filesSearchRequest struct {
	BusinessScene string `json:"businessScene,omitempty"`
	Keyword       string `json:"keyword,omitempty"`
	MimeType      string `json:"mimeType,omitempty"`
	PageNum       int    `json:"pageNum,omitempty"`
	PageSize      int    `json:"pageSize,omitempty"`
}
