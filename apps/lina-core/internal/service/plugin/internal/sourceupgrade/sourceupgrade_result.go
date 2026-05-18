// This file contains source-plugin upgrade result message localization helpers.

package sourceupgrade

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
)

// setSourceUpgradeResultMessage stores a stable message key, parameters, and
// localized text on one source-plugin upgrade result.
func setSourceUpgradeResultMessage(
	ctx context.Context,
	translator sourceUpgradeI18nService,
	result *SourceUpgradeResult,
	messageKey string,
	fallback string,
	params map[string]any,
) {
	if result == nil {
		return
	}
	result.MessageKey = strings.TrimSpace(messageKey)
	result.MessageParams = cloneSourceUpgradeMessageParams(params)

	template := strings.TrimSpace(fallback)
	if translator != nil && result.MessageKey != "" {
		template = translator.Translate(ctx, result.MessageKey, fallback)
	}
	result.Message = bizerr.Format(template, params)
}

// cloneSourceUpgradeMessageParams copies message parameters before exposing
// them on the stable source-upgrade result contract.
func cloneSourceUpgradeMessageParams(params map[string]any) map[string]any {
	if len(params) == 0 {
		return nil
	}
	cloned := make(map[string]any, len(params))
	for key, value := range params {
		cloned[key] = value
	}
	return cloned
}
