// Package i18nresource loads host and plugin-owned i18n JSON resources without
// coupling callers to the runtime i18n service package.
package i18nresource

import (
	"context"
	"encoding/json"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/pkg/logger"
)

// LayoutMode controls which locale resource layout is loaded from one filesystem.
type LayoutMode string

const (
	// LayoutModeLocaleFile loads only `<subdir>/<locale>.json`.
	LayoutModeLocaleFile LayoutMode = "locale-file"
	// LayoutModeLocaleFileAndDirectory loads `<subdir>/<locale>.json` plus
	// every JSON file under `<subdir>/<locale>/`, with directory files applied
	// after the root file in lexical path order.
	LayoutModeLocaleFileAndDirectory LayoutMode = "locale-file-and-directory"
)

// PluginScope controls whether plugin resource keys are namespace restricted.
type PluginScope string

const (
	// PluginScopeOpen allows plugin resources to contribute any runtime key.
	PluginScopeOpen PluginScope = "open"
	// PluginScopeRestrictedToPluginNamespace only allows keys under
	// `plugins.<sanitizedPluginID>.`.
	PluginScopeRestrictedToPluginNamespace PluginScope = "restricted-to-plugin-namespace"
)

// ValueMode controls how JSON leaf values are converted into flat catalog values.
type ValueMode string

const (
	// ValueModeStringifyScalars stringifies scalar JSON leaf values.
	ValueModeStringifyScalars ValueMode = "stringify-scalars"
	// ValueModeStringOnly rejects non-string JSON leaf values.
	ValueModeStringOnly ValueMode = "string-only"
)

// SourcePlugin describes the source-plugin metadata needed to load embedded i18n resources.
type SourcePlugin interface {
	// ID returns the stable plugin identifier.
	ID() string
	// GetEmbeddedFiles returns the plugin-owned embedded filesystem.
	GetEmbeddedFiles() fs.FS
}

// LocaleAsset stores one already-extracted dynamic plugin i18n asset.
type LocaleAsset struct {
	Locale  string // Locale is the asset locale code.
	Content string // Content is one JSON locale bundle.
}

// ReleaseRef stores one dynamic-plugin release's already-extracted locale assets.
type ReleaseRef struct {
	PluginID string        // PluginID is the stable plugin identifier.
	Assets   []LocaleAsset // Assets stores locale bundle snapshots.
}

// KeyFilter decides whether one flat key should remain in the loaded catalog.
type KeyFilter func(key string) bool

// ResourceLoader loads host, source-plugin, and dynamic-plugin locale resources.
type ResourceLoader struct {
	HostFS        fs.FS                 // HostFS stores host-owned embedded resources.
	SourcePlugins func() []SourcePlugin // SourcePlugins returns registered source plugins.
	Subdir        string                // Subdir is the slash-separated locale resource directory.
	PluginScope   PluginScope           // PluginScope restricts plugin-owned keys when needed.
	LayoutMode    LayoutMode            // LayoutMode selects the supported filesystem layout.
	ValueMode     ValueMode             // ValueMode selects JSON scalar conversion behavior.
	KeyFilter     KeyFilter             // KeyFilter optionally removes disallowed flat keys.
}

// LoadHostBundle loads one host-owned locale bundle.
func (l ResourceLoader) LoadHostBundle(ctx context.Context, locale string) map[string]string {
	return l.loadFilesystemBundle(ctx, l.HostFS, locale)
}

// LoadSourcePluginBundles loads locale bundles from registered source plugins.
func (l ResourceLoader) LoadSourcePluginBundles(ctx context.Context, locale string) map[string]map[string]string {
	bundles := make(map[string]map[string]string)
	rawSourcePlugins := l.sourcePlugins()
	if len(rawSourcePlugins) == 0 {
		return bundles
	}

	sourcePlugins := make([]SourcePlugin, 0, len(rawSourcePlugins))
	for _, sourcePlugin := range rawSourcePlugins {
		if sourcePlugin == nil || sourcePlugin.GetEmbeddedFiles() == nil || strings.TrimSpace(sourcePlugin.ID()) == "" {
			continue
		}
		sourcePlugins = append(sourcePlugins, sourcePlugin)
	}
	sort.Slice(sourcePlugins, func(i, j int) bool {
		return sourcePlugins[i].ID() < sourcePlugins[j].ID()
	})
	for _, sourcePlugin := range sourcePlugins {
		pluginID := strings.TrimSpace(sourcePlugin.ID())
		bundle := l.filterPluginBundle(ctx, pluginID, l.loadFilesystemBundle(ctx, sourcePlugin.GetEmbeddedFiles(), locale))
		if len(bundle) > 0 {
			bundles[pluginID] = bundle
		}
	}
	return bundles
}

// LoadDynamicPluginBundles loads locale bundles from already-extracted dynamic plugin release assets.
func (l ResourceLoader) LoadDynamicPluginBundles(ctx context.Context, locale string, releases []ReleaseRef) map[string]map[string]string {
	bundles := make(map[string]map[string]string)
	normalizedLocale := strings.TrimSpace(locale)
	for _, release := range releases {
		pluginID := strings.TrimSpace(release.PluginID)
		if pluginID == "" {
			continue
		}
		pluginBundle := make(map[string]string)
		for _, asset := range release.Assets {
			if strings.TrimSpace(asset.Locale) != normalizedLocale || strings.TrimSpace(asset.Content) == "" {
				continue
			}
			assetBundle, err := ParseCatalog([]byte(asset.Content), l.valueMode())
			if err != nil {
				logger.Warningf(ctx, "parse dynamic plugin i18n resource failed plugin=%s locale=%s err=%v", pluginID, locale, err)
				continue
			}
			mergeCatalog(pluginBundle, assetBundle)
		}
		pluginBundle = l.filterPluginBundle(ctx, pluginID, pluginBundle)
		if len(pluginBundle) > 0 {
			bundles[pluginID] = pluginBundle
		}
	}
	return bundles
}

// ParseCatalog parses one JSON locale resource into a flat dotted-key catalog.
func ParseCatalog(content []byte, valueMode ValueMode) (map[string]string, error) {
	result := make(map[string]interface{})
	if len(content) == 0 {
		return map[string]string{}, nil
	}
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, err
	}

	flatMessages := make(map[string]string)
	if err := flattenCatalogValue("", result, flatMessages, normalizeValueMode(valueMode)); err != nil {
		return nil, err
	}
	return flatMessages, nil
}

// SanitizeKeyPart normalizes one dotted-key segment for structured i18n keys.
func SanitizeKeyPart(part string) string {
	trimmedPart := strings.TrimSpace(part)
	trimmedPart = strings.Trim(trimmedPart, "{}")
	trimmedPart = strings.ReplaceAll(trimmedPart, "-", "_")
	trimmedPart = keyInvalidCharsPattern.ReplaceAllString(trimmedPart, "_")
	trimmedPart = strings.Trim(trimmedPart, "_")
	if trimmedPart == "" {
		return "empty"
	}
	return trimmedPart
}

// loadFilesystemBundle loads one locale bundle from a filesystem using the loader layout.
func (l ResourceLoader) loadFilesystemBundle(ctx context.Context, filesystem fs.FS, locale string) map[string]string {
	if filesystem == nil {
		return map[string]string{}
	}
	bundle := make(map[string]string)
	mergeCatalog(bundle, l.loadFilesystemBundleFile(ctx, filesystem, path.Join(l.subdir(), strings.TrimSpace(locale)+".json")))

	if l.layoutMode() != LayoutModeLocaleFileAndDirectory {
		return bundle
	}
	localeDir := path.Join(l.subdir(), strings.TrimSpace(locale))
	entries := make([]string, 0)
	if err := fs.WalkDir(filesystem, localeDir, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if filePath == localeDir {
				return nil
			}
			return walkErr
		}
		if entry == nil || entry.IsDir() || !strings.HasSuffix(filePath, ".json") {
			return nil
		}
		entries = append(entries, filePath)
		return nil
	}); err != nil {
		logger.Warningf(ctx, "scan i18n resource directory failed locale=%s dir=%s err=%v", locale, localeDir, err)
		return bundle
	}
	sort.Strings(entries)
	for _, entryPath := range entries {
		mergeCatalog(bundle, l.loadFilesystemBundleFile(ctx, filesystem, entryPath))
	}
	return bundle
}

// loadFilesystemBundleFile loads one JSON file from a filesystem.
func (l ResourceLoader) loadFilesystemBundleFile(ctx context.Context, filesystem fs.FS, filePath string) map[string]string {
	content, err := fs.ReadFile(filesystem, filePath)
	if err != nil || len(content) == 0 {
		return map[string]string{}
	}
	bundle, err := ParseCatalog(content, l.valueMode())
	if err != nil {
		logger.Warningf(ctx, "parse i18n resource failed path=%s err=%v", filePath, err)
		return map[string]string{}
	}
	return l.filterCatalog(ctx, "", bundle)
}

// sourcePlugins returns the registered source plugins from the configured provider.
func (l ResourceLoader) sourcePlugins() []SourcePlugin {
	if l.SourcePlugins == nil {
		return []SourcePlugin{}
	}
	return l.SourcePlugins()
}

// filterPluginBundle applies namespace and key filters to one plugin-owned bundle.
func (l ResourceLoader) filterPluginBundle(ctx context.Context, pluginID string, source map[string]string) map[string]string {
	return l.filterCatalog(ctx, pluginID, source)
}

// filterCatalog applies the loader's plugin scope and key filter.
func (l ResourceLoader) filterCatalog(ctx context.Context, pluginID string, source map[string]string) map[string]string {
	if len(source) == 0 {
		return map[string]string{}
	}
	target := make(map[string]string, len(source))
	namespacePrefix := ""
	if l.pluginScope() == PluginScopeRestrictedToPluginNamespace && strings.TrimSpace(pluginID) != "" {
		namespacePrefix = "plugins." + SanitizeKeyPart(pluginID) + "."
	}
	for key, value := range source {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		if namespacePrefix != "" && !strings.HasPrefix(trimmedKey, namespacePrefix) {
			logger.Warningf(ctx, "ignore i18n resource key outside plugin namespace plugin=%s key=%s", pluginID, trimmedKey)
			continue
		}
		if l.KeyFilter != nil && !l.KeyFilter(trimmedKey) {
			logger.Warningf(ctx, "ignore i18n resource key rejected by loader filter plugin=%s key=%s", pluginID, trimmedKey)
			continue
		}
		target[trimmedKey] = value
	}
	return target
}

// flattenCatalogValue flattens one nested JSON value into dotted keys.
func flattenCatalogValue(prefix string, value interface{}, output map[string]string, valueMode ValueMode) error {
	switch typedValue := value.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(typedValue))
		for key := range typedValue {
			if strings.TrimSpace(key) == "" {
				continue
			}
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			left := strings.TrimSpace(keys[i])
			right := strings.TrimSpace(keys[j])
			if left == right {
				return keys[i] < keys[j]
			}
			return left < right
		})
		for _, key := range keys {
			trimmedKey := strings.TrimSpace(key)
			nextPrefix := trimmedKey
			if prefix != "" {
				nextPrefix = prefix + "." + trimmedKey
			}
			if err := flattenCatalogValue(nextPrefix, typedValue[key], output, valueMode); err != nil {
				return err
			}
		}
	case string:
		if prefix != "" {
			output[prefix] = typedValue
		}
	default:
		if prefix == "" {
			return nil
		}
		if valueMode == ValueModeStringOnly {
			return gerror.New("i18n resource values must be strings or objects")
		}
		output[prefix] = gconv.String(typedValue)
	}
	return nil
}

// mergeCatalog merges source values into target values.
func mergeCatalog(target map[string]string, source map[string]string) {
	for key, value := range source {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		target[trimmedKey] = value
	}
}

// subdir returns the normalized resource subdirectory.
func (l ResourceLoader) subdir() string {
	return strings.Trim(strings.TrimSpace(l.Subdir), "/")
}

// layoutMode returns the configured layout mode with a file-only default.
func (l ResourceLoader) layoutMode() LayoutMode {
	if l.LayoutMode == "" {
		return LayoutModeLocaleFile
	}
	return l.LayoutMode
}

// pluginScope returns the configured plugin scope with an open default.
func (l ResourceLoader) pluginScope() PluginScope {
	if l.PluginScope == "" {
		return PluginScopeOpen
	}
	return l.PluginScope
}

// valueMode returns the configured value mode with a stringify default.
func (l ResourceLoader) valueMode() ValueMode {
	return normalizeValueMode(l.ValueMode)
}

// normalizeValueMode returns the configured value mode with a stringify default.
func normalizeValueMode(valueMode ValueMode) ValueMode {
	if valueMode == "" {
		return ValueModeStringifyScalars
	}
	return valueMode
}

var keyInvalidCharsPattern = regexp.MustCompile(`[^A-Za-z0-9_]+`)
