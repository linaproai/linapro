package auth

import (
	"context"

	v1 "lina-core/api/auth/v1"
	authsvc "lina-core/internal/service/auth"
)

// Login handles user login.
func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	out, err := c.authSvc.Login(ctx, authsvc.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &v1.LoginRes{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		PreToken:     out.PreToken,
		Tenants:      toLoginTenants(out.Tenants),
	}, nil
}

// toLoginTenants converts service tenant candidates into API DTOs.
func toLoginTenants(items []authsvc.TenantInfo) []*v1.LoginTenantEntity {
	if len(items) == 0 {
		return nil
	}
	out := make([]*v1.LoginTenantEntity, 0, len(items))
	for _, item := range items {
		out = append(out, &v1.LoginTenantEntity{
			Id:     item.Id,
			Code:   item.Code,
			Name:   item.Name,
			Status: item.Status,
		})
	}
	return out
}
