// demo_storage.go implements plugin-owned attachment storage helpers and
// uninstall-time cleanup for the source-plugin sample.

package demo

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/grand"
)

// Attachment storage constants define the storage namespace and default upload
// ceiling used by the sample source plugin.
const (
	demoAttachmentStorageNamespace = "plugin-demo-source"
	defaultUploadMaxSizeMB         = 10
)

// PurgeStorageData clears plugin-owned attachment files before uninstall SQL drops the data table.
func (s *serviceImpl) PurgeStorageData(ctx context.Context) error {
	paths, err := listAllAttachmentPaths(ctx)
	if err != nil {
		return err
	}
	for _, path := range paths {
		if err = deleteDemoAttachmentFile(ctx, path); err != nil {
			return err
		}
	}
	storageRoot, err := resolveDemoAttachmentStorageRoot(ctx)
	if err != nil {
		return err
	}
	if gfile.Exists(storageRoot) {
		if err = gfile.Remove(storageRoot); err != nil {
			return gerror.Wrap(err, "清理源码插件示例附件存储目录失败")
		}
	}
	return nil
}

// saveDemoAttachmentFile stores one optional uploaded attachment into the
// plugin-owned storage area.
func saveDemoAttachmentFile(
	ctx context.Context,
	file *ghttp.UploadFile,
) (originalName string, relativePath string, err error) {
	if file == nil {
		return "", "", nil
	}
	if err = validateDemoAttachmentFileSize(ctx, file); err != nil {
		return "", "", err
	}

	sanitizedName := sanitizeAttachmentFilename(file.Filename)
	source, err := file.Open()
	if err != nil {
		return "", "", gerror.Wrap(err, "打开源码插件示例附件失败")
	}
	defer func() {
		closeErr := source.Close()
		if err == nil && closeErr != nil {
			err = gerror.Wrap(closeErr, "关闭源码插件示例附件失败")
		}
	}()

	storageRoot, err := resolveDemoAttachmentStorageRoot(ctx)
	if err != nil {
		return "", "", err
	}

	now := gtime.Now()
	dir := filepath.Join(now.Format("Y"), now.Format("m"))
	targetDir := gfile.Join(storageRoot, dir)
	if err = gfile.Mkdir(targetDir); err != nil {
		return "", "", gerror.Wrap(err, "创建源码插件示例附件目录失败")
	}

	ext := gfile.ExtName(sanitizedName)
	storedName := fmt.Sprintf("%s_%s", now.Format("Ymd_His"), grand.S(8))
	if ext != "" {
		storedName += "." + gstr.ToLower(ext)
	}
	fullPath := gfile.Join(targetDir, storedName)

	targetFile, err := os.Create(fullPath)
	if err != nil {
		return "", "", gerror.Wrap(err, "创建源码插件示例附件文件失败")
	}
	defer func() {
		closeErr := targetFile.Close()
		if err == nil && closeErr != nil {
			err = gerror.Wrap(closeErr, "关闭源码插件示例附件文件失败")
		}
	}()

	if _, err = io.Copy(targetFile, source); err != nil {
		_ = os.Remove(fullPath)
		return "", "", gerror.Wrap(err, "写入源码插件示例附件失败")
	}

	return sanitizedName, gfile.Join(dir, storedName), nil
}

// deleteDemoAttachmentFile removes one stored attachment when its relative path
// is present.
func deleteDemoAttachmentFile(ctx context.Context, relativePath string) error {
	if strings.TrimSpace(relativePath) == "" {
		return nil
	}
	fullPath, err := buildDemoAttachmentFullPath(ctx, relativePath)
	if err != nil {
		return err
	}
	if !gfile.Exists(fullPath) {
		return nil
	}
	if err = gfile.Remove(fullPath); err != nil {
		return gerror.Wrap(err, "删除源码插件示例附件文件失败")
	}
	return nil
}

// buildDemoAttachmentFullPath resolves one relative attachment path against the
// plugin-owned storage root.
func buildDemoAttachmentFullPath(ctx context.Context, relativePath string) (string, error) {
	storageRoot, err := resolveDemoAttachmentStorageRoot(ctx)
	if err != nil {
		return "", err
	}
	return gfile.Join(storageRoot, relativePath), nil
}

// resolveDemoAttachmentStorageRoot resolves the plugin-owned attachment storage
// root under the configured upload path.
func resolveDemoAttachmentStorageRoot(ctx context.Context) (string, error) {
	uploadPath := strings.TrimSpace(g.Cfg().MustGet(ctx, "upload.path", "temp/upload").String())
	if uploadPath == "" {
		uploadPath = "temp/upload"
	}
	return filepath.Clean(gfile.Join(uploadPath, demoAttachmentStorageNamespace)), nil
}

// validateDemoAttachmentFileSize enforces the runtime-configured attachment
// upload ceiling.
func validateDemoAttachmentFileSize(ctx context.Context, file *ghttp.UploadFile) error {
	if file == nil {
		return nil
	}
	maxSizeMB := g.Cfg().MustGet(ctx, "upload.maxSize", defaultUploadMaxSizeMB).Int64()
	if maxSizeMB <= 0 {
		maxSizeMB = defaultUploadMaxSizeMB
	}
	if file.Size > maxSizeMB*1024*1024 {
		return gerror.Newf("附件大小不能超过%dMB", maxSizeMB)
	}
	return nil
}

// sanitizeAttachmentFilename strips unsafe characters and truncates overly long
// attachment filenames.
func sanitizeAttachmentFilename(filename string) string {
	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, "\x00", "")
	if strings.TrimSpace(filename) == "" {
		return "attachment"
	}

	disallowed := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, item := range disallowed {
		filename = strings.ReplaceAll(filename, item, "_")
	}
	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		name := filename[:255-len(ext)]
		filename = name + ext
	}
	return filename
}
