// This file implements dynamic-plugin package upload and projects runtime
// lifecycle flags into typed public API contracts.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/statusflag"
)

// UploadDynamicPackage uploads one dynamic wasm package into runtime storage.
func (c *ControllerV1) UploadDynamicPackage(ctx context.Context, req *v1.UploadDynamicPackageReq) (res *v1.UploadDynamicPackageRes, err error) {
	r := g.RequestFromCtx(ctx)
	uploadFile := r.GetUploadFile("file")
	out, err := c.pluginSvc.UploadDynamicPackage(ctx, &pluginsvc.DynamicUploadInput{
		File:             uploadFile,
		OverwriteSupport: req.OverwriteSupport == statusflag.Yes,
	})
	if err != nil {
		return nil, err
	}

	return &v1.UploadDynamicPackageRes{
		Id:          out.Id,
		Name:        out.Name,
		Version:     out.Version,
		Type:        v1.PluginType(out.Type),
		RuntimeKind: out.RuntimeKind,
		RuntimeAbi:  out.RuntimeABI,
		Installed:   statusflag.Installation(out.Installed),
		Enabled:     statusflag.Enabled(out.Enabled),
	}, nil
}
