// This file tests storage host service codec round trips.

package pluginbridge

import "testing"

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
