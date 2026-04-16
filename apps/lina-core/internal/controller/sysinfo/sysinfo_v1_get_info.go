package sysinfo

import (
	"context"

	"lina-core/api/sysinfo/v1"
)

// GetInfo returns system information
func (c *ControllerV1) GetInfo(ctx context.Context, req *v1.GetInfoReq) (res *v1.GetInfoRes, err error) {
	info, err := c.sysInfoSvc.GetInfo(ctx)
	if err != nil {
		return nil, err
	}

	res = &v1.GetInfoRes{
		GoVersion:   info.GoVersion,
		GfVersion:   info.GfVersion,
		Os:          info.Os,
		Arch:        info.Arch,
		DbVersion:   info.DbVersion,
		StartTime:   info.StartTime,
		RunDuration: info.RunDuration,
	}

	// Map backend components
	for _, c := range info.BackendComponents {
		res.BackendComponents = append(res.BackendComponents, v1.ComponentInfo{
			Name:        c.Name,
			Version:     c.Version,
			Url:         c.Url,
			Description: c.Description,
		})
	}

	// Map frontend components
	for _, c := range info.FrontendComponents {
		res.FrontendComponents = append(res.FrontendComponents, v1.ComponentInfo{
			Name:        c.Name,
			Version:     c.Version,
			Url:         c.Url,
			Description: c.Description,
		})
	}

	return res, nil
}
