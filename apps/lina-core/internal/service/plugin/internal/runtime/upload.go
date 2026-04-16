// This file handles runtime wasm package uploads and writes validated runtime
// artifacts into the configured runtime storage directory for later discovery,
// installation, and review.

package runtime

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/closeutil"
)

// DynamicUploadInput defines input for uploading a runtime wasm package.
type DynamicUploadInput struct {
	// File is the uploaded runtime wasm package.
	File *ghttp.UploadFile
	// OverwriteSupport allows replacing a not-installed runtime package.
	OverwriteSupport bool
}

// DynamicUploadOutput defines output for uploading a runtime wasm package.
type DynamicUploadOutput struct {
	// Id is the plugin identifier embedded in the runtime package.
	Id string
	// Name is the display name embedded in the runtime package.
	Name string
	// Version is the plugin version embedded in the runtime package.
	Version string
	// Type is the normalized top-level plugin type.
	Type string
	// RuntimeKind is the validated runtime artifact kind.
	RuntimeKind string
	// RuntimeABI is the validated runtime ABI version.
	RuntimeABI string
	// Installed is the synchronized installation status after upload.
	Installed int
	// Enabled is the synchronized enablement status after upload.
	Enabled int
}

// UploadDynamicPackage validates one runtime wasm package and writes it into the
// configured plugin.dynamic.storagePath directory.
func (s *serviceImpl) UploadDynamicPackage(ctx context.Context, in *DynamicUploadInput) (out *DynamicUploadOutput, err error) {
	if in == nil || in.File == nil {
		return nil, gerror.New("请上传动态插件文件")
	}

	source, err := in.File.Open()
	if err != nil {
		return nil, gerror.Wrap(err, "打开动态插件文件失败")
	}
	defer closeutil.Close(source, &err, "关闭动态插件上传文件失败")

	content, err := io.ReadAll(source)
	if err != nil {
		return nil, gerror.Wrap(err, "读取动态插件文件失败")
	}
	if len(content) == 0 {
		return nil, gerror.New("动态插件文件不能为空")
	}
	return s.storeUploadedPackage(
		ctx,
		normalizeUploadFilename(in.File.Filename),
		content,
		in.OverwriteSupport,
	)
}

// normalizeUploadFilename ensures the filename ends with .wasm.
func normalizeUploadFilename(filename string) string {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return "runtime-plugin.wasm"
	}
	if strings.EqualFold(gfile.ExtName(filename), ".wasm") {
		return filename
	}

	// Some browser upload pipelines downgrade the filename to a generic blob
	// name even when the file content is valid WebAssembly. Runtime validation
	// still checks the wasm header and embedded Lina metadata so we only
	// normalize the display name here instead of rejecting the upload early.
	return filepath.Base(filename) + ".wasm"
}

func (s *serviceImpl) storeUploadedPackage(
	ctx context.Context,
	filename string,
	content []byte,
	overwriteSupport bool,
) (*DynamicUploadOutput, error) {
	artifact, err := s.ParseRuntimeWasmArtifactContent(filename, content)
	if err != nil {
		return nil, err
	}
	if artifact.Manifest == nil {
		return nil, gerror.New("动态插件嵌入清单不能为空")
	}

	manifest := &catalog.Manifest{
		ID:              strings.TrimSpace(artifact.Manifest.ID),
		Name:            strings.TrimSpace(artifact.Manifest.Name),
		Version:         strings.TrimSpace(artifact.Manifest.Version),
		Type:            catalog.NormalizeType(artifact.Manifest.Type).String(),
		Description:     strings.TrimSpace(artifact.Manifest.Description),
		Menus:           artifact.Manifest.Menus,
		RuntimeArtifact: artifact,
	}
	if err = s.catalogSvc.ValidateUploadedRuntimeManifest(manifest); err != nil {
		return nil, err
	}

	storageDir, err := s.catalogSvc.RuntimeStorageDir(ctx)
	if err != nil {
		return nil, err
	}
	targetPath := filepath.Join(storageDir, buildArtifactFileName(manifest.ID))

	registry, err := s.catalogSvc.GetRegistry(ctx, manifest.ID)
	if err != nil {
		return nil, err
	}
	registry, err = s.reconcileRegistryArtifactState(ctx, registry)
	if err != nil {
		return nil, err
	}
	if registry != nil && catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return nil, gerror.New("已存在同名源码插件，不允许上传动态插件覆盖")
	}

	allowInstalledUpgradeOverwrite := false
	if registry != nil && registry.Installed == catalog.InstalledYes {
		compareResult, compareErr := catalog.CompareSemanticVersions(manifest.Version, registry.Version)
		if compareErr != nil {
			return nil, compareErr
		}
		if compareResult <= 0 {
			return nil, gerror.New("已安装的动态插件只允许上传更高版本作为待切换 release")
		}
		allowInstalledUpgradeOverwrite = true
	}
	if conflictPath, conflictErr := s.findDuplicateArtifactPath(storageDir, manifest.ID, targetPath); conflictErr != nil {
		return nil, conflictErr
	} else if conflictPath != "" {
		return nil, gerror.Newf("动态插件目录存在重复的插件ID %s，请先移除冲突文件: %s", manifest.ID, conflictPath)
	}
	if gfile.Exists(targetPath) && !overwriteSupport && !allowInstalledUpgradeOverwrite {
		return nil, gerror.New("动态插件文件已存在，请开启覆盖后重试")
	}
	if err = gfile.Mkdir(storageDir); err != nil {
		return nil, gerror.Wrap(err, "创建动态插件存储目录失败")
	}

	backupContent := []byte(nil)
	targetExisted := gfile.Exists(targetPath)
	if targetExisted {
		backupContent = gfile.GetBytes(targetPath)
	}
	if err = gfile.PutBytes(targetPath, content); err != nil {
		return nil, gerror.Wrap(err, "写入动态插件产物失败")
	}
	reloadedManifest, err := s.catalogSvc.LoadManifestFromArtifactPath(targetPath)
	if err != nil {
		if restoreErr := restoreArtifactBackup(targetPath, targetExisted, backupContent); restoreErr != nil {
			return nil, gerror.Wrapf(err, "解析上传后的动态插件失败，且恢复备份失败: %v", restoreErr)
		}
		return nil, err
	}
	s.invalidateFrontendBundle(ctx, reloadedManifest.ID, "runtime_package_uploaded")

	syncedRegistry, err := s.catalogSvc.SyncManifest(ctx, reloadedManifest)
	if err != nil {
		if restoreErr := restoreArtifactBackup(targetPath, targetExisted, backupContent); restoreErr != nil {
			return nil, gerror.Wrapf(err, "同步动态插件清单失败，且恢复备份失败: %v", restoreErr)
		}
		return nil, err
	}

	return &DynamicUploadOutput{
		Id:          reloadedManifest.ID,
		Name:        reloadedManifest.Name,
		Version:     reloadedManifest.Version,
		Type:        reloadedManifest.Type,
		RuntimeKind: reloadedManifest.RuntimeArtifact.RuntimeKind,
		RuntimeABI:  reloadedManifest.RuntimeArtifact.ABIVersion,
		Installed:   syncedRegistry.Installed,
		Enabled:     syncedRegistry.Status,
	}, nil
}

// StoreUploadedPackage is the exported form of storeUploadedPackage for cross-package access.
func (s *serviceImpl) StoreUploadedPackage(ctx context.Context, filename string, content []byte, overwriteSupport bool) (*DynamicUploadOutput, error) {
	return s.storeUploadedPackage(ctx, filename, content, overwriteSupport)
}

// findDuplicateArtifactPath scans the storage directory for any .wasm file other than
// targetPath that embeds the same plugin ID. Returns the conflicting path if found.
func (s *serviceImpl) findDuplicateArtifactPath(storageDir string, pluginID string, targetPath string) (string, error) {
	if !gfile.Exists(storageDir) || !gfile.IsDir(storageDir) {
		return "", nil
	}

	artifactFiles, err := gfile.ScanDirFile(storageDir, "*.wasm", false)
	if err != nil {
		return "", err
	}
	for _, artifactPath := range artifactFiles {
		if filepath.Clean(artifactPath) == filepath.Clean(targetPath) {
			continue
		}
		artifact, parseErr := s.ParseRuntimeWasmArtifact(artifactPath)
		if parseErr != nil {
			return "", gerror.Wrapf(parseErr, "解析现有动态插件文件失败: %s", artifactPath)
		}
		if artifact.Manifest != nil && strings.TrimSpace(artifact.Manifest.ID) == pluginID {
			return artifactPath, nil
		}
	}
	return "", nil
}

// restoreArtifactBackup reverts a failed upload by restoring the previous artifact file.
func restoreArtifactBackup(targetPath string, targetExisted bool, backupContent []byte) error {
	if targetExisted {
		if err := gfile.PutBytes(targetPath, backupContent); err != nil {
			return gerror.Wrap(err, "恢复动态插件备份文件失败")
		}
		return nil
	}
	if err := gfile.Remove(targetPath); err != nil {
		return gerror.Wrap(err, "删除失败的动态插件产物失败")
	}
	return nil
}
