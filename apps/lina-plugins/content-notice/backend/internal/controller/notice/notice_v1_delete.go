// notice_v1_delete.go implements the controller method that deletes notice
// records by their comma-separated identifiers.

package notice

import (
	"context"

	v1 "lina-plugin-content-notice/backend/api/notice/v1"
)

// Delete deletes a notice
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	err = c.noticeSvc.Delete(ctx, req.Ids)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}
