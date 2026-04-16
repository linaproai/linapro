// This file validates plugin manifests for structural correctness and
// validates uploaded runtime manifests from WASM artifacts.

package catalog

import (
	"crypto/sha256"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/pkg/pluginfs"
)

const (
	menuDefaultVisible = 1
	menuDefaultStatus  = 1
	menuDefaultIsFrame = 0
	menuDefaultIsCache = 0
)

// ValidateManifest validates required fields and structural constraints in a plugin manifest.
// For source plugins it additionally checks for go.mod and backend/plugin.go.
// For dynamic plugins it optionally validates the runtime artifact via ArtifactParser.
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
		return gerror.Newf("插件清单缺少id: %s", fileLabel)
	}
	if manifest.Name == "" {
		return gerror.Newf("插件清单缺少name: %s", fileLabel)
	}
	if manifest.Version == "" {
		return gerror.Newf("插件清单缺少version: %s", fileLabel)
	}
	if manifest.Type == "" {
		manifest.Type = TypeSource.String()
	} else {
		manifest.Type = NormalizeType(manifest.Type).String()
	}
	if !IsSupportedType(manifest.Type) {
		return gerror.Newf("插件类型仅支持 source/dynamic: %s", fileLabel)
	}
	if !ManifestIDPattern.MatchString(manifest.ID) {
		return gerror.Newf("插件ID需使用kebab-case风格: %s", manifest.ID)
	}
	if err := ValidateManifestSemanticVersion(manifest.Version); err != nil {
		return gerror.Wrapf(err, "插件版本不合法: %s", fileLabel)
	}
	if err := ValidateManifestMenus(manifest); err != nil {
		return gerror.Wrapf(err, "插件菜单元数据不合法: %s", fileLabel)
	}
	if NormalizeType(manifest.Type) == TypeSource {
		if manifest.SourcePlugin != nil && strings.TrimSpace(manifest.SourcePlugin.ID) != "" && manifest.ID != manifest.SourcePlugin.ID {
			return gerror.Newf("源码插件嵌入清单 ID 与注册插件 ID 不一致: %s != %s", manifest.ID, manifest.SourcePlugin.ID)
		}
		goModPath := filepath.Join(rootDir, "go.mod")
		if !HasSourcePluginEmbeddedFiles(manifest) && !gfile.Exists(goModPath) {
			return gerror.Newf("源码插件目录缺少 go.mod: %s", rootDir)
		}
		backendEntryPath := filepath.Join(rootDir, "backend", "plugin.go")
		if !HasSourcePluginEmbeddedFiles(manifest) && !gfile.Exists(backendEntryPath) {
			return gerror.Newf("源码插件目录缺少 backend/plugin.go: %s", rootDir)
		}
	} else if s.artifactParser != nil {
		if err := s.artifactParser.ValidateRuntimeArtifact(manifest, rootDir); err != nil {
			// Tolerate a missing artifact during local development/scan so dynamic
			// plugins remain visible even before make wasm is run.
			if !strings.Contains(err.Error(), "缺少") {
				return gerror.Wrapf(err, "动态插件产物校验失败: %s", filePath)
			}
		}
	}
	if embeddedFiles := GetSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		if err := pluginfs.ValidateSQLPathsFromFS(embeddedFiles, s.ListInstallSQLPaths(manifest), false); err != nil {
			return gerror.Wrapf(err, "插件清单 install SQL 约束不合法: %s", fileLabel)
		}
		if err := pluginfs.ValidateSQLPathsFromFS(embeddedFiles, s.ListUninstallSQLPaths(manifest), true); err != nil {
			return gerror.Wrapf(err, "插件清单 uninstall SQL 约束不合法: %s", fileLabel)
		}
		if err := pluginfs.ValidateVuePathsFromFS(embeddedFiles, s.ListFrontendPagePaths(manifest), "frontend/pages/"); err != nil {
			return gerror.Wrapf(err, "插件清单 frontend page 约束不合法: %s", fileLabel)
		}
		if err := pluginfs.ValidateVuePathsFromFS(embeddedFiles, s.ListFrontendSlotPaths(manifest), "frontend/slots/"); err != nil {
			return gerror.Wrapf(err, "插件清单 frontend slot 约束不合法: %s", fileLabel)
		}
		return nil
	}
	if err := pluginfs.ValidateSQLPaths(rootDir, s.ListInstallSQLPaths(manifest), false); err != nil {
		return gerror.Wrapf(err, "插件清单 install SQL 约束不合法: %s", fileLabel)
	}
	if err := pluginfs.ValidateSQLPaths(rootDir, s.ListUninstallSQLPaths(manifest), true); err != nil {
		return gerror.Wrapf(err, "插件清单 uninstall SQL 约束不合法: %s", fileLabel)
	}
	if err := pluginfs.ValidateVuePaths(rootDir, s.ListFrontendPagePaths(manifest), "frontend/pages/"); err != nil {
		return gerror.Wrapf(err, "插件清单 frontend page 约束不合法: %s", fileLabel)
	}
	if err := pluginfs.ValidateVuePaths(rootDir, s.ListFrontendSlotPaths(manifest), "frontend/slots/"); err != nil {
		return gerror.Wrapf(err, "插件清单 frontend slot 约束不合法: %s", fileLabel)
	}
	return nil
}

// ValidateUploadedRuntimeManifest validates the identity fields extracted from a WASM artifact manifest.
func (s *serviceImpl) ValidateUploadedRuntimeManifest(manifest *Manifest) error {
	if manifest == nil {
		return gerror.New("动态插件清单不能为空")
	}
	manifest.Type = NormalizeType(manifest.Type).String()
	if manifest.Type != TypeDynamic.String() {
		return gerror.New("动态插件类型必须是 dynamic")
	}
	if manifest.ID == "" || !ManifestIDPattern.MatchString(manifest.ID) {
		return gerror.New("动态插件 ID 非法")
	}
	if manifest.Name == "" {
		return gerror.New("动态插件名称不能为空")
	}
	if err := ValidateManifestSemanticVersion(manifest.Version); err != nil {
		return err
	}
	return ValidateManifestMenus(manifest)
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
			return gerror.Newf("第 %d 个菜单声明不能为空", index+1)
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
			return gerror.Newf("第 %d 个菜单声明缺少 key", index+1)
		}
		if spec.Name == "" {
			return gerror.Newf("插件菜单缺少 name: %s", spec.Key)
		}
		if !IsSupportedMenuType(NormalizeMenuType(spec.Type)) {
			return gerror.Newf("插件菜单类型仅支持 D/M/B: %s", spec.Key)
		}
		if spec.ParentKey == spec.Key {
			return gerror.Newf("插件菜单 parent_key 不能指向自己: %s", spec.Key)
		}
		pluginID := parsePluginIDFromMenuKey(spec.Key)
		if pluginID == "" || pluginID != manifest.ID {
			return gerror.Newf("插件菜单 key 必须使用当前插件前缀 plugin:%s:* : %s", manifest.ID, spec.Key)
		}
		if parentPluginID := parsePluginIDFromMenuKey(spec.ParentKey); parentPluginID != "" && parentPluginID != manifest.ID {
			return gerror.Newf("插件菜单 parent_key 不允许引用其他插件菜单: %s -> %s", spec.Key, spec.ParentKey)
		}
		if _, ok := declaredKeys[spec.Key]; ok {
			return gerror.Newf("插件菜单 key 重复: %s", spec.Key)
		}
		declaredKeys[spec.Key] = struct{}{}

		if _, err := normalizeMenuFlag(spec.Visible, menuDefaultVisible); err != nil {
			return gerror.Wrapf(err, "插件菜单 visible 非法: %s", spec.Key)
		}
		if _, err := normalizeMenuFlag(spec.Status, menuDefaultStatus); err != nil {
			return gerror.Wrapf(err, "插件菜单 status 非法: %s", spec.Key)
		}
		if _, err := normalizeMenuFlag(spec.IsFrame, menuDefaultIsFrame); err != nil {
			return gerror.Wrapf(err, "插件菜单 is_frame 非法: %s", spec.Key)
		}
		if _, err := normalizeMenuFlag(spec.IsCache, menuDefaultIsCache); err != nil {
			return gerror.Wrapf(err, "插件菜单 is_cache 非法: %s", spec.Key)
		}
		if _, err := buildMenuQueryParam(spec); err != nil {
			return gerror.Wrapf(err, "插件菜单 query 非法: %s", spec.Key)
		}
	}

	for _, spec := range manifest.Menus {
		if spec == nil || spec.ParentKey == "" {
			continue
		}
		parentPluginID := parsePluginIDFromMenuKey(spec.ParentKey)
		if parentPluginID != manifest.ID {
			continue
		}
		if _, ok := declaredKeys[spec.ParentKey]; !ok {
			return gerror.Newf("插件菜单引用了未声明的 parent_key: %s -> %s", spec.Key, spec.ParentKey)
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
		return 0, gerror.New("仅支持 0 或 1")
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
	if !strings.HasPrefix(key, MenuKeyPrefix) {
		return ""
	}
	withoutPrefix := key[len(MenuKeyPrefix):]
	if idx := strings.Index(withoutPrefix, ":"); idx > 0 {
		return withoutPrefix[:idx]
	}
	return ""
}

// sha256sum is an internal helper for generating SHA-256 checksums.
func sha256sum(data []byte) [32]byte {
	return sha256.Sum256(data)
}
