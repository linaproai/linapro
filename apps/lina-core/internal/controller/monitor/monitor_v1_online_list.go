package monitor

import (
	"context"

	v1 "lina-core/api/monitor/v1"
	"lina-core/internal/service/session"
)

// OnlineList queries online user list with pagination
func (c *ControllerV1) OnlineList(ctx context.Context, req *v1.OnlineListReq) (res *v1.OnlineListRes, err error) {
	result, err := c.sessionStore().ListPage(ctx, &session.ListFilter{
		Username: req.Username,
		Ip:       req.Ip,
	}, req.PageNum, req.PageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*v1.OnlineUserItem, 0, len(result.Items))
	for _, s := range result.Items {
		items = append(items, &v1.OnlineUserItem{
			TokenId:   s.TokenId,
			Username:  s.Username,
			DeptName:  s.DeptName,
			Ip:        s.Ip,
			Browser:   s.Browser,
			Os:        s.Os,
			LoginTime: s.LoginTime.Format("Y-m-d H:i:s"),
		})
	}

	return &v1.OnlineListRes{
		Items: items,
		Total: result.Total,
	}, nil
}
