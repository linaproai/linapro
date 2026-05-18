package dict

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/dict/v1"
	dictsvc "lina-core/internal/service/dict"
)

// TypeExport exports dictionary types to Excel.
func (c *ControllerV1) TypeExport(ctx context.Context, req *v1.TypeExportReq) (res *v1.TypeExportRes, err error) {
	data, err := c.dictSvc.Export(ctx, dictsvc.ExportInput{
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
	r.Response.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''%E5%AD%97%E5%85%B8%E7%B1%BB%E5%9E%8B%E5%AF%BC%E5%87%BA.xlsx")
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}
