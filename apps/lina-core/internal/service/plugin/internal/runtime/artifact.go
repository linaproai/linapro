// This file implements the catalog.ArtifactParser interface: reading and validating
// WASM artifact files, extracting embedded custom sections, and building review-friendly
// checksums and remarks for plugin governance.

package runtime

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/pluginfs"
)

// DynamicKindWasm is the only supported runtime artifact kind.
const DynamicKindWasm = "wasm"

// dynamicKindWasm is the package-private alias kept for internal references.
const dynamicKindWasm = DynamicKindWasm

// IsMissingArtifactError reports whether err signals a missing runtime artifact.
func IsMissingArtifactError(err error) bool {
	return isMissingArtifactError(err)
}

// BuildArtifactFileName returns the canonical wasm filename for a plugin ID.
func BuildArtifactFileName(pluginID string) string {
	return buildArtifactFileName(pluginID)
}

// BuildArtifactRelativePath returns the canonical relative path for a plugin's wasm artifact.
func BuildArtifactRelativePath(pluginID string) string {
	return buildArtifactRelativePath(pluginID)
}

// artifactMissingError marks the "wasm not generated yet" state so that discovery
// can keep dynamic plugins visible while lifecycle actions stay strict.
type artifactMissingError struct {
	rootDir      string
	relativePath string
}

// Error returns the actionable missing-artifact message used by lifecycle guards.
func (e *artifactMissingError) Error() string {
	return fmt.Sprintf("动态插件目录缺少 %s: %s", e.relativePath, e.rootDir)
}

// buildArtifactFileName returns the canonical wasm filename for one plugin ID.
func buildArtifactFileName(pluginID string) string {
	normalizedID := strings.TrimSpace(pluginID)
	if normalizedID == "" {
		return "plugin.wasm"
	}
	return normalizedID + ".wasm"
}

// buildArtifactRelativePath returns the canonical relative runtime artifact path.
func buildArtifactRelativePath(pluginID string) string {
	return filepath.Join("runtime", buildArtifactFileName(pluginID))
}

// resolveArtifactPath resolves the current or legacy runtime artifact path
// inside a plugin root and reports a typed missing-artifact error otherwise.
func resolveArtifactPath(rootDir string, pluginID string) (string, error) {
	relativePath := filepath.ToSlash(buildArtifactRelativePath(pluginID))
	candidatePath := filepath.Join(rootDir, buildArtifactRelativePath(pluginID))
	if gfile.Exists(candidatePath) {
		return candidatePath, nil
	}

	legacyPath := filepath.Join(rootDir, "runtime", "plugin.wasm")
	if gfile.Exists(legacyPath) {
		return legacyPath, nil
	}

	return candidatePath, &artifactMissingError{
		rootDir:      rootDir,
		relativePath: relativePath,
	}
}

// isMissingArtifactError reports whether the error indicates a missing wasm artifact.
func isMissingArtifactError(err error) bool {
	var target *artifactMissingError
	return errors.As(err, &target)
}

// ParseRuntimeWasmArtifact reads one WASM artifact file and extracts all embedded custom sections.
// It implements the catalog.ArtifactParser interface.
func (s *serviceImpl) ParseRuntimeWasmArtifact(filePath string) (*catalog.ArtifactSpec, error) {
	content := gfile.GetBytes(filePath)
	if len(content) == 0 {
		return nil, gerror.Newf("动态插件产物为空: %s", filePath)
	}
	return s.ParseRuntimeWasmArtifactContent(filePath, content)
}

// ParseRuntimeWasmArtifactContent parses one WASM artifact from an in-memory byte slice.
// It implements the catalog.ArtifactParser interface.
func (s *serviceImpl) ParseRuntimeWasmArtifactContent(filePath string, content []byte) (*catalog.ArtifactSpec, error) {
	sections, err := parseWasmCustomSections(content)
	if err != nil {
		return nil, gerror.Wrapf(err, "解析动态插件产物失败: %s", filePath)
	}

	manifestSection, ok := sections[pluginbridge.WasmSectionManifest]
	if !ok {
		return nil, gerror.Newf("动态插件缺少自定义节 %s: %s", pluginbridge.WasmSectionManifest, filePath)
	}
	runtimeSection, ok := sections[pluginbridge.WasmSectionRuntime]
	if !ok {
		runtimeSection, ok = sections[pluginbridge.WasmSectionLegacyRuntime]
	}
	if !ok {
		return nil, gerror.Newf("动态插件缺少自定义节 %s: %s", pluginbridge.WasmSectionRuntime, filePath)
	}

	embeddedManifest := &catalog.ArtifactManifest{}
	if err = unmarshalRuntimeArtifactSection(manifestSection, embeddedManifest); err != nil {
		return nil, gerror.Wrapf(err, "解析动态插件嵌入清单失败: %s", filePath)
	}
	if strings.TrimSpace(embeddedManifest.ID) == "" ||
		strings.TrimSpace(embeddedManifest.Name) == "" ||
		strings.TrimSpace(embeddedManifest.Version) == "" ||
		strings.TrimSpace(embeddedManifest.Type) == "" {
		return nil, gerror.Newf("动态插件嵌入清单缺少必填字段: %s", filePath)
	}

	runtimeMetadata := &pluginbridge.RuntimeArtifactMetadata{}
	if err = unmarshalRuntimeArtifactSection(runtimeSection, runtimeMetadata); err != nil {
		return nil, gerror.Wrapf(err, "解析动态插件运行时元数据失败: %s", filePath)
	}

	frontendAssets, err := parseRuntimeArtifactFrontendAssets(filePath, sections, pluginbridge.WasmSectionFrontendAssets)
	if err != nil {
		return nil, err
	}
	runtimeI18NAssets, err := parseRuntimeArtifactLocaleJSONAssets(filePath, sections, pluginbridge.WasmSectionI18NAssets)
	if err != nil {
		return nil, err
	}
	apiDocI18NAssets, err := parseRuntimeArtifactLocaleJSONAssets(filePath, sections, pluginbridge.WasmSectionAPIDocI18NAssets)
	if err != nil {
		return nil, err
	}
	installSQLAssets, err := parseRuntimeArtifactSQLAssets(filePath, sections, pluginbridge.WasmSectionInstallSQL)
	if err != nil {
		return nil, err
	}
	uninstallSQLAssets, err := parseRuntimeArtifactSQLAssets(filePath, sections, pluginbridge.WasmSectionUninstallSQL)
	if err != nil {
		return nil, err
	}
	hookSpecs, err := parseRuntimeArtifactHookSpecs(filePath, embeddedManifest.ID, sections)
	if err != nil {
		return nil, err
	}
	resourceSpecs, err := parseRuntimeArtifactResourceSpecs(filePath, embeddedManifest.ID, sections)
	if err != nil {
		return nil, err
	}
	routeContracts, err := parseRuntimeArtifactRouteContracts(filePath, embeddedManifest.ID, sections)
	if err != nil {
		return nil, err
	}
	bridgeSpec, err := parseRuntimeArtifactBridgeSpec(filePath, sections)
	if err != nil {
		return nil, err
	}
	if err = rejectDeprecatedRuntimeArtifactCapabilities(filePath, sections); err != nil {
		return nil, err
	}
	hostServices, err := parseRuntimeArtifactHostServices(filePath, sections)
	if err != nil {
		return nil, err
	}
	// Runtime capability checks remain in place, but the capability set is now
	// derived from the single hostServices snapshot instead of a second embedded section.
	capabilities := pluginbridge.CapabilitiesFromHostServices(hostServices)

	runtimeKind := strings.TrimSpace(strings.ToLower(runtimeMetadata.RuntimeKind))
	if runtimeKind == "" {
		runtimeKind = dynamicKindWasm
	}
	if runtimeKind != dynamicKindWasm {
		return nil, gerror.Newf("动态插件产物类型仅支持 wasm: %s", runtimeKind)
	}

	abiVersion := strings.TrimSpace(strings.ToLower(runtimeMetadata.ABIVersion))
	if abiVersion == "" {
		return nil, gerror.Newf("动态插件缺少 ABI 版本: %s", filePath)
	}
	if abiVersion != pluginbridge.SupportedABIVersion {
		return nil, gerror.Newf("动态插件 ABI 版本不受支持: %s", runtimeMetadata.ABIVersion)
	}

	totalSQLAssetCount := len(installSQLAssets) + len(uninstallSQLAssets)
	if runtimeMetadata.SQLAssetCount > 0 && runtimeMetadata.SQLAssetCount != totalSQLAssetCount {
		return nil, gerror.Newf(
			"动态插件 SQL 资源数量与元数据不一致: metadata=%d actual=%d",
			runtimeMetadata.SQLAssetCount,
			totalSQLAssetCount,
		)
	}
	if runtimeMetadata.SQLAssetCount <= 0 {
		runtimeMetadata.SQLAssetCount = totalSQLAssetCount
	}
	if runtimeMetadata.FrontendAssetCount > 0 && runtimeMetadata.FrontendAssetCount != len(frontendAssets) {
		return nil, gerror.Newf(
			"动态插件前端资源数量与元数据不一致: metadata=%d actual=%d",
			runtimeMetadata.FrontendAssetCount,
			len(frontendAssets),
		)
	}
	if runtimeMetadata.FrontendAssetCount <= 0 {
		runtimeMetadata.FrontendAssetCount = len(frontendAssets)
	}
	if runtimeMetadata.I18NAssetCount > 0 && runtimeMetadata.I18NAssetCount != len(runtimeI18NAssets) {
		return nil, gerror.Newf(
			"动态插件运行时 i18n 资源数量与元数据不一致: metadata=%d actual=%d",
			runtimeMetadata.I18NAssetCount,
			len(runtimeI18NAssets),
		)
	}
	if runtimeMetadata.I18NAssetCount <= 0 {
		runtimeMetadata.I18NAssetCount = len(runtimeI18NAssets)
	}
	if runtimeMetadata.APIDocI18NAssetCount > 0 && runtimeMetadata.APIDocI18NAssetCount != len(apiDocI18NAssets) {
		return nil, gerror.Newf(
			"动态插件 apidoc i18n 资源数量与元数据不一致: metadata=%d actual=%d",
			runtimeMetadata.APIDocI18NAssetCount,
			len(apiDocI18NAssets),
		)
	}
	if runtimeMetadata.APIDocI18NAssetCount <= 0 {
		runtimeMetadata.APIDocI18NAssetCount = len(apiDocI18NAssets)
	}
	if runtimeMetadata.RouteCount > 0 && runtimeMetadata.RouteCount != len(routeContracts) {
		return nil, gerror.Newf(
			"动态插件路由数量与元数据不一致: metadata=%d actual=%d",
			runtimeMetadata.RouteCount,
			len(routeContracts),
		)
	}
	if runtimeMetadata.RouteCount <= 0 {
		runtimeMetadata.RouteCount = len(routeContracts)
	}

	return &catalog.ArtifactSpec{
		Path:                 filePath,
		Checksum:             fmt.Sprintf("%x", sha256.Sum256(content)),
		RuntimeKind:          runtimeKind,
		ABIVersion:           abiVersion,
		FrontendAssetCount:   maxInt(runtimeMetadata.FrontendAssetCount, 0),
		I18NAssetCount:       maxInt(runtimeMetadata.I18NAssetCount, 0),
		APIDocI18NAssetCount: maxInt(runtimeMetadata.APIDocI18NAssetCount, 0),
		SQLAssetCount:        maxInt(runtimeMetadata.SQLAssetCount, 0),
		RouteCount:           maxInt(runtimeMetadata.RouteCount, 0),
		Manifest:             embeddedManifest,
		FrontendAssets:       frontendAssets,
		InstallSQLAssets:     installSQLAssets,
		UninstallSQLAssets:   uninstallSQLAssets,
		HookSpecs:            hookSpecs,
		ResourceSpecs:        resourceSpecs,
		RouteContracts:       routeContracts,
		BridgeSpec:           bridgeSpec,
		Capabilities:         capabilities,
		HostServices:         hostServices,
	}, nil
}

// ValidateRuntimeArtifact loads and validates the WASM artifact for a dynamic plugin source directory.
// It implements the catalog.ArtifactParser interface.
func (s *serviceImpl) ValidateRuntimeArtifact(manifest *catalog.Manifest, rootDir string) error {
	artifactPath, err := resolveArtifactPath(rootDir, manifest.ID)
	if err != nil {
		return err
	}

	artifact, err := s.ParseRuntimeWasmArtifact(artifactPath)
	if err != nil {
		return err
	}
	if artifact.Manifest == nil {
		return gerror.Newf("动态插件缺少嵌入清单: %s", artifactPath)
	}

	artifact.Manifest.Type = catalog.NormalizeType(artifact.Manifest.Type).String()
	if catalog.NormalizeType(artifact.Manifest.Type) != catalog.TypeDynamic {
		return gerror.Newf("动态插件嵌入清单类型必须是 dynamic: %s", artifactPath)
	}
	if manifest.ID != artifact.Manifest.ID {
		return gerror.Newf("动态插件嵌入清单 ID 与 plugin.yaml 不一致: %s != %s", artifact.Manifest.ID, manifest.ID)
	}
	if manifest.Name != artifact.Manifest.Name {
		return gerror.Newf("动态插件嵌入清单名称与 plugin.yaml 不一致: %s != %s", artifact.Manifest.Name, manifest.Name)
	}
	if manifest.Version != artifact.Manifest.Version {
		return gerror.Newf("动态插件嵌入清单版本与 plugin.yaml 不一致: %s != %s", artifact.Manifest.Version, manifest.Version)
	}

	manifest.RuntimeArtifact = artifact
	return nil
}

// ensureArtifactAvailable ensures the WASM artifact is present for lifecycle operations.
func (s *serviceImpl) ensureArtifactAvailable(manifest *catalog.Manifest, actionLabel string) error {
	if manifest == nil {
		return gerror.New("插件清单不能为空")
	}
	if catalog.NormalizeType(manifest.Type) != catalog.TypeDynamic {
		return nil
	}
	if manifest.RuntimeArtifact != nil {
		return nil
	}

	if err := s.ValidateRuntimeArtifact(manifest, manifest.RootDir); err != nil {
		if isMissingArtifactError(err) {
			return gerror.Newf(
				"动态插件缺少 %s，无法%s；请先执行 make wasm p=%s 生成产物",
				filepath.ToSlash(buildArtifactRelativePath(manifest.ID)),
				actionLabel,
				manifest.ID,
			)
		}
		return gerror.Wrapf(err, "动态插件产物校验失败，无法%s", actionLabel)
	}
	return nil
}

// buildPluginRegistryChecksum returns the SHA-256 checksum of the plugin artifact or manifest.
func (s *serviceImpl) buildPluginRegistryChecksum(manifest *catalog.Manifest) string {
	if manifest == nil {
		return ""
	}
	if manifest.RuntimeArtifact != nil {
		return manifest.RuntimeArtifact.Checksum
	}
	content, err := s.catalogSvc.ReadSourcePluginManifestContent(manifest)
	if err != nil || len(content) == 0 {
		return ""
	}
	return fmt.Sprintf("%x", sha256.Sum256(content))
}

// buildRuntimeArtifactRemark summarizes runtime WASM metadata for governance review.
func buildRuntimeArtifactRemark(manifest *catalog.Manifest) string {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return ""
	}
	return fmt.Sprintf(
		"The host validated one %s runtime artifact using ABI %s with %d embedded frontend assets, %d install SQL assets, %d uninstall SQL assets, and %d dynamic routes declared.",
		manifest.RuntimeArtifact.RuntimeKind,
		manifest.RuntimeArtifact.ABIVersion,
		manifest.RuntimeArtifact.FrontendAssetCount,
		len(manifest.RuntimeArtifact.InstallSQLAssets),
		len(manifest.RuntimeArtifact.UninstallSQLAssets),
		len(manifest.RuntimeArtifact.RouteContracts),
	)
}

// unmarshalRuntimeArtifactSection decodes one JSON-encoded custom section payload.
func unmarshalRuntimeArtifactSection(content []byte, target interface{}) error {
	if err := json.Unmarshal(content, target); err == nil {
		return nil
	}
	return gerror.New("动态插件自定义节仅支持 JSON 编码")
}

// parseWasmCustomSections extracts custom sections from a wasm binary by name.
func parseWasmCustomSections(content []byte) (map[string][]byte, error) {
	if len(content) < 8 {
		return nil, gerror.New("wasm 文件长度不足")
	}
	if string(content[:4]) != "\x00asm" {
		return nil, gerror.New("wasm 文件头非法")
	}
	if content[4] != 0x01 || content[5] != 0x00 || content[6] != 0x00 || content[7] != 0x00 {
		return nil, gerror.New("wasm 版本非法")
	}

	sections := make(map[string][]byte)
	cursor := 8
	for cursor < len(content) {
		sectionID := content[cursor]
		cursor++

		sectionSize, nextCursor, err := readWasmULEB128(content, cursor)
		if err != nil {
			return nil, err
		}
		cursor = nextCursor

		end := cursor + int(sectionSize)
		if end > len(content) {
			return nil, gerror.New("wasm 节长度越界")
		}

		if sectionID == 0 {
			nameLength, nameCursor, err := readWasmULEB128(content, cursor)
			if err != nil {
				return nil, err
			}
			nameEnd := nameCursor + int(nameLength)
			if nameEnd > end {
				return nil, gerror.New("wasm 自定义节名称越界")
			}

			sectionName := string(content[nameCursor:nameEnd])
			sectionPayload := make([]byte, end-nameEnd)
			copy(sectionPayload, content[nameEnd:end])
			sections[sectionName] = sectionPayload
		}

		cursor = end
	}
	return sections, nil
}

// readWasmULEB128 decodes one unsigned LEB128 value from the wasm byte stream.
func readWasmULEB128(content []byte, start int) (uint32, int, error) {
	var (
		value uint32
		shift uint
	)

	cursor := start
	for {
		if cursor >= len(content) {
			return 0, cursor, gerror.New("wasm ULEB128 数据越界")
		}
		current := content[cursor]
		cursor++

		value |= uint32(current&0x7f) << shift
		if current&0x80 == 0 {
			return value, cursor, nil
		}

		shift += 7
		if shift > 28 {
			return 0, cursor, gerror.New("wasm ULEB128 数值过大")
		}
	}
}

// maxInt clamps value to the given lower bound.
func maxInt(value int, lowerBound int) int {
	if value < lowerBound {
		return lowerBound
	}
	return value
}

// runtimeArtifactLocaleJSONAsset stores one locale JSON payload embedded in a
// dynamic plugin artifact.
type runtimeArtifactLocaleJSONAsset struct {
	Locale  string `json:"locale"`
	Content string `json:"content"`
}

// parseRuntimeArtifactLocaleJSONAssets validates locale JSON assets embedded
// for runtime UI i18n or API-documentation i18n.
func parseRuntimeArtifactLocaleJSONAssets(
	filePath string,
	sections map[string][]byte,
	sectionName string,
) ([]*runtimeArtifactLocaleJSONAsset, error) {
	sectionContent, ok := sections[sectionName]
	if !ok {
		return []*runtimeArtifactLocaleJSONAsset{}, nil
	}

	assets := make([]*runtimeArtifactLocaleJSONAsset, 0)
	if err := json.Unmarshal(sectionContent, &assets); err != nil {
		return nil, gerror.Wrapf(err, "解析动态插件 i18n 自定义节失败 section=%s: %s", sectionName, filePath)
	}
	for _, asset := range assets {
		if asset == nil {
			return nil, gerror.Newf("动态插件 i18n 自定义节存在空项 section=%s: %s", sectionName, filePath)
		}
		asset.Locale = strings.TrimSpace(asset.Locale)
		asset.Content = strings.TrimSpace(asset.Content)
		if asset.Locale == "" || asset.Content == "" {
			return nil, gerror.Newf("动态插件 i18n 自定义节缺少 locale 或 content section=%s: %s", sectionName, filePath)
		}
		if err := validateRuntimeArtifactLocaleJSONContent(sectionName, asset.Content); err != nil {
			return nil, gerror.Wrapf(err, "解析动态插件 i18n 资源内容失败 section=%s locale=%s: %s", sectionName, asset.Locale, filePath)
		}
	}
	return assets, nil
}

// validateRuntimeArtifactLocaleJSONContent validates one locale JSON payload.
// Runtime UI and API-documentation i18n both accept nested JSON authoring, and
// both keep string leaves after normalizing to flat structured keys.
func validateRuntimeArtifactLocaleJSONContent(sectionName string, content string) error {
	var bundle map[string]interface{}
	if err := json.Unmarshal([]byte(content), &bundle); err != nil {
		return err
	}
	return validateRuntimeArtifactI18NMessageValue(bundle)
}

// validateRuntimeArtifactI18NMessageValue verifies nested runtime i18n assets
// contain JSON objects with string leaves only.
func validateRuntimeArtifactI18NMessageValue(value interface{}) error {
	switch typedValue := value.(type) {
	case map[string]interface{}:
		for _, item := range typedValue {
			if err := validateRuntimeArtifactI18NMessageValue(item); err != nil {
				return err
			}
		}
		return nil
	case string:
		return nil
	default:
		return gerror.New("运行时 i18n 资源值必须是字符串或对象")
	}
}

// parseRuntimeArtifactSQLAssets restores embedded SQL assets and validates
// their canonical file-style keys.
func parseRuntimeArtifactSQLAssets(
	filePath string,
	sections map[string][]byte,
	sectionName string,
) ([]*catalog.ArtifactSQLAsset, error) {
	sectionContent, ok := sections[sectionName]
	if !ok {
		return []*catalog.ArtifactSQLAsset{}, nil
	}

	assets := make([]*catalog.ArtifactSQLAsset, 0)
	if err := json.Unmarshal(sectionContent, &assets); err != nil {
		return nil, gerror.Wrapf(err, "解析动态插件 SQL 自定义节失败: %s", filePath)
	}
	for _, asset := range assets {
		if asset == nil {
			return nil, gerror.Newf("动态插件 SQL 自定义节存在空项: %s", filePath)
		}
		asset.Key = strings.TrimSpace(asset.Key)
		asset.Content = strings.TrimSpace(asset.Content)
		if asset.Key == "" || asset.Content == "" {
			return nil, gerror.Newf("动态插件 SQL 自定义节缺少 key 或 content: %s", filePath)
		}
		if strings.Contains(asset.Key, "/") || strings.Contains(asset.Key, "\\") {
			return nil, gerror.Newf("动态插件 SQL 资源键不能包含路径分隔符: %s", asset.Key)
		}
		if !pluginfs.IsValidSQLFileName(asset.Key) {
			return nil, gerror.Newf("动态插件 SQL 资源键不符合命名规则: %s", asset.Key)
		}
	}
	return assets, nil
}

// parseRuntimeArtifactHookSpecs restores and validates embedded hook specs.
func parseRuntimeArtifactHookSpecs(
	filePath string,
	pluginID string,
	sections map[string][]byte,
) ([]*catalog.HookSpec, error) {
	content, ok := sections[pluginbridge.WasmSectionBackendHooks]
	if !ok {
		return []*catalog.HookSpec{}, nil
	}

	items := make([]*catalog.HookSpec, 0)
	if err := json.Unmarshal(content, &items); err != nil {
		return nil, gerror.Wrapf(err, "解析动态插件后端 Hook 契约失败: %s", filePath)
	}
	for _, item := range items {
		if err := catalog.ValidateHookSpec(pluginID, item, filePath); err != nil {
			return nil, err
		}
	}
	return catalog.CloneHookSpecs(items), nil
}

// parseRuntimeArtifactResourceSpecs restores and validates embedded resource specs.
func parseRuntimeArtifactResourceSpecs(
	filePath string,
	pluginID string,
	sections map[string][]byte,
) ([]*catalog.ResourceSpec, error) {
	content, ok := sections[pluginbridge.WasmSectionBackendResources]
	if !ok {
		return []*catalog.ResourceSpec{}, nil
	}

	items := make([]*catalog.ResourceSpec, 0)
	if err := json.Unmarshal(content, &items); err != nil {
		return nil, gerror.Wrapf(err, "解析动态插件后端资源契约失败: %s", filePath)
	}
	cloned := make([]*catalog.ResourceSpec, 0, len(items))
	for _, item := range items {
		if err := catalog.ValidateResourceSpec(pluginID, item, filePath); err != nil {
			return nil, err
		}
		cloned = append(cloned, catalog.CloneResourceSpec(item))
	}
	return cloned, nil
}

// parseRuntimeArtifactRouteContracts restores and validates embedded route contracts.
func parseRuntimeArtifactRouteContracts(
	filePath string,
	pluginID string,
	sections map[string][]byte,
) ([]*pluginbridge.RouteContract, error) {
	content, ok := sections[pluginbridge.WasmSectionBackendRoutes]
	if !ok {
		return []*pluginbridge.RouteContract{}, nil
	}

	items := make([]*pluginbridge.RouteContract, 0)
	if err := json.Unmarshal(content, &items); err != nil {
		return nil, gerror.Wrapf(err, "解析动态插件后端动态路由契约失败: %s", filePath)
	}
	if err := pluginbridge.ValidateRouteContracts(pluginID, items); err != nil {
		return nil, gerror.Wrapf(err, "校验动态插件后端动态路由契约失败: %s", filePath)
	}
	return items, nil
}

// parseRuntimeArtifactBridgeSpec restores and validates the optional bridge spec.
func parseRuntimeArtifactBridgeSpec(
	filePath string,
	sections map[string][]byte,
) (*pluginbridge.BridgeSpec, error) {
	content, ok := sections[pluginbridge.WasmSectionBackendBridge]
	if !ok {
		return nil, nil
	}

	spec := &pluginbridge.BridgeSpec{}
	if err := json.Unmarshal(content, spec); err != nil {
		return nil, gerror.Wrapf(err, "解析动态插件 bridge 契约失败: %s", filePath)
	}
	if err := pluginbridge.ValidateBridgeSpec(spec); err != nil {
		return nil, gerror.Wrapf(err, "校验动态插件 bridge 契约失败: %s", filePath)
	}
	return spec, nil
}

// rejectDeprecatedRuntimeArtifactCapabilities fails fast when old capability
// sections are still embedded alongside the structured host-service contract.
func rejectDeprecatedRuntimeArtifactCapabilities(
	filePath string,
	sections map[string][]byte,
) error {
	content, ok := sections[pluginbridge.WasmSectionBackendCapabilities]
	if !ok {
		return nil
	}

	var items []string
	if err := json.Unmarshal(content, &items); err == nil {
		normalizedItems := pluginbridge.NormalizeCapabilities(items)
		if len(normalizedItems) > 0 {
			return gerror.Newf(
				"动态插件产物已废弃自定义节 %s，请删除旧 capabilities 声明后重新构建: %s (%s)",
				pluginbridge.WasmSectionBackendCapabilities,
				filePath,
				strings.Join(normalizedItems, ", "),
			)
		}
	}
	return gerror.Newf(
		"动态插件产物已废弃自定义节 %s，请删除旧 capabilities 声明后重新构建: %s",
		pluginbridge.WasmSectionBackendCapabilities,
		filePath,
	)
}

// parseRuntimeArtifactHostServices restores and validates embedded host-service declarations.
func parseRuntimeArtifactHostServices(
	filePath string,
	sections map[string][]byte,
) ([]*pluginbridge.HostServiceSpec, error) {
	content, ok := sections[pluginbridge.WasmSectionBackendHostServices]
	if !ok {
		return []*pluginbridge.HostServiceSpec{}, nil
	}

	items := make([]*pluginbridge.HostServiceSpec, 0)
	if err := json.Unmarshal(content, &items); err != nil {
		return nil, gerror.Wrapf(err, "解析动态插件宿主服务声明失败: %s", filePath)
	}
	if err := pluginbridge.ValidateHostServiceSpecs(items); err != nil {
		return nil, gerror.Wrapf(err, "校验动态插件宿主服务声明失败: %s", filePath)
	}
	return pluginbridge.NormalizeHostServiceSpecs(items), nil
}

// parseRuntimeArtifactFrontendAssets restores embedded frontend assets and
// decodes their base64-encoded content payloads.
func parseRuntimeArtifactFrontendAssets(
	filePath string,
	sections map[string][]byte,
	sectionName string,
) ([]*catalog.ArtifactFrontendAsset, error) {
	content, ok := sections[sectionName]
	if !ok {
		return []*catalog.ArtifactFrontendAsset{}, nil
	}

	assets := make([]*catalog.ArtifactFrontendAsset, 0)
	if err := json.Unmarshal(content, &assets); err != nil {
		return nil, gerror.Wrapf(err, "解析动态插件前端资源失败: %s", filePath)
	}

	for _, asset := range assets {
		if asset == nil {
			return nil, gerror.Newf("动态插件前端资源不能为空: %s", filePath)
		}
		asset.Path = normalizeAssetPath(asset.Path)
		if asset.Path == "" {
			return nil, gerror.Newf("动态插件前端资源路径不能为空: %s", filePath)
		}
		if asset.ContentBase64 == "" {
			return nil, gerror.Newf("动态插件前端资源内容不能为空: %s", asset.Path)
		}

		decoded, err := base64.StdEncoding.DecodeString(asset.ContentBase64)
		if err != nil {
			return nil, gerror.Wrapf(err, "解析动态插件前端资源内容失败: %s", asset.Path)
		}
		if len(decoded) == 0 {
			return nil, gerror.Newf("动态插件前端资源内容不能为空: %s", asset.Path)
		}
		asset.Content = decoded
	}
	return assets, nil
}

// normalizeAssetPath normalizes a relative frontend asset path into canonical form.
func normalizeAssetPath(relativePath string) string {
	normalizedPath := strings.TrimSpace(relativePath)
	normalizedPath = strings.ReplaceAll(normalizedPath, "\\", "/")
	normalizedPath = strings.TrimPrefix(normalizedPath, "/")
	normalizedPath = strings.TrimPrefix(normalizedPath, "./")
	normalizedPath = strings.TrimSpace(normalizedPath)
	if normalizedPath == "" {
		return ""
	}
	normalizedPath = filepath.ToSlash(filepath.Clean(normalizedPath))
	if normalizedPath == "." || normalizedPath == ".." || strings.HasPrefix(normalizedPath, "../") {
		return ""
	}
	return normalizedPath
}
