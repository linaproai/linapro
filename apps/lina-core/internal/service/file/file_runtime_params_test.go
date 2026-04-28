// This file verifies runtime file-upload behaviors driven by managed
// sys_config parameters.

package file

import (
	"context"
	"mime/multipart"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	hostconfig "lina-core/internal/service/config"
	"lina-core/pkg/bizerr"
)

// TestUploadRejectsFileExceedingRuntimeMaxSize verifies managed upload size
// settings are enforced before storage begins.
func TestUploadRejectsFileExceedingRuntimeMaxSize(t *testing.T) {
	withRuntimeParamValue(t, hostconfig.RuntimeParamKeyUploadMaxSize, "1")

	svc := New()
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
