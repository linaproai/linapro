// This file implements the v1 system-config import HTTP handler.

package config

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/config/v1"
	"lina-core/internal/service/sysconfig"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/closeutil"
)

// ConfigImport imports configs from an Excel file.
func (c *ControllerV1) ConfigImport(ctx context.Context, req *v1.ConfigImportReq) (res *v1.ConfigImportRes, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, bizerr.NewCode(sysconfig.CodeSysConfigImportFileRequired)
	}

	// Get updateSupport flag
	updateSupport := r.Get("updateSupport").String() == "1" || r.Get("updateSupport").String() == "true"

	f, err := file.Open()
	if err != nil {
		return nil, bizerr.WrapCode(err, sysconfig.CodeSysConfigImportFileReadFailed)
	}
	defer closeutil.Close(ctx, f, &err, "close config import file failed")

	result, err := c.sysConfigSvc.Import(ctx, f, updateSupport)
	if err != nil {
		return nil, err
	}

	failList := make([]v1.ConfigImportFailItem, 0, len(result.FailList))
	for _, item := range result.FailList {
		failList = append(failList, v1.ConfigImportFailItem{
			Row:    item.Row,
			Reason: item.Reason,
		})
	}

	return &v1.ConfigImportRes{
		Success:  result.Success,
		Fail:     result.Fail,
		FailList: failList,
	}, nil
}
