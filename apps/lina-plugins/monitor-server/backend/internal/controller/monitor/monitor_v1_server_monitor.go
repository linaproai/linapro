package monitor

import (
	"context"

	v1 "lina-plugin-monitor-server/backend/api/monitor/v1"
)

// ServerMonitor returns server-monitor information.
func (c *ControllerV1) ServerMonitor(ctx context.Context, req *v1.ServerMonitorReq) (res *v1.ServerMonitorRes, err error) {
	nodes, err := c.monitorSvc.GetLatest(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}

	dbInfo := c.monitorSvc.GetDBInfo(ctx)
	items := make([]*v1.ServerNodeInfo, 0, len(nodes))
	for _, node := range nodes {
		item := &v1.ServerNodeInfo{NodeName: node.NodeName, NodeIp: node.NodeIp, CollectAt: node.CollectAt}
		if node.Data.Server != nil {
			item.Server = &v1.ServerBasic{
				Hostname:  node.Data.Server.Hostname,
				OS:        node.Data.Server.OS,
				Arch:      node.Data.Server.Arch,
				BootTime:  node.Data.Server.BootTime,
				Uptime:    node.Data.Server.Uptime,
				StartTime: node.Data.Server.StartTime,
			}
		}
		if node.Data.CPU != nil {
			item.CPU = &v1.CPUMetrics{
				Cores:        node.Data.CPU.Cores,
				ModelName:    node.Data.CPU.ModelName,
				UsagePercent: node.Data.CPU.UsagePercent,
			}
		}
		if node.Data.Memory != nil {
			item.Memory = &v1.MemoryMetrics{
				Total:        node.Data.Memory.Total,
				Used:         node.Data.Memory.Used,
				Available:    node.Data.Memory.Available,
				UsagePercent: node.Data.Memory.UsagePercent,
			}
		}
		for _, disk := range node.Data.Disks {
			item.Disks = append(item.Disks, &v1.DiskMetrics{
				Path:         disk.Path,
				FsType:       disk.FsType,
				Total:        disk.Total,
				Used:         disk.Used,
				Free:         disk.Free,
				UsagePercent: disk.UsagePercent,
			})
		}
		if node.Data.Network != nil {
			item.Network = &v1.NetMetrics{
				BytesSent: node.Data.Network.BytesSent,
				BytesRecv: node.Data.Network.BytesRecv,
				SendRate:  node.Data.Network.SendRate,
				RecvRate:  node.Data.Network.RecvRate,
			}
		}
		if node.Data.GoInfo != nil {
			item.GoInfo = &v1.GoMetrics{
				Version:       node.Data.GoInfo.Version,
				Goroutines:    node.Data.GoInfo.Goroutines,
				ProcessCPU:    node.Data.GoInfo.ProcessCPU,
				ProcessMemory: node.Data.GoInfo.ProcessMemory,
				GCPauseNs:     node.Data.GoInfo.GCPauseNs,
				GfVersion:     node.Data.GoInfo.GfVersion,
				ServiceUptime: node.Data.GoInfo.ServiceUptime,
			}
		}
		items = append(items, item)
	}

	return &v1.ServerMonitorRes{
		Nodes: items,
		DBInfo: &v1.DBMetrics{
			Version:      dbInfo.Version,
			MaxOpenConns: dbInfo.MaxOpenConns,
			OpenConns:    dbInfo.OpenConns,
			InUse:        dbInfo.InUse,
			Idle:         dbInfo.Idle,
		},
	}, nil
}
