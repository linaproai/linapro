package water

import (
	"context"

	"lina-plugin-water/backend/api/water/v1"
	watersvc "lina-plugin-water/backend/internal/service/water"
)

// SubmitSnap submits one asynchronous watermark snapshot task.
func (c *ControllerV1) SubmitSnap(ctx context.Context, req *v1.SubmitSnapReq) (res *v1.SubmitSnapRes, err error) {
	out, err := c.waterSvc.SubmitSnap(ctx, watersvc.SubmitSnapInput{
		DeviceType:  req.DeviceType,
		DeviceId:    req.DeviceId,
		ErrorCode:   req.ErrorCode,
		DeviceCode:  req.DeviceCode,
		ChannelCode: req.ChannelCode,
		DeviceIdx:   req.DeviceIdx,
		Image:       req.Image,
		ImageName:   req.ImageName,
		ImagePath:   req.ImagePath,
		AccessNode:  req.AccessNode,
		AcceptNode:  req.AcceptNode,
		UploadUrl:   req.UploadUrl,
		CallbackUrl: req.CallbackUrl,
		Url:         req.Url,
		User:        req.User,
		Tenant:      req.Tenant,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SubmitSnapRes{
		Success: out.Success,
		TaskId:  out.TaskId,
		Status:  string(out.Status),
	}, nil
}
