// This file verifies storage-backed access for uploaded file URLs.

package file

import (
	"context"
	"io"
	"strings"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
)

// fakeAccessStorage records object reads and returns deterministic content.
type fakeAccessStorage struct {
	content string // content returned by Get
	getPath string // getPath records the path requested by OpenByPath
}

// Put is unused by access tests and returns an empty object path.
func (s *fakeAccessStorage) Put(_ context.Context, _ string, _ io.Reader) (string, error) {
	return "", nil
}

// Get returns the configured test content and records the requested object path.
func (s *fakeAccessStorage) Get(_ context.Context, path string) (io.ReadCloser, error) {
	s.getPath = path
	return io.NopCloser(strings.NewReader(s.content)), nil
}

// Delete is unused by access tests and performs no storage mutation.
func (s *fakeAccessStorage) Delete(_ context.Context, _ string) error {
	return nil
}

// Url returns the upload URL shape for a storage object path.
func (s *fakeAccessStorage) Url(_ context.Context, path string) string {
	return "/api/v1/uploads/" + path
}

// TestOpenByPathRejectsParentTraversalWithoutStorageAccess verifies unsafe URL
// path segments are rejected before metadata lookup or storage access.
func TestOpenByPathRejectsParentTraversalWithoutStorageAccess(t *testing.T) {
	storage := &fakeAccessStorage{}
	svc := &serviceImpl{storage: storage}

	_, err := svc.OpenByPath(context.Background(), "../secret.txt")
	if err == nil {
		t.Fatal("expected parent traversal path to be rejected")
	}
	messageErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured file error, got %T %v", err, err)
	}
	if !messageErr.Matches(CodeFileNotFound) {
		t.Fatalf("expected %s, got %s", CodeFileNotFound.RuntimeCode(), messageErr.RuntimeCode())
	}
	if storage.getPath != "" {
		t.Fatalf("expected storage not to be read, got path %q", storage.getPath)
	}
}

// TestOpenByPathReadsThroughStorageBackendWithoutUserContext verifies public
// upload URL access resolves metadata before reading through the storage backend.
func TestOpenByPathReadsThroughStorageBackendWithoutUserContext(t *testing.T) {
	ctx := context.Background()
	storagePath := "e2e/storage-backed-access.txt"
	storage := &fakeAccessStorage{content: "stored-content"}
	adminUserID := mustQueryFileAccessAdminUserID(t, ctx)
	svc := &serviceImpl{
		storage:   storage,
		bizCtxSvc: fileAccessStaticBizCtx{},
	}

	result, err := dao.SysFile.Ctx(ctx).Data(do.SysFile{
		Name:      "storage-backed-access.txt",
		Original:  "storage-backed-access.txt",
		Suffix:    "txt",
		Scene:     "other",
		Size:      int64(len(storage.content)),
		Hash:      "storage-backed-access-hash",
		Url:       "/api/v1/uploads/" + storagePath,
		Path:      storagePath,
		Engine:    EngineLocal,
		CreatedBy: adminUserID,
	}).Insert()
	if err != nil {
		t.Fatalf("insert file metadata: %v", err)
	}
	fileID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read inserted file metadata id: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := dao.SysFile.Ctx(ctx).Unscoped().Where(do.SysFile{Id: fileID}).Delete(); cleanupErr != nil {
			t.Fatalf("cleanup file metadata: %v", cleanupErr)
		}
	})

	output, err := svc.OpenByPath(ctx, "/"+storagePath)
	if err != nil {
		t.Fatalf("open file by storage path: %v", err)
	}
	defer func() {
		if closeErr := output.Reader.Close(); closeErr != nil {
			t.Fatalf("close opened file stream: %v", closeErr)
		}
	}()

	body, err := io.ReadAll(output.Reader)
	if err != nil {
		t.Fatalf("read opened file stream: %v", err)
	}
	if string(body) != storage.content {
		t.Fatalf("expected storage content %q, got %q", storage.content, string(body))
	}
	if storage.getPath != storagePath {
		t.Fatalf("expected storage path %q, got %q", storagePath, storage.getPath)
	}
	if output.ContentType != "application/octet-stream" {
		t.Fatalf("expected default content type, got %q", output.ContentType)
	}
}

// fileAccessStaticBizCtx returns a fixed request business context for file tests.
type fileAccessStaticBizCtx struct {
	ctx *model.Context
}

// Init is unused by file service tests because they inject context directly.
func (s fileAccessStaticBizCtx) Init(_ *ghttp.Request, _ *model.Context) {}

// Get returns the configured business context.
func (s fileAccessStaticBizCtx) Get(context.Context) *model.Context { return s.ctx }

// SetLocale is unused by file service tests.
func (s fileAccessStaticBizCtx) SetLocale(context.Context, string) {}

// SetUser is unused by file service tests.
func (s fileAccessStaticBizCtx) SetUser(context.Context, string, int, string, int) {}

// SetUserAccess is unused by file service tests.
func (s fileAccessStaticBizCtx) SetUserAccess(context.Context, int, bool, int) {}

// mustQueryFileAccessAdminUserID resolves the built-in administrator user ID for data-scope tests.
func mustQueryFileAccessAdminUserID(t *testing.T, ctx context.Context) int {
	t.Helper()

	var admin *entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Username: "admin"}).
		Scan(&admin); err != nil {
		t.Fatalf("query built-in admin user: %v", err)
	}
	if admin == nil {
		t.Fatal("expected built-in admin user to exist")
	}
	return admin.Id
}
