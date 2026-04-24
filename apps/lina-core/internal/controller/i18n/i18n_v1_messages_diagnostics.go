// This file handles the i18n source diagnostics endpoint.

package i18n

import (
	"context"

	v1 "lina-core/api/i18n/v1"
)

// DiagnoseMessages returns effective source diagnostics for one locale.
func (c *ControllerV1) DiagnoseMessages(ctx context.Context, req *v1.DiagnoseMessagesReq) (res *v1.DiagnoseMessagesRes, err error) {
	locale := c.i18nSvc.ResolveLocale(ctx, req.Locale)
	defaultLocale := c.i18nSvc.GetLocale(context.Background())
	items := c.i18nSvc.DiagnoseMessages(ctx, req.Locale, req.KeyPrefix)
	responseItems := make([]v1.MessageDiagnosticItem, 0, len(items))
	for _, item := range items {
		responseItems = append(responseItems, v1.MessageDiagnosticItem{
			Key:             item.Key,
			Value:           item.Value,
			RequestedLocale: item.RequestedLocale,
			EffectiveLocale: item.EffectiveLocale,
			FromFallback:    item.FromFallback,
			SourceType:      item.Source.Type,
			SourceKey:       item.Source.ScopeKey,
		})
	}
	return &v1.DiagnoseMessagesRes{
		Locale:        locale,
		DefaultLocale: defaultLocale,
		Total:         len(responseItems),
		Items:         responseItems,
	}, nil
}
