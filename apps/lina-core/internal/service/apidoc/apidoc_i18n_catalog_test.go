// This file verifies that dedicated API-documentation i18n bundles cover the
// current OpenAPI source metadata without relying on runtime UI i18n resources.

package apidoc

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"unicode"
)

var openAPIMetadataTagPattern = regexp.MustCompile(`([A-Za-z0-9_-]+):"((?:\\.|[^"\\])*)"`)

// containsCJK reports whether a string contains Han characters.
func containsCJK(value string) bool {
	for _, item := range value {
		if unicode.Is(unicode.Han, item) {
			return true
		}
	}
	return false
}

// TestOpenAPIMetadataUsesEnglishSourceText prevents new host, source-plugin, or
// dynamic-plugin API DTO documentation strings from using Chinese source text.
func TestOpenAPIMetadataUsesEnglishSourceText(t *testing.T) {
	values := collectOpenAPISourceMetadataStrings(t)
	var chineseValues []string
	var opaquePlaceholders []string
	for _, value := range values {
		if isOpenAPIExampleMetadata(value) {
			continue
		}
		if containsCJK(value.Value) {
			chineseValues = append(chineseValues, value.Location+": "+value.Value)
			continue
		}
		if isOpaqueOpenAPIPlaceholder(value) {
			opaquePlaceholders = append(opaquePlaceholders, value.Location+": "+value.Value)
		}
	}
	if len(chineseValues) > 0 || len(opaquePlaceholders) > 0 {
		t.Fatalf(
			"OpenAPI metadata source text must be readable English: chinese=%d opaque=%d\nchinese:\n%s\nopaque:\n%s",
			len(chineseValues),
			len(opaquePlaceholders),
			strings.Join(limitStrings(chineseValues, 20), "\n"),
			strings.Join(limitStrings(opaquePlaceholders, 20), "\n"),
		)
	}
}

// TestOpenAPII18nBundlesCoverCurrentMetadata verifies that English API docs use
// source metadata directly while non-English API docs keep complete structured
// apidoc bundle coverage for hand-authored API metadata.
func TestOpenAPII18nBundlesCoverCurrentMetadata(t *testing.T) {
	repoRoot := locateRepositoryRoot(t)
	sourceEn := readOpenAPIJSONBundle(t, filepath.Join(repoRoot, "apps/lina-core/manifest/i18n/apidoc/en-US.json"))
	sourceZh := readOpenAPIJSONBundle(t, filepath.Join(repoRoot, "apps/lina-core/manifest/i18n/apidoc/zh-CN.json"))
	packedEn := readOpenAPIJSONBundle(t, filepath.Join(repoRoot, "apps/lina-core/internal/packed/manifest/i18n/apidoc/en-US.json"))
	packedZh := readOpenAPIJSONBundle(t, filepath.Join(repoRoot, "apps/lina-core/internal/packed/manifest/i18n/apidoc/zh-CN.json"))
	pluginEn := readOpenAPIPluginJSONBundles(t, repoRoot, "en-US")
	pluginZh := readOpenAPIPluginJSONBundles(t, repoRoot, "zh-CN")
	mergedZh := cloneOpenAPIMessageCatalog(sourceZh)
	for _, bundle := range pluginZh {
		mergeOpenAPIMessageCatalog(mergedZh, bundle)
	}

	assertOpenAPIBundlesMirror(t, "en-US", sourceEn, packedEn)
	assertOpenAPIBundlesMirror(t, "zh-CN", sourceZh, packedZh)
	assertOpenAPIEnglishBundlePlaceholder(t, sourceEn)
	assertOpenAPIPluginEnglishBundlesArePlaceholders(t, pluginEn)
	assertOpenAPIBundleUsesStructuredKeys(t, "en-US", sourceEn)
	assertOpenAPIBundleUsesStructuredKeys(t, "zh-CN", sourceZh)
	for pluginID, bundle := range pluginZh {
		assertOpenAPIBundleUsesStructuredKeys(t, "zh-CN plugin "+pluginID, bundle)
	}
	assertOpenAPIBundleDoesNotTranslateGeneratedEntityMetadata(t, "en-US", sourceEn)
	assertOpenAPIBundleDoesNotTranslateGeneratedEntityMetadata(t, "zh-CN", sourceZh)
	for pluginID, bundle := range pluginEn {
		assertOpenAPIBundleDoesNotTranslateGeneratedEntityMetadata(t, "en-US plugin "+pluginID, bundle)
	}
	for pluginID, bundle := range pluginZh {
		assertOpenAPIBundleDoesNotTranslateGeneratedEntityMetadata(t, "zh-CN plugin "+pluginID, bundle)
	}
	assertOpenAPIHostBundleDoesNotOwnPluginKeys(t, sourceZh)
	assertOpenAPIPluginBundlesOwnOnlyPluginKeys(t, pluginZh)

	values := collectOpenAPISourceMetadataStrings(t)
	values = append(values,
		openAPIMetadataValue{Key: "fixed", Value: "Dynamic plugin route execution failed", Location: "plugin/openapi"},
		openAPIMetadataValue{Key: "fixed", Value: "Dynamic plugin route response", Location: "plugin/openapi"},
		openAPIMetadataValue{Key: "fixed", Value: "Dynamic plugin route bridge is not executable", Location: "plugin/openapi"},
	)

	var missingChinese []string
	var sourceTextKeys []string
	for _, value := range values {
		if strings.TrimSpace(value.Value) == "" {
			continue
		}
		if !isOpenAPITranslatableMetadata(value) {
			continue
		}
		if _, ok := mergedZh[value.Value]; ok {
			sourceTextKeys = append(sourceTextKeys, "zh-CN: "+value.Location+": "+value.Value)
		}
	}
	for _, value := range collectGeneratedEntityDescriptionStrings(t) {
		if strings.TrimSpace(value.Value) == "" {
			continue
		}
		if _, ok := mergedZh[value.Value]; ok {
			sourceTextKeys = append(sourceTextKeys, "zh-CN generated: "+value.Location+": "+value.Value)
		}
	}

	requiredKeys := collectOpenAPITranslatableStructuredKeys(t)
	requiredKeys = append(requiredKeys,
		"core.openapi.info.title",
		"core.openapi.info.description",
		"core.openapi.securitySchemes.BearerAuth.dc",
		"core.openapi.servers.0.dc",
		"core.api.auth.v1.LoginReq.meta.tags",
		"core.api.auth.v1.LoginReq.fields.username.dc",
		"core.api.user.v1.ListReq.fields.pageNum.dc",
		"plugins.monitor_loginlog.api.loginlog.v1.ListReq.meta.tags",
		"plugins.plugin_demo_dynamic.paths.get.backend_summary.meta.summary",
	)
	for _, key := range requiredKeys {
		if strings.TrimSpace(resolveOpenAPITestMessage(mergedZh, key)) == "" {
			missingChinese = append(missingChinese, "required structured key: "+key)
		}
	}

	if len(missingChinese) > 0 || len(sourceTextKeys) > 0 {
		t.Fatalf(
			"OpenAPI apidoc i18n bundles are incomplete: missingChinese=%d sourceTextKeys=%d\nmissing zh-CN:\n%s\nsource-text keys:\n%s",
			len(missingChinese),
			len(sourceTextKeys),
			strings.Join(limitStrings(missingChinese, 20), "\n"),
			strings.Join(limitStrings(sourceTextKeys, 20), "\n"),
		)
	}
}

// resolveOpenAPITestMessage mirrors runtime lookup for exact keys plus common
// fallback keys so coverage checks accept intentionally deduplicated metadata.
func resolveOpenAPITestMessage(bundle map[string]string, key string) string {
	if value := strings.TrimSpace(bundle[key]); value != "" {
		return value
	}
	for _, fallbackKey := range openAPICommonFallbackKeys(key) {
		if value := strings.TrimSpace(bundle[fallbackKey]); value != "" {
			return value
		}
	}
	return ""
}

// TestOpenAPIBundlesAreSeparatedFromRuntimeI18n verifies apidoc source-string
// keys are not stored in runtime UI translation bundles.
func TestOpenAPIBundlesAreSeparatedFromRuntimeI18n(t *testing.T) {
	repoRoot := locateRepositoryRoot(t)
	apiBundle := readOpenAPIJSONBundle(t, filepath.Join(repoRoot, "apps/lina-core/manifest/i18n/apidoc/zh-CN.json"))
	for _, bundle := range readOpenAPIPluginJSONBundles(t, repoRoot, "zh-CN") {
		mergeOpenAPIMessageCatalog(apiBundle, bundle)
	}
	apiKeys := make(map[string]struct{}, len(apiBundle))
	for key := range apiBundle {
		apiKeys[key] = struct{}{}
	}

	var mixed []string
	for _, root := range []string{
		filepath.Join(repoRoot, "apps/lina-core/manifest/i18n"),
		filepath.Join(repoRoot, "apps/lina-core/internal/packed/manifest/i18n"),
		filepath.Join(repoRoot, "apps/lina-plugins"),
	} {
		if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".json") {
				return nil
			}
			normalizedPath := filepath.ToSlash(path)
			if !strings.Contains(normalizedPath, "/manifest/i18n/") {
				return nil
			}
			if strings.Contains(normalizedPath, "/manifest/i18n/apidoc/") {
				return nil
			}
			runtimeBundle := readOpenAPIJSONBundle(t, path)
			for key := range runtimeBundle {
				if _, ok := apiKeys[key]; ok {
					mixed = append(mixed, normalizedPath+": "+key)
				}
			}
			return nil
		}); err != nil {
			t.Fatalf("scan runtime i18n root %s failed: %v", root, err)
		}
	}
	if len(mixed) > 0 {
		t.Fatalf("apidoc source-string keys must stay out of runtime i18n bundles:\n%s", strings.Join(limitStrings(mixed, 20), "\n"))
	}
}

// TestOpenAPIHostCatalogMergeIgnoresGeneratedEntityKeys ensures generated
// entity metadata is never translated through host apidoc JSON resources.
func TestOpenAPIHostCatalogMergeIgnoresGeneratedEntityKeys(t *testing.T) {
	target := map[string]string{}
	mergeOpenAPIMessageCatalog(target, map[string]string{
		"core.api.auth.v1.LoginReq.meta.summary":             "用户登录",
		"core.internal.model.entity.SysConfig.fields.id.dc":  "参数ID",
		"core.internal.model.entity.SysConfig.fields.key.dc": "参数键名",
	})

	if got := target["core.api.auth.v1.LoginReq.meta.summary"]; got != "用户登录" {
		t.Fatalf("expected hand-authored API key to merge, got %q", got)
	}
	if _, ok := target["core.internal.model.entity.SysConfig.fields.id.dc"]; ok {
		t.Fatal("expected generated entity id key to be ignored")
	}
	if _, ok := target["core.internal.model.entity.SysConfig.fields.key.dc"]; ok {
		t.Fatal("expected generated entity key key to be ignored")
	}
}

// TestOpenAPIPluginCatalogMergeRejectsForeignNamespaces ensures plugin-owned
// bundles cannot override host or sibling-plugin documentation strings at
// runtime even if a malformed bundle is supplied.
func TestOpenAPIPluginCatalogMergeRejectsForeignNamespaces(t *testing.T) {
	target := map[string]string{}
	mergeOpenAPIPluginMessageCatalog(context.Background(), target, "plugin-demo-source", map[string]string{
		"plugins.plugin_demo_source.api.demo.v1.PingReq.meta.summary":                  "查询源码插件示例公开 ping",
		"plugins.plugin_demo_source.backend.internal.model.entity.Record.fields.id.dc": "不应合并",
		"plugins.other_plugin.api.demo.v1.PingReq.meta.summary":                        "不应合并",
		"core.openapi.info.title": "不应合并",
	})

	if got := target["plugins.plugin_demo_source.api.demo.v1.PingReq.meta.summary"]; got != "查询源码插件示例公开 ping" {
		t.Fatalf("expected plugin-owned apidoc key to merge, got %q", got)
	}
	if _, ok := target["plugins.plugin_demo_source.backend.internal.model.entity.Record.fields.id.dc"]; ok {
		t.Fatal("expected generated entity apidoc key from plugin bundle to be ignored")
	}
	if _, ok := target["plugins.other_plugin.api.demo.v1.PingReq.meta.summary"]; ok {
		t.Fatal("expected sibling-plugin apidoc key to be ignored")
	}
	if _, ok := target["core.openapi.info.title"]; ok {
		t.Fatal("expected host apidoc key from plugin bundle to be ignored")
	}
}

// TestOpenAPIBundleLoaderSupportsNestedSplitFiles verifies apidoc bundles can
// be maintained as one locale root file plus multiple nested JSON files.
func TestOpenAPIBundleLoaderSupportsNestedSplitFiles(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"manifest/i18n/apidoc/zh-CN.json": &fstest.MapFile{Data: []byte(`{
  "core": {
    "openapi": {
      "info": {
        "title": "根文件标题"
      }
    }
  }
}`)},
		"manifest/i18n/apidoc/zh-CN/core-api-auth.json": &fstest.MapFile{Data: []byte(`{
  "core": {
    "api": {
      "auth": {
        "v1": {
          "LoginReq": {
            "meta": {
              "summary": "用户登录"
            }
          }
        }
      }
    }
  }
}`)},
		"manifest/i18n/apidoc/zh-CN/override.json": &fstest.MapFile{Data: []byte(`{
  "core.openapi.info.title": "扁平覆盖标题"
}`)},
	}

	bundle := loadOpenAPIEmbeddedBundle(context.Background(), filesystem, "manifest/i18n/apidoc", "zh-CN")
	if got := bundle["core.openapi.info.title"]; got != "扁平覆盖标题" {
		t.Fatalf("expected split flat key to override root nested key, got %q", got)
	}
	if got := bundle["core.api.auth.v1.LoginReq.meta.summary"]; got != "用户登录" {
		t.Fatalf("expected nested split file key to load, got %q", got)
	}
}

// TestOpenAPICommonFallbackKeys verifies generated wrapper metadata can share
// common apidoc translations without repeating the same exact key per API.
func TestOpenAPICommonFallbackKeys(t *testing.T) {
	t.Parallel()

	localizer := &openAPILocalizer{
		catalog: map[string]string{
			"core.common.responses.fields.code.dc": "错误码",
			"core.common.schemas.response.dc":      "按接口定义返回的结果数据",
			"core.common.fields.pageNum.dc":        "页码",
		},
	}

	if got := localizer.translate("core.api.user.v1.ListReq.responses.200.content.application_json.fields.code.dc", "Code"); got != "错误码" {
		t.Fatalf("expected standard response code fallback, got %q", got)
	}
	if got := localizer.translate("core.api.user.v1.ListRes.schema.dc", "Response data as defined by the API contract"); got != "按接口定义返回的结果数据" {
		t.Fatalf("expected response schema fallback, got %q", got)
	}
	if got := localizer.translate("core.api.user.v1.ListReq.fields.pageNum.dc", "Page number"); got != "页码" {
		t.Fatalf("expected pageNum fallback, got %q", got)
	}
}

// openAPIMetadataValue records one extracted metadata string and its source
// location for actionable failure messages.
type openAPIMetadataValue struct {
	Key      string
	Value    string
	Location string
}

// collectOpenAPISourceMetadataStrings scans API DTOs and metadata config values
// that are expected to use readable English as their source text.
func collectOpenAPISourceMetadataStrings(t *testing.T) []openAPIMetadataValue {
	t.Helper()

	repoRoot := locateRepositoryRoot(t)
	scanRoots := []string{
		filepath.Join(repoRoot, "apps/lina-core/api"),
		filepath.Join(repoRoot, "apps/lina-plugins"),
		filepath.Join(repoRoot, "apps/lina-core/manifest/config"),
	}
	packedConfigRoot := filepath.Join(repoRoot, "apps/lina-core/internal/packed/manifest/config")
	if _, err := os.Stat(packedConfigRoot); err == nil {
		scanRoots = append(scanRoots, packedConfigRoot)
	}
	return collectOpenAPIMetadataValues(t, scanRoots, shouldScanOpenAPISourceMetadataFile)
}

// collectOpenAPITranslatableStructuredKeys scans hand-authored host and plugin
// API DTO source files and returns the stable apidoc keys expected in
// non-English translation bundles.
func collectOpenAPITranslatableStructuredKeys(t *testing.T) []string {
	t.Helper()

	repoRoot := locateRepositoryRoot(t)
	keySet := make(map[string]struct{})
	scanRoots := []string{
		filepath.Join(repoRoot, "apps/lina-core/api"),
		filepath.Join(repoRoot, "apps/lina-plugins"),
	}
	for _, root := range scanRoots {
		if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !shouldScanOpenAPISourceMetadataFile(path) || !strings.HasSuffix(path, ".go") {
				return nil
			}
			fileKeys, err := collectOpenAPITranslatableStructuredKeysFromFile(repoRoot, path)
			if err != nil {
				return err
			}
			for _, key := range fileKeys {
				keySet[key] = struct{}{}
			}
			return nil
		}); err != nil {
			t.Fatalf("scan OpenAPI structured keys root %s failed: %v", root, err)
		}
	}

	keys := make([]string, 0, len(keySet))
	for key := range keySet {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// collectOpenAPITranslatableStructuredKeysFromFile parses one API source file
// and maps route metadata and DTO field descriptions to stable apidoc keys.
func collectOpenAPITranslatableStructuredKeysFromFile(repoRoot string, path string) ([]string, error) {
	componentPrefix := openAPIComponentPrefixForAPIFile(repoRoot, path)
	if componentPrefix == "" {
		return nil, nil
	}

	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		return nil, err
	}

	var keys []string
	ast.Inspect(parsedFile, func(node ast.Node) bool {
		typeSpec, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return false
		}
		componentKey := componentPrefix + "." + typeSpec.Name.Name
		for _, field := range structType.Fields.List {
			if field.Tag == nil {
				continue
			}
			tag := parseOpenAPIASTStructTag(field.Tag.Value)
			if strings.TrimSpace(tag.Get("tags")) != "" {
				keys = append(keys, componentKey+".meta.tags")
			}
			if strings.TrimSpace(tag.Get("summary")) != "" {
				keys = append(keys, componentKey+".meta.summary")
			}
			if strings.TrimSpace(tag.Get("dc")) != "" && strings.TrimSpace(tag.Get("path")) != "" {
				keys = append(keys, componentKey+".meta.dc", componentKey+".schema.dc")
			}
			if strings.TrimSpace(tag.Get("dc")) == "" && strings.TrimSpace(tag.Get("description")) == "" {
				continue
			}
			jsonName := openAPIJSONNameFromStructTag(tag)
			if jsonName == "" {
				continue
			}
			keys = append(keys, buildOpenAPIFieldKey(componentKey+".schema", jsonName)+".dc")
		}
		return false
	})
	return keys, nil
}

// openAPIComponentPrefixForAPIFile converts an API source file path into the
// component-key prefix used by GoFrame schema names.
func openAPIComponentPrefixForAPIFile(repoRoot string, path string) string {
	relativePath, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return ""
	}
	normalizedPath := filepath.ToSlash(relativePath)
	packagePath := filepath.ToSlash(filepath.Dir(normalizedPath))
	switch {
	case strings.HasPrefix(normalizedPath, "apps/lina-core/api/"):
		packagePath = strings.TrimPrefix(packagePath, "apps/lina-core/")
		return "core." + sanitizeOpenAPIKey(strings.ReplaceAll(packagePath, "/", "."))
	case strings.HasPrefix(normalizedPath, "apps/lina-plugins/"):
		parts := strings.Split(normalizedPath, "/")
		if len(parts) < 6 || parts[0] != "apps" || parts[1] != "lina-plugins" || parts[3] != "backend" {
			return ""
		}
		pluginID := sanitizeOpenAPIKeyPart(parts[2])
		pluginPackagePath := strings.Join(parts[4:len(parts)-1], ".")
		if !strings.HasPrefix(pluginPackagePath, "api.") {
			return ""
		}
		return "plugins." + pluginID + "." + sanitizeOpenAPIKey(pluginPackagePath)
	default:
		return ""
	}
}

// parseOpenAPIASTStructTag decodes a raw AST struct tag literal.
func parseOpenAPIASTStructTag(raw string) reflect.StructTag {
	decoded, err := strconv.Unquote(raw)
	if err != nil {
		return reflect.StructTag(strings.Trim(raw, "`"))
	}
	return reflect.StructTag(decoded)
}

// openAPIJSONNameFromStructTag returns the public field name used in OpenAPI
// schemas and parameters.
func openAPIJSONNameFromStructTag(tag reflect.StructTag) string {
	jsonName := strings.TrimSpace(tag.Get("json"))
	if jsonName == "" {
		jsonName = strings.TrimSpace(tag.Get("p"))
	}
	if name, _, ok := strings.Cut(jsonName, ","); ok {
		jsonName = name
	}
	if jsonName == "-" {
		return ""
	}
	return strings.TrimSpace(jsonName)
}

// collectGeneratedEntityDescriptionStrings scans generated entity descriptions
// that can still be projected into schemas but are not hand-authored API DTO
// source text.
func collectGeneratedEntityDescriptionStrings(t *testing.T) []openAPIMetadataValue {
	t.Helper()

	repoRoot := locateRepositoryRoot(t)
	scanRoots := []string{
		filepath.Join(repoRoot, "apps/lina-core/internal/model/entity"),
		filepath.Join(repoRoot, "apps/lina-plugins"),
	}
	return collectOpenAPIMetadataValues(t, scanRoots, shouldScanGeneratedEntityMetadataFile)
}

// collectOpenAPIMetadataValues walks selected roots and extracts OpenAPI
// metadata values from Go struct tags and YAML metadata files.
func collectOpenAPIMetadataValues(
	t *testing.T,
	scanRoots []string,
	shouldScan func(string) bool,
) []openAPIMetadataValue {
	t.Helper()

	values := make(map[string]openAPIMetadataValue)
	for _, root := range scanRoots {
		if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !shouldScan(path) {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for _, value := range extractOpenAPIMetadataStrings(path, string(content)) {
				if strings.TrimSpace(value.Value) == "" {
					continue
				}
				values[value.Key+"\x00"+value.Value] = value
			}
			return nil
		}); err != nil {
			t.Fatalf("scan OpenAPI metadata root %s failed: %v", root, err)
		}
	}

	items := make([]openAPIMetadataValue, 0, len(values))
	for _, value := range values {
		items = append(items, value)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Value == items[j].Value {
			return items[i].Location < items[j].Location
		}
		return items[i].Value < items[j].Value
	})
	return items
}

// locateRepositoryRoot resolves the monorepo root from this test file path so
// the coverage scan is independent of the current go test working directory.
func locateRepositoryRoot(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve current test file path failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../../../../.."))
}

// shouldScanOpenAPISourceMetadataFile reports whether one file can contribute
// hand-authored API source metadata.
func shouldScanOpenAPISourceMetadataFile(path string) bool {
	normalizedPath := filepath.ToSlash(path)
	switch {
	case strings.HasSuffix(normalizedPath, ".go"):
		return strings.Contains(normalizedPath, "/api/")
	case strings.HasSuffix(normalizedPath, ".yaml"), strings.HasSuffix(normalizedPath, ".yml"):
		return strings.Contains(normalizedPath, "/manifest/config/")
	default:
		return false
	}
}

// shouldScanGeneratedEntityMetadataFile reports whether one file can contribute
// generated entity schema descriptions.
func shouldScanGeneratedEntityMetadataFile(path string) bool {
	normalizedPath := filepath.ToSlash(path)
	return strings.HasSuffix(normalizedPath, ".go") &&
		strings.Contains(normalizedPath, "/internal/model/entity/")
}

// extractOpenAPIMetadataStrings extracts metadata values from Go struct tags or
// metadata YAML lines that can be projected into the API document.
func extractOpenAPIMetadataStrings(path string, content string) []openAPIMetadataValue {
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		return extractOpenAPIYAMLMetadataStrings(path, content)
	}
	return extractOpenAPIGoMetadataStrings(path, content)
}

// extractOpenAPIGoMetadataStrings extracts route, schema, example, and
// validation-message text from Go struct tags.
func extractOpenAPIGoMetadataStrings(path string, content string) []openAPIMetadataValue {
	scannedTagKeys := map[string]struct{}{
		"brief":       {},
		"dc":          {},
		"description": {},
		"eg":          {},
		"example":     {},
		"name":        {},
		"summary":     {},
		"tags":        {},
	}
	var values []openAPIMetadataValue
	for _, match := range openAPIMetadataTagPattern.FindAllStringSubmatch(content, -1) {
		key := match[1]
		value := decodeOpenAPITagValue(match[2])
		if _, ok := scannedTagKeys[key]; ok {
			values = append(values, openAPIMetadataValue{
				Key:      key,
				Value:    value,
				Location: filepath.ToSlash(path) + ":" + key,
			})
		}
	}
	return values
}

// decodeOpenAPITagValue mirrors Go struct-tag unescaping so the catalog
// coverage test checks the actual strings that GoFrame sees at runtime.
func decodeOpenAPITagValue(raw string) string {
	value, err := strconv.Unquote(`"` + raw + `"`)
	if err != nil {
		return raw
	}
	return value
}

// extractOpenAPIYAMLMetadataStrings extracts metadata fields from the host
// packaged metadata config.
func extractOpenAPIYAMLMetadataStrings(path string, content string) []openAPIMetadataValue {
	var values []openAPIMetadataValue
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		for _, key := range []string{"description", "serverDescription", "title"} {
			if !strings.HasPrefix(trimmed, key+":") {
				continue
			}
			_, value, ok := strings.Cut(trimmed, ":")
			if ok {
				values = append(values, openAPIMetadataValue{
					Key:      key,
					Value:    strings.Trim(strings.TrimSpace(value), `"'`),
					Location: filepath.ToSlash(path) + ":" + key,
				})
			}
		}
	}
	return values
}

// isOpaqueOpenAPIPlaceholder reports whether a documentation tag appears to be
// an i18n key rather than readable English source text.
func isOpaqueOpenAPIPlaceholder(value openAPIMetadataValue) bool {
	if isOpenAPIExampleMetadata(value) {
		return false
	}
	trimmed := strings.TrimSpace(value.Value)
	if strings.Contains(trimmed, " ") || strings.Contains(trimmed, "/") {
		return false
	}
	return regexp.MustCompile(`^[a-z][a-z0-9_-]*(\.[a-z0-9_-]+){2,}$`).MatchString(trimmed)
}

// readOpenAPIJSONBundle reads a JSON translation bundle from disk.
func readOpenAPIJSONBundle(t *testing.T, path string) map[string]string {
	t.Helper()

	bundle := make(map[string]string)
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("read i18n bundle %s failed: %v", path, err)
	}
	if err == nil {
		parsedBundle, parseErr := parseOpenAPIMessageCatalogJSON(content)
		if parseErr != nil {
			t.Fatalf("parse i18n bundle %s failed: %v", path, parseErr)
		}
		mergeOpenAPIMessageCatalog(bundle, parsedBundle)
	}

	if locale := strings.TrimSuffix(filepath.Base(path), ".json"); locale != filepath.Base(path) {
		localeDir := filepath.Join(filepath.Dir(path), locale)
		if _, statErr := os.Stat(localeDir); statErr == nil {
			entries := make([]string, 0)
			if walkErr := filepath.WalkDir(localeDir, func(filePath string, entry os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if entry.IsDir() || !strings.HasSuffix(filePath, ".json") {
					return nil
				}
				entries = append(entries, filePath)
				return nil
			}); walkErr != nil {
				t.Fatalf("scan i18n bundle directory %s failed: %v", localeDir, walkErr)
			}
			sort.Strings(entries)
			for _, entryPath := range entries {
				mergeOpenAPIMessageCatalog(bundle, readOpenAPIJSONBundle(t, entryPath))
			}
		} else if !os.IsNotExist(statErr) {
			t.Fatalf("stat i18n bundle directory %s failed: %v", localeDir, statErr)
		}
	}
	return bundle
}

// readOpenAPIPluginJSONBundles reads plugin-owned apidoc bundles for one
// locale from each plugin's manifest/i18n/apidoc directory.
func readOpenAPIPluginJSONBundles(t *testing.T, repoRoot string, locale string) map[string]map[string]string {
	t.Helper()

	result := make(map[string]map[string]string)
	pluginsRoot := filepath.Join(repoRoot, "apps/lina-plugins")
	entries, err := os.ReadDir(pluginsRoot)
	if err != nil {
		t.Fatalf("read plugin root %s failed: %v", pluginsRoot, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pluginID := entry.Name()
		bundlePath := filepath.Join(pluginsRoot, pluginID, "manifest/i18n/apidoc", locale+".json")
		bundleDir := filepath.Join(pluginsRoot, pluginID, "manifest/i18n/apidoc", locale)
		if _, fileErr := os.Stat(bundlePath); fileErr != nil {
			if _, dirErr := os.Stat(bundleDir); dirErr != nil {
				if os.IsNotExist(fileErr) && os.IsNotExist(dirErr) {
					continue
				}
				if !os.IsNotExist(fileErr) {
					t.Fatalf("stat plugin apidoc bundle %s failed: %v", bundlePath, fileErr)
				}
				t.Fatalf("stat plugin apidoc bundle dir %s failed: %v", bundleDir, dirErr)
			}
			if !os.IsNotExist(fileErr) {
				t.Fatalf("stat plugin apidoc bundle %s failed: %v", bundlePath, fileErr)
			}
		}
		if bundle := readOpenAPIJSONBundle(t, bundlePath); len(bundle) > 0 || locale == "en-US" {
			result[pluginID] = bundle
			continue
		}
		if _, dirErr := os.Stat(bundleDir); dirErr == nil {
			result[pluginID] = readOpenAPIJSONBundle(t, bundlePath)
		} else if !os.IsNotExist(dirErr) {
			t.Fatalf("stat plugin apidoc bundle dir %s failed: %v", bundleDir, dirErr)
		} else {
			if _, statErr := os.Stat(bundlePath); statErr != nil {
				continue
			}
		}
	}
	return result
}

// assertOpenAPIEnglishBundlePlaceholder ensures English docs are driven by API
// source metadata instead of duplicating source strings in the apidoc bundle.
func assertOpenAPIEnglishBundlePlaceholder(t *testing.T, bundle map[string]string) {
	t.Helper()

	if len(bundle) > 0 {
		t.Fatalf("apidoc en-US bundle must stay an empty placeholder, got %d keys", len(bundle))
	}
}

// assertOpenAPIPluginEnglishBundlesArePlaceholders ensures plugin en-US apidoc
// bundles follow the same source-text rule as the host bundle.
func assertOpenAPIPluginEnglishBundlesArePlaceholders(t *testing.T, bundles map[string]map[string]string) {
	t.Helper()

	var invalid []string
	for pluginID, bundle := range bundles {
		if len(bundle) > 0 {
			invalid = append(invalid, pluginID)
		}
	}
	if len(invalid) > 0 {
		sort.Strings(invalid)
		t.Fatalf("plugin apidoc en-US bundles must stay empty placeholders:\n%s", strings.Join(invalid, "\n"))
	}
}

// assertOpenAPIBundleUsesStructuredKeys ensures apidoc catalogs never use
// display text as JSON keys.
func assertOpenAPIBundleUsesStructuredKeys(t *testing.T, locale string, bundle map[string]string) {
	t.Helper()

	var invalid []string
	for key, value := range bundle {
		if !isStructuredOpenAPIBundleKey(key) {
			invalid = append(invalid, key)
		}
		if hasOpenAPIExampleKeyPart(key) {
			invalid = append(invalid, key)
		}
		if containsCJK(value) && locale == "en-US" {
			invalid = append(invalid, key+" => "+value)
		}
	}
	if len(invalid) > 0 {
		sort.Strings(invalid)
		t.Fatalf("apidoc %s bundle must use structured non-display keys:\n%s", locale, strings.Join(limitStrings(invalid, 20), "\n"))
	}
}

// assertOpenAPIBundleDoesNotTranslateGeneratedEntityMetadata keeps generated
// DAO/entity and database-source metadata out of apidoc translation resources.
func assertOpenAPIBundleDoesNotTranslateGeneratedEntityMetadata(t *testing.T, locale string, bundle map[string]string) {
	t.Helper()

	var invalid []string
	for key := range bundle {
		if isGeneratedEntityOpenAPIKey(key) {
			invalid = append(invalid, key)
		}
	}
	if len(invalid) > 0 {
		sort.Strings(invalid)
		t.Fatalf("apidoc %s bundle must not translate generated entity metadata:\n%s", locale, strings.Join(limitStrings(invalid, 20), "\n"))
	}
}

// assertOpenAPIHostBundleDoesNotOwnPluginKeys keeps plugin-owned translations
// out of the host apidoc resource.
func assertOpenAPIHostBundleDoesNotOwnPluginKeys(t *testing.T, bundle map[string]string) {
	t.Helper()

	var invalid []string
	for key := range bundle {
		if strings.HasPrefix(key, "plugins.") {
			invalid = append(invalid, key)
		}
	}
	if len(invalid) > 0 {
		sort.Strings(invalid)
		t.Fatalf("host apidoc bundle must not own plugin keys:\n%s", strings.Join(limitStrings(invalid, 20), "\n"))
	}
}

// assertOpenAPIPluginBundlesOwnOnlyPluginKeys ensures each plugin keeps only
// keys under its own stable plugin namespace.
func assertOpenAPIPluginBundlesOwnOnlyPluginKeys(t *testing.T, bundles map[string]map[string]string) {
	t.Helper()

	var invalid []string
	for pluginID, bundle := range bundles {
		prefix := "plugins." + sanitizeOpenAPIKeyPart(pluginID) + "."
		for key := range bundle {
			if !strings.HasPrefix(key, prefix) {
				invalid = append(invalid, pluginID+": "+key)
			}
		}
	}
	if len(invalid) > 0 {
		sort.Strings(invalid)
		t.Fatalf("plugin apidoc bundles must only own their plugin namespace:\n%s", strings.Join(limitStrings(invalid, 20), "\n"))
	}
}

// isOpenAPITranslatableMetadata reports whether a source metadata tag should be
// backed by an apidoc i18n resource. Example values intentionally stay raw.
func isOpenAPITranslatableMetadata(value openAPIMetadataValue) bool {
	if isOpenAPIExampleMetadata(value) {
		return false
	}
	return true
}

// hasOpenAPIExampleKeyPart reports whether a structured key still contains the
// removed example-value translation namespace.
func hasOpenAPIExampleKeyPart(key string) bool {
	for _, part := range strings.Split(key, ".") {
		if part == "eg" || part == "example" || part == "examples" {
			return true
		}
	}
	return false
}

// isOpenAPIExampleMetadata reports whether a struct tag is an example value
// that must stay outside API-documentation localization.
func isOpenAPIExampleMetadata(value openAPIMetadataValue) bool {
	return value.Key == "eg" || value.Key == "example"
}

// isStructuredOpenAPIBundleKey reports whether a key belongs to the stable
// structural apidoc namespace.
func isStructuredOpenAPIBundleKey(key string) bool {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" || strings.ContainsAny(trimmed, " /") || containsCJK(trimmed) {
		return false
	}
	return strings.HasPrefix(trimmed, "core.") || strings.HasPrefix(trimmed, "plugins.")
}

// assertOpenAPIBundlesMirror ensures packed resources stay in sync with the
// editable manifest resources.
func assertOpenAPIBundlesMirror(t *testing.T, locale string, source map[string]string, packed map[string]string) {
	t.Helper()

	var mismatches []string
	for key, sourceValue := range source {
		if packedValue, ok := packed[key]; !ok || packedValue != sourceValue {
			mismatches = append(mismatches, key)
		}
	}
	for key := range packed {
		if _, ok := source[key]; !ok {
			mismatches = append(mismatches, key)
		}
	}
	if len(mismatches) > 0 {
		sort.Strings(mismatches)
		t.Fatalf("packed apidoc %s bundle must mirror manifest bundle:\n%s", locale, strings.Join(limitStrings(mismatches, 20), "\n"))
	}
}

// limitStrings caps long failure reports while still showing actionable
// examples in test output.
func limitStrings(values []string, limit int) []string {
	if len(values) <= limit {
		return values
	}
	return values[:limit]
}
