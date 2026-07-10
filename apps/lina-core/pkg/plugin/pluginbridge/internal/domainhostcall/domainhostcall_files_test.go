// This file verifies guest-side file host-service clients encode dynamic file
// write requests through the shared JSON host-call transport.

package domainhostcall

import (
	"bytes"
	"encoding/json"
	"testing"

	"lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// filesHostCallRecorder records guest files host-service invocations.
type filesHostCallRecorder struct {
	service string
	method  string
	upload  filesUploadRequest
	storage filesCreateFromStorageRequest
}

// invoke decodes file JSON requests and writes deterministic projections.
func (r *filesHostCallRecorder) invoke(service string, method string, request []byte, out any) error {
	r.service = service
	r.method = method
	envelope, err := protocol.UnmarshalHostServiceJSONRequest(request)
	if err != nil {
		return err
	}
	switch method {
	case protocol.HostServiceMethodFilesUpload:
		if err = json.Unmarshal(envelope.Value, &r.upload); err != nil {
			return err
		}
	case protocol.HostServiceMethodFilesCreateFromStorage:
		if err = json.Unmarshal(envelope.Value, &r.storage); err != nil {
			return err
		}
	}
	if projection, ok := out.(*filecap.FileInfo); ok {
		*projection = filecap.FileInfo{ID: "file-1", Name: "created.txt", BusinessScene: "plugin"}
	}
	return nil
}

// TestFilesUploadCallsPublishedHostService verifies Upload uses files.upload.
func TestFilesUploadCallsPublishedHostService(t *testing.T) {
	recorder := &filesHostCallRecorder{}
	service := Files(recorder.invoke)

	projection, err := service.Upload(t.Context(), filecap.UploadInput{
		Filename:      "created.txt",
		BusinessScene: "plugin",
		Reader:        bytes.NewReader([]byte("content")),
		SizeBytes:     int64(len("content")),
	})
	if err != nil {
		t.Fatalf("upload through host service: %v", err)
	}
	if projection == nil || projection.ID != "file-1" {
		t.Fatalf("unexpected projection: %#v", projection)
	}
	if recorder.service != protocol.HostServiceFiles || recorder.method != protocol.HostServiceMethodFilesUpload {
		t.Fatalf("unexpected host service call: %s.%s", recorder.service, recorder.method)
	}
	if recorder.upload.Filename != "created.txt" || string(recorder.upload.Body) != "content" || recorder.upload.SizeBytes != int64(len("content")) {
		t.Fatalf("unexpected upload request: %#v", recorder.upload)
	}
}

// TestFilesCreateFromStorageCallsPublishedHostService verifies CreateFromStorage uses files.create_from_storage.
func TestFilesCreateFromStorageCallsPublishedHostService(t *testing.T) {
	recorder := &filesHostCallRecorder{}
	service := Files(recorder.invoke)

	projection, err := service.CreateFromStorage(t.Context(), filecap.CreateFromStorageInput{
		StoragePath:   "exports/source.txt",
		Filename:      "source.txt",
		BusinessScene: "export",
		SizeBytes:     12,
	})
	if err != nil {
		t.Fatalf("create from storage through host service: %v", err)
	}
	if projection == nil || projection.ID != "file-1" {
		t.Fatalf("unexpected projection: %#v", projection)
	}
	if recorder.service != protocol.HostServiceFiles || recorder.method != protocol.HostServiceMethodFilesCreateFromStorage {
		t.Fatalf("unexpected host service call: %s.%s", recorder.service, recorder.method)
	}
	if recorder.storage.StoragePath != "exports/source.txt" || recorder.storage.Filename != "source.txt" || recorder.storage.SizeBytes != 12 {
		t.Fatalf("unexpected create-from-storage request: %#v", recorder.storage)
	}
}
