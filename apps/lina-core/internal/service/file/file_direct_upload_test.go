// This file tests file-center direct upload session lifecycle and proxy fallback.

package file

import (
	"context"
	"strings"
	"testing"
	"time"

	"lina-core/internal/service/config"
	storagesvc "lina-core/internal/service/storage"
	"lina-core/pkg/plugin/capability/storagecap"
)

type directUploadTestConfig struct {
	config.Service
	maxSize      int64
	directUrlTTL time.Duration
}

func (c *directUploadTestConfig) GetUploadMaxSize(context.Context) (int64, error) {
	if c.maxSize <= 0 {
		return 20, nil
	}
	return c.maxSize, nil
}

func (c *directUploadTestConfig) GetUploadDirectUrlTTL(context.Context) (time.Duration, error) {
	if c.directUrlTTL > 0 {
		return c.directUrlTTL, nil
	}
	return time.Hour, nil
}

// directUploadTestStorage implements only the storage methods used by direct-upload tests.
type directUploadTestStorage struct {
	access    *storagecap.DirectAccess
	accessErr error
	statFound bool
	statSize  int64
	putKey    string
	provider  string
}

func (s *directUploadTestStorage) Put(context.Context, storagesvc.PutInput) (*storagesvc.PutOutput, error) {
	return nil, nil
}
func (s *directUploadTestStorage) Get(context.Context, storagesvc.GetInput) (*storagesvc.GetOutput, error) {
	return &storagesvc.GetOutput{Found: false}, nil
}
func (s *directUploadTestStorage) Delete(context.Context, storagesvc.DeleteInput) error { return nil }
func (s *directUploadTestStorage) DeleteMany(context.Context, storagesvc.DeleteManyInput) error {
	return nil
}
func (s *directUploadTestStorage) List(context.Context, storagesvc.ListInput) (*storagesvc.ListOutput, error) {
	return &storagesvc.ListOutput{}, nil
}
func (s *directUploadTestStorage) ListCursor(context.Context, storagesvc.ListCursorInput) (*storagesvc.ListCursorOutput, error) {
	return &storagesvc.ListCursorOutput{}, nil
}
func (s *directUploadTestStorage) BatchStat(context.Context, storagesvc.BatchStatInput) (*storagesvc.BatchStatOutput, error) {
	return &storagesvc.BatchStatOutput{}, nil
}

func (s *directUploadTestStorage) CreateDirectAccess(_ context.Context, in storagesvc.DirectAccessInput) (*storagesvc.DirectAccessOutput, error) {
	if s.accessErr != nil {
		return nil, s.accessErr
	}
	s.putKey = in.Key
	access := s.access
	if access == nil {
		access = &storagecap.DirectAccess{Mode: storagecap.DirectAccessModeProxy, Operation: in.Operation}
	}
	provider := s.provider
	if provider == "" {
		provider = storagecap.LocalProviderID
	}
	return &storagesvc.DirectAccessOutput{
		Access:      access,
		ProviderID:  provider,
		ProviderKey: "files/" + in.Key,
	}, nil
}

func (s *directUploadTestStorage) Stat(context.Context, storagesvc.StatInput) (*storagesvc.StatOutput, error) {
	if !s.statFound {
		return &storagesvc.StatOutput{Found: false}, nil
	}
	return &storagesvc.StatOutput{
		Found: true,
		Object: &storagesvc.Object{
			Key:  s.putKey,
			Size: s.statSize,
		},
	}, nil
}

func TestDirectUploadInitReturnsProxyWithoutSession(t *testing.T) {
	t.Parallel()
	storage := &directUploadTestStorage{
		access: &storagecap.DirectAccess{Mode: storagecap.DirectAccessModeProxy, Operation: storagecap.DirectAccessOpPut},
	}
	svc := &serviceImpl{
		configSvc:      &directUploadTestConfig{maxSize: 20},
		storage:        storage,
		directSessions: newDirectUploadSessionStore(),
	}
	out, err := svc.DirectUploadInit(context.Background(), &DirectUploadInitInput{
		Scene:    "other",
		FileName: "demo.txt",
		Size:     12,
	})
	if err != nil {
		t.Fatalf("DirectUploadInit: %v", err)
	}
	if out.UploadSessionID != "" {
		t.Fatalf("proxy mode must not create session, got %q", out.UploadSessionID)
	}
	if !storagecap.IsProxyDirectAccess(out.Access) {
		t.Fatalf("expected proxy access, got %+v", out.Access)
	}
}

func TestDirectUploadSessionExpiryOnComplete(t *testing.T) {
	t.Parallel()
	store := newDirectUploadSessionStore()
	store.put(&directUploadSession{
		ID:          "sess-expired",
		TenantID:    0,
		StoragePath: "0/2026/07/demo.txt",
		Size:        4,
		ExpiresAt:   time.Now().UTC().Add(-time.Minute),
	})
	svc := &serviceImpl{
		configSvc:      &directUploadTestConfig{maxSize: 20},
		storage:        &directUploadTestStorage{},
		directSessions: store,
	}
	_, err := svc.DirectUploadComplete(context.Background(), &DirectUploadCompleteInput{
		UploadSessionID: "sess-expired",
	})
	if err == nil {
		t.Fatal("expected expired session error")
	}
	if !strings.Contains(err.Error(), "expired") && !strings.Contains(err.Error(), "FILE_DIRECT_SESSION_EXPIRED") {
		// bizerr message may be English template text
		if code := err.Error(); !strings.Contains(strings.ToUpper(code), "EXPIRED") && !strings.Contains(code, "SESSION") {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestDirectUploadCompleteRequiresObject(t *testing.T) {
	t.Parallel()
	store := newDirectUploadSessionStore()
	store.put(&directUploadSession{
		ID:           "sess-ok",
		TenantID:     0,
		OriginalName: "demo.txt",
		Suffix:       "txt",
		Scene:        "other",
		StoragePath:  "0/2026/07/demo.txt",
		Size:         4,
		ProviderID:   "linapro-storage-s3",
		ExpiresAt:    time.Now().UTC().Add(10 * time.Minute),
	})
	svc := &serviceImpl{
		configSvc: &directUploadTestConfig{maxSize: 20},
		storage: &directUploadTestStorage{
			statFound: false,
		},
		directSessions: store,
	}
	_, err := svc.DirectUploadComplete(context.Background(), &DirectUploadCompleteInput{
		UploadSessionID: "sess-ok",
	})
	if err == nil {
		t.Fatal("expected complete validation failure when object missing")
	}
}
