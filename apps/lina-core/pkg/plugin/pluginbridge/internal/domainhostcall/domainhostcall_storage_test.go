// This file verifies storage host-service client helper behavior that remains
// internal to the domainhostcall implementation.

package domainhostcall

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

type storageHostCallRecord struct {
	service     string
	method      string
	resourceRef string
	body        []byte
	offset      int64
}

type storageHostCallRecorder struct {
	calls          []storageHostCallRecord
	nextChunkError bool
	uploadID       string
}

func (r *storageHostCallRecorder) invoke(service string, method string, resourceRef string, _ string, request []byte) ([]byte, error) {
	if r.uploadID == "" {
		r.uploadID = "upload-1"
	}
	record := storageHostCallRecord{service: service, method: method, resourceRef: resourceRef}
	switch method {
	case protocol.HostServiceMethodStoragePut:
		decoded, err := protocol.UnmarshalHostServiceStoragePutRequest(request)
		if err != nil {
			return nil, err
		}
		record.body = append([]byte(nil), decoded.Body...)
		r.calls = append(r.calls, record)
		return protocol.MarshalHostServiceStoragePutResponse(&protocol.HostServiceStoragePutResponse{
			Object: &protocol.HostServiceStorageObject{Path: decoded.Path, Size: int64(len(decoded.Body))},
		}), nil
	case protocol.HostServiceMethodStoragePutInit:
		decoded, err := protocol.UnmarshalHostServiceStoragePutInitRequest(request)
		if err != nil {
			return nil, err
		}
		record.body = []byte(decoded.Path)
		r.calls = append(r.calls, record)
		return protocol.MarshalHostServiceStoragePutInitResponse(&protocol.HostServiceStoragePutInitResponse{UploadID: r.uploadID}), nil
	case protocol.HostServiceMethodStoragePutChunk:
		decoded, err := protocol.UnmarshalHostServiceStoragePutChunkRequest(request)
		if err != nil {
			return nil, err
		}
		record.offset = decoded.Offset
		record.body = append([]byte(nil), decoded.Body...)
		r.calls = append(r.calls, record)
		if r.nextChunkError {
			return nil, errors.New("chunk failed")
		}
		return protocol.MarshalHostServiceStoragePutChunkResponse(&protocol.HostServiceStoragePutChunkResponse{
			NextOffset: decoded.Offset + int64(len(decoded.Body)),
		}), nil
	case protocol.HostServiceMethodStoragePutCommit:
		decoded, err := protocol.UnmarshalHostServiceStoragePutCommitRequest(request)
		if err != nil {
			return nil, err
		}
		record.offset = decoded.Size
		r.calls = append(r.calls, record)
		return protocol.MarshalHostServiceStoragePutCommitResponse(&protocol.HostServiceStoragePutCommitResponse{
			Object: &protocol.HostServiceStorageObject{Path: decoded.Path, Size: decoded.Size},
		}), nil
	case protocol.HostServiceMethodStoragePutAbort:
		decoded, err := protocol.UnmarshalHostServiceStoragePutAbortRequest(request)
		if err != nil {
			return nil, err
		}
		record.body = []byte(decoded.UploadID)
		r.calls = append(r.calls, record)
		return nil, nil
	default:
		return nil, errors.New("unexpected method: " + method)
	}
}

type unknownSizeReader struct{ *bytes.Reader }

type noProgressReader struct{}

func (noProgressReader) Read(_ []byte) (int, error) {
	return 0, nil
}

// TestStoragePutUsesDirectHostCallForSmallKnownBodies verifies known small
// bodies keep using the existing storage.put method.
func TestStoragePutUsesDirectHostCallForSmallKnownBodies(t *testing.T) {
	recorder := &storageHostCallRecorder{}
	service := Storage(recorder.invoke)

	output, err := service.Put(t.Context(), storagecap.PutInput{
		Path:        "reports/small.txt",
		Body:        bytes.NewReader([]byte("small")),
		Size:        int64(len("small")),
		ContentType: "text/plain",
		Overwrite:   true,
	})
	if err != nil {
		t.Fatalf("put failed: %v", err)
	}
	if output == nil || output.Object == nil || output.Object.Size != int64(len("small")) {
		t.Fatalf("output: got %#v", output)
	}
	if len(recorder.calls) != 1 || recorder.calls[0].method != protocol.HostServiceMethodStoragePut {
		t.Fatalf("expected one direct put call, got %#v", recorder.calls)
	}
	if string(recorder.calls[0].body) != "small" {
		t.Fatalf("direct body: got %q want small", recorder.calls[0].body)
	}
}

// TestStoragePutUsesChunkedHostCallsForUnknownSizeBodies verifies unknown-size
// readers are streamed through init/chunk/commit without a full read-ahead.
func TestStoragePutUsesChunkedHostCallsForUnknownSizeBodies(t *testing.T) {
	recorder := &storageHostCallRecorder{}
	service := Storage(recorder.invoke)
	body := &unknownSizeReader{Reader: bytes.NewReader([]byte("streamed-body"))}

	output, err := service.Put(t.Context(), storagecap.PutInput{
		Path:        "reports/stream.bin",
		Body:        body,
		Size:        -1,
		ContentType: "application/octet-stream",
		Overwrite:   true,
	})
	if err != nil {
		t.Fatalf("put failed: %v", err)
	}
	if output == nil || output.Object == nil || output.Object.Size != int64(len("streamed-body")) {
		t.Fatalf("output: got %#v", output)
	}
	methods := storageHostCallMethods(recorder.calls)
	want := []string{
		protocol.HostServiceMethodStoragePutInit,
		protocol.HostServiceMethodStoragePutChunk,
		protocol.HostServiceMethodStoragePutCommit,
	}
	if !equalStringSlices(methods, want) {
		t.Fatalf("methods: got %#v want %#v", methods, want)
	}
	if string(recorder.calls[1].body) != "streamed-body" {
		t.Fatalf("chunk body: got %q want streamed-body", recorder.calls[1].body)
	}
}

// TestStoragePutAbortsChunkedUploadOnChunkFailure verifies chunk failures trigger
// best-effort abort before returning the original error.
func TestStoragePutAbortsChunkedUploadOnChunkFailure(t *testing.T) {
	recorder := &storageHostCallRecorder{nextChunkError: true}
	service := Storage(recorder.invoke)

	_, err := service.Put(t.Context(), storagecap.PutInput{
		Path: "reports/fail.bin",
		Body: io.LimitReader(bytes.NewReader([]byte("fail")), int64(len("fail"))),
		Size: -1,
	})
	if err == nil {
		t.Fatal("expected chunk error")
	}
	methods := storageHostCallMethods(recorder.calls)
	want := []string{
		protocol.HostServiceMethodStoragePutInit,
		protocol.HostServiceMethodStoragePutChunk,
		protocol.HostServiceMethodStoragePutAbort,
	}
	if !equalStringSlices(methods, want) {
		t.Fatalf("methods: got %#v want %#v", methods, want)
	}
}

// TestStoragePutAbortsChunkedUploadOnNoProgressReader verifies stalled readers
// do not spin forever and still clean up the upload session.
func TestStoragePutAbortsChunkedUploadOnNoProgressReader(t *testing.T) {
	recorder := &storageHostCallRecorder{}
	service := Storage(recorder.invoke)

	_, err := service.Put(t.Context(), storagecap.PutInput{
		Path: "reports/stalled.bin",
		Body: noProgressReader{},
		Size: -1,
	})
	if !errors.Is(err, io.ErrNoProgress) {
		t.Fatalf("expected ErrNoProgress, got %v", err)
	}
	methods := storageHostCallMethods(recorder.calls)
	want := []string{
		protocol.HostServiceMethodStoragePutInit,
		protocol.HostServiceMethodStoragePutAbort,
	}
	if !equalStringSlices(methods, want) {
		t.Fatalf("methods: got %#v want %#v", methods, want)
	}
}

// TestStorageListEffectiveLimit verifies guest storage list responses expose
// the same bounded limit semantics as storagecap.Service implementations.
func TestStorageListEffectiveLimit(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   int
		want int
	}{
		{name: "default", in: 0, want: storagecap.DefaultListLimit},
		{name: "negative default", in: -1, want: storagecap.DefaultListLimit},
		{name: "bounded", in: 10, want: 10},
		{name: "max", in: storagecap.MaxListLimit + 1, want: storagecap.MaxListLimit},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := storageListEffectiveLimit(c.in); got != c.want {
				t.Fatalf("storageListEffectiveLimit(%d) = %d, want %d", c.in, got, c.want)
			}
		})
	}
}

func storageHostCallMethods(calls []storageHostCallRecord) []string {
	methods := make([]string, 0, len(calls))
	for _, call := range calls {
		methods = append(methods, call.method)
	}
	return methods
}

func equalStringSlices(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
