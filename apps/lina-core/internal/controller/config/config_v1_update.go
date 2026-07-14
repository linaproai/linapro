package config

import (
	"context"

	v1 "lina-core/api/config/v1"
	"lina-core/internal/service/sysconfig"
	"lina-core/pkg/configvaluetype"
)

// Update updates the specified config item.
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	input := sysconfig.UpdateInput{
		Id:     req.Id,
		Name:   req.Name,
		Key:    req.Key,
		Value:  req.Value,
		Remark: req.Remark,
	}
	if req.ValueType != nil {
		valueType := req.ValueType.String()
		input.ValueType = &valueType
	}
	if req.Options != nil {
		options := make([]configvaluetype.Option, 0, len(*req.Options))
		for _, option := range *req.Options {
			options = append(options, configvaluetype.Option{
				Label: option.Label,
				Value: option.Value,
			})
		}
		input.Options = &options
	}
	err = c.sysConfigSvc.Update(ctx, input)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}
