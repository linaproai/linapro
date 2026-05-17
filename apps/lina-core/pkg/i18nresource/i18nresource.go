// Package i18nresource loads host and plugin-owned i18n JSON resources without
// coupling callers to the runtime i18n service package.
package i18nresource

import "io/fs"

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
	// ID returns the stable plugin identifier used to namespace plugin-owned
	// translation keys and diagnostic logs. Empty identifiers are ignored by the
	// loader.
	ID() string
	// GetEmbeddedFiles returns the plugin-owned embedded filesystem containing
	// locale JSON resources. Nil filesystems are skipped; the loader only reads
	// resources and does not mutate plugin manifests or runtime cache state.
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
	LocaleSubdir  string                // LocaleSubdir optionally narrows loading to a child directory under the locale.
	Recursive     bool                  // Recursive scans JSON files below the target directory recursively.
	PluginScope   PluginScope           // PluginScope restricts plugin-owned keys when needed.
	ValueMode     ValueMode             // ValueMode selects JSON scalar conversion behavior.
	KeyFilter     KeyFilter             // KeyFilter optionally removes disallowed flat keys.
}
