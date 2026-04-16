// demo_v1_record_delete.go implements the plugin-demo-source record delete HTTP handler.

package demo

import (
	"context"

	"lina-plugin-demo-source/backend/api/demo/v1"
)

// DeleteRecord deletes one demo record and its attachment file.
func (c *ControllerV1) DeleteRecord(
	ctx context.Context,
	req *v1.DeleteRecordReq,
) (res *v1.DeleteRecordRes, err error) {
	if err = c.demoSvc.DeleteRecord(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteRecordRes{Id: req.Id}, nil
}
