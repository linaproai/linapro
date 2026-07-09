// This file tests storage host service codec round trips.

package protocol

import "testing"

// TestHostServiceStoragePutRequestRoundTrip verifies storage put requests keep
// path, content, and overwrite semantics through the codec.
func TestHostServiceStoragePutRequestRoundTrip(t *testing.T) {
	original := &HostServiceStoragePutRequest{
		Path:        "reports/demo.json",
		Body:        []byte(`{"ok":true}`),
		ContentType: "application/json",
		Overwrite:   true,
	}

	data := MarshalHostServiceStoragePutRequest(original)
	decoded, err := UnmarshalHostServiceStoragePutRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Path != original.Path {
		t.Fatalf("path: got %q want %q", decoded.Path, original.Path)
	}
	if string(decoded.Body) != string(original.Body) {
		t.Fatalf("body: got %q want %q", decoded.Body, original.Body)
	}
	if decoded.ContentType != original.ContentType {
		t.Fatalf("contentType: got %q want %q", decoded.ContentType, original.ContentType)
	}
	if !decoded.Overwrite {
		t.Fatal("overwrite: expected true")
	}
}

// TestHostServiceStorageChunkedPutRoundTrip verifies chunked storage upload
// DTOs keep session, offset, body, and commit metadata through the codec.
func TestHostServiceStorageChunkedPutRoundTrip(t *testing.T) {
	initReq := &HostServiceStoragePutInitRequest{
		Path:        "reports/demo.bin",
		ContentType: "application/octet-stream",
		Overwrite:   true,
	}
	initData := MarshalHostServiceStoragePutInitRequest(initReq)
	initDecoded, err := UnmarshalHostServiceStoragePutInitRequest(initData)
	if err != nil {
		t.Fatalf("init request unmarshal failed: %v", err)
	}
	if initDecoded.Path != initReq.Path || initDecoded.ContentType != initReq.ContentType || !initDecoded.Overwrite {
		t.Fatalf("init request: got %#v want %#v", initDecoded, initReq)
	}

	initResp := &HostServiceStoragePutInitResponse{UploadID: "upload-1"}
	initRespData := MarshalHostServiceStoragePutInitResponse(initResp)
	initRespDecoded, err := UnmarshalHostServiceStoragePutInitResponse(initRespData)
	if err != nil {
		t.Fatalf("init response unmarshal failed: %v", err)
	}
	if initRespDecoded.UploadID != initResp.UploadID {
		t.Fatalf("init response uploadId: got %q want %q", initRespDecoded.UploadID, initResp.UploadID)
	}

	chunkReq := &HostServiceStoragePutChunkRequest{
		Path:     "reports/demo.bin",
		UploadID: "upload-1",
		Offset:   5,
		Body:     []byte("chunk"),
	}
	chunkData := MarshalHostServiceStoragePutChunkRequest(chunkReq)
	chunkDecoded, err := UnmarshalHostServiceStoragePutChunkRequest(chunkData)
	if err != nil {
		t.Fatalf("chunk request unmarshal failed: %v", err)
	}
	if chunkDecoded.Path != chunkReq.Path || chunkDecoded.UploadID != chunkReq.UploadID || chunkDecoded.Offset != chunkReq.Offset || string(chunkDecoded.Body) != string(chunkReq.Body) {
		t.Fatalf("chunk request: got %#v want %#v", chunkDecoded, chunkReq)
	}

	chunkResp := &HostServiceStoragePutChunkResponse{NextOffset: 10}
	chunkRespData := MarshalHostServiceStoragePutChunkResponse(chunkResp)
	chunkRespDecoded, err := UnmarshalHostServiceStoragePutChunkResponse(chunkRespData)
	if err != nil {
		t.Fatalf("chunk response unmarshal failed: %v", err)
	}
	if chunkRespDecoded.NextOffset != chunkResp.NextOffset {
		t.Fatalf("chunk response nextOffset: got %d want %d", chunkRespDecoded.NextOffset, chunkResp.NextOffset)
	}

	commitReq := &HostServiceStoragePutCommitRequest{
		Path:     "reports/demo.bin",
		UploadID: "upload-1",
		Size:     10,
	}
	commitData := MarshalHostServiceStoragePutCommitRequest(commitReq)
	commitDecoded, err := UnmarshalHostServiceStoragePutCommitRequest(commitData)
	if err != nil {
		t.Fatalf("commit request unmarshal failed: %v", err)
	}
	if commitDecoded.Path != commitReq.Path || commitDecoded.UploadID != commitReq.UploadID || commitDecoded.Size != commitReq.Size {
		t.Fatalf("commit request: got %#v want %#v", commitDecoded, commitReq)
	}

	commitResp := &HostServiceStoragePutCommitResponse{
		Object: &HostServiceStorageObject{
			Path:        "reports/demo.bin",
			Size:        10,
			ContentType: "application/octet-stream",
			Visibility:  HostServiceStorageVisibilityPrivate,
		},
	}
	commitRespData := MarshalHostServiceStoragePutCommitResponse(commitResp)
	commitRespDecoded, err := UnmarshalHostServiceStoragePutCommitResponse(commitRespData)
	if err != nil {
		t.Fatalf("commit response unmarshal failed: %v", err)
	}
	if commitRespDecoded.Object == nil || commitRespDecoded.Object.Path != commitResp.Object.Path || commitRespDecoded.Object.Size != commitResp.Object.Size {
		t.Fatalf("commit response object: got %#v want %#v", commitRespDecoded.Object, commitResp.Object)
	}

	abortReq := &HostServiceStoragePutAbortRequest{Path: "reports/demo.bin", UploadID: "upload-1"}
	abortData := MarshalHostServiceStoragePutAbortRequest(abortReq)
	abortDecoded, err := UnmarshalHostServiceStoragePutAbortRequest(abortData)
	if err != nil {
		t.Fatalf("abort request unmarshal failed: %v", err)
	}
	if abortDecoded.Path != abortReq.Path || abortDecoded.UploadID != abortReq.UploadID {
		t.Fatalf("abort request: got %#v want %#v", abortDecoded, abortReq)
	}
}

// TestHostServiceStorageGetResponseRoundTrip verifies storage get responses
// keep found flags, object metadata, and body payloads through the codec.
func TestHostServiceStorageGetResponseRoundTrip(t *testing.T) {
	original := &HostServiceStorageGetResponse{
		Found: true,
		Object: &HostServiceStorageObject{
			Path:        "reports/demo.json",
			Size:        12,
			ContentType: "application/json",
			UpdatedAt:   "2026-04-14T10:00:00Z",
			Visibility:  HostServiceStorageVisibilityPrivate,
		},
		Body: []byte(`{"ok":true}`),
	}

	data := MarshalHostServiceStorageGetResponse(original)
	decoded, err := UnmarshalHostServiceStorageGetResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !decoded.Found {
		t.Fatal("found: expected true")
	}
	if decoded.Object == nil || decoded.Object.Path != original.Object.Path {
		t.Fatalf("object: got %#v want %#v", decoded.Object, original.Object)
	}
	if string(decoded.Body) != string(original.Body) {
		t.Fatalf("body: got %q want %q", decoded.Body, original.Body)
	}
}

// TestHostServiceStorageListResponseRoundTrip verifies storage list responses
// preserve ordered object metadata snapshots.
func TestHostServiceStorageListResponseRoundTrip(t *testing.T) {
	original := &HostServiceStorageListResponse{
		Objects: []*HostServiceStorageObject{
			{
				Path:        "reports/a.json",
				Size:        10,
				ContentType: "application/json",
				UpdatedAt:   "2026-04-14T10:00:00Z",
				Visibility:  HostServiceStorageVisibilityPrivate,
			},
			{
				Path:        "reports/b.txt",
				Size:        8,
				ContentType: "text/plain",
				UpdatedAt:   "2026-04-14T10:00:01Z",
				Visibility:  HostServiceStorageVisibilityPublic,
			},
		},
	}

	data := MarshalHostServiceStorageListResponse(original)
	decoded, err := UnmarshalHostServiceStorageListResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(decoded.Objects) != 2 {
		t.Fatalf("objects: got %d want 2", len(decoded.Objects))
	}
	if decoded.Objects[1].Visibility != HostServiceStorageVisibilityPublic {
		t.Fatalf("visibility: got %q want %q", decoded.Objects[1].Visibility, HostServiceStorageVisibilityPublic)
	}
}

// TestHostServiceStorageStatResponseRoundTrip verifies storage stat responses
// preserve found flags and object metadata snapshots.
func TestHostServiceStorageStatResponseRoundTrip(t *testing.T) {
	original := &HostServiceStorageStatResponse{
		Found: true,
		Object: &HostServiceStorageObject{
			Path:        "reports/demo.json",
			Size:        12,
			ContentType: "application/json",
			UpdatedAt:   "2026-04-14T10:00:00Z",
			Visibility:  HostServiceStorageVisibilityPrivate,
		},
	}

	data := MarshalHostServiceStorageStatResponse(original)
	decoded, err := UnmarshalHostServiceStorageStatResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !decoded.Found {
		t.Fatal("found: expected true")
	}
	if decoded.Object == nil || decoded.Object.Size != original.Object.Size {
		t.Fatalf("object: got %#v want %#v", decoded.Object, original.Object)
	}
}
