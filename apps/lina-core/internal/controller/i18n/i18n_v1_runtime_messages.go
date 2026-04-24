// This file handles the runtime i18n message-bundle endpoint.

package i18n

import (
	"context"

	v1 "lina-core/api/i18n/v1"
)

// RuntimeMessages returns the aggregated runtime translation bundle for one locale.
func (c *ControllerV1) RuntimeMessages(ctx context.Context, req *v1.RuntimeMessagesReq) (res *v1.RuntimeMessagesRes, err error) {
	locale := c.i18nSvc.ResolveLocale(ctx, req.Lang)
	return &v1.RuntimeMessagesRes{
		Locale:   locale,
		Messages: c.i18nSvc.BuildRuntimeMessages(ctx, locale),
	}, nil
}
