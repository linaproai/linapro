package config

import (
	"context"

	v1 "lina-core/api/config/v1"
	"lina-core/internal/service/sysconfig"
	"lina-core/pkg/configvaluetype"
)

// Create creates a new config item.
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	options := make([]configvaluetype.Option, 0, len(req.Options))
	for _, option := range req.Options {
		options = append(options, configvaluetype.Option{
			Label: option.Label,
			Value: option.Value,
		})
	}
	id, err := c.sysConfigSvc.Create(ctx, sysconfig.CreateInput{
		Name:      req.Name,
		Key:       req.Key,
		Value:     req.Value,
		ValueType: req.ValueType.String(),
		Options:   options,
		Remark:    req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Id: id}, nil
}
