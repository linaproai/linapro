// This file covers runtime package upload-size validation behaviors.

package runtime_test

import (
	"context"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
)

// fixedUploadSizeProvider lets tests override the runtime upload ceiling.
type fixedUploadSizeProvider struct {
	maxSizeMB int64
}

// GetUploadMaxSize returns the configured upload ceiling used by the test runtime.
func (p fixedUploadSizeProvider) GetUploadMaxSize(_ context.Context) (int64, error) {
	return p.maxSizeMB, nil
}

// TestUploadDynamicPackageRejectsFileExceedingRuntimeMaxSize verifies that the
// runtime upload ceiling is enforced before parsing the artifact payload.
func TestUploadDynamicPackageRejectsFileExceedingRuntimeMaxSize(t *testing.T) {
	services := testutil.NewServices()
	services.Runtime.SetUploadSizeProvider(fixedUploadSizeProvider{maxSizeMB: 1})

	_, err := services.Runtime.UploadDynamicPackage(context.Background(), &runtime.DynamicUploadInput{
		File: &ghttp.UploadFile{
			FileHeader: &multipart.FileHeader{
				Filename: "too-large.wasm",
				Size:     2 * 1024 * 1024,
			},
		},
	})
	if err == nil {
		t.Fatal("expected oversized runtime package upload to fail")
	}
	if !strings.Contains(err.Error(), "文件大小不能超过1MB") {
		t.Fatalf("expected friendly runtime upload size error, got %v", err)
	}
}
