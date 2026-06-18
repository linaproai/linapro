// This file implements the guest-side storage host-service client using the
// injected raw host-service invoker and existing protocol codecs.

package domainhostcall

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

const (
	// storageSinglePutMaxBytes bounds guest memory used by the direct put path.
	storageSinglePutMaxBytes int64 = 1 * 1024 * 1024
	// storagePutChunkBytes bounds each chunk sent through one host-service call.
	storagePutChunkBytes = 1 * 1024 * 1024
)

// storageService adapts the storage host service to storagecap.Service.
type storageService struct{ baseService }

// Storage creates the storage domain guest client.
func Storage(invoker HostServiceInvoker) storagecap.Service {
	return &storageService{baseService: newBaseServiceWithHostService(nil, invoker)}
}

// Put writes one governed storage object under the given logical path.
func (s *storageService) Put(_ context.Context, in storagecap.PutInput) (*storagecap.PutOutput, error) {
	body := in.Body
	if body == nil {
		body = bytes.NewReader(nil)
	}
	if in.Size >= 0 && in.Size <= storageSinglePutMaxBytes {
		content, tooLarge, err := readStorageDirectBody(body, storageSinglePutMaxBytes)
		if err != nil {
			return nil, err
		}
		if !tooLarge {
			return s.putDirect(in, content)
		}
		body = io.MultiReader(bytes.NewReader(content), body)
	}
	return s.putChunked(in, body)
}

func (s *storageService) putDirect(in storagecap.PutInput, body []byte) (*storagecap.PutOutput, error) {
	request := &protocol.HostServiceStoragePutRequest{
		Path:        in.Path,
		Body:        body,
		ContentType: in.ContentType,
		Overwrite:   in.Overwrite,
	}
	payload, err := s.callHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStoragePut,
		in.Path,
		"",
		protocol.MarshalHostServiceStoragePutRequest(request),
	)
	if err != nil {
		return nil, err
	}
	response, err := protocol.UnmarshalHostServiceStoragePutResponse(payload)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil
	}
	return &storagecap.PutOutput{Object: storageObjectFromWire(response.Object)}, nil
}

func (s *storageService) putChunked(in storagecap.PutInput, body io.Reader) (output *storagecap.PutOutput, err error) {
	initPayload, err := s.callHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStoragePutInit,
		in.Path,
		"",
		protocol.MarshalHostServiceStoragePutInitRequest(&protocol.HostServiceStoragePutInitRequest{
			Path:        in.Path,
			ContentType: in.ContentType,
			Overwrite:   in.Overwrite,
		}),
	)
	if err != nil {
		return nil, err
	}
	initResponse, err := protocol.UnmarshalHostServiceStoragePutInitResponse(initPayload)
	if err != nil {
		return nil, err
	}
	if initResponse == nil || initResponse.UploadID == "" {
		return nil, gerror.New("storage upload init response missing upload id")
	}

	uploadID := initResponse.UploadID
	abortUpload := true
	defer func() {
		if !abortUpload {
			return
		}
		abortErr := s.abortStorageUpload(in.Path, uploadID)
		if err == nil && abortErr != nil {
			err = abortErr
		}
	}()

	buffer := make([]byte, storagePutChunkBytes)
	var offset int64
	for {
		n, readErr := body.Read(buffer)
		if n > 0 {
			nextOffset, chunkErr := s.putChunk(in.Path, uploadID, offset, buffer[:n])
			if chunkErr != nil {
				return nil, chunkErr
			}
			expectedOffset := offset + int64(n)
			if nextOffset != expectedOffset {
				return nil, gerror.Newf("storage upload next offset mismatch: got %d want %d", nextOffset, expectedOffset)
			}
			offset = nextOffset
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, readErr
		}
		if n == 0 {
			return nil, io.ErrNoProgress
		}
	}

	commitPayload, err := s.callHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStoragePutCommit,
		in.Path,
		"",
		protocol.MarshalHostServiceStoragePutCommitRequest(&protocol.HostServiceStoragePutCommitRequest{
			Path:     in.Path,
			UploadID: uploadID,
			Size:     offset,
		}),
	)
	if err != nil {
		return nil, err
	}
	commitResponse, err := protocol.UnmarshalHostServiceStoragePutCommitResponse(commitPayload)
	if err != nil {
		return nil, err
	}
	abortUpload = false
	if commitResponse == nil {
		return nil, nil
	}
	return &storagecap.PutOutput{Object: storageObjectFromWire(commitResponse.Object)}, nil
}

func (s *storageService) putChunk(path string, uploadID string, offset int64, body []byte) (int64, error) {
	payload, err := s.callHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStoragePutChunk,
		path,
		"",
		protocol.MarshalHostServiceStoragePutChunkRequest(&protocol.HostServiceStoragePutChunkRequest{
			Path:     path,
			UploadID: uploadID,
			Offset:   offset,
			Body:     body,
		}),
	)
	if err != nil {
		return 0, err
	}
	response, err := protocol.UnmarshalHostServiceStoragePutChunkResponse(payload)
	if err != nil {
		return 0, err
	}
	if response == nil {
		return 0, gerror.New("storage upload chunk response is empty")
	}
	return response.NextOffset, nil
}

func (s *storageService) abortStorageUpload(path string, uploadID string) error {
	_, err := s.callHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStoragePutAbort,
		path,
		"",
		protocol.MarshalHostServiceStoragePutAbortRequest(&protocol.HostServiceStoragePutAbortRequest{
			Path:     path,
			UploadID: uploadID,
		}),
	)
	return err
}

func readStorageDirectBody(body io.Reader, maxBytes int64) ([]byte, bool, error) {
	limited := io.LimitReader(body, maxBytes+1)
	content, err := io.ReadAll(limited)
	if err != nil {
		return nil, false, err
	}
	if int64(len(content)) > maxBytes {
		return content, true, nil
	}
	return content, false, nil
}

// Get reads one governed storage object under the given logical path.
func (s *storageService) Get(_ context.Context, in storagecap.GetInput) (*storagecap.GetOutput, error) {
	request := &protocol.HostServiceStorageGetRequest{Path: in.Path}
	payload, err := s.callHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStorageGet,
		in.Path,
		"",
		protocol.MarshalHostServiceStorageGetRequest(request),
	)
	if err != nil {
		return nil, err
	}
	response, err := protocol.UnmarshalHostServiceStorageGetResponse(payload)
	if err != nil {
		return nil, err
	}
	if response == nil || !response.Found {
		return &storagecap.GetOutput{Found: false}, nil
	}
	return &storagecap.GetOutput{
		Object: storageObjectFromWire(response.Object),
		Body:   io.NopCloser(bytes.NewReader(response.Body)),
		Found:  true,
	}, nil
}

// Delete removes one governed storage object under the given logical path.
func (s *storageService) Delete(_ context.Context, in storagecap.DeleteInput) error {
	request := &protocol.HostServiceStorageDeleteRequest{Path: in.Path}
	_, err := s.callHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStorageDelete,
		in.Path,
		"",
		protocol.MarshalHostServiceStorageDeleteRequest(request),
	)
	return err
}

// DeleteMany removes governed storage objects under explicit logical paths.
func (s *storageService) DeleteMany(_ context.Context, in storagecap.DeleteManyInput) error {
	return s.callHostServiceJSONRequest(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStorageDeleteBatch,
		storageBatchResourceRef(in.Paths),
		"",
		storageBatchPathsRequest{Paths: in.Paths},
		nil,
	)
}

// List lists governed storage objects under one logical path prefix.
func (s *storageService) List(_ context.Context, in storagecap.ListInput) (*storagecap.ListOutput, error) {
	request := &protocol.HostServiceStorageListRequest{
		Prefix: in.Prefix,
		Limit:  uint32(in.Limit),
	}
	payload, err := s.callHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStorageList,
		in.Prefix,
		"",
		protocol.MarshalHostServiceStorageListRequest(request),
	)
	if err != nil {
		return nil, err
	}
	response, err := protocol.UnmarshalHostServiceStorageListResponse(payload)
	if err != nil {
		return nil, err
	}
	output := &storagecap.ListOutput{
		Objects: []*storagecap.Object{},
		Limit:   storageListEffectiveLimit(in.Limit),
	}
	if response == nil {
		return output, nil
	}
	for _, object := range response.Objects {
		output.Objects = append(output.Objects, storageObjectFromWire(object))
	}
	return output, nil
}

// ListCursor lists governed storage objects under one logical path prefix with cursor pagination.
func (s *storageService) ListCursor(_ context.Context, in storagecap.ListCursorInput) (*storagecap.ListCursorOutput, error) {
	response := &storageListCursorResponse{}
	err := s.callHostServiceJSONRequest(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStorageListCursor,
		in.Prefix,
		"",
		storageListCursorRequest{
			Prefix: in.Prefix,
			Cursor: in.Cursor,
			Limit:  in.Limit,
		},
		response,
	)
	if err != nil {
		return nil, err
	}
	output := &storagecap.ListCursorOutput{
		Objects:    []*storagecap.Object{},
		NextCursor: response.NextCursor,
		Limit:      response.Limit,
	}
	if output.Limit <= 0 {
		output.Limit = storageListEffectiveLimit(in.Limit)
	}
	for _, object := range response.Objects {
		output.Objects = append(output.Objects, storageObjectFromWire(object))
	}
	return output, nil
}

// Stat reads metadata for one governed storage object under the given logical path.
func (s *storageService) Stat(_ context.Context, in storagecap.StatInput) (*storagecap.StatOutput, error) {
	request := &protocol.HostServiceStorageStatRequest{Path: in.Path}
	payload, err := s.callHostService(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStorageStat,
		in.Path,
		"",
		protocol.MarshalHostServiceStorageStatRequest(request),
	)
	if err != nil {
		return nil, err
	}
	response, err := protocol.UnmarshalHostServiceStorageStatResponse(payload)
	if err != nil {
		return nil, err
	}
	if response == nil || !response.Found {
		return &storagecap.StatOutput{Found: false}, nil
	}
	return &storagecap.StatOutput{Object: storageObjectFromWire(response.Object), Found: true}, nil
}

// BatchStat reads governed storage metadata for explicit logical paths.
func (s *storageService) BatchStat(_ context.Context, in storagecap.BatchStatInput) (*storagecap.BatchStatOutput, error) {
	response := &storageBatchStatResponse{}
	err := s.callHostServiceJSONRequest(
		protocol.HostServiceStorage,
		protocol.HostServiceMethodStorageStatBatch,
		storageBatchResourceRef(in.Paths),
		"",
		storageBatchPathsRequest{Paths: in.Paths},
		response,
	)
	if err != nil {
		return nil, err
	}
	output := &storagecap.BatchStatOutput{
		Objects:      []*storagecap.Object{},
		MissingPaths: append([]string(nil), response.MissingPaths...),
	}
	for _, object := range response.Objects {
		output.Objects = append(output.Objects, storageObjectFromWire(object))
	}
	return output, nil
}

// ProviderStatuses is not exposed through the dynamic storage host-service
// transport. Source plugins can call the host-side storagecap.Service directly.
func (s *storageService) ProviderStatuses(_ context.Context) ([]*storagecap.ProviderStatus, error) {
	return nil, errHostCallsUnavailable
}

func storageObjectFromWire(object *protocol.HostServiceStorageObject) *storagecap.Object {
	if object == nil {
		return nil
	}
	return &storagecap.Object{
		Path:        object.Path,
		Size:        object.Size,
		ContentType: object.ContentType,
		UpdatedAt:   parseWireTime(object.UpdatedAt),
		Visibility:  object.Visibility,
	}
}

func storageListEffectiveLimit(limit int) int {
	if limit <= 0 {
		return storagecap.DefaultListLimit
	}
	if limit > storagecap.MaxListLimit {
		return storagecap.MaxListLimit
	}
	return limit
}

func storageBatchResourceRef(paths []string) string {
	return batchResourceRef(paths)
}

func parseWireTime(value string) *time.Time {
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil
	}
	return &parsed
}

type storageBatchPathsRequest struct {
	Paths []string `json:"paths"`
}

type storageBatchStatResponse struct {
	Objects      []*protocol.HostServiceStorageObject `json:"objects"`
	MissingPaths []string                             `json:"missingPaths,omitempty"`
}

type storageListCursorRequest struct {
	Prefix string `json:"prefix"`
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

type storageListCursorResponse struct {
	Objects    []*protocol.HostServiceStorageObject `json:"objects"`
	NextCursor string                               `json:"nextCursor,omitempty"`
	Limit      int                                  `json:"limit,omitempty"`
}
