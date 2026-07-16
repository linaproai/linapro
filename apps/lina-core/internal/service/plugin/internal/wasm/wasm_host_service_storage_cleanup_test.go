// This file tests governed storage cleanup wiring for dynamic-plugin host services.

package wasm

import (
	"context"
	"strings"
	"testing"

	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// cleanupStorageService records cleanup delete calls for lifecycle cleanup tests.
type cleanupStorageService struct {
	objects      []*storagecap.Object
	deletedPaths []string
	listLimit    int
}

// Put is unused by cleanup tests.
func (*cleanupStorageService) Put(context.Context, storagecap.PutInput) (*storagecap.PutOutput, error) {
	return nil, nil
}

// Get is unused by cleanup tests.
func (*cleanupStorageService) Get(context.Context, storagecap.GetInput) (*storagecap.GetOutput, error) {
	return nil, nil
}

// Delete records one deleted logical path.
func (s *cleanupStorageService) Delete(_ context.Context, in storagecap.DeleteInput) error {
	s.deletedPaths = append(s.deletedPaths, in.Path)
	remaining := s.objects[:0]
	for _, object := range s.objects {
		if object == nil || object.Path == in.Path {
			continue
		}
		remaining = append(remaining, object)
	}
	s.objects = remaining
	return nil
}

// DeleteMany records deleted logical paths.
func (s *cleanupStorageService) DeleteMany(ctx context.Context, in storagecap.DeleteManyInput) error {
	for _, path := range in.Paths {
		if err := s.Delete(ctx, storagecap.DeleteInput{Path: path}); err != nil {
			return err
		}
	}
	return nil
}

// List returns the configured cleanup objects.
func (s *cleanupStorageService) List(_ context.Context, in storagecap.ListInput) (*storagecap.ListOutput, error) {
	limit := in.Limit
	if s.listLimit > 0 && s.listLimit < limit {
		limit = s.listLimit
	}
	if limit <= 0 || limit > len(s.objects) {
		limit = len(s.objects)
	}
	return &storagecap.ListOutput{Objects: append([]*storagecap.Object(nil), s.objects[:limit]...), Limit: limit}, nil
}

// ListCursor returns cleanup objects with the same bounded behavior as List.
func (s *cleanupStorageService) ListCursor(ctx context.Context, in storagecap.ListCursorInput) (*storagecap.ListCursorOutput, error) {
	output, err := s.List(ctx, storagecap.ListInput{Prefix: in.Prefix, Limit: in.Limit})
	if err != nil {
		return nil, err
	}
	return &storagecap.ListCursorOutput{Objects: output.Objects, Limit: output.Limit}, nil
}

// Stat is unused by cleanup tests.
func (*cleanupStorageService) Stat(context.Context, storagecap.StatInput) (*storagecap.StatOutput, error) {
	return nil, nil
}

// BatchStat is unused by cleanup tests.
func (*cleanupStorageService) BatchStat(context.Context, storagecap.BatchStatInput) (*storagecap.BatchStatOutput, error) {
	return &storagecap.BatchStatOutput{}, nil
}

// ProviderStatuses is unused by cleanup tests.
func (*cleanupStorageService) ProviderStatuses(context.Context) ([]*storagecap.ProviderStatus, error) {
	return nil, nil
}

// CreateDirectPut is unused by cleanup tests.
func (*cleanupStorageService) CreateDirectPut(_ context.Context, in storagecap.DirectPutInput) (*storagecap.DirectPutOutput, error) {
	return &storagecap.DirectPutOutput{
		Access: &storagecap.DirectAccess{Mode: storagecap.DirectAccessModeProxy, Operation: storagecap.DirectAccessOpPut},
		Path:   in.Path,
	}, nil
}

// ConfirmDirectPut is unused by cleanup tests.
func (*cleanupStorageService) ConfirmDirectPut(context.Context, storagecap.ConfirmDirectPutInput) (*storagecap.ConfirmDirectPutOutput, error) {
	return nil, nil
}

// CreateDirectGet is unused by cleanup tests.
func (*cleanupStorageService) CreateDirectGet(_ context.Context, in storagecap.DirectGetInput) (*storagecap.DirectGetOutput, error) {
	return &storagecap.DirectGetOutput{
		Access: &storagecap.DirectAccess{Mode: storagecap.DirectAccessModeProxy, Operation: storagecap.DirectAccessOpGet},
		Path:   in.Path,
	}, nil
}

func (*cleanupStorageService) SupportsMultipart(context.Context) (bool, error) { return false, nil }
func (*cleanupStorageService) CreateMultipart(context.Context, storagecap.MultipartCreateInput) (*storagecap.MultipartCreateOutput, error) {
	return nil, storagecap.NewMultipartUnsupportedError()
}
func (*cleanupStorageService) UploadPart(context.Context, storagecap.MultipartPartInput) (*storagecap.MultipartPartOutput, error) {
	return nil, storagecap.NewMultipartUnsupportedError()
}
func (*cleanupStorageService) CompleteMultipart(context.Context, storagecap.MultipartCompleteInput) (*storagecap.MultipartCompleteOutput, error) {
	return nil, storagecap.NewMultipartUnsupportedError()
}
func (*cleanupStorageService) AbortMultipart(context.Context, storagecap.MultipartAbortInput) error {
	return storagecap.NewMultipartUnsupportedError()
}
func (*cleanupStorageService) CreateMultipartPartAccess(context.Context, storagecap.MultipartPartAccessInput) (*storagecap.MultipartPartAccessOutput, error) {
	return nil, storagecap.NewMultipartUnsupportedError()
}

// TestPurgeAuthorizedStoragePathsRequiresStorageService verifies lifecycle
// cleanup fails explicitly when storage capability wiring is missing.
func TestPurgeAuthorizedStoragePathsRequiresStorageService(t *testing.T) {
	err := PurgeAuthorizedStoragePaths(context.Background(), nil, nil)
	if err == nil {
		t.Fatal("expected missing storage service to fail cleanup")
	}
	if !strings.Contains(err.Error(), "storage cleanup service is not configured") {
		t.Fatalf("expected storage service error, got %v", err)
	}
}

// TestPurgeAuthorizedStoragePathsDeletesAuthorizedObjects verifies cleanup uses
// storagecap.Service rather than local filesystem internals.
func TestPurgeAuthorizedStoragePathsDeletesAuthorizedObjects(t *testing.T) {
	service := &cleanupStorageService{
		objects: []*storagecap.Object{
			{Path: "reports/a.json"},
			{Path: "reports/b.json"},
		},
	}
	err := PurgeAuthorizedStoragePaths(
		context.Background(),
		service,
		[]*protocol.HostServiceSpec{{
			Service: protocol.HostServiceStorage,
			Paths:   []string{"reports/"},
		}},
	)
	if err != nil {
		t.Fatalf("purge authorized paths failed: %v", err)
	}
	if strings.Join(service.deletedPaths, ",") != "reports/a.json,reports/b.json" {
		t.Fatalf("expected cleanup deletes through storage service, got %#v", service.deletedPaths)
	}
}

// TestPurgeAuthorizedStoragePathsDeletesMultiplePages verifies prefix cleanup
// keeps listing until the authorized prefix is empty.
func TestPurgeAuthorizedStoragePathsDeletesMultiplePages(t *testing.T) {
	service := &cleanupStorageService{
		objects: []*storagecap.Object{
			{Path: "reports/a.json"},
			{Path: "reports/b.json"},
			{Path: "reports/c.json"},
		},
		listLimit: 2,
	}
	err := PurgeAuthorizedStoragePaths(
		context.Background(),
		service,
		[]*protocol.HostServiceSpec{{
			Service: protocol.HostServiceStorage,
			Paths:   []string{"reports/"},
		}},
	)
	if err != nil {
		t.Fatalf("purge multiple pages failed: %v", err)
	}
	if strings.Join(service.deletedPaths, ",") != "reports/a.json,reports/b.json,reports/c.json" {
		t.Fatalf("expected cleanup to delete all pages, got %#v", service.deletedPaths)
	}
}
