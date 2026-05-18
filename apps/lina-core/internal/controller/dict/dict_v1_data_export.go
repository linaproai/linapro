package dict

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/dict/v1"
	dictsvc "lina-core/internal/service/dict"
)

// DataExport exports dictionary data to Excel.
func (c *ControllerV1) DataExport(ctx context.Context, req *v1.DataExportReq) (res *v1.DataExportRes, err error) {
	data, err := c.dictSvc.DataExport(ctx, dictsvc.DataExportInput{
		DictType: req.DictType,
		Label:    req.Label,
		Ids:      req.Ids,
	})
	if err != nil {
		return nil, err
	}
	r := g.RequestFromCtx(ctx)
	// Use RFC 5987 format for Chinese filename support
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''%E5%AD%97%E5%85%B8%E6%95%B0%E6%8D%AE%E5%AF%BC%E5%87%BA.xlsx")
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}
