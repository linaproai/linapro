// This file localizes plugin-runtime display metadata while keeping plugin
// ownership rules inside the runtime package.

package runtime

import (
	"context"
	"strings"
)

// localizePluginMetadata returns localized plugin name and description values.
func (s *serviceImpl) localizePluginMetadata(
	ctx context.Context,
	id string,
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
	localizedName := s.i18nSvc.Translate(ctx, "plugin."+trimmedID+".name", name)
	localizedDescription := s.i18nSvc.Translate(ctx, "plugin."+trimmedID+".description", description)
	return localizedName, localizedDescription
}
