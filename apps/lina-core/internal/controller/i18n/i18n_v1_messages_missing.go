// This file handles the missing i18n message diagnostics endpoint.

package i18n

import (
	"context"

	v1 "lina-core/api/i18n/v1"
)

// MissingMessages returns translation keys missing from one locale.
func (c *ControllerV1) MissingMessages(ctx context.Context, req *v1.MissingMessagesReq) (res *v1.MissingMessagesRes, err error) {
	locale := c.i18nSvc.ResolveLocale(ctx, req.Locale)
	defaultLocale := c.i18nSvc.GetLocale(context.Background())
	items := c.i18nSvc.CheckMissingMessages(ctx, req.Locale, req.KeyPrefix)
	responseItems := make([]v1.MissingMessageItem, 0, len(items))
	for _, item := range items {
		responseItems = append(responseItems, v1.MissingMessageItem{
			Key:          item.Key,
			DefaultValue: item.DefaultValue,
			SourceType:   item.Source.Type,
			SourceKey:    item.Source.ScopeKey,
		})
	}
	return &v1.MissingMessagesRes{
		Locale:        locale,
		DefaultLocale: defaultLocale,
		Total:         len(responseItems),
		Items:         responseItems,
	}, nil
}
