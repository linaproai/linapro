// This file verifies runtime file-upload behaviors driven by managed
// sys_config parameters.

package file

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	storagesvc "lina-core/internal/service/storage"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/storagecap"
)

// TestStorageUploadBodyOmitsCloser ensures host-owned spool files are not exposed
// as io.Closer to storage backends (COS SDK TeeReader closes request bodies).
func TestStorageUploadBodyOmitsCloser(t *testing.T) {
	spooled, err := os.CreateTemp("", "lina-file-upload-body-*")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	t.Cleanup(func() {
		_ = spooled.Close()
		_ = os.Remove(spooled.Name())
	})
	body := storageUploadBody{ReadSeeker: spooled}
	if _, ok := any(body).(io.Closer); ok {
		t.Fatal("storageUploadBody must not implement io.Closer")
	}
	if _, ok := any(body).(io.ReadSeeker); !ok {
		t.Fatal("storageUploadBody must remain an io.ReadSeeker for storage retries")
	}
}

// TestCleanupSpooledUploadIgnoresAlreadyClosedFile verifies cleanup does not map
// "file already closed" into FILE_READ_RESET_FAILED after a backend closed the fd.
func TestCleanupSpooledUploadIgnoresAlreadyClosedFile(t *testing.T) {
	spooled, err := os.CreateTemp("", "lina-file-upload-cleanup-*")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	name := spooled.Name()
	if closeErr := spooled.Close(); closeErr != nil {
		t.Fatalf("close temp file: %v", closeErr)
	}
	var cleanupErr error
	cleanupSpooledUpload(spooled, &cleanupErr)
	if cleanupErr != nil {
		t.Fatalf("expected cleanup to ignore already-closed file, got %v", cleanupErr)
	}
	if _, statErr := os.Stat(name); !os.IsNotExist(statErr) {
		t.Fatalf("expected temp file removed, stat err=%v", statErr)
	}
}

// TestCreateFromReaderSurvivesBodyClosingStorage simulates cloud SDKs that close
// the request body after Put and asserts the host upload still succeeds.
func TestCreateFromReaderSurvivesBodyClosingStorage(t *testing.T) {
	var (
		ctx       = context.Background()
		tenantID  = int(time.Now().UnixNano()%1000000 + 5000000)
		tenantCtx = datascope.WithTenantScope(ctx, tenantID)
		userID    = insertFileScopeUser(t, ctx, "file-upload-closing-body")
	)
	t.Cleanup(func() { cleanupFileScopeUsers(t, ctx, []int{userID}) })

	content := "cos-upload-body-close-regression"
	storage := &closingBodyUploadStorage{}
	svc := &serviceImpl{
		configSvc: hostconfig.New(),
		storage:   storage,
		bizCtxSvc: fileScopeStaticBizCtx{ctx: &model.Context{TenantId: tenantID, UserId: userID}},
	}

	output, err := svc.CreateFromReader(tenantCtx, &CreateFromReaderInput{
		Filename:  "cos-image.png",
		Scene:     "other",
		Reader:    strings.NewReader(content),
		SizeBytes: int64(len(content)),
	})
	if err != nil {
		t.Fatalf("create from reader with body-closing storage: %v", err)
	}
	t.Cleanup(func() { cleanupFileScopeRecords(t, ctx, []int64{output.Id}) })

	if storage.putContent != content {
		t.Fatalf("expected stored content %q, got %q", content, storage.putContent)
	}
	if storage.bodyWasCloser {
		t.Fatal("expected host to pass a non-closer body to storage Put")
	}
	if path.Base(output.Name) == "" || output.Size != int64(len(content)) {
		t.Fatalf("unexpected upload output %#v", output)
	}
}

// closingBodyUploadStorage mimics cloud SDK Put behavior: drain the body, then
// Close it when the body implements io.Closer (as COS TeeReader does).
type closingBodyUploadStorage struct {
	storagesvc.Service
	putContent    string
	bodyWasCloser bool
}

// Put records content and closes ReadCloser bodies to reproduce the COS failure.
func (s *closingBodyUploadStorage) Put(_ context.Context, in storagesvc.PutInput) (*storagesvc.PutOutput, error) {
	content, err := io.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}
	s.putContent = string(content)
	if closer, ok := in.Body.(io.Closer); ok {
		s.bodyWasCloser = true
		if closeErr := closer.Close(); closeErr != nil {
			return nil, closeErr
		}
	}
	return &storagesvc.PutOutput{Object: &storagesvc.Object{Key: in.Key, Size: int64(len(content))}}, nil
}

// TestCreateFromReaderFailsClosedOnStorageProviderConflict verifies multi-cloud
// provider conflict rejects every upload path, including hash-dedup reuse that
// would otherwise skip storage.Put (file upload vs image upload inconsistency).
func TestCreateFromReaderFailsClosedOnStorageProviderConflict(t *testing.T) {
	var (
		ctx      = context.Background()
		tenantID = 0
		userID   = insertFileScopeUser(t, ctx, "file-upload-conflict")
	)
	t.Cleanup(func() { cleanupFileScopeUsers(t, ctx, []int{userID}) })

	aID := fmt.Sprintf("file-upload-conflict-a-%d", time.Now().UnixNano())
	bID := fmt.Sprintf("file-upload-conflict-b-%d", time.Now().UnixNano()+1)
	for _, id := range []string{aID, bID} {
		if err := storagecap.Provide(id, func(context.Context, storagecap.ProviderEnv) (storagecap.Provider, error) {
			return &fileConflictNoopProvider{}, nil
		}); err != nil {
			t.Fatalf("provide %s: %v", id, err)
		}
	}

	local := &fileTenantUploadStorage{}
	storage := storagesvc.NewResolvingService(
		local,
		fileConflictRuntime{available: map[string]bool{aID: true, bID: true}},
		&fileConflictNoopProvider{},
	)
	svc := &serviceImpl{
		configSvc: hostconfig.New(),
		storage:   storage,
		bizCtxSvc: fileScopeStaticBizCtx{ctx: &model.Context{TenantId: tenantID, UserId: userID}},
	}
	tenantCtx := datascope.WithTenantScope(ctx, tenantID)

	// Seed a hash-dedup candidate while backend is still writable (plain local).
	seedSvc := &serviceImpl{
		configSvc: hostconfig.New(),
		storage:   &fileTenantUploadStorage{},
		bizCtxSvc: fileScopeStaticBizCtx{ctx: &model.Context{TenantId: tenantID, UserId: userID}},
	}
	content := "duplicate-file-upload-conflict-body"
	seeded, err := seedSvc.CreateFromReader(tenantCtx, &CreateFromReaderInput{
		Filename:  "dup-conflict.txt",
		Scene:     "other",
		Reader:    strings.NewReader(content),
		SizeBytes: int64(len(content)),
	})
	if err != nil {
		t.Fatalf("seed upload: %v", err)
	}
	t.Cleanup(func() { cleanupFileScopeRecords(t, ctx, []int64{seeded.Id}) })

	// Same bytes hit hash reuse; must still fail because multiple providers are active.
	_, err = svc.CreateFromReader(tenantCtx, &CreateFromReaderInput{
		Filename:  "dup-conflict-again.txt",
		Scene:     "other",
		Reader:    strings.NewReader(content),
		SizeBytes: int64(len(content)),
	})
	if err == nil {
		t.Fatal("expected hash-reuse upload to fail when multiple storage providers are enabled")
	}
	if !bizerr.Is(err, CodeFileStorageConflict) {
		t.Fatalf("expected CodeFileStorageConflict, got %v", err)
	}
}

type fileConflictRuntime struct {
	available map[string]bool
}

func (r fileConflictRuntime) ProviderPluginAvailable(_ context.Context, pluginID string) bool {
	return r.available[pluginID]
}

type fileConflictNoopProvider struct{}

func (*fileConflictNoopProvider) Put(context.Context, storagecap.ProviderPutInput) (*storagecap.ProviderObject, error) {
	return &storagecap.ProviderObject{}, nil
}
func (*fileConflictNoopProvider) Get(context.Context, storagecap.ProviderGetInput) (*storagecap.ProviderGetOutput, error) {
	return &storagecap.ProviderGetOutput{Found: false}, nil
}
func (*fileConflictNoopProvider) Delete(context.Context, storagecap.ProviderDeleteInput) error {
	return nil
}
func (*fileConflictNoopProvider) DeleteMany(context.Context, storagecap.ProviderDeleteManyInput) error {
	return nil
}
func (*fileConflictNoopProvider) List(context.Context, storagecap.ProviderListInput) (*storagecap.ProviderListOutput, error) {
	return &storagecap.ProviderListOutput{}, nil
}
func (*fileConflictNoopProvider) ListCursor(context.Context, storagecap.ProviderListCursorInput) (*storagecap.ProviderListCursorOutput, error) {
	return &storagecap.ProviderListCursorOutput{}, nil
}
func (*fileConflictNoopProvider) Stat(context.Context, storagecap.ProviderStatInput) (*storagecap.ProviderStatOutput, error) {
	return &storagecap.ProviderStatOutput{Found: false}, nil
}
func (*fileConflictNoopProvider) BatchStat(context.Context, storagecap.ProviderBatchStatInput) (*storagecap.ProviderBatchStatOutput, error) {
	return &storagecap.ProviderBatchStatOutput{}, nil
}

// TestUploadRejectsFileExceedingRuntimeMaxSize verifies managed upload size
// settings are enforced before storage begins.
func TestUploadRejectsFileExceedingRuntimeMaxSize(t *testing.T) {
	withRuntimeParamValue(t, hostconfig.RuntimeParamKeyUploadMaxSize, "1")

	var (
		bizCtxSvc = bizctx.New()
		orgCapSvc = orgspi.New(nil, nil, nil)
		svc       = New(hostconfig.New(), nil, bizCtxSvc, nil, datascope.New(bizCtxSvc, nil, orgCapSvc.Scope()))
	)
	_, err := svc.Upload(context.Background(), &UploadInput{
		File: &ghttp.UploadFile{
			FileHeader: &multipart.FileHeader{
				Filename: "too-large.txt",
				Size:     2 * 1024 * 1024,
			},
		},
		Scene: "other",
	})
	if err == nil {
		t.Fatal("expected oversized upload to fail")
	}
	messageErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured file upload error, got %T %v", err, err)
	}
	if !messageErr.Matches(CodeFileTooLarge) {
		t.Fatalf("expected %s, got %s", CodeFileTooLarge.RuntimeCode(), messageErr.RuntimeCode())
	}
	if messageErr.Params()["maxSizeMB"] != int64(1) {
		t.Fatalf("expected maxSizeMB=1, got %v", messageErr.Params()["maxSizeMB"])
	}
}

// withRuntimeParamValue temporarily overrides one protected runtime parameter
// and restores the original sys_config record during cleanup.
func withRuntimeParamValue(t *testing.T, key string, value string) {
	t.Helper()

	ctx := context.Background()
	original, err := queryRuntimeParam(ctx, key)
	if err != nil {
		t.Fatalf("query runtime param %s: %v", key, err)
	}

	if original == nil {
		_, err = dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
			Name:   key,
			Key:    key,
			Value:  value,
			Remark: "test override",
		}).Insert()
		if err != nil {
			t.Fatalf("insert runtime param %s: %v", key, err)
		}
		markRuntimeParamChanged(t, ctx)
		t.Cleanup(func() {
			if _, cleanupErr := dao.SysConfig.Ctx(ctx).Unscoped().Where(do.SysConfig{Key: key}).Delete(); cleanupErr != nil {
				t.Fatalf("cleanup runtime param %s: %v", key, cleanupErr)
			}
			markRuntimeParamChanged(t, ctx)
		})
		return
	}

	_, err = dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Id: original.Id}).
		Data(do.SysConfig{Value: value}).
		Update()
	if err != nil {
		t.Fatalf("update runtime param %s: %v", key, err)
	}
	markRuntimeParamChanged(t, ctx)
	t.Cleanup(func() {
		_, cleanupErr := dao.SysConfig.Ctx(ctx).
			Unscoped().
			Where(do.SysConfig{Id: original.Id}).
			Data(do.SysConfig{
				Name:   original.Name,
				Key:    original.Key,
				Value:  original.Value,
				Remark: original.Remark,
			}).
			Update()
		if cleanupErr != nil {
			t.Fatalf("restore runtime param %s: %v", key, cleanupErr)
		}
		markRuntimeParamChanged(t, ctx)
	})
}

// markRuntimeParamChanged bumps the runtime-parameter revision for tests after
// direct sys_config mutations.
func markRuntimeParamChanged(t *testing.T, ctx context.Context) {
	t.Helper()

	if err := hostconfig.New().MarkRuntimeParamsChanged(ctx); err != nil {
		t.Fatalf("mark runtime params changed: %v", err)
	}
}

// queryRuntimeParam loads one sys_config record by protected runtime-parameter key.
func queryRuntimeParam(ctx context.Context, key string) (*entity.SysConfig, error) {
	var runtimeParam *entity.SysConfig
	err := dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Key: key}).
		Scan(&runtimeParam)
	if err != nil {
		return nil, err
	}
	return runtimeParam, nil
}
