package servermon

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	cpuutil "github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	netutil "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/logger"
)

// MonitorData represents all collected server metrics.
type MonitorData struct {
	Server  *ServerInfo    `json:"server"`  // Server information
	CPU     *CPUInfo       `json:"cpu"`     // CPU information
	Memory  *MemoryInfo    `json:"memory"`  // Memory information
	Disks   []*DiskInfo    `json:"disks"`   // Disk list
	Network *NetworkInfo   `json:"network"` // Network information
	GoInfo  *GoRuntimeInfo `json:"goInfo"`  // Go runtime information
}

// ServerInfo represents server basic information.
type ServerInfo struct {
	Hostname  string `json:"hostname"`  // Hostname
	OS        string `json:"os"`        // Operating system
	Arch      string `json:"arch"`      // System architecture
	BootTime  string `json:"bootTime"`  // System boot time
	Uptime    uint64 `json:"uptime"`    // System uptime (seconds)
	StartTime string `json:"startTime"` // Service start time
}

// CPUInfo represents CPU metrics.
type CPUInfo struct {
	Cores        int     `json:"cores"`        // Number of CPU cores
	ModelName    string  `json:"modelName"`    // CPU model name
	UsagePercent float64 `json:"usagePercent"` // CPU usage percentage
}

// MemoryInfo represents memory metrics.
type MemoryInfo struct {
	Total        uint64  `json:"total"`        // Total memory (bytes)
	Used         uint64  `json:"used"`         // Used memory (bytes)
	Available    uint64  `json:"available"`    // Available memory (bytes)
	UsagePercent float64 `json:"usagePercent"` // Memory usage percentage
}

// DiskInfo represents disk metrics.
type DiskInfo struct {
	Path         string  `json:"path"`         // Mount path
	FsType       string  `json:"fsType"`       // Filesystem type
	Total        uint64  `json:"total"`        // Total space (bytes)
	Used         uint64  `json:"used"`         // Used space (bytes)
	Free         uint64  `json:"free"`         // Free space (bytes)
	UsagePercent float64 `json:"usagePercent"` // Usage percentage
}

// NetworkInfo represents network metrics.
type NetworkInfo struct {
	BytesSent uint64  `json:"bytesSent"` // Bytes sent
	BytesRecv uint64  `json:"bytesRecv"` // Bytes received
	SendRate  float64 `json:"sendRate"`  // Send rate (bytes/second)
	RecvRate  float64 `json:"recvRate"`  // Receive rate (bytes/second)
}

// GoRuntimeInfo represents Go runtime metrics.
type GoRuntimeInfo struct {
	Version       string  `json:"version"`       // Go version
	Goroutines    int     `json:"goroutines"`    // Number of goroutines
	ProcessCPU    float64 `json:"processCpu"`    // Process CPU usage
	ProcessMemory float64 `json:"processMemory"` // Process memory usage
	GCPauseNs     uint64  `json:"gcPauseNs"`     // GC pause time (nanoseconds)
	GfVersion     string  `json:"gfVersion"`     // GoFrame version
	ServiceUptime string  `json:"serviceUptime"` // Service uptime
}

// DBInfo represents database metrics.
type DBInfo struct {
	Version      string `json:"version"`      // Database version
	MaxOpenConns int    `json:"maxOpenConns"` // Max connections
	OpenConns    int    `json:"openConns"`    // Open connections
	InUse        int    `json:"inUse"`        // Connections in use
	Idle         int    `json:"idle"`         // Idle connections
}

// Service defines the servermon service contract.
type Service interface {
	// CollectAndStore collects metrics and stores them in the database.
	// This method is designed to be called by the cron service.
	// Uses Save() with unique key constraint to ensure each node only has one record,
	// with updated_at automatically maintained by the database.
	CollectAndStore(ctx context.Context)
	// Collect gathers all server metrics.
	Collect(ctx context.Context) *MonitorData
	// GetDBInfo collects database metrics on-demand.
	GetDBInfo(ctx context.Context) *DBInfo
	// GetLatest returns the latest monitor records for each node.
	GetLatest(ctx context.Context, nodeName string) ([]*NodeMonitorData, error)
	// CleanupStale deletes monitor records that haven't been updated
	// within the specified threshold duration.
	// This method is called by the cron service for scheduled cleanup.
	CleanupStale(ctx context.Context, threshold time.Duration) (int64, error)
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	startTime     time.Time               // Service start time
	lastNetBytes  *netutil.IOCountersStat // Last network statistics
	lastCollectAt time.Time               // Last collection time
}

// New creates a new Service.
func New() Service {
	return &serviceImpl{
		startTime: time.Now(),
	}
}

// CollectAndStore collects metrics and stores them in the database.
// This method is designed to be called by the cron service.
// Uses Save() with unique key constraint to ensure each node only has one record,
// with updated_at automatically maintained by the database.
func (s *serviceImpl) CollectAndStore(ctx context.Context) {
	data := s.Collect(ctx)
	jsonData, err := gjson.Encode(data)
	if err != nil {
		logger.Errorf(ctx, "Failed to encode monitor data: %v", err)
		return
	}

	nodeName := ""
	if hostname, hostnameErr := os.Hostname(); hostnameErr == nil {
		nodeName = hostname
	} else {
		logger.Warningf(ctx, "resolve monitor node hostname failed: %v", hostnameErr)
	}
	nodeIp := getLocalIP()

	// Use Save() with unique key (node_name, node_ip) to upsert.
	// Do not set CreatedAt/UpdatedAt manually - let GoFrame/MySQL handle them:
	// - created_at: auto-filled on INSERT, preserved on UPDATE
	// - updated_at: auto-updated by MySQL ON UPDATE CURRENT_TIMESTAMP
	_, err = dao.SysServerMonitor.Ctx(ctx).Data(do.SysServerMonitor{
		NodeName: nodeName,
		NodeIp:   nodeIp,
		Data:     string(jsonData),
	}).Save()
	if err != nil {
		logger.Errorf(ctx, "Failed to store monitor data: %v", err)
	}
}

// Collect gathers all server metrics.
func (s *serviceImpl) Collect(ctx context.Context) *MonitorData {
	data := &MonitorData{}
	data.Server = s.collectServer(ctx)
	data.CPU = s.collectCPU()
	data.Memory = s.collectMemory()
	data.Disks = s.collectDisks()
	data.Network = s.collectNetwork()
	data.GoInfo = s.collectGoRuntime()
	return data
}

func (s *serviceImpl) collectServer(ctx context.Context) *ServerInfo {
	hostname := ""
	if resolvedHostname, err := os.Hostname(); err == nil {
		hostname = resolvedHostname
	} else {
		logger.Warningf(ctx, "resolve hostname failed: %v", err)
	}
	info, err := host.Info()
	if err != nil {
		logger.Warningf(ctx, "collect host info failed: %v", err)
		info = nil
	}
	bootTime := ""
	var uptime uint64
	if info != nil {
		bootTime = time.Unix(int64(info.BootTime), 0).Format("2006-01-02 15:04:05")
		uptime = info.Uptime
	}
	return &ServerInfo{
		Hostname:  hostname,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		BootTime:  bootTime,
		Uptime:    uptime,
		StartTime: s.startTime.Format("2006-01-02 15:04:05"),
	}
}

func (s *serviceImpl) collectCPU() *CPUInfo {
	info := &CPUInfo{}
	info.Cores = runtime.NumCPU()
	cpuInfos, err := cpuutil.Info()
	if err == nil && len(cpuInfos) > 0 {
		info.ModelName = cpuInfos[0].ModelName
	}
	percents, err := cpuutil.Percent(time.Second, false)
	if err == nil && len(percents) > 0 {
		info.UsagePercent = percents[0]
	}
	return info
}

func (s *serviceImpl) collectMemory() *MemoryInfo {
	v, err := mem.VirtualMemory()
	if err != nil {
		return &MemoryInfo{}
	}
	return &MemoryInfo{
		Total:        v.Total,
		Used:         v.Used,
		Available:    v.Available,
		UsagePercent: v.UsedPercent,
	}
}

// virtualFsTypes lists filesystem types to exclude (common in containers).
var virtualFsTypes = map[string]bool{
	"overlay":  true,
	"tmpfs":    true,
	"devtmpfs": true,
	"devfs":    true,
	"proc":     true,
	"sysfs":    true,
	"cgroup":   true,
	"cgroup2":  true,
	"squashfs": true,
	"aufs":     true,
	"shm":      true,
	"nsfs":     true,
	"fuse":     true,
}

func (s *serviceImpl) collectDisks() []*DiskInfo {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil
	}
	var disks []*DiskInfo
	for _, p := range partitions {
		// Skip virtual/pseudo filesystems
		if virtualFsTypes[p.Fstype] {
			continue
		}
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil || usage.Total == 0 {
			continue
		}
		disks = append(disks, &DiskInfo{
			Path:         p.Mountpoint,
			FsType:       p.Fstype,
			Total:        usage.Total,
			Used:         usage.Used,
			Free:         usage.Free,
			UsagePercent: usage.UsedPercent,
		})
	}
	return disks
}

func (s *serviceImpl) collectNetwork() *NetworkInfo {
	counters, err := netutil.IOCounters(false)
	if err != nil || len(counters) == 0 {
		return &NetworkInfo{}
	}
	current := &counters[0]
	info := &NetworkInfo{
		BytesSent: current.BytesSent,
		BytesRecv: current.BytesRecv,
	}

	// Calculate rate from previous sample
	if s.lastNetBytes != nil && !s.lastCollectAt.IsZero() {
		elapsed := time.Since(s.lastCollectAt).Seconds()
		if elapsed > 0 {
			info.SendRate = float64(current.BytesSent-s.lastNetBytes.BytesSent) / elapsed
			info.RecvRate = float64(current.BytesRecv-s.lastNetBytes.BytesRecv) / elapsed
		}
	}
	s.lastNetBytes = current
	s.lastCollectAt = time.Now()
	return info
}

func (s *serviceImpl) collectGoRuntime() *GoRuntimeInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	info := &GoRuntimeInfo{
		Version:    runtime.Version(),
		Goroutines: runtime.NumGoroutine(),
		GCPauseNs:  m.PauseNs[(m.NumGC+255)%256],
		GfVersion:  "v2.10.0",
	}

	// Collect process CPU and memory usage
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err == nil {
		if cpuPercent, err := proc.CPUPercent(); err == nil {
			info.ProcessCPU = cpuPercent
		}
		if memPercent, err := proc.MemoryPercent(); err == nil {
			info.ProcessMemory = float64(memPercent)
		}
	}

	// Calculate service uptime
	duration := time.Since(s.startTime)
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	mins := int(duration.Minutes()) % 60
	parts := make([]string, 0, 3)
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d天", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d小时", hours))
	}
	if mins > 0 {
		parts = append(parts, fmt.Sprintf("%d分钟", mins))
	}
	if len(parts) == 0 {
		info.ServiceUptime = "刚启动"
	} else {
		info.ServiceUptime = strings.Join(parts, " ")
	}

	return info
}

// GetDBInfo collects database metrics on-demand.
func (s *serviceImpl) GetDBInfo(ctx context.Context) *DBInfo {
	info := &DBInfo{}

	// Get database version
	result, err := g.DB().GetValue(ctx, "SELECT VERSION()")
	if err == nil {
		info.Version = result.String()
	}

	// Get connection pool stats
	statsItems := g.DB().GetCore().Stats(ctx)
	if len(statsItems) > 0 {
		stats := statsItems[0].Stats()
		info.MaxOpenConns = stats.MaxOpenConnections
		info.OpenConns = stats.OpenConnections
		info.InUse = stats.InUse
		info.Idle = stats.Idle
	}

	return info
}

// GetLatest returns the latest monitor records for each node.
func (s *serviceImpl) GetLatest(ctx context.Context, nodeName string) ([]*NodeMonitorData, error) {
	cols := dao.SysServerMonitor.Columns()
	m := dao.SysServerMonitor.Ctx(ctx)
	if nodeName != "" {
		m = m.Where(cols.NodeName, nodeName)
	}

	// Get all records ordered by updated_at desc
	var allRecords []*entity.SysServerMonitor
	err := m.OrderDesc(cols.UpdatedAt).Scan(&allRecords)
	if err != nil {
		return nil, err
	}

	// Group by node, keep latest for each
	seen := make(map[string]bool)
	var result []*NodeMonitorData
	for _, record := range allRecords {
		key := record.NodeName + "|" + record.NodeIp
		if seen[key] {
			continue
		}
		seen[key] = true

		var data MonitorData
		if err := gjson.DecodeTo([]byte(record.Data), &data); err != nil {
			continue
		}

		// Use updated_at as the latest report time (when data was last updated)
		updatedAt := record.UpdatedAt
		if updatedAt == nil {
			updatedAt = record.CreatedAt
		}

		result = append(result, &NodeMonitorData{
			NodeName:  record.NodeName,
			NodeIp:    record.NodeIp,
			Data:      &data,
			CollectAt: updatedAt.Format("Y-m-d H:i:s"),
		})
	}
	return result, nil
}

// NodeMonitorData wraps monitor data with node info.
type NodeMonitorData struct {
	NodeName  string       `json:"nodeName"`  // Node name
	NodeIp    string       `json:"nodeIp"`    // Node IP address
	Data      *MonitorData `json:"data"`      // Monitor data
	CollectAt string       `json:"collectAt"` // Collection time
}

// getLocalIP returns the first non-loopback IPv4 address.
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}
	return "unknown"
}

// CleanupStale deletes monitor records that haven't been updated
// within the specified threshold duration.
// This method is called by the cron service for scheduled cleanup.
func (s *serviceImpl) CleanupStale(ctx context.Context, threshold time.Duration) (int64, error) {
	cutoff := time.Now().Add(-threshold)
	cols := dao.SysServerMonitor.Columns()

	result, err := dao.SysServerMonitor.Ctx(ctx).
		WhereLT(cols.UpdatedAt, cutoff).
		Delete()
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}
