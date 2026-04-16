package monitor

import (
	"context"

	v1 "lina-core/api/monitor/v1"
)

// ServerMonitor returns server monitor information
func (c *ControllerV1) ServerMonitor(ctx context.Context, req *v1.ServerMonitorReq) (res *v1.ServerMonitorRes, err error) {
	nodes, err := c.serverMonSvc.GetLatest(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}

	dbInfo := c.serverMonSvc.GetDBInfo(ctx)

	items := make([]*v1.ServerNodeInfo, 0, len(nodes))
	for _, n := range nodes {
		item := &v1.ServerNodeInfo{
			NodeName:  n.NodeName,
			NodeIp:    n.NodeIp,
			CollectAt: n.CollectAt,
		}
		if n.Data.Server != nil {
			item.Server = &v1.ServerBasic{
				Hostname:  n.Data.Server.Hostname,
				OS:        n.Data.Server.OS,
				Arch:      n.Data.Server.Arch,
				BootTime:  n.Data.Server.BootTime,
				Uptime:    n.Data.Server.Uptime,
				StartTime: n.Data.Server.StartTime,
			}
		}
		if n.Data.CPU != nil {
			item.CPU = &v1.CPUMetrics{
				Cores:        n.Data.CPU.Cores,
				ModelName:    n.Data.CPU.ModelName,
				UsagePercent: n.Data.CPU.UsagePercent,
			}
		}
		if n.Data.Memory != nil {
			item.Memory = &v1.MemoryMetrics{
				Total:        n.Data.Memory.Total,
				Used:         n.Data.Memory.Used,
				Available:    n.Data.Memory.Available,
				UsagePercent: n.Data.Memory.UsagePercent,
			}
		}
		if n.Data.Disks != nil {
			for _, d := range n.Data.Disks {
				item.Disks = append(item.Disks, &v1.DiskMetrics{
					Path:         d.Path,
					FsType:       d.FsType,
					Total:        d.Total,
					Used:         d.Used,
					Free:         d.Free,
					UsagePercent: d.UsagePercent,
				})
			}
		}
		if n.Data.Network != nil {
			item.Network = &v1.NetMetrics{
				BytesSent: n.Data.Network.BytesSent,
				BytesRecv: n.Data.Network.BytesRecv,
				SendRate:  n.Data.Network.SendRate,
				RecvRate:  n.Data.Network.RecvRate,
			}
		}
		if n.Data.GoInfo != nil {
			item.GoInfo = &v1.GoMetrics{
				Version:       n.Data.GoInfo.Version,
				Goroutines:    n.Data.GoInfo.Goroutines,
				ProcessCPU:    n.Data.GoInfo.ProcessCPU,
				ProcessMemory: n.Data.GoInfo.ProcessMemory,
				GCPauseNs:     n.Data.GoInfo.GCPauseNs,
				GfVersion:     n.Data.GoInfo.GfVersion,
				ServiceUptime: n.Data.GoInfo.ServiceUptime,
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
