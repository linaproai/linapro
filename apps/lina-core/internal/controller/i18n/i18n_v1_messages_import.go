// This file handles the flat i18n message import endpoint.

package i18n

import (
	"context"

	v1 "lina-core/api/i18n/v1"
	i18nsvc "lina-core/internal/service/i18n"
)

// ImportMessages imports flat translation messages into database overrides.
func (c *ControllerV1) ImportMessages(ctx context.Context, req *v1.ImportMessagesReq) (res *v1.ImportMessagesRes, err error) {
	output, err := c.maintainer.ImportMessages(ctx, i18nsvc.MessageImportInput{
		Locale:    req.Locale,
		ScopeType: req.ScopeType,
		ScopeKey:  req.ScopeKey,
		Overwrite: req.Overwrite,
		Remark:    req.Remark,
		Messages:  req.Messages,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ImportMessagesRes{
		Locale:   output.Locale,
		Imported: output.Imported,
		Created:  output.Created,
		Updated:  output.Updated,
		Skipped:  output.Skipped,
	}, nil
}
