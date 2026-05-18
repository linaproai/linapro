package config

import (
	"context"

	v1 "lina-core/api/config/v1"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/sysconfig"
)

// List queries config items with pagination and filters.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.sysConfigSvc.List(ctx, sysconfig.ListInput{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		Name:      req.Name,
		Key:       req.Key,
		BeginTime: req.BeginTime,
		EndTime:   req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.ConfigItem, 0, len(out.List))
	for _, row := range out.List {
		item := configItem(row)
		list = append(list, &item)
	}
	return &v1.ListRes{List: list, Total: out.Total}, nil
}

// configItem maps a config entity to the API-safe response DTO.
func configItem(cfg *entity.SysConfig) v1.ConfigItem {
	if cfg == nil {
		return v1.ConfigItem{}
	}
	return v1.ConfigItem{
		Id:        cfg.Id,
		Name:      cfg.Name,
		Key:       cfg.Key,
		Value:     cfg.Value,
		IsBuiltin: cfg.IsBuiltin,
		Remark:    cfg.Remark,
		CreatedAt: cfg.CreatedAt,
		UpdatedAt: cfg.UpdatedAt,
	}
}
