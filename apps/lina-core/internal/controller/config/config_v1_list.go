// This file maps system-configuration list projections into public API DTOs,
// including shared flag and tenant-override contract types.

package config

import (
	"context"

	v1 "lina-core/api/config/v1"
	"lina-core/internal/service/sysconfig"
	"lina-core/pkg/apitime"
	"lina-core/pkg/configvaluetype"
	"lina-core/pkg/fallbackoverride"
	"lina-core/pkg/statusflag"
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

// configItem maps a config projection to the API-safe response DTO.
func configItem(item *sysconfig.ConfigProjection) v1.ConfigItem {
	if item == nil || item.SysConfig == nil {
		return v1.ConfigItem{}
	}
	valueType := configvaluetype.Normalize(item.ValueType)
	options := make([]v1.ConfigValueOption, 0)
	for _, option := range parseConfigOptions(item.Options) {
		options = append(options, v1.ConfigValueOption{
			Label: option.Label,
			Value: option.Value,
		})
	}
	return v1.ConfigItem{
		Id:             item.Id,
		Name:           item.Name,
		Key:            item.Key,
		Value:          item.Value,
		ValueType:      valueType,
		Options:        options,
		IsBuiltin:      statusflag.YesNo(item.IsBuiltin),
		Remark:         item.Remark,
		SourceTenantId: item.SourceTenantId,
		IsFallback:     item.IsFallback,
		CanEdit:        item.CanEdit,
		CanOverride:    item.CanOverride,
		OverrideMode:   fallbackoverride.Mode(item.OverrideMode),
		CreatedAt:      apitime.Milli(item.CreatedAt),
		UpdatedAt:      apitime.Milli(item.UpdatedAt),
	}
}

// parseConfigOptions decodes stored options JSON into option structs.
func parseConfigOptions(raw string) []configvaluetype.Option {
	options, err := configvaluetype.ParseOptions(raw)
	if err != nil || len(options) == 0 {
		return []configvaluetype.Option{}
	}
	return options
}
