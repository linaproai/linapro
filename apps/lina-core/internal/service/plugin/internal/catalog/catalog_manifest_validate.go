// This file validates plugin manifests for structural correctness and
// validates uploaded runtime manifests from WASM artifacts.

package catalog

import (
	"crypto/sha256"
	"encoding/json"
	"io/fs"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"

	pluginv1 "lina-core/api/plugin/v1"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/resourcefs"
)

// defaultPublicAssetIndex is the fallback directory index file for one
// public_assets declaration when plugin.yaml does not specify index.
const defaultPublicAssetIndex = "index.html"

// DefaultPublicAssetIndex returns the directory index fallback used when one
// public_assets declaration omits the index field.
func DefaultPublicAssetIndex() string {
	return defaultPublicAssetIndex
}

// Default menu flag values applied when manifest menu fields are omitted.
const (
	menuDefaultVisible = 1
	menuDefaultStatus  = 1
	menuDefaultIsFrame = 0
	menuDefaultIsCache = 0
)

// ValidateManifest validates required fields and structural constraints in a plugin manifest.
// For source plugins it additionally checks for go.mod and backend/plugin.go.
// For dynamic plugins it optionally validates the runtime artifact via ArtifactParser.
//
//nolint:cyclop // Manifest validation is intentionally explicit so each plugin.yaml field returns a precise error.
func (s *serviceImpl) ValidateManifest(manifest *Manifest, filePath string) error {
	rootDir := filepath.Dir(filePath)
	if strings.TrimSpace(filePath) == "" && strings.TrimSpace(manifest.RootDir) != "" {
		rootDir = manifest.RootDir
	}
	fileLabel := strings.TrimSpace(filePath)
	if fileLabel == "" {
		fileLabel = strings.TrimSpace(manifest.ManifestPath)
	}
	if fileLabel == "" {
		fileLabel = manifest.ID
	}

	if manifest.ID == "" {
		return gerror.Newf("plugin manifest is missing id: %s", fileLabel)
	}
	if manifest.Name == "" {
		return gerror.Newf("plugin manifest is missing name: %s", fileLabel)
	}
	if manifest.Version == "" {
		return gerror.Newf("plugin manifest is missing version: %s", fileLabel)
	}
	if manifest.Type == "" {
		manifest.Type = pluginv1.PluginTypeSource.String()
	} else {
		manifest.Type = plugintypes.NormalizeType(manifest.Type).String()
	}
	if !plugintypes.IsSupportedType(manifest.Type) {
		return gerror.Newf("plugin type only supports source/dynamic: %s", fileLabel)
	}
	manifest.Distribution = plugintypes.NormalizeDistribution(manifest.Distribution).String()
	if !plugintypes.IsSupportedDistribution(manifest.Distribution) {
		return gerror.Newf("plugin distribution only supports managed/builtin: %s", fileLabel)
	}
	if err := s.hydrateManifestTenantGovernanceFromFile(manifest, filePath); err != nil {
		return gerror.Wrapf(err, "plugin tenant governance metadata cannot be loaded: %s", fileLabel)
	}
	if err := normalizeManifestTenantGovernance(manifest); err != nil {
		return gerror.Wrapf(err, "plugin tenant governance metadata is invalid: %s", fileLabel)
	}
	manifest.ID = strings.TrimSpace(manifest.ID)
	if err := plugintypes.ValidatePluginID(manifest.ID); err != nil {
		return gerror.Wrapf(err, "plugin ID is invalid: %s", fileLabel)
	}
	if err := plugintypes.ValidateManifestSemanticVersion(manifest.Version); err != nil {
		return gerror.Wrapf(err, "plugin version is invalid: %s", fileLabel)
	}
	if err := ValidateManifestMenus(manifest); err != nil {
		return gerror.Wrapf(err, "plugin menu metadata is invalid: %s", fileLabel)
	}
	if err := plugintypes.ValidateDependencySpec(manifest.ID, manifest.Dependencies); err != nil {
		return gerror.Wrapf(err, "plugin dependency metadata is invalid: %s", fileLabel)
	}
	if err := normalizeManifestHostServices(manifest); err != nil {
		return gerror.Wrapf(err, "plugin host service metadata is invalid: %s", fileLabel)
	}
	if err := ValidateOwnerHostServiceDependencies(manifest); err != nil {
		return gerror.Wrapf(err, "plugin owner host service dependency metadata is invalid: %s", fileLabel)
	}
	if plugintypes.NormalizeType(manifest.Type) == pluginv1.PluginTypeSource {
		if manifest.SourcePlugin != nil && strings.TrimSpace(manifest.SourcePlugin.ID()) != "" {
			registeredPluginID := strings.TrimSpace(manifest.SourcePlugin.ID())
			if err := plugintypes.ValidatePluginID(registeredPluginID); err != nil {
				return gerror.Wrapf(err, "source plugin registered ID is invalid: %s", registeredPluginID)
			}
			if manifest.ID != registeredPluginID {
				return gerror.Newf("source plugin embedded manifest ID does not match registered plugin ID: %s != %s", manifest.ID, registeredPluginID)
			}
		}
		if plugintypes.NormalizeDistribution(manifest.Distribution) == pluginv1.PluginDistributionBuiltin && manifest.SourcePlugin == nil {
			return gerror.Newf("builtin source plugin must be registered in source plugin registry: %s", manifest.ID)
		}
		goModPath := filepath.Join(rootDir, "go.mod")
		if !HasSourcePluginEmbeddedFiles(manifest) && !gfile.Exists(goModPath) {
			return gerror.Newf("source plugin directory is missing go.mod: %s", rootDir)
		}
		backendEntryPath := filepath.Join(rootDir, "backend", "plugin.go")
		if !HasSourcePluginEmbeddedFiles(manifest) && !gfile.Exists(backendEntryPath) {
			return gerror.Newf("source plugin directory is missing backend/plugin.go: %s", rootDir)
		}
	} else if plugintypes.NormalizeDistribution(manifest.Distribution) == pluginv1.PluginDistributionBuiltin {
		return gerror.Newf("builtin plugin distribution only supports source plugins: %s", fileLabel)
	} else if err := validateRuntimeArtifact(manifest, rootDir); err != nil {
		// Tolerate a missing artifact during local development/scan so dynamic
		// plugins remain visible even before make wasm is run.
		if !isMissingArtifactError(err) {
			return gerror.Wrapf(err, "dynamic plugin artifact validation failed: %s", filePath)
		}
	}
	if err := ValidatePublicAssets(manifest, rootDir); err != nil {
		return gerror.Wrapf(err, "plugin public asset metadata is invalid: %s", fileLabel)
	}
	if embeddedFiles := GetSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		if err := resourcefs.ValidateSQLPathsFromFS(embeddedFiles, s.ListInstallSQLPaths(manifest), false); err != nil {
			return gerror.Wrapf(err, "plugin manifest install SQL constraint is invalid: %s", fileLabel)
		}
		if err := resourcefs.ValidateSQLPathsFromFS(embeddedFiles, s.ListUninstallSQLPaths(manifest), true); err != nil {
			return gerror.Wrapf(err, "plugin manifest uninstall SQL constraint is invalid: %s", fileLabel)
		}
		if err := resourcefs.ValidateVuePathsFromFS(embeddedFiles, s.ListFrontendPagePaths(manifest), "frontend/pages/"); err != nil {
			return gerror.Wrapf(err, "plugin manifest frontend page constraint is invalid: %s", fileLabel)
		}
		if err := resourcefs.ValidateVuePathsFromFS(embeddedFiles, s.ListFrontendSlotPaths(manifest), "frontend/slots/"); err != nil {
			return gerror.Wrapf(err, "plugin manifest frontend slot constraint is invalid: %s", fileLabel)
		}
		return nil
	}
	if err := resourcefs.ValidateSQLPaths(rootDir, s.ListInstallSQLPaths(manifest), false); err != nil {
		return gerror.Wrapf(err, "plugin manifest install SQL constraint is invalid: %s", fileLabel)
	}
	if err := resourcefs.ValidateSQLPaths(rootDir, s.ListUninstallSQLPaths(manifest), true); err != nil {
		return gerror.Wrapf(err, "plugin manifest uninstall SQL constraint is invalid: %s", fileLabel)
	}
	if err := resourcefs.ValidateVuePaths(rootDir, s.ListFrontendPagePaths(manifest), "frontend/pages/"); err != nil {
		return gerror.Wrapf(err, "plugin manifest frontend page constraint is invalid: %s", fileLabel)
	}
	if err := resourcefs.ValidateVuePaths(rootDir, s.ListFrontendSlotPaths(manifest), "frontend/slots/"); err != nil {
		return gerror.Wrapf(err, "plugin manifest frontend slot constraint is invalid: %s", fileLabel)
	}
	return nil
}

// ValidateUploadedRuntimeManifest validates the identity fields extracted from a WASM artifact manifest.
func (s *serviceImpl) ValidateUploadedRuntimeManifest(manifest *Manifest) error {
	if manifest == nil {
		return gerror.New("dynamic plugin manifest cannot be nil")
	}
	manifest.Type = plugintypes.NormalizeType(manifest.Type).String()
	if manifest.Type != pluginv1.PluginTypeDynamic.String() {
		return gerror.New("dynamic plugin type must be dynamic")
	}
	manifest.Distribution = plugintypes.NormalizeDistribution(manifest.Distribution).String()
	if !plugintypes.IsSupportedDistribution(manifest.Distribution) {
		return gerror.New("dynamic plugin distribution only supports managed or builtin")
	}
	if plugintypes.NormalizeDistribution(manifest.Distribution) == pluginv1.PluginDistributionBuiltin {
		return gerror.New("dynamic plugin cannot declare builtin distribution")
	}
	if err := normalizeManifestTenantGovernance(manifest); err != nil {
		return err
	}
	manifest.ID = strings.TrimSpace(manifest.ID)
	if err := plugintypes.ValidatePluginID(manifest.ID); err != nil {
		return gerror.Wrap(err, "dynamic plugin ID is invalid")
	}
	if manifest.Name == "" {
		return gerror.New("dynamic plugin name cannot be empty")
	}
	if err := plugintypes.ValidateManifestSemanticVersion(manifest.Version); err != nil {
		return err
	}
	if err := plugintypes.ValidateDependencySpec(manifest.ID, manifest.Dependencies); err != nil {
		return err
	}
	if err := normalizeManifestHostServices(manifest); err != nil {
		return err
	}
	if err := ValidateOwnerHostServiceDependencies(manifest); err != nil {
		return err
	}
	if err := ValidatePublicAssets(manifest, manifest.RootDir); err != nil {
		return err
	}
	return ValidateManifestMenus(manifest)
}

// ValidatePublicAssets validates and normalizes plugin-declared public asset
// directories. Source plugins must point at real embedded or filesystem
// directories. Dynamic plugins must point at prefixes present in runtime
// frontend assets; frontend assets are not public unless a declaration matches.
func ValidatePublicAssets(manifest *Manifest, rootDir string) error {
	if manifest == nil || len(manifest.PublicAssets) == 0 {
		return nil
	}

	mounts := make([]string, 0, len(manifest.PublicAssets))
	for index, spec := range manifest.PublicAssets {
		if spec == nil {
			return gerror.Newf("public_assets declaration %d cannot be nil", index+1)
		}
		source, err := normalizePublicAssetSource(spec.Source)
		if err != nil {
			return gerror.Wrapf(err, "public_assets declaration %d source is invalid", index+1)
		}
		mount, err := normalizePublicAssetMount(spec.Mount)
		if err != nil {
			return gerror.Wrapf(err, "public_assets declaration %d mount is invalid", index+1)
		}
		indexFile, err := normalizePublicAssetIndex(spec.Index)
		if err != nil {
			return gerror.Wrapf(err, "public_assets declaration %d index is invalid", index+1)
		}
		if publicAssetMountOverlaps(mounts, mount) {
			return gerror.Newf("public_assets mount overlaps another declaration: %s", mount)
		}
		if err = validatePublicAssetSourceExists(manifest, rootDir, source); err != nil {
			return gerror.Wrapf(err, "public_assets source does not exist: %s", source)
		}
		spec.Source = source
		spec.Mount = mount
		spec.Index = indexFile
		mounts = append(mounts, mount)
	}
	return nil
}

// normalizePublicAssetSource converts one declaration source into a safe
// plugin-relative path. The declaration is treated as the plugin author's
// explicit publication boundary; containment and existence checks are enforced
// later against the plugin backing store.
func normalizePublicAssetSource(value string) (string, error) {
	normalized, err := normalizePublicAssetRelativePath(value, false)
	if err != nil {
		return "", err
	}
	return normalized, nil
}

// normalizePublicAssetMount converts one URL mount into a relative path under
// the plugin version root. Empty and "/" both mean the version root.
func normalizePublicAssetMount(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "/" {
		return "", nil
	}
	return normalizePublicAssetRelativePath(trimmed, true)
}

// normalizePublicAssetIndex converts one directory index value into a safe
// declaration-relative file path. Empty values keep the historical index.html
// default.
func normalizePublicAssetIndex(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return defaultPublicAssetIndex, nil
	}
	if strings.HasSuffix(strings.ReplaceAll(trimmed, "\\", "/"), "/") {
		return "", gerror.Newf("index must be a file name under the declared source root: %s", value)
	}
	normalized, err := normalizePublicAssetRelativePath(trimmed, false)
	if err != nil {
		return "", err
	}
	if path.Base(normalized) != normalized {
		return "", gerror.Newf("index must be a file name under the declared source root: %s", value)
	}
	return normalized, nil
}

// normalizePublicAssetRelativePath is the shared strict normalizer for source
// and mount paths. It rejects URLs, absolute paths, traversal, and wildcards.
func normalizePublicAssetRelativePath(value string, allowRoot bool) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		if allowRoot {
			return "", nil
		}
		return "", gerror.New("path cannot be empty")
	}
	if strings.Contains(trimmed, "://") {
		return "", gerror.Newf("path cannot be a URL: %s", value)
	}
	if parsed, err := url.Parse(trimmed); err == nil && parsed != nil && parsed.Scheme != "" {
		return "", gerror.Newf("path cannot be a URL: %s", value)
	}
	if filepath.IsAbs(trimmed) || strings.HasPrefix(trimmed, "/") {
		if allowRoot && trimmed == "/" {
			return "", nil
		}
		return "", gerror.Newf("path must be relative: %s", value)
	}
	if strings.ContainsAny(trimmed, "*?") || strings.Contains(trimmed, "#") {
		return "", gerror.Newf("path contains unsupported characters: %s", value)
	}
	normalized := path.Clean(strings.ReplaceAll(trimmed, "\\", "/"))
	normalized = strings.TrimPrefix(normalized, "./")
	if normalized == "." {
		if allowRoot {
			return "", nil
		}
		return "", gerror.New("path cannot be empty")
	}
	if normalized == ".." || strings.HasPrefix(normalized, "../") {
		return "", gerror.Newf("path escapes plugin root: %s", value)
	}
	return strings.Trim(normalized, "/"), nil
}

// publicAssetMountOverlaps reports duplicate or parent-child mount collisions.
func publicAssetMountOverlaps(existing []string, candidate string) bool {
	candidate = strings.Trim(candidate, "/")
	for _, item := range existing {
		current := strings.Trim(item, "/")
		if current == candidate {
			return true
		}
		if current == "" || candidate == "" {
			return true
		}
		if strings.HasPrefix(current, candidate+"/") || strings.HasPrefix(candidate, current+"/") {
			return true
		}
	}
	return false
}

// validatePublicAssetSourceExists checks declaration sources against the
// relevant plugin backing store.
func validatePublicAssetSourceExists(manifest *Manifest, rootDir string, source string) error {
	if manifest == nil {
		return gerror.New("plugin manifest cannot be nil")
	}
	if manifest.RuntimeArtifact != nil {
		if runtimePublicAssetSourceExists(manifest.RuntimeArtifact.FrontendAssets, source) {
			return nil
		}
		return gerror.Newf("dynamic runtime frontend asset prefix does not exist: %s", source)
	}
	if embeddedFiles := GetSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		if err := resourcefs.ValidateNoSymlinkPathFromFS(embeddedFiles, source); err != nil {
			return err
		}
		stat, err := fs.Stat(embeddedFiles, source)
		if err != nil {
			return err
		}
		if !stat.IsDir() {
			return gerror.Newf("public_assets source must be a directory: %s", source)
		}
		return nil
	}
	resolvedPath, err := resourcefs.ResolveResourcePath(rootDir, source)
	if err != nil {
		return err
	}
	if !gfile.IsDir(resolvedPath) {
		return gerror.Newf("public_assets source must be a directory: %s", source)
	}
	return nil
}

// runtimePublicAssetSourceExists reports whether the dynamic artifact contains
// at least one frontend asset under source.
func runtimePublicAssetSourceExists(assets []*ArtifactFrontendAsset, source string) bool {
	normalizedSource := strings.Trim(normalizeRuntimeAssetPath(source), "/")
	for _, asset := range assets {
		if asset == nil {
			continue
		}
		assetPath := normalizeRuntimeAssetPath(asset.Path)
		if assetPath == "" {
			continue
		}
		if normalizedSource == "" || assetPath == normalizedSource || strings.HasPrefix(assetPath, normalizedSource+"/") {
			return true
		}
	}
	return false
}

// normalizeRuntimeAssetPath normalizes dynamic artifact frontend asset keys.
func normalizeRuntimeAssetPath(value string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	normalized = strings.TrimPrefix(normalized, "/")
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = path.Clean(normalized)
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") {
		return ""
	}
	return strings.Trim(normalized, "/")
}

// hydrateManifestTenantGovernanceFromFile fills governance fields from the
// authoring manifest when tests or callers pass a partial in-memory object.
func (s *serviceImpl) hydrateManifestTenantGovernanceFromFile(manifest *Manifest, filePath string) error {
	if manifest == nil || manifest.RuntimeArtifact != nil {
		return nil
	}
	filePath = strings.TrimSpace(filePath)
	if filePath == "" || strings.Contains(filePath, "://") || !gfile.Exists(filePath) {
		return nil
	}
	fileManifest := &Manifest{}
	if err := s.LoadManifestFromYAML(filePath, fileManifest); err != nil {
		return err
	}
	if strings.TrimSpace(manifest.ScopeNature) == "" {
		manifest.ScopeNature = fileManifest.ScopeNature
	}
	if manifest.SupportsMultiTenant == nil {
		manifest.SupportsMultiTenant = fileManifest.SupportsMultiTenant
	}
	if strings.TrimSpace(manifest.DefaultInstallMode) == "" {
		manifest.DefaultInstallMode = fileManifest.DefaultInstallMode
	}
	if len(manifest.PublicAssets) == 0 && len(fileManifest.PublicAssets) > 0 {
		manifest.PublicAssets = ClonePublicAssetSpecs(fileManifest.PublicAssets)
	}
	return nil
}

// normalizeManifestTenantGovernance validates and normalizes the tenant
// governance fields carried by plugin.yaml.
func normalizeManifestTenantGovernance(manifest *Manifest) error {
	if manifest == nil {
		return nil
	}
	if manifest.RuntimeArtifact != nil && manifest.RuntimeArtifact.Manifest != nil {
		manifest.ScopeNature = strings.TrimSpace(manifest.RuntimeArtifact.Manifest.ScopeNature)
		manifest.SupportsMultiTenant = manifest.RuntimeArtifact.Manifest.SupportsMultiTenant
		manifest.DefaultInstallMode = strings.TrimSpace(manifest.RuntimeArtifact.Manifest.DefaultInstallMode)
	}
	scope := strings.TrimSpace(manifest.ScopeNature)
	if scope == "" {
		scope = pluginv1.ScopeNatureTenantAware.String()
	} else if !plugintypes.IsSupportedScopeNature(scope) {
		return gerror.Newf("scope_nature only supports platform_only/tenant_aware: %s", manifest.ScopeNature)
	}
	manifest.ScopeNature = plugintypes.NormalizeScopeNature(scope).String()

	if manifest.SupportsMultiTenant == nil {
		return gerror.New("supports_multi_tenant is required")
	}
	if manifest.ScopeNature == pluginv1.ScopeNaturePlatformOnly.String() && *manifest.SupportsMultiTenant {
		return gerror.New("supports_multi_tenant cannot be true when scope_nature is platform_only")
	}

	mode := strings.TrimSpace(manifest.DefaultInstallMode)
	if manifest.ScopeNature == pluginv1.ScopeNaturePlatformOnly.String() {
		manifest.DefaultInstallMode = pluginv1.InstallModeGlobal.String()
		return nil
	}
	if !*manifest.SupportsMultiTenant {
		manifest.DefaultInstallMode = pluginv1.InstallModeGlobal.String()
		return nil
	}
	if mode == "" {
		mode = pluginv1.InstallModeTenantScoped.String()
	} else if !plugintypes.IsSupportedInstallMode(mode) {
		return gerror.Newf("default_install_mode only supports global/tenant_scoped: %s", manifest.DefaultInstallMode)
	}
	manifest.DefaultInstallMode = plugintypes.NormalizeInstallMode(mode).String()
	return nil
}

// ValidateManifestMenus validates the structural constraints of all menu declarations in a manifest.
// It normalizes menu field values in-place and returns the first validation error encountered.
func ValidateManifestMenus(manifest *Manifest) error {
	if manifest == nil || len(manifest.Menus) == 0 {
		return nil
	}

	declaredKeys := make(map[string]struct{}, len(manifest.Menus))
	for index, spec := range manifest.Menus {
		if spec == nil {
			return gerror.Newf("menu declaration %d cannot be nil", index+1)
		}

		spec.Key = strings.TrimSpace(spec.Key)
		spec.ParentKey = strings.TrimSpace(spec.ParentKey)
		spec.Name = strings.TrimSpace(spec.Name)
		spec.Path = strings.TrimSpace(spec.Path)
		spec.Component = strings.TrimSpace(spec.Component)
		spec.Perms = strings.TrimSpace(spec.Perms)
		spec.Icon = strings.TrimSpace(spec.Icon)
		spec.Type = NormalizeMenuType(spec.Type).String()
		spec.QueryParam = strings.TrimSpace(spec.QueryParam)
		spec.Remark = strings.TrimSpace(spec.Remark)

		if spec.Key == "" {
			return gerror.Newf("menu declaration %d is missing key", index+1)
		}
		if spec.Name == "" {
			return gerror.Newf("plugin menu is missing name: %s", spec.Key)
		}
		if !IsSupportedMenuType(NormalizeMenuType(spec.Type)) {
			return gerror.Newf("plugin menu type only supports D/M/B: %s", spec.Key)
		}
		if spec.ParentKey == spec.Key {
			return gerror.Newf("plugin menu parent_key cannot point to itself: %s", spec.Key)
		}
		pluginID := parsePluginIDFromMenuKey(spec.Key)
		if pluginID == "" || pluginID != manifest.ID {
			return gerror.Newf("plugin menu key must use current plugin prefix plugin:%s:* : %s", manifest.ID, spec.Key)
		}
		if _, ok := declaredKeys[spec.Key]; ok {
			return gerror.Newf("plugin menu key is duplicated: %s", spec.Key)
		}
		declaredKeys[spec.Key] = struct{}{}

		if _, err := normalizeMenuFlag(spec.Visible, menuDefaultVisible); err != nil {
			return gerror.Wrapf(err, "plugin menu visible is invalid: %s", spec.Key)
		}
		if _, err := normalizeMenuFlag(spec.Status, menuDefaultStatus); err != nil {
			return gerror.Wrapf(err, "plugin menu status is invalid: %s", spec.Key)
		}
		if _, err := normalizeMenuFlag(spec.IsFrame, menuDefaultIsFrame); err != nil {
			return gerror.Wrapf(err, "plugin menu is_frame is invalid: %s", spec.Key)
		}
		if _, err := normalizeMenuFlag(spec.IsCache, menuDefaultIsCache); err != nil {
			return gerror.Wrapf(err, "plugin menu is_cache is invalid: %s", spec.Key)
		}
		if _, err := buildMenuQueryParam(spec); err != nil {
			return gerror.Wrapf(err, "plugin menu query is invalid: %s", spec.Key)
		}
	}

	for _, spec := range manifest.Menus {
		if spec == nil || spec.ParentKey == "" {
			continue
		}
		if parsePluginIDFromMenuKey(spec.ParentKey) != manifest.ID {
			continue
		}
		if _, ok := declaredKeys[spec.ParentKey]; !ok {
			return gerror.Newf("plugin menu references undeclared parent_key: %s -> %s", spec.Key, spec.ParentKey)
		}
	}

	return nil
}

// normalizeMenuFlag validates and returns a plugin menu integer flag (0 or 1).
func normalizeMenuFlag(value *int, defaultValue int) (int, error) {
	if value == nil {
		return defaultValue, nil
	}
	if *value != 0 && *value != 1 {
		return 0, gerror.New("only 0 or 1 is supported")
	}
	return *value, nil
}

// buildMenuQueryParam serializes the query map or query_param field of a menu spec.
func buildMenuQueryParam(spec *MenuSpec) (string, error) {
	if spec == nil {
		return "", nil
	}
	if strings.TrimSpace(spec.QueryParam) != "" {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(spec.QueryParam), &payload); err != nil {
			return "", err
		}
		if len(payload) == 0 {
			return "", nil
		}
		content, err := json.Marshal(payload)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}
	if len(spec.Query) == 0 {
		return "", nil
	}
	content, err := json.Marshal(spec.Query)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// parsePluginIDFromMenuKey extracts the plugin ID portion from a "plugin:<id>:*" menu key.
func parsePluginIDFromMenuKey(key string) string {
	key = strings.TrimSpace(key)
	if !strings.HasPrefix(key, plugintypes.MenuKeyPrefix) {
		return ""
	}
	withoutPrefix := key[len(plugintypes.MenuKeyPrefix):]
	if idx := strings.Index(withoutPrefix, ":"); idx > 0 {
		return withoutPrefix[:idx]
	}
	return ""
}

// sha256sum is an internal helper for generating SHA-256 checksums.
func sha256sum(data []byte) [32]byte {
	return sha256.Sum256(data)
}
