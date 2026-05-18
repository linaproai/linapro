package file

import (
	"context"

	v1 "lina-core/api/file/v1"
)

// UsageScenes returns file usage scene list
func (c *ControllerV1) UsageScenes(ctx context.Context, req *v1.UsageScenesReq) (res *v1.UsageScenesRes, err error) {
	out, err := c.fileSvc.UsageScenes(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.UsageSceneItem, len(out))
	for i, item := range out {
		items[i] = &v1.UsageSceneItem{
			Value: item.Value,
			Label: item.Label,
		}
	}
	return &v1.UsageScenesRes{
		List: items,
	}, nil
}
