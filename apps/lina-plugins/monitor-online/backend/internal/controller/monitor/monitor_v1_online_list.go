package monitor

import (
	"context"

	v1 "lina-plugin-monitor-online/backend/api/monitor/v1"
	monitorsvc "lina-plugin-monitor-online/backend/internal/service/monitor"
)

// OnlineList queries the online-user list through the published host session seam.
func (c *ControllerV1) OnlineList(ctx context.Context, req *v1.OnlineListReq) (res *v1.OnlineListRes, err error) {
	out, err := c.monitorSvc.List(ctx, monitorsvc.ListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Username: req.Username,
		Ip:       req.Ip,
	})
	if err != nil {
		return nil, err
	}

	items := make([]*v1.OnlineUserItem, 0, len(out.Items))
	for _, session := range out.Items {
		loginTime := ""
		if session.LoginTime != nil {
			loginTime = session.LoginTime.Format("Y-m-d H:i:s")
		}
		items = append(items, &v1.OnlineUserItem{
			TokenId:   session.TokenId,
			Username:  session.Username,
			DeptName:  session.DeptName,
			Ip:        session.Ip,
			Browser:   session.Browser,
			Os:        session.Os,
			LoginTime: loginTime,
		})
	}

	return &v1.OnlineListRes{Items: items, Total: out.Total}, nil
}
