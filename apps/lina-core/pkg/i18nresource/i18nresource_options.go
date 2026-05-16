// This file implements ResourceLoader option defaults and path normalization helpers.

package i18nresource

import "strings"

// subdir returns the normalized resource subdirectory.
func (l ResourceLoader) subdir() string {
	return strings.Trim(strings.TrimSpace(l.Subdir), "/")
}

// localeSubdir returns the normalized locale-scoped subdirectory.
func (l ResourceLoader) localeSubdir() string {
	return strings.Trim(strings.TrimSpace(l.LocaleSubdir), "/")
}

// layoutMode returns the configured layout mode with a locale-directory default.
func (l ResourceLoader) layoutMode() LayoutMode {
	if l.LayoutMode == "" {
		return LayoutModeLocaleDirectory
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
