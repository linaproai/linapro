// This file localizes plugin-runtime display metadata while keeping plugin
// ownership rules inside the runtime package.

package runtime

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
)

// localizePluginMetadata returns localized plugin name and description values.
func (s *serviceImpl) localizePluginMetadata(
	ctx context.Context,
	id string,
	pluginType string,
	name string,
	description string,
) (string, string) {
	if s == nil || s.i18nSvc == nil {
		return name, description
	}
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return name, description
	}
	nameKey := "plugin." + trimmedID + ".name"
	descriptionKey := "plugin." + trimmedID + ".description"
	if catalog.NormalizeType(pluginType) == catalog.TypeDynamic {
		localizedName := s.i18nSvc.TranslateDynamicPluginSourceText(ctx, trimmedID, nameKey, name)
		localizedDescription := s.i18nSvc.TranslateDynamicPluginSourceText(ctx, trimmedID, descriptionKey, description)
		return localizedName, localizedDescription
	}
	localizedName := s.i18nSvc.Translate(ctx, nameKey, name)
	localizedDescription := s.i18nSvc.Translate(ctx, descriptionKey, description)
	return localizedName, localizedDescription
}
