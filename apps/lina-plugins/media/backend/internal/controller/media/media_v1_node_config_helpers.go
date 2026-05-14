// This file contains controller response converters for media node config resources.

package media

import (
	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// nodeOutputToItem converts service node output to API item.
func nodeOutputToItem(out *mediasvc.NodeOutput) *v1.NodeListItem {
	if out == nil {
		return &v1.NodeListItem{}
	}
	return &v1.NodeListItem{
		Id:         out.Id,
		NodeNum:    out.NodeNum,
		Name:       out.Name,
		QnUrl:      out.QnUrl,
		BasicUrl:   out.BasicUrl,
		DnUrl:      out.DnUrl,
		CreatorId:  out.CreatorId,
		CreateTime: out.CreateTime,
		UpdaterId:  out.UpdaterId,
		UpdateTime: out.UpdateTime,
	}
}

// deviceNodeOutputToItem converts service device-node output to API item.
func deviceNodeOutputToItem(out *mediasvc.DeviceNodeOutput) *v1.DeviceNodeListItem {
	if out == nil {
		return &v1.DeviceNodeListItem{}
	}
	return &v1.DeviceNodeListItem{
		DeviceId: out.DeviceId,
		NodeNum:  out.NodeNum,
		NodeName: out.NodeName,
	}
}

// tenantStreamConfigOutputToItem converts service tenant stream output to API item.
func tenantStreamConfigOutputToItem(out *mediasvc.TenantStreamConfigOutput) *v1.TenantStreamConfigListItem {
	if out == nil {
		return &v1.TenantStreamConfigListItem{}
	}
	return &v1.TenantStreamConfigListItem{
		TenantId:      out.TenantId,
		MaxConcurrent: out.MaxConcurrent,
		NodeNum:       out.NodeNum,
		NodeName:      out.NodeName,
		Enable:        out.Enable,
		CreatorId:     out.CreatorId,
		CreateTime:    out.CreateTime,
		UpdaterId:     out.UpdaterId,
		UpdateTime:    out.UpdateTime,
	}
}
