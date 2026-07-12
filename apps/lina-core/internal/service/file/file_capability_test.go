// This file verifies plugin file capabilities create governed sys_file records
// through the file owner while keeping plugin object storage as a scoped source.

package file

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	storagesvc "lina-core/internal/service/storage"
	internalbizctx "lina-core/internal/service/bizctx"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// TestCreateFromReaderCreatesFileRecord verifies the host file owner can create
// sys_file metadata from a non-seekable stream.
func TestCreateFromReaderCreatesFileRecord(t *testing.T) {
	var (
		ctx       = context.Background()
		tenantID  = int(time.Now().UnixNano()%1000000 + 1000)
		tenantCtx = datascope.WithTenantScope(ctx, tenantID)
		userID    = insertFileScopeUser(t, ctx, "file-reader-current")
	)
	t.Cleanup(func() { cleanupFileScopeUsers(t, ctx, []int{userID}) })

	content := uniqueFileScopeName("reader-content")
	storage := &fileCapabilityTestStorage{}
	svc := &serviceImpl{
		configSvc: hostconfig.New(),
		storage:   storage,
		bizCtxSvc: fileScopeStaticBizCtx{ctx: &model.Context{TenantId: tenantID, UserId: userID}},
	}

	output, err := svc.CreateFromReader(tenantCtx, &CreateFromReaderInput{
		Filename:  "reader-upload.txt",
		Scene:     "plugin",
		Reader:    strings.NewReader(content),
		SizeBytes: int64(len(content)),
	})
	if err != nil {
		t.Fatalf("create file from reader: %v", err)
	}
	t.Cleanup(func() { cleanupFileScopeRecords(t, ctx, []int64{output.Id}) })

	record := queryCapabilityFileRecord(t, ctx, output.Id)
	if record.TenantId != tenantID || record.CreatedBy != int64(userID) {
		t.Fatalf("expected tenant/user %d/%d, got %d/%d", tenantID, userID, record.TenantId, record.CreatedBy)
	}
	if record.Original != "reader-upload.txt" || record.Scene != "plugin" || record.Size != int64(len(content)) {
		t.Fatalf("unexpected file metadata: %#v", record)
	}
	if record.Hash != sha256Hex(content) {
		t.Fatalf("expected hash %s, got %s", sha256Hex(content), record.Hash)
	}
	if !strings.HasSuffix(storage.putName, ".txt") || storage.putContent != content {
		t.Fatalf("expected stored txt content %q, got name=%q content=%q", content, storage.putName, storage.putContent)
	}
}

// TestCreateFromReaderUsesCapabilityBizContext verifies dynamic host-call
// contexts can provide tenant and uploader without an HTTP model.Context value.
func TestCreateFromReaderUsesCapabilityBizContext(t *testing.T) {
	var (
		ctx      = context.Background()
		tenantID = int(time.Now().UnixNano()%1000000 + 4000000)
		userID   = insertFileScopeUser(t, ctx, "file-reader-capctx-current")
	)
	t.Cleanup(func() { cleanupFileScopeUsers(t, ctx, []int{userID}) })

	var (
		content = uniqueFileScopeName("reader-capctx-content")
		capCtx  = bizctxcap.WithCurrentContext(ctx, bizctxcap.CurrentContext{TenantID: tenantID, UserID: userID})
		storage = &fileCapabilityTestStorage{}
	)
	svc := &serviceImpl{
		configSvc: hostconfig.New(),
		storage:   storage,
		bizCtxSvc: internalbizctx.New(),
	}

	output, err := svc.CreateFromReader(capCtx, &CreateFromReaderInput{
		Filename:  "capctx-upload.txt",
		Scene:     "dynamic",
		Reader:    strings.NewReader(content),
		SizeBytes: int64(len(content)),
	})
	if err != nil {
		t.Fatalf("create file from capability context: %v", err)
	}
	t.Cleanup(func() { cleanupFileScopeRecords(t, ctx, []int64{output.Id}) })

	record := queryCapabilityFileRecord(t, ctx, output.Id)
	if record.TenantId != tenantID || record.CreatedBy != int64(userID) {
		t.Fatalf("expected capability tenant/user %d/%d, got %d/%d", tenantID, userID, record.TenantId, record.CreatedBy)
	}
}

// TestCapabilityUploadCreatesFileRecord verifies plugin Files.Upload records to sys_file.
func TestCapabilityUploadCreatesFileRecord(t *testing.T) {
	var (
		ctx       = context.Background()
		tenantID  = int(time.Now().UnixNano()%1000000 + 2000000)
		tenantCtx = datascope.WithTenantScope(ctx, tenantID)
		userID    = insertFileScopeUser(t, ctx, "file-cap-upload-current")
	)
	t.Cleanup(func() { cleanupFileScopeUsers(t, ctx, []int{userID}) })

	content := uniqueFileScopeName("cap-upload-content")
	storage := &fileCapabilityTestStorage{}
	svc := &serviceImpl{
		configSvc: hostconfig.New(),
		storage:   storage,
		bizCtxSvc: fileScopeStaticBizCtx{ctx: &model.Context{TenantId: tenantID, UserId: userID}},
	}
	adapter := &fileCapabilityAdapter{owner: svc}

	projection, err := adapter.Upload(tenantCtx, filecap.UploadInput{
		Filename:      "plugin-upload.log",
		BusinessScene: "plugin-log",
		Reader:        bytes.NewBufferString(content),
		SizeBytes:     int64(len(content)),
	})
	if err != nil {
		t.Fatalf("plugin file upload: %v", err)
	}
	fileID := parseCapabilityFileIDForTest(t, projection.ID)
	t.Cleanup(func() { cleanupFileScopeRecords(t, ctx, []int64{fileID}) })

	record := queryCapabilityFileRecord(t, ctx, fileID)
	if record.Scene != "plugin-log" || record.TenantId != tenantID || record.CreatedBy != int64(userID) {
		t.Fatalf("unexpected uploaded record: %#v", record)
	}
	if projection.Name != "plugin-upload.log" || projection.BusinessScene != "plugin-log" || projection.SizeBytes != int64(len(content)) {
		t.Fatalf("unexpected projection: %#v", projection)
	}
}

// TestCapabilityCreateFromStorageCopiesPluginObject verifies storage promotion
// copies plugin storage content into file-center storage and leaves the source untouched.
func TestCapabilityCreateFromStorageCopiesPluginObject(t *testing.T) {
	var (
		ctx       = context.Background()
		tenantID  = int(time.Now().UnixNano()%1000000 + 3000000)
		tenantCtx = datascope.WithTenantScope(ctx, tenantID)
		userID    = insertFileScopeUser(t, ctx, "file-cap-storage-current")
	)
	t.Cleanup(func() { cleanupFileScopeUsers(t, ctx, []int{userID}) })

	var (
		content       = uniqueFileScopeName("storage-source-content")
		fileStorage   = &fileCapabilityTestStorage{}
		pluginStorage = &fileCapabilityPluginStorage{content: content}
	)
	svc := &serviceImpl{
		configSvc: hostconfig.New(),
		storage:   fileStorage,
		bizCtxSvc: fileScopeStaticBizCtx{ctx: &model.Context{TenantId: tenantID, UserId: userID}},
	}
	adapter := &fileCapabilityAdapter{owner: svc, storage: pluginStorage}

	projection, err := adapter.CreateFromStorage(tenantCtx, filecap.CreateFromStorageInput{
		StoragePath:   "exports/source.txt",
		BusinessScene: "plugin-export",
	})
	if err != nil {
		t.Fatalf("create file from plugin storage: %v", err)
	}
	fileID := parseCapabilityFileIDForTest(t, projection.ID)
	t.Cleanup(func() { cleanupFileScopeRecords(t, ctx, []int64{fileID}) })

	record := queryCapabilityFileRecord(t, ctx, fileID)
	if record.Original != "source.txt" || record.Scene != "plugin-export" {
		t.Fatalf("unexpected file metadata: %#v", record)
	}
	if pluginStorage.getPath != "exports/source.txt" {
		t.Fatalf("expected plugin storage get path exports/source.txt, got %q", pluginStorage.getPath)
	}
	if pluginStorage.deleteCount != 0 {
		t.Fatalf("expected copy semantics without deleting source, got deletes=%d", pluginStorage.deleteCount)
	}
	if fileStorage.putContent != content {
		t.Fatalf("expected file-center storage content %q, got %q", content, fileStorage.putContent)
	}
}

// TestCapabilityDetailRejectsTenantInvisibleFileBeforeOwnerDetail verifies
// plugin tenant filtering happens before host-owned detail projection.
func TestCapabilityDetailRejectsTenantInvisibleFileBeforeOwnerDetail(t *testing.T) {
	var (
		ctx             = context.Background()
		recordTenantID  = 710001
		requestTenantID = 710002
	)
	fileID, err := dao.SysFile.Ctx(ctx).Data(do.SysFile{
		TenantId:  recordTenantID,
		Name:      "tenant-hidden.txt",
		Original:  "tenant-hidden.txt",
		Suffix:    "txt",
		Scene:     DefaultFileSceneOther,
		Size:      1,
		Hash:      uniqueFileScopeName("cap-detail-hidden-hash"),
		Url:       "/api/v1/uploads/tenant-hidden.txt",
		Path:      "tenant-hidden.txt",
		Engine:    EngineLocal,
		CreatedBy: 0,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert tenant-hidden file: %v", err)
	}
	t.Cleanup(func() { cleanupFileScopeRecords(t, ctx, []int64{fileID}) })

	owner := &fileCapabilityDetailOwner{}
	adapter := &fileCapabilityAdapter{
		owner:        owner,
		tenantFilter: fileCapabilityTenantFilter{tenantID: requestTenantID},
	}
	_, err = adapter.Detail(ctx, filecap.FileID(strconv.FormatInt(fileID, 10)))
	if !bizerr.Is(err, capmodel.CodeCapabilityDenied) {
		t.Fatalf("expected capability denied for tenant-hidden detail, got %v", err)
	}
	if owner.detailCalled {
		t.Fatal("expected owner detail not to be called before tenant visibility passes")
	}
}

// TestCapabilityBatchGetRejectsOversizedInput verifies batch file reads are bounded.
func TestCapabilityBatchGetRejectsOversizedInput(t *testing.T) {
	ids := make([]filecap.FileID, filecap.MaxBatchGetFiles+1)
	for i := range ids {
		ids[i] = filecap.FileID(strconv.Itoa(i + 1))
	}
	_, err := (&fileCapabilityAdapter{}).BatchGet(context.Background(), ids)
	if !bizerr.Is(err, capmodel.CodeCapabilityLimitExceeded) {
		t.Fatalf("expected capability limit error, got %v", err)
	}
}

// TestFileMimeHelpersCoverSharedMappings verifies projections and filters use
// the same file suffix to MIME map.
func TestFileMimeHelpersCoverSharedMappings(t *testing.T) {
	cases := map[string]string{
		"jpg":  "image/jpeg",
		"webp": "image/webp",
		"svg":  "image/svg+xml",
		"pdf":  "application/pdf",
	}
	for suffix, mimeType := range cases {
		if got := mimeTypeFromSuffix(suffix); got != mimeType {
			t.Fatalf("expected suffix %s mime %s, got %s", suffix, mimeType, got)
		}
		if !stringSliceContains(suffixesForMimeType(mimeType), suffix) {
			t.Fatalf("expected mime %s suffix list to include %s", mimeType, suffix)
		}
	}
	if got := contentTypeForSuffix("unknown"); got != "application/octet-stream" {
		t.Fatalf("expected unknown suffix to use octet-stream, got %s", got)
	}
}

// fileCapabilityTestStorage records writes made by the host file owner.
type fileCapabilityTestStorage struct {
	storagesvc.Service
	putName    string
	putContent string
}

// Put drains the file-center upload stream and returns a deterministic path.
func (s *fileCapabilityTestStorage) Put(_ context.Context, in storagesvc.PutInput) (*storagesvc.PutOutput, error) {
	content, err := io.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}
	s.putName = path.Base(in.Key)
	s.putContent = string(content)
	return &storagesvc.PutOutput{Object: &storagesvc.Object{Key: in.Key, Size: int64(len(content))}}, nil
}

// Get returns the previously written file-center content.
func (s *fileCapabilityTestStorage) Get(_ context.Context, in storagesvc.GetInput) (*storagesvc.GetOutput, error) {
	return &storagesvc.GetOutput{
		Object: &storagesvc.Object{Key: in.Key, Size: int64(len(s.putContent))},
		Body:   io.NopCloser(strings.NewReader(s.putContent)),
		Found:  true,
	}, nil
}


// fileCapabilityPluginStorage is a plugin-private storage source fake.
type fileCapabilityPluginStorage struct {
	content     string
	getPath     string
	deleteCount int
}

// Put is unused by file promotion tests.
func (s *fileCapabilityPluginStorage) Put(context.Context, storagecap.PutInput) (*storagecap.PutOutput, error) {
	return nil, nil
}

// Get returns the configured plugin-private object content.
func (s *fileCapabilityPluginStorage) Get(_ context.Context, in storagecap.GetInput) (*storagecap.GetOutput, error) {
	s.getPath = in.Path
	return &storagecap.GetOutput{
		Object: &storagecap.Object{Path: in.Path, Size: int64(len(s.content))},
		Body:   io.NopCloser(strings.NewReader(s.content)),
		Found:  true,
	}, nil
}

// Delete records source deletion attempts so copy semantics can be asserted.
func (s *fileCapabilityPluginStorage) Delete(context.Context, storagecap.DeleteInput) error {
	s.deleteCount++
	return nil
}

// DeleteMany is unused by file promotion tests.
func (s *fileCapabilityPluginStorage) DeleteMany(context.Context, storagecap.DeleteManyInput) error {
	return nil
}

// List is unused by file promotion tests.
func (s *fileCapabilityPluginStorage) List(context.Context, storagecap.ListInput) (*storagecap.ListOutput, error) {
	return nil, nil
}

// ListCursor is unused by file promotion tests.
func (s *fileCapabilityPluginStorage) ListCursor(context.Context, storagecap.ListCursorInput) (*storagecap.ListCursorOutput, error) {
	return nil, nil
}

// Stat is unused by file promotion tests.
func (s *fileCapabilityPluginStorage) Stat(context.Context, storagecap.StatInput) (*storagecap.StatOutput, error) {
	return nil, nil
}

// BatchStat is unused by file promotion tests.
func (s *fileCapabilityPluginStorage) BatchStat(context.Context, storagecap.BatchStatInput) (*storagecap.BatchStatOutput, error) {
	return nil, nil
}

// ProviderStatuses is unused by file promotion tests.
func (s *fileCapabilityPluginStorage) ProviderStatuses(context.Context) ([]*storagecap.ProviderStatus, error) {
	return nil, nil
}

// fileCapabilityDetailOwner records whether detail fallback was reached.
type fileCapabilityDetailOwner struct {
	detailCalled bool
}

// Upload is unused by capability detail tests.
func (o *fileCapabilityDetailOwner) Upload(context.Context, *UploadInput) (*UploadOutput, error) {
	return nil, nil
}

// CreateFromReader is unused by capability detail tests.
func (o *fileCapabilityDetailOwner) CreateFromReader(context.Context, *CreateFromReaderInput) (*UploadOutput, error) {
	return nil, nil
}

// List is unused by capability detail tests.
func (o *fileCapabilityDetailOwner) List(context.Context, *ListInput) (*ListOutput, error) {
	return nil, nil
}

// Info is unused by capability detail tests.
func (o *fileCapabilityDetailOwner) Info(context.Context, int64) (*entity.SysFile, error) {
	return nil, nil
}

// InfoByIds is unused by capability detail tests.
func (o *fileCapabilityDetailOwner) InfoByIds(context.Context, []int64) ([]*entity.SysFile, error) {
	return nil, nil
}

// Delete is unused by capability detail tests.
func (o *fileCapabilityDetailOwner) Delete(context.Context, []int64) error {
	return nil
}

// OpenByID is unused by capability detail tests.
func (o *fileCapabilityDetailOwner) OpenByID(context.Context, int64) (*OpenOutput, error) {
	return nil, nil
}

// OpenByPath is unused by capability detail tests.
func (o *fileCapabilityDetailOwner) OpenByPath(context.Context, string) (*OpenOutput, error) {
	return nil, nil
}

// UsageScenes is unused by capability detail tests.
func (o *fileCapabilityDetailOwner) UsageScenes(context.Context) ([]*UsageScenesOutput, error) {
	return nil, nil
}

// Suffixes is unused by capability detail tests.
func (o *fileCapabilityDetailOwner) Suffixes(context.Context) ([]*SuffixesOutput, error) {
	return nil, nil
}

// Detail records that host owner detail projection was reached.
func (o *fileCapabilityDetailOwner) Detail(_ context.Context, id int64) (*DetailOutput, error) {
	o.detailCalled = true
	return &DetailOutput{
		SysFile: &entity.SysFile{
			Id:       id,
			Original: "tenant-hidden.txt",
			Suffix:   "txt",
			Scene:    DefaultFileSceneOther,
		},
	}, nil
}

// fileCapabilityTenantFilter returns a fixed tenant filter context.
type fileCapabilityTenantFilter struct {
	tenantID int
}

// Context returns the configured tenant filter context.
func (f fileCapabilityTenantFilter) Context(context.Context) tenantcap.TenantFilterContext {
	return tenantcap.TenantFilterContext{TenantID: f.tenantID}
}

// queryCapabilityFileRecord loads one sys_file row for capability tests.
func queryCapabilityFileRecord(t *testing.T, ctx context.Context, id int64) *entity.SysFile {
	t.Helper()

	var record *entity.SysFile
	if err := dao.SysFile.Ctx(ctx).Where(do.SysFile{Id: id}).Scan(&record); err != nil {
		t.Fatalf("query capability file record: %v", err)
	}
	if record == nil {
		t.Fatalf("expected sys_file %d to exist", id)
	}
	return record
}

// parseCapabilityFileIDForTest parses a projection file ID.
func parseCapabilityFileIDForTest(t *testing.T, id filecap.FileID) int64 {
	t.Helper()

	parsed, err := parseFileID(id)
	if err != nil {
		t.Fatalf("parse capability file id %q: %v", id, err)
	}
	return parsed
}

// sha256Hex returns the SHA-256 hex digest for test content.
func sha256Hex(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

// stringSliceContains reports whether a slice contains one string.
func stringSliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
