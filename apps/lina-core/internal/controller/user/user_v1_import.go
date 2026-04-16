// This file implements the v1 user import HTTP handler.

package user

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/user/v1"
	"lina-core/pkg/closeutil"
)

// Import imports users
func (c *ControllerV1) Import(ctx context.Context, req *v1.ImportReq) (res *v1.ImportRes, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, gerror.New("请选择要导入的文件")
	}

	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer closeutil.Close(f, &err, "关闭用户导入文件失败")

	result, err := c.userSvc.Import(ctx, f)
	if err != nil {
		return nil, err
	}

	failList := make([]v1.ImportFailItem, 0, len(result.FailList))
	for _, item := range result.FailList {
		failList = append(failList, v1.ImportFailItem{
			Row:    item.Row,
			Reason: item.Reason,
		})
	}

	return &v1.ImportRes{
		Success:  result.Success,
		Fail:     result.Fail,
		FailList: failList,
	}, nil
}
