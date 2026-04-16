package dict

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/dict/v1"
	dictsvc "lina-core/internal/service/dict"
)

// Export exports dictionary types and data together to Excel.
func (c *ControllerV1) Export(ctx context.Context, req *v1.ExportReq) (res *v1.ExportRes, err error) {
	data, err := c.dictSvc.CombinedExport(ctx, dictsvc.CombinedExportInput{
		Name: req.Name,
		Type: req.Type,
		Ids:  req.Ids,
	})
	if err != nil {
		return nil, err
	}

	r := g.RequestFromCtx(ctx)
	// Use RFC 5987 format for Chinese filename support
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''%E5%AD%97%E5%85%B8%E7%AE%A1%E7%90%86%E5%AF%BC%E5%87%BA.xlsx")
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}