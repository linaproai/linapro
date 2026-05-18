// This file implements the v1 user import HTTP handler.

package user

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/user/v1"
	usersvc "lina-core/internal/service/user"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/closeutil"
)

// Import imports users
func (c *ControllerV1) Import(ctx context.Context, req *v1.ImportReq) (res *v1.ImportRes, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, bizerr.NewCode(usersvc.CodeUserImportFileRequired)
	}

	f, err := file.Open()
	if err != nil {
		return nil, bizerr.WrapCode(err, usersvc.CodeUserImportExcelParseFailed)
	}
	defer closeutil.Close(ctx, f, &err, "close user import file failed")

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
