// This file implements the guest-side files capability hostcall client.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// filesService adapts file projection reads to host services.
type filesService struct{ baseService }

// Files creates the file-domain guest client.
func Files(invoker Invoker) filecap.Service {
	return filesService{baseService: newBaseService(invoker)}
}

// BatchGet returns visible file projections and opaque missing IDs.
func (s filesService) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []filecap.FileID) (*capmodel.BatchResult[*filecap.FileProjection, filecap.FileID], error) {
	out := &capmodel.BatchResult[*filecap.FileProjection, filecap.FileID]{Items: map[filecap.FileID]*filecap.FileProjection{}}
	err := s.callJSONRequest(protocol.HostServiceFiles, protocol.HostServiceMethodFilesBatchGet, idsRequest{IDs: fileIDsToStrings(ids)}, out)
	return out, err
}

// Search returns one bounded page of visible file projections.
func (s filesService) Search(_ context.Context, _ capmodel.CapabilityContext, input filecap.SearchInput) (*capmodel.PageResult[*filecap.FileProjection], error) {
	out := &capmodel.PageResult[*filecap.FileProjection]{Items: []*filecap.FileProjection{}}
	err := s.callJSONRequest(protocol.HostServiceFiles, protocol.HostServiceMethodFilesSearch, filesSearchRequest{
		BusinessScene: input.BusinessScene,
		Keyword:       input.Keyword,
		MimeType:      input.MimeType,
		PageNum:       input.Page.PageNum,
		PageSize:      input.Page.PageSize,
	}, out)
	return out, err
}

// EnsureVisible rejects when any requested file is absent or invisible.
func (s filesService) EnsureVisible(_ context.Context, _ capmodel.CapabilityContext, ids []filecap.FileID) error {
	return s.callJSONRequest(protocol.HostServiceFiles, protocol.HostServiceMethodFilesEnsureVisible, idsRequest{IDs: fileIDsToStrings(ids)}, nil)
}

// fileIDsToStrings converts file IDs to transport strings.
func fileIDsToStrings(ids []filecap.FileID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
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

var _ filecap.Service = (*filesService)(nil)
