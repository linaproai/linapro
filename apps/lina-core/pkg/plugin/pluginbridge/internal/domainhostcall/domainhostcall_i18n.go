// This file implements guest-side runtime translation capability hostcall
// clients. It keeps i18n lookup DTOs next to translation methods.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/i18ncap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// i18nService adapts runtime translation reads to host services.
type i18nService struct{ baseService }

// I18n creates the runtime translation guest client.
func I18n(invoker Invoker) i18ncap.Service {
	return i18nService{baseService: newBaseService(invoker)}
}

// GetLocale returns the effective request locale.
func (s i18nService) GetLocale(_ context.Context) string {
	var out string
	if err := s.callJSONRequest(protocol.HostServiceI18n, protocol.HostServiceMethodI18nGetLocale, nil, &out); err != nil {
		return ""
	}
	return out
}

// Translate returns the localized value for one key and fallback.
func (s i18nService) Translate(_ context.Context, key string, fallback string) string {
	var out string
	if err := s.callJSONRequest(protocol.HostServiceI18n, protocol.HostServiceMethodI18nTranslate, translateRequest{Key: key, Fallback: fallback}, &out); err != nil {
		return fallback
	}
	return out
}

// FindMessageKeys returns runtime message keys matching prefix and keyword.
func (s i18nService) FindMessageKeys(_ context.Context, prefix string, keyword string) []string {
	var out []string
	if err := s.callJSONRequest(protocol.HostServiceI18n, protocol.HostServiceMethodI18nFindMessageKeys, findMessagesRequest{Prefix: prefix, Keyword: keyword}, &out); err != nil {
		return nil
	}
	return out
}

// translateRequest carries one runtime translation lookup.
type translateRequest struct {
	Key      string `json:"key"`
	Fallback string `json:"fallback"`
}

// findMessagesRequest carries runtime message-key search parameters.
type findMessagesRequest struct {
	Prefix  string `json:"prefix"`
	Keyword string `json:"keyword"`
}

var _ i18ncap.Service = (*i18nService)(nil)
