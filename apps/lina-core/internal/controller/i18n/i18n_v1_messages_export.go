// This file handles the flat i18n message export endpoint.

package i18n

import (
	"context"

	v1 "lina-core/api/i18n/v1"
)

// ExportMessages exports flat runtime messages for one locale.
func (c *ControllerV1) ExportMessages(ctx context.Context, req *v1.ExportMessagesReq) (res *v1.ExportMessagesRes, err error) {
	output := c.i18nSvc.ExportMessages(ctx, req.Locale, req.Raw)
	return &v1.ExportMessagesRes{
		Locale:        output.Locale,
		DefaultLocale: output.DefaultLocale,
		Mode:          output.Mode,
		Total:         len(output.Messages),
		Messages:      output.Messages,
	}, nil
}
