// demo_summary.go implements the summary payload for the demo service.

package demo

import (
	"context"
	"strings"
)

// Summary message constants define the runtime i18n key and English source
// fallback returned by the sample source plugin backend.
const (
	summaryMessageKey = "plugin.plugin-demo-source.page.summaryMessage"
	summaryMessage    = "This short description comes from the plugin-demo-source API and verifies that a source plugin menu page can read backend data."
)

// SummaryOutput defines one concise plugin summary payload.
type SummaryOutput struct {
	// Message is the concise page introduction returned from the plugin API.
	Message string
}

// Summary returns one concise plugin summary payload.
func (s *serviceImpl) Summary(ctx context.Context) (out *SummaryOutput, err error) {
	return &SummaryOutput{
		Message: s.translate(ctx, summaryMessageKey, summaryMessage),
	}, nil
}

// translate resolves one plugin runtime i18n key and falls back to English
// source text when the current language bundle does not define it.
func (s *serviceImpl) translate(ctx context.Context, key string, fallback string) string {
	if s == nil || s.i18nSvc == nil || strings.TrimSpace(key) == "" {
		return fallback
	}
	return s.i18nSvc.Translate(ctx, key, fallback)
}
