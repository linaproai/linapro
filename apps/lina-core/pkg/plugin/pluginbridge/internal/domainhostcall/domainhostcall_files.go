// This file implements the guest-side files capability hostcall client.

package domainhostcall

import (
	"bytes"
	"context"
	"io"

	"lina-core/pkg/bizerr"
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
func (s filesService) BatchGet(_ context.Context, ids []filecap.FileID) (*capmodel.BatchResult[*filecap.FileInfo, filecap.FileID], error) {
	out := &capmodel.BatchResult[*filecap.FileInfo, filecap.FileID]{Items: map[filecap.FileID]*filecap.FileInfo{}}
	err := s.callJSONRequest(protocol.HostServiceFiles, protocol.HostServiceMethodFilesBatchGet, idsRequest{IDs: fileIDsToStrings(ids)}, out)
	return out, err
}

// Get returns one visible file projection through the registered batch-read method.
func (s filesService) Get(ctx context.Context, id filecap.FileID) (*filecap.FileInfo, error) {
	result, err := s.BatchGet(ctx, []filecap.FileID{id})
	if err != nil || result == nil {
		return nil, err
	}
	if item, ok := result.Items[id]; ok {
		return item, nil
	}
	return nil, nil
}

// Detail is not published as a dynamic files host-service method.
func (s filesService) Detail(context.Context, filecap.FileID) (*filecap.DetailInfo, error) {
	return nil, unsupportedDynamicMethodError("files.detail")
}

// List returns one bounded page of visible file projections.
func (s filesService) List(_ context.Context, input filecap.ListInput) (*capmodel.PageResult[*filecap.FileInfo], error) {
	out := &capmodel.PageResult[*filecap.FileInfo]{Items: []*filecap.FileInfo{}}
	err := s.callJSONRequest(protocol.HostServiceFiles, protocol.HostServiceMethodFilesList, filesListRequest{
		BusinessScene: input.BusinessScene,
		Keyword:       input.Keyword,
		MimeType:      input.MimeType,
		PageNum:       input.Page.PageNum,
		PageSize:      input.Page.PageSize,
	}, out)
	return out, err
}

// ListScenes is not published as a dynamic files host-service method.
func (s filesService) ListScenes(context.Context) ([]*filecap.Option, error) {
	return nil, unsupportedDynamicMethodError("files.scenes.list")
}

// ListSuffixes is not published as a dynamic files host-service method.
func (s filesService) ListSuffixes(context.Context) ([]*filecap.Option, error) {
	return nil, unsupportedDynamicMethodError("files.suffixes.list")
}

// Open is not published as a dynamic files host-service method.
func (s filesService) Open(context.Context, filecap.FileID) (*filecap.FileContent, error) {
	return nil, unsupportedDynamicMethodError("files.open")
}

// EnsureVisible rejects when any requested file is absent or invisible.
func (s filesService) EnsureVisible(_ context.Context, ids []filecap.FileID) error {
	return s.callJSONRequest(protocol.HostServiceFiles, protocol.HostServiceMethodFilesEnsureVisible, idsRequest{IDs: fileIDsToStrings(ids)}, nil)
}

// Upload creates one host file record through a bounded direct JSON payload.
func (s filesService) Upload(_ context.Context, input filecap.UploadInput) (*filecap.FileInfo, error) {
	body, err := readDirectFileUploadBody(input.Reader)
	if err != nil {
		return nil, err
	}
	sizeBytes := input.SizeBytes
	if sizeBytes <= 0 {
		sizeBytes = int64(len(body))
	}
	out := &filecap.FileInfo{}
	err = s.callJSONRequest(protocol.HostServiceFiles, protocol.HostServiceMethodFilesUpload, filesUploadRequest{
		Filename:      input.Filename,
		BusinessScene: input.BusinessScene,
		Body:          body,
		SizeBytes:     sizeBytes,
	}, out)
	return out, err
}

// CreateFromStorage creates one host file record from plugin-private storage.
func (s filesService) CreateFromStorage(_ context.Context, input filecap.CreateFromStorageInput) (*filecap.FileInfo, error) {
	out := &filecap.FileInfo{}
	err := s.callJSONRequest(protocol.HostServiceFiles, protocol.HostServiceMethodFilesCreateFromStorage, filesCreateFromStorageRequest{
		StoragePath:   input.StoragePath,
		Filename:      input.Filename,
		BusinessScene: input.BusinessScene,
		SizeBytes:     input.SizeBytes,
	}, out)
	return out, err
}

// UpdateMetadata is not published as a dynamic files host-service method.
func (s filesService) UpdateMetadata(context.Context, filecap.UpdateMetadataInput) error {
	return unsupportedDynamicMethodError("files.metadata.update")
}

// Delete is not published as a dynamic files host-service method.
func (s filesService) Delete(context.Context, filecap.FileID) error {
	return unsupportedDynamicMethodError("files.delete")
}

// DeleteMany is not published as a dynamic files host-service method.
func (s filesService) DeleteMany(context.Context, []filecap.FileID) error {
	return unsupportedDynamicMethodError("files.delete_many")
}

// fileIDsToStrings converts file IDs to transport strings.
func fileIDsToStrings(ids []filecap.FileID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

// filesListRequest carries governed file list parameters.
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

// readDirectFileUploadBody reads one bounded direct-upload request body.
func readDirectFileUploadBody(reader io.Reader) ([]byte, error) {
	if reader == nil {
		reader = bytes.NewReader(nil)
	}
	content, err := io.ReadAll(io.LimitReader(reader, filecap.MaxDirectUploadBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(content)) > filecap.MaxDirectUploadBytes {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", filecap.MaxDirectUploadBytes))
	}
	return content, nil
}

var _ filecap.Service = (*filesService)(nil)
