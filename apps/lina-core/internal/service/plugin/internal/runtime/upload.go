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

// bytesPerMegabyte converts the configured `sys.upload.maxSize` value from MB
// into raw bytes when upload.go compares the runtime file-size ceiling against
// the actual uploaded artifact size.
const bytesPerMegabyte int64 = 1024 * 1024

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
		return nil, gerror.New("dynamic plugin file is required")
	}
	if err = s.validateUploadedPackageSize(ctx, in.File.Size); err != nil {
		return nil, err
	}

	source, err := in.File.Open()
	if err != nil {
		return nil, gerror.Wrap(err, "open dynamic plugin file failed")
	}
	defer closeutil.Close(ctx, source, &err, "close dynamic plugin upload file failed")

	content, err := io.ReadAll(source)
	if err != nil {
		return nil, gerror.Wrap(err, "read dynamic plugin file failed")
	}
	if len(content) == 0 {
		return nil, gerror.New("dynamic plugin file cannot be empty")
	}
	if err = s.validateUploadedPackageSize(ctx, int64(len(content))); err != nil {
		return nil, err
	}
	return s.storeUploadedPackage(
		ctx,
		normalizeUploadFilename(in.File.Filename),
		content,
		in.OverwriteSupport,
	)
}

// validateUploadedPackageSize enforces the runtime-effective upload ceiling for one package upload.
func (s *serviceImpl) validateUploadedPackageSize(ctx context.Context, sizeBytes int64) error {
	if sizeBytes <= 0 || s == nil || s.uploadSize == nil {
		return nil
	}

	uploadMaxSizeMB, err := s.uploadSize.GetUploadMaxSize(ctx)
	if err != nil {
		return err
	}
	if uploadMaxSizeMB <= 0 {
		return nil
	}
	if sizeBytes <= uploadMaxSizeMB*bytesPerMegabyte {
		return nil
	}
	return gerror.Newf("file size cannot exceed %dMB", uploadMaxSizeMB)
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

// storeUploadedPackage validates one runtime artifact upload and writes it into
// the canonical storage location after registry safety checks.
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
		return nil, gerror.New("dynamic plugin embedded manifest cannot be nil")
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
		return nil, gerror.New("a source plugin with the same ID already exists; dynamic plugin upload cannot overwrite it")
	}

	allowInstalledUpgradeOverwrite := false
	if registry != nil && registry.Installed == catalog.InstalledYes {
		compareResult, compareErr := catalog.CompareSemanticVersions(manifest.Version, registry.Version)
		if compareErr != nil {
			return nil, compareErr
		}
		if compareResult <= 0 {
			return nil, gerror.New("installed dynamic plugins can only upload a higher version as the pending release")
		}
		allowInstalledUpgradeOverwrite = true
	}
	if conflictPath, conflictErr := s.findDuplicateArtifactPath(storageDir, manifest.ID, targetPath); conflictErr != nil {
		return nil, conflictErr
	} else if conflictPath != "" {
		return nil, gerror.Newf("dynamic plugin storage contains duplicate plugin ID %s, remove conflicting file first: %s", manifest.ID, conflictPath)
	}
	if gfile.Exists(targetPath) && !overwriteSupport && !allowInstalledUpgradeOverwrite {
		return nil, gerror.New("dynamic plugin file already exists; enable overwrite and retry")
	}
	if err = gfile.Mkdir(storageDir); err != nil {
		return nil, gerror.Wrap(err, "create dynamic plugin storage directory failed")
	}

	backupContent := []byte(nil)
	targetExisted := gfile.Exists(targetPath)
	if targetExisted {
		backupContent = gfile.GetBytes(targetPath)
	}
	if err = gfile.PutBytes(targetPath, content); err != nil {
		return nil, gerror.Wrap(err, "write dynamic plugin artifact failed")
	}
	reloadedManifest, err := s.catalogSvc.LoadManifestFromArtifactPath(targetPath)
	if err != nil {
		if restoreErr := restoreArtifactBackup(targetPath, targetExisted, backupContent); restoreErr != nil {
			return nil, gerror.Wrapf(err, "parse uploaded dynamic plugin failed and restoring backup failed: %v", restoreErr)
		}
		return nil, err
	}
	s.invalidateRuntimeCaches(ctx, reloadedManifest.ID, "runtime_package_uploaded")

	syncedRegistry, err := s.catalogSvc.SyncManifest(ctx, reloadedManifest)
	if err != nil {
		if restoreErr := restoreArtifactBackup(targetPath, targetExisted, backupContent); restoreErr != nil {
			return nil, gerror.Wrapf(err, "sync dynamic plugin manifest failed and restoring backup failed: %v", restoreErr)
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
			return "", gerror.Wrapf(parseErr, "parse existing dynamic plugin file failed: %s", artifactPath)
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
			return gerror.Wrap(err, "restore dynamic plugin backup file failed")
		}
		return nil
	}
	if err := gfile.Remove(targetPath); err != nil {
		return gerror.Wrap(err, "delete failed dynamic plugin artifact failed")
	}
	return nil
}
