package file

import (
	"context"

	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	v1 "lina-core/api/file/v1"
)

// InfoByIds returns file information by ID list
func (c *ControllerV1) InfoByIds(ctx context.Context, req *v1.InfoByIdsReq) (res *v1.InfoByIdsRes, err error) {
	idStrs := gstr.SplitAndTrim(req.Ids, ",")
	ids := make([]int64, 0, len(idStrs))
	for _, s := range idStrs {
		ids = append(ids, gconv.Int64(s))
	}
	files, err := c.fileSvc.InfoByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	return &v1.InfoByIdsRes{List: files}, nil
}
