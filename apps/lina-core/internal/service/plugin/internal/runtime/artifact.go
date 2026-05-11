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
	"lina-core/pkg/bizerr"
	bridgeartifact "lina-core/pkg/pluginbridge/artifact"
	bridgecontract "lina-core/pkg/pluginbridge/contract"
	bridgehostservice "lina-core/pkg/pluginbridge/hostservice"
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
	return fmt.Sprintf("Dynamic plugin directory is missing %s: %s", e.relativePath, e.rootDir)
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
		return nil, gerror.Newf("Dynamic plugin artifact is empty: %s", filePath)
	}
	return s.ParseRuntimeWasmArtifactContent(filePath, content)
}

// ParseRuntimeWasmArtifactContent parses one WASM artifact from an in-memory byte slice.
// It implements the catalog.ArtifactParser interface.
func (s *serviceImpl) ParseRuntimeWasmArtifactContent(filePath string, content []byte) (*catalog.ArtifactSpec, error) {
	sections, err := bridgeartifact.ListCustomSections(content)
	if err != nil {
		return nil, gerror.Wrapf(err, "Failed to parse dynamic plugin artifact: %s", filePath)
	}

	manifestSection, ok := sections[bridgeartifact.WasmSectionManifest]
	if !ok {
		return nil, gerror.Newf("Dynamic plugin artifact is missing custom section %s: %s", bridgeartifact.WasmSectionManifest, filePath)
	}
	runtimeSection, ok := sections[bridgeartifact.WasmSectionRuntime]
	if !ok {
		runtimeSection, ok = sections[bridgeartifact.WasmSectionLegacyRuntime]
	}
	if !ok {
		return nil, gerror.Newf("Dynamic plugin artifact is missing custom section %s: %s", bridgeartifact.WasmSectionRuntime, filePath)
	}

	embeddedManifest := &catalog.ArtifactManifest{}
	if err = unmarshalRuntimeArtifactSection(manifestSection, embeddedManifest); err != nil {
		return nil, gerror.Wrapf(err, "Failed to parse dynamic plugin embedded manifest: %s", filePath)
	}
	if strings.TrimSpace(embeddedManifest.ID) == "" ||
		strings.TrimSpace(embeddedManifest.Name) == "" ||
		strings.TrimSpace(embeddedManifest.Version) == "" ||
		strings.TrimSpace(embeddedManifest.Type) == "" {
		return nil, gerror.Newf("Dynamic plugin embedded manifest is missing required fields: %s", filePath)
	}
	embeddedManifest.ScopeNature = strings.TrimSpace(embeddedManifest.ScopeNature)
	embeddedManifest.DefaultInstallMode = strings.TrimSpace(embeddedManifest.DefaultInstallMode)

	runtimeMetadata := &bridgeartifact.RuntimeArtifactMetadata{}
	if err = unmarshalRuntimeArtifactSection(runtimeSection, runtimeMetadata); err != nil {
		return nil, gerror.Wrapf(err, "Failed to parse dynamic plugin runtime metadata: %s", filePath)
	}

	frontendAssets, err := parseRuntimeArtifactFrontendAssets(filePath, sections, bridgeartifact.WasmSectionFrontendAssets)
	if err != nil {
		return nil, err
	}
	runtimeI18NAssets, err := parseRuntimeArtifactLocaleJSONAssets(filePath, sections, bridgeartifact.WasmSectionI18NAssets)
	if err != nil {
		return nil, err
	}
	apiDocI18NAssets, err := parseRuntimeArtifactLocaleJSONAssets(filePath, sections, bridgeartifact.WasmSectionAPIDocI18NAssets)
	if err != nil {
		return nil, err
	}
	installSQLAssets, err := parseRuntimeArtifactSQLAssets(filePath, sections, bridgeartifact.WasmSectionInstallSQL)
	if err != nil {
		return nil, err
	}
	uninstallSQLAssets, err := parseRuntimeArtifactSQLAssets(filePath, sections, bridgeartifact.WasmSectionUninstallSQL)
	if err != nil {
		return nil, err
	}
	mockSQLAssets, err := parseRuntimeArtifactSQLAssets(filePath, sections, bridgeartifact.WasmSectionMockSQL)
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
	capabilities := bridgehostservice.CapabilitiesFromHostServices(hostServices)

	runtimeKind := strings.TrimSpace(strings.ToLower(runtimeMetadata.RuntimeKind))
	if runtimeKind == "" {
		runtimeKind = dynamicKindWasm
	}
	if runtimeKind != dynamicKindWasm {
		return nil, gerror.Newf("Dynamic plugin artifact runtime kind must be wasm: %s", runtimeKind)
	}

	abiVersion := strings.TrimSpace(strings.ToLower(runtimeMetadata.ABIVersion))
	if abiVersion == "" {
		return nil, gerror.Newf("Dynamic plugin artifact is missing ABI version: %s", filePath)
	}
	if abiVersion != bridgecontract.SupportedABIVersion {
		return nil, gerror.Newf("Dynamic plugin ABI version is not supported: %s", runtimeMetadata.ABIVersion)
	}

	// SQLAssetCount tallies install + uninstall + mock together so a single
	// metadata field can validate the artifact's overall SQL footprint while
	// each direction still has its own typed slice for per-phase execution.
	totalSQLAssetCount := len(installSQLAssets) + len(uninstallSQLAssets) + len(mockSQLAssets)
	if runtimeMetadata.SQLAssetCount > 0 && runtimeMetadata.SQLAssetCount != totalSQLAssetCount {
		return nil, gerror.Newf(
			"Dynamic plugin SQL asset count does not match metadata: metadata=%d actual=%d",
			runtimeMetadata.SQLAssetCount,
			totalSQLAssetCount,
		)
	}
	if runtimeMetadata.SQLAssetCount <= 0 {
		runtimeMetadata.SQLAssetCount = totalSQLAssetCount
	}
	if runtimeMetadata.MockSQLAssetCount > 0 && runtimeMetadata.MockSQLAssetCount != len(mockSQLAssets) {
		return nil, gerror.Newf(
			"Dynamic plugin mock SQL asset count does not match metadata: metadata=%d actual=%d",
			runtimeMetadata.MockSQLAssetCount,
			len(mockSQLAssets),
		)
	}
	if runtimeMetadata.MockSQLAssetCount <= 0 {
		runtimeMetadata.MockSQLAssetCount = len(mockSQLAssets)
	}
	if runtimeMetadata.FrontendAssetCount > 0 && runtimeMetadata.FrontendAssetCount != len(frontendAssets) {
		return nil, gerror.Newf(
			"Dynamic plugin frontend asset count does not match metadata: metadata=%d actual=%d",
			runtimeMetadata.FrontendAssetCount,
			len(frontendAssets),
		)
	}
	if runtimeMetadata.FrontendAssetCount <= 0 {
		runtimeMetadata.FrontendAssetCount = len(frontendAssets)
	}
	if runtimeMetadata.I18NAssetCount > 0 && runtimeMetadata.I18NAssetCount != len(runtimeI18NAssets) {
		return nil, gerror.Newf(
			"Dynamic plugin runtime i18n asset count does not match metadata: metadata=%d actual=%d",
			runtimeMetadata.I18NAssetCount,
			len(runtimeI18NAssets),
		)
	}
	if runtimeMetadata.I18NAssetCount <= 0 {
		runtimeMetadata.I18NAssetCount = len(runtimeI18NAssets)
	}
	if runtimeMetadata.APIDocI18NAssetCount > 0 && runtimeMetadata.APIDocI18NAssetCount != len(apiDocI18NAssets) {
		return nil, gerror.Newf(
			"Dynamic plugin apidoc i18n asset count does not match metadata: metadata=%d actual=%d",
			runtimeMetadata.APIDocI18NAssetCount,
			len(apiDocI18NAssets),
		)
	}
	if runtimeMetadata.APIDocI18NAssetCount <= 0 {
		runtimeMetadata.APIDocI18NAssetCount = len(apiDocI18NAssets)
	}
	if runtimeMetadata.RouteCount > 0 && runtimeMetadata.RouteCount != len(routeContracts) {
		return nil, gerror.Newf(
			"Dynamic plugin route count does not match metadata: metadata=%d actual=%d",
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
		MockSQLAssets:        mockSQLAssets,
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
		return gerror.Newf("Dynamic plugin artifact is missing embedded manifest: %s", artifactPath)
	}

	artifact.Manifest.Type = catalog.NormalizeType(artifact.Manifest.Type).String()
	if catalog.NormalizeType(artifact.Manifest.Type) != catalog.TypeDynamic {
		return gerror.Newf("Dynamic plugin embedded manifest type must be dynamic: %s", artifactPath)
	}
	if manifest.ID != artifact.Manifest.ID {
		return gerror.Newf("Dynamic plugin embedded manifest ID does not match plugin.yaml: %s != %s", artifact.Manifest.ID, manifest.ID)
	}
	if manifest.Name != artifact.Manifest.Name {
		return gerror.Newf("Dynamic plugin embedded manifest name does not match plugin.yaml: %s != %s", artifact.Manifest.Name, manifest.Name)
	}
	if manifest.Version != artifact.Manifest.Version {
		return gerror.Newf("Dynamic plugin embedded manifest version does not match plugin.yaml: %s != %s", artifact.Manifest.Version, manifest.Version)
	}

	manifest.RuntimeArtifact = artifact
	return nil
}

// ensureArtifactAvailable ensures the WASM artifact is present for lifecycle operations.
func (s *serviceImpl) ensureArtifactAvailable(manifest *catalog.Manifest, actionLabel string) error {
	if manifest == nil {
		return bizerr.NewCode(CodeDynamicPluginManifestRequired)
	}
	if catalog.NormalizeType(manifest.Type) != catalog.TypeDynamic {
		return nil
	}
	if manifest.RuntimeArtifact != nil {
		return nil
	}

	if err := s.ValidateRuntimeArtifact(manifest, manifest.RootDir); err != nil {
		if isMissingArtifactError(err) {
			return bizerr.NewCode(
				CodeDynamicPluginArtifactMissing,
				bizerr.P("artifactPath", filepath.ToSlash(buildArtifactRelativePath(manifest.ID))),
				bizerr.P("action", actionLabel),
				bizerr.P("pluginId", manifest.ID),
			)
		}
		return bizerr.WrapCode(
			err,
			CodeDynamicPluginArtifactValidateFailed,
			bizerr.P("action", actionLabel),
		)
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
		"The host validated one %s runtime artifact using ABI %s with %d embedded frontend assets, %d install SQL assets, %d uninstall SQL assets, %d mock SQL assets, and %d dynamic routes declared.",
		manifest.RuntimeArtifact.RuntimeKind,
		manifest.RuntimeArtifact.ABIVersion,
		manifest.RuntimeArtifact.FrontendAssetCount,
		len(manifest.RuntimeArtifact.InstallSQLAssets),
		len(manifest.RuntimeArtifact.UninstallSQLAssets),
		len(manifest.RuntimeArtifact.MockSQLAssets),
		len(manifest.RuntimeArtifact.RouteContracts),
	)
}

// unmarshalRuntimeArtifactSection decodes one JSON-encoded custom section payload.
func unmarshalRuntimeArtifactSection(content []byte, target interface{}) error {
	if err := json.Unmarshal(content, target); err == nil {
		return nil
	}
	return gerror.New("Dynamic plugin custom sections support JSON encoding only")
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
		return nil, gerror.Wrapf(err, "Failed to parse dynamic plugin i18n custom section section=%s: %s", sectionName, filePath)
	}
	for _, asset := range assets {
		if asset == nil {
			return nil, gerror.Newf("Dynamic plugin i18n custom section contains a null item section=%s: %s", sectionName, filePath)
		}
		asset.Locale = strings.TrimSpace(asset.Locale)
		asset.Content = strings.TrimSpace(asset.Content)
		if asset.Locale == "" || asset.Content == "" {
			return nil, gerror.Newf("Dynamic plugin i18n custom section is missing locale or content section=%s: %s", sectionName, filePath)
		}
		if err := validateRuntimeArtifactLocaleJSONContent(sectionName, asset.Content); err != nil {
			return nil, gerror.Wrapf(err, "Failed to parse dynamic plugin i18n resource content section=%s locale=%s: %s", sectionName, asset.Locale, filePath)
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
		return gerror.New("Runtime i18n resource values must be strings or objects")
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
		return nil, gerror.Wrapf(err, "Failed to parse dynamic plugin SQL custom section: %s", filePath)
	}
	for _, asset := range assets {
		if asset == nil {
			return nil, gerror.Newf("Dynamic plugin SQL custom section contains a null item: %s", filePath)
		}
		asset.Key = strings.TrimSpace(asset.Key)
		asset.Content = strings.TrimSpace(asset.Content)
		if asset.Key == "" || asset.Content == "" {
			return nil, gerror.Newf("Dynamic plugin SQL custom section is missing key or content: %s", filePath)
		}
		if strings.Contains(asset.Key, "/") || strings.Contains(asset.Key, "\\") {
			return nil, gerror.Newf("Dynamic plugin SQL asset key cannot contain path separators: %s", asset.Key)
		}
		if !pluginfs.IsValidSQLFileName(asset.Key) {
			return nil, gerror.Newf("Dynamic plugin SQL asset key does not match the naming rule: %s", asset.Key)
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
	content, ok := sections[bridgeartifact.WasmSectionBackendHooks]
	if !ok {
		return []*catalog.HookSpec{}, nil
	}

	items := make([]*catalog.HookSpec, 0)
	if err := json.Unmarshal(content, &items); err != nil {
		return nil, gerror.Wrapf(err, "Failed to parse dynamic plugin backend hook contracts: %s", filePath)
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
	content, ok := sections[bridgeartifact.WasmSectionBackendResources]
	if !ok {
		return []*catalog.ResourceSpec{}, nil
	}

	items := make([]*catalog.ResourceSpec, 0)
	if err := json.Unmarshal(content, &items); err != nil {
		return nil, gerror.Wrapf(err, "Failed to parse dynamic plugin backend resource contracts: %s", filePath)
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
) ([]*bridgecontract.RouteContract, error) {
	content, ok := sections[bridgeartifact.WasmSectionBackendRoutes]
	if !ok {
		return []*bridgecontract.RouteContract{}, nil
	}

	items := make([]*bridgecontract.RouteContract, 0)
	if err := json.Unmarshal(content, &items); err != nil {
		return nil, gerror.Wrapf(err, "Failed to parse dynamic plugin backend route contracts: %s", filePath)
	}
	if err := bridgecontract.ValidateRouteContracts(pluginID, items); err != nil {
		return nil, gerror.Wrapf(err, "Failed to validate dynamic plugin backend route contracts: %s", filePath)
	}
	return items, nil
}

// parseRuntimeArtifactBridgeSpec restores and validates the optional bridge spec.
func parseRuntimeArtifactBridgeSpec(
	filePath string,
	sections map[string][]byte,
) (*bridgecontract.BridgeSpec, error) {
	content, ok := sections[bridgeartifact.WasmSectionBackendBridge]
	if !ok {
		return nil, nil
	}

	spec := &bridgecontract.BridgeSpec{}
	if err := json.Unmarshal(content, spec); err != nil {
		return nil, gerror.Wrapf(err, "Failed to parse dynamic plugin bridge contract: %s", filePath)
	}
	if err := bridgecontract.ValidateBridgeSpec(spec); err != nil {
		return nil, gerror.Wrapf(err, "Failed to validate dynamic plugin bridge contract: %s", filePath)
	}
	return spec, nil
}

// rejectDeprecatedRuntimeArtifactCapabilities fails fast when old capability
// sections are still embedded alongside the structured host-service contract.
func rejectDeprecatedRuntimeArtifactCapabilities(
	filePath string,
	sections map[string][]byte,
) error {
	content, ok := sections[bridgeartifact.WasmSectionBackendCapabilities]
	if !ok {
		return nil
	}

	var items []string
	if err := json.Unmarshal(content, &items); err == nil {
		normalizedItems := bridgehostservice.NormalizeCapabilities(items)
		if len(normalizedItems) > 0 {
			return gerror.Newf(
				"Dynamic plugin artifact contains deprecated custom section %s; remove the legacy capabilities declaration and rebuild: %s (%s)",
				bridgeartifact.WasmSectionBackendCapabilities,
				filePath,
				strings.Join(normalizedItems, ", "),
			)
		}
	}
	return gerror.Newf(
		"Dynamic plugin artifact contains deprecated custom section %s; remove the legacy capabilities declaration and rebuild: %s",
		bridgeartifact.WasmSectionBackendCapabilities,
		filePath,
	)
}

// parseRuntimeArtifactHostServices restores and validates embedded host-service declarations.
func parseRuntimeArtifactHostServices(
	filePath string,
	sections map[string][]byte,
) ([]*bridgehostservice.HostServiceSpec, error) {
	content, ok := sections[bridgeartifact.WasmSectionBackendHostServices]
	if !ok {
		return []*bridgehostservice.HostServiceSpec{}, nil
	}

	items := make([]*bridgehostservice.HostServiceSpec, 0)
	if err := json.Unmarshal(content, &items); err != nil {
		return nil, gerror.Wrapf(err, "Failed to parse dynamic plugin host-service declarations: %s", filePath)
	}
	if err := bridgehostservice.ValidateHostServiceSpecs(items); err != nil {
		return nil, gerror.Wrapf(err, "Failed to validate dynamic plugin host-service declarations: %s", filePath)
	}
	normalized, err := bridgehostservice.NormalizeHostServiceSpecs(items)
	if err != nil {
		return nil, gerror.Wrapf(err, "Failed to normalize dynamic plugin host-service declarations: %s", filePath)
	}
	return normalized, nil
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
		return nil, gerror.Wrapf(err, "Failed to parse dynamic plugin frontend assets: %s", filePath)
	}

	for _, asset := range assets {
		if asset == nil {
			return nil, gerror.Newf("Dynamic plugin frontend asset cannot be null: %s", filePath)
		}
		asset.Path = normalizeAssetPath(asset.Path)
		if asset.Path == "" {
			return nil, gerror.Newf("Dynamic plugin frontend asset path cannot be empty: %s", filePath)
		}
		if asset.ContentBase64 == "" {
			return nil, gerror.Newf("Dynamic plugin frontend asset content cannot be empty: %s", asset.Path)
		}

		decoded, err := base64.StdEncoding.DecodeString(asset.ContentBase64)
		if err != nil {
			return nil, gerror.Wrapf(err, "Failed to parse dynamic plugin frontend asset content: %s", asset.Path)
		}
		if len(decoded) == 0 {
			return nil, gerror.Newf("Dynamic plugin frontend asset content cannot be empty: %s", asset.Path)
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
