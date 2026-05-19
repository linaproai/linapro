// This file verifies data-scope enforcement for file management operations.

package file

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/orgcap"
	rolesvc "lina-core/internal/service/role"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/pkg/bizerr"
)

// TestFileDataScopeFiltersListDetailDownloadDeleteAndSuffixes verifies file
// metadata, binary access, deletion, and aggregation all respect uploader scope.
func TestFileDataScopeFiltersListDetailDownloadDeleteAndSuffixes(t *testing.T) {
	ctx := context.Background()
	currentUserID := insertFileScopeUser(t, ctx, "file-scope-current")
	otherUserID := insertFileScopeUser(t, ctx, "file-scope-other")
	roleID := insertFileScopeRole(t, ctx, "file-scope-self", 3)
	t.Cleanup(func() {
		cleanupFileScopeUsers(t, ctx, []int{currentUserID, otherUserID})
		cleanupFileScopeRoles(t, ctx, []int{roleID})
	})
	insertFileScopeUserRole(t, ctx, currentUserID, roleID)

	visibleID := insertFileScopeRecord(t, ctx, currentUserID, "visible", "txt")
	hiddenID := insertFileScopeRecord(t, ctx, otherUserID, "hidden", "pdf")
	t.Cleanup(func() { cleanupFileScopeRecords(t, ctx, []int64{visibleID, hiddenID}) })

	storage := &fileDataScopeStorage{content: "visible-content"}
	bizCtxSvc := fileScopeStaticBizCtx{ctx: &model.Context{UserId: currentUserID}}
	orgCapSvc := orgcap.New(nil)
	roleSvc := newFileDataScopeRoleService(bizCtxSvc, orgCapSvc)
	svc := &serviceImpl{
		storage:   storage,
		bizCtxSvc: bizCtxSvc,
		dictSvc:   nil,
		scopeSvc:  datascope.New(bizCtxSvc, roleSvc, orgCapSvc),
	}

	out, err := svc.List(ctx, &ListInput{PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list files: %v", err)
	}
	if out.Total != 1 || len(out.List) != 1 || out.List[0].SysFile.Id != visibleID {
		t.Fatalf("expected only visible file, got total=%d ids=%v", out.Total, fileScopeListIDs(out.List))
	}

	if _, err = svc.Info(ctx, hiddenID); !bizerr.Is(err, CodeFileDataScopeDenied) {
		t.Fatalf("expected hidden info to be denied, got %v", err)
	}
	if _, err = svc.OpenByID(ctx, hiddenID); !bizerr.Is(err, CodeFileDataScopeDenied) {
		t.Fatalf("expected hidden download to be denied, got %v", err)
	}
	if err = svc.Delete(ctx, fmt.Sprintf("%d", hiddenID)); !bizerr.Is(err, CodeFileDataScopeDenied) {
		t.Fatalf("expected hidden delete to be denied, got %v", err)
	}
	if count := countFileScopeRecord(t, ctx, hiddenID); count != 1 {
		t.Fatalf("expected hidden file to remain after denied delete, count=%d", count)
	}

	suffixes, err := svc.Suffixes(ctx)
	if err != nil {
		t.Fatalf("list suffixes: %v", err)
	}
	if len(suffixes) != 1 || suffixes[0].Value != "txt" {
		t.Fatalf("expected only visible suffix txt, got %#v", suffixes)
	}
}

// TestTenantUploadPersistsCurrentTenantAndListsInTenantScope verifies uploaded
// file metadata uses the current tenant so tenant list filters can find it.
func TestTenantUploadPersistsCurrentTenantAndListsInTenantScope(t *testing.T) {
	ctx := context.Background()
	tenantID := 77
	tenantCtx := datascope.WithTenantForTest(ctx, tenantID)
	currentUserID := insertFileScopeUser(t, ctx, "file-tenant-upload-current")
	t.Cleanup(func() { cleanupFileScopeUsers(t, ctx, []int{currentUserID}) })

	bizCtxSvc := fileScopeStaticBizCtx{ctx: &model.Context{TenantId: tenantID, UserId: currentUserID}}
	storage := &fileTenantUploadStorage{}
	scopeSvc := datascope.New(bizCtxSvc, fileTenantUploadAccessProvider{userID: currentUserID}, nil)
	svc := &serviceImpl{
		configSvc: hostconfig.New(),
		storage:   storage,
		bizCtxSvc: bizCtxSvc,
		scopeSvc:  scopeSvc,
	}

	uploadFile := buildFileScopeUploadFile(t, "tenant-upload.png", "tenant image content")
	uploaded, err := svc.Upload(tenantCtx, &UploadInput{File: uploadFile, Scene: "other"})
	if err != nil {
		t.Fatalf("upload tenant file: %v", err)
	}
	t.Cleanup(func() { cleanupFileScopeRecords(t, ctx, []int64{uploaded.Id}) })

	var record *entity.SysFile
	if err = dao.SysFile.Ctx(ctx).Where(do.SysFile{Id: uploaded.Id}).Scan(&record); err != nil {
		t.Fatalf("query uploaded file metadata: %v", err)
	}
	if record == nil {
		t.Fatal("expected uploaded file metadata to exist")
	}
	if record.TenantId != tenantID {
		t.Fatalf("expected uploaded file tenant_id=%d, got %d", tenantID, record.TenantId)
	}

	out, err := svc.List(tenantCtx, &ListInput{PageNum: 1, PageSize: 20, Original: "tenant-upload.png"})
	if err != nil {
		t.Fatalf("list tenant uploaded files: %v", err)
	}
	if out.Total != 1 || len(out.List) != 1 || out.List[0].SysFile.Id != uploaded.Id {
		t.Fatalf("expected uploaded file %d in tenant list, got total=%d ids=%v", uploaded.Id, out.Total, fileScopeListIDs(out.List))
	}
}

// newFileDataScopeRoleService builds the explicit role dependency used by
// file data-scope tests.
func newFileDataScopeRoleService(bizCtxSvc bizctx.Service, orgCapSvc orgcap.Service) rolesvc.Service {
	configSvc := hostconfig.New()
	i18nSvc := i18nsvc.New(bizCtxSvc, configSvc, cachecoord.Default(nil))
	tenantSvc := tenantcapsvc.New(nil, bizCtxSvc)
	roleSvc := rolesvc.New(nil, bizCtxSvc, configSvc, i18nSvc, nil, orgCapSvc, tenantSvc)
	roleSvc.SetDataScopeService(datascope.New(bizCtxSvc, roleSvc, orgCapSvc))
	return roleSvc
}

// fileDataScopeStorage is a deterministic storage fake for data-scope tests.
type fileDataScopeStorage struct {
	content string
}

// Put is unused by file data-scope tests.
func (s *fileDataScopeStorage) Put(context.Context, string, io.Reader) (string, error) {
	return "", nil
}

// Get returns deterministic content.
func (s *fileDataScopeStorage) Get(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(s.content)), nil
}

// Delete performs no storage mutation.
func (s *fileDataScopeStorage) Delete(context.Context, string) error { return nil }

// Url returns one deterministic URL path.
func (s *fileDataScopeStorage) Url(_ context.Context, path string) string {
	return "/api/v1/uploads/" + path
}

// fileTenantUploadStorage records file upload writes without touching disk.
type fileTenantUploadStorage struct{}

// Put drains the uploaded stream and returns a deterministic storage path.
func (s *fileTenantUploadStorage) Put(_ context.Context, filename string, data io.Reader) (string, error) {
	if _, err := io.Copy(io.Discard, data); err != nil {
		return "", err
	}
	return "tenant-upload/" + filename, nil
}

// Get is unused by the upload tenant test.
func (s *fileTenantUploadStorage) Get(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}

// Delete is unused by the upload tenant test.
func (s *fileTenantUploadStorage) Delete(context.Context, string) error { return nil }

// Url returns the upload URL shape for the stored path.
func (s *fileTenantUploadStorage) Url(_ context.Context, path string) string {
	return "/api/v1/uploads/" + path
}

// fileTenantUploadAccessProvider grants tenant-wide visibility for the upload
// regression test without relying on persisted role fixtures.
type fileTenantUploadAccessProvider struct {
	userID int
}

// GetUserDataScopeSnapshot returns a tenant-wide data-scope snapshot.
func (p fileTenantUploadAccessProvider) GetUserDataScopeSnapshot(context.Context, int) (*datascope.AccessSnapshot, error) {
	return &datascope.AccessSnapshot{UserID: p.userID, Scope: datascope.ScopeTenant}, nil
}

// fileScopeStaticBizCtx returns a fixed business context.
type fileScopeStaticBizCtx struct {
	ctx *model.Context
}

// Init is unused by file data-scope tests.
func (s fileScopeStaticBizCtx) Init(_ *ghttp.Request, _ *model.Context) {}

// Get returns the configured business context.
func (s fileScopeStaticBizCtx) Get(context.Context) *model.Context { return s.ctx }

// SetLocale is unused by file data-scope tests.
func (s fileScopeStaticBizCtx) SetLocale(context.Context, string) {}

// SetUser is unused by file data-scope tests.
func (s fileScopeStaticBizCtx) SetUser(context.Context, string, int, string, int) {}

// SetTenant is unused by file data-scope tests.
func (s fileScopeStaticBizCtx) SetTenant(context.Context, int) {}

// SetImpersonation is unused by file data-scope tests.
func (s fileScopeStaticBizCtx) SetImpersonation(context.Context, int, int, bool, bool) {}

// SetUserAccess is unused by file data-scope tests.
func (s fileScopeStaticBizCtx) SetUserAccess(context.Context, int, bool, int) {}

// insertFileScopeUser inserts one temporary user.
func insertFileScopeUser(t *testing.T, ctx context.Context, prefix string) int {
	t.Helper()

	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: uniqueFileScopeName(prefix),
		Password: "hashed",
		Nickname: prefix,
		Status:   1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert file-scope user: %v", err)
	}
	return int(id)
}

// insertFileScopeRole inserts one temporary role.
func insertFileScopeRole(t *testing.T, ctx context.Context, prefix string, scope int) int {
	t.Helper()

	id, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		Name:      uniqueFileScopeName(prefix),
		Key:       uniqueFileScopeName(prefix + "-key"),
		Sort:      99,
		DataScope: scope,
		Status:    1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert file-scope role: %v", err)
	}
	return int(id)
}

// insertFileScopeUserRole binds a user to a role.
func insertFileScopeUserRole(t *testing.T, ctx context.Context, userID int, roleID int) {
	t.Helper()

	if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{UserId: userID, RoleId: roleID}).Insert(); err != nil {
		t.Fatalf("insert file-scope user role: %v", err)
	}
}

// insertFileScopeRecord inserts one temporary file metadata row.
func insertFileScopeRecord(t *testing.T, ctx context.Context, ownerID int, name string, suffix string) int64 {
	t.Helper()

	id, err := dao.SysFile.Ctx(ctx).Data(do.SysFile{
		Name:      name + "." + suffix,
		Original:  name + "." + suffix,
		Suffix:    suffix,
		Scene:     "other",
		Size:      1,
		Hash:      uniqueFileScopeName("hash"),
		Url:       "/api/v1/uploads/" + name + "." + suffix,
		Path:      name + "." + suffix,
		Engine:    EngineLocal,
		CreatedBy: ownerID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert file-scope record: %v", err)
	}
	return id
}

// uniqueFileScopeName returns one collision-resistant identifier.
func uniqueFileScopeName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// cleanupFileScopeUsers removes temporary users.
func cleanupFileScopeUsers(t *testing.T, ctx context.Context, ids []int) {
	t.Helper()
	if len(ids) == 0 {
		return
	}
	if _, err := dao.SysUserRole.Ctx(ctx).WhereIn(dao.SysUserRole.Columns().UserId, ids).Delete(); err != nil {
		t.Fatalf("cleanup file-scope user roles: %v", err)
	}
	if _, err := dao.SysUser.Ctx(ctx).Unscoped().WhereIn(dao.SysUser.Columns().Id, ids).Delete(); err != nil {
		t.Fatalf("cleanup file-scope users: %v", err)
	}
}

// cleanupFileScopeRoles removes temporary roles.
func cleanupFileScopeRoles(t *testing.T, ctx context.Context, ids []int) {
	t.Helper()
	if len(ids) == 0 {
		return
	}
	if _, err := dao.SysRole.Ctx(ctx).Unscoped().WhereIn(dao.SysRole.Columns().Id, ids).Delete(); err != nil {
		t.Fatalf("cleanup file-scope roles: %v", err)
	}
}

// cleanupFileScopeRecords removes temporary file records.
func cleanupFileScopeRecords(t *testing.T, ctx context.Context, ids []int64) {
	t.Helper()
	if len(ids) == 0 {
		return
	}
	if _, err := dao.SysFile.Ctx(ctx).Unscoped().WhereIn(dao.SysFile.Columns().Id, ids).Delete(); err != nil {
		t.Fatalf("cleanup file-scope records: %v", err)
	}
}

// countFileScopeRecord counts visible file records by ID.
func countFileScopeRecord(t *testing.T, ctx context.Context, id int64) int {
	t.Helper()
	count, err := dao.SysFile.Ctx(ctx).Where(do.SysFile{Id: id}).Count()
	if err != nil {
		t.Fatalf("count file-scope record: %v", err)
	}
	return count
}

// buildFileScopeUploadFile creates one real multipart file header so Upload can
// exercise FileHeader.Open just like an HTTP request would.
func buildFileScopeUploadFile(t *testing.T, filename string, content string) *ghttp.UploadFile {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err = part.Write([]byte(content)); err != nil {
		t.Fatalf("write multipart file content: %v", err)
	}
	if err = writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	reader := multipart.NewReader(bytes.NewReader(body.Bytes()), writer.Boundary())
	form, err := reader.ReadForm(int64(body.Len()))
	if err != nil {
		t.Fatalf("read multipart form: %v", err)
	}
	t.Cleanup(func() {
		if cleanupErr := form.RemoveAll(); cleanupErr != nil {
			t.Fatalf("cleanup multipart form: %v", cleanupErr)
		}
	})

	files := form.File["file"]
	if len(files) != 1 {
		t.Fatalf("expected one multipart file, got %d", len(files))
	}
	return &ghttp.UploadFile{FileHeader: files[0]}
}

// fileScopeListIDs returns file IDs from list items.
func fileScopeListIDs(items []*ListOutputItem) []int64 {
	ids := make([]int64, 0, len(items))
	for _, item := range items {
		if item == nil || item.SysFile == nil {
			continue
		}
		ids = append(ids, item.SysFile.Id)
	}
	return ids
}
