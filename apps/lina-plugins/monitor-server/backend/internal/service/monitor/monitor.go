// Package monitor implements server monitoring collection, storage, query,
// and cleanup services for the monitor-server source plugin. It owns the
// plugin_monitor_server table and runtime sampling logic instead of depending
// on host-internal server-monitor services.
package monitor

import (
	"context"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	cpuutil "github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	netutil "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"

	"lina-core/pkg/dialect"
	"lina-core/pkg/logger"
	"lina-plugin-monitor-server/backend/internal/dao"
	"lina-plugin-monitor-server/backend/internal/model/do"
	entitymodel "lina-plugin-monitor-server/backend/internal/model/entity"
)

// Storage metadata constants for server-monitor persistence.
const (
	colNodeName  = "node_name"
	colNodeIp    = "node_ip"
	colCreatedAt = "created_at"
	colUpdatedAt = "updated_at"
)

// MonitorData represents all collected server metrics.
type MonitorData struct {
	Server  *ServerInfo    `json:"server"`
	CPU     *CPUInfo       `json:"cpu"`
	Memory  *MemoryInfo    `json:"memory"`
	Disks   []*DiskInfo    `json:"disks"`
	Network *NetworkInfo   `json:"network"`
	GoInfo  *GoRuntimeInfo `json:"goInfo"`
}

// ServerInfo represents server basic information.
type ServerInfo struct {
	Hostname  string `json:"hostname"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	BootTime  string `json:"bootTime"`
	Uptime    uint64 `json:"uptime"`
	StartTime string `json:"startTime"`
}

// CPUInfo represents CPU metrics.
type CPUInfo struct {
	Cores        int     `json:"cores"`
	ModelName    string  `json:"modelName"`
	UsagePercent float64 `json:"usagePercent"`
}

// MemoryInfo represents memory metrics.
type MemoryInfo struct {
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Available    uint64  `json:"available"`
	UsagePercent float64 `json:"usagePercent"`
}

// DiskInfo represents disk metrics.
type DiskInfo struct {
	Path         string  `json:"path"`
	FsType       string  `json:"fsType"`
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Free         uint64  `json:"free"`
	UsagePercent float64 `json:"usagePercent"`
}

// NetworkInfo represents network metrics.
type NetworkInfo struct {
	BytesSent uint64  `json:"bytesSent"`
	BytesRecv uint64  `json:"bytesRecv"`
	SendRate  float64 `json:"sendRate"`
	RecvRate  float64 `json:"recvRate"`
}

// GoRuntimeInfo represents Go runtime metrics.
type GoRuntimeInfo struct {
	Version       string  `json:"version"`
	Goroutines    int     `json:"goroutines"`
	ProcessCPU    float64 `json:"processCpu"`
	ProcessMemory float64 `json:"processMemory"`
	GCPauseNs     uint64  `json:"gcPauseNs"`
	GfVersion     string  `json:"gfVersion"`
	ServiceUptime string  `json:"serviceUptime"`
}

// DBInfo represents database metrics.
type DBInfo struct {
	Version      string `json:"version"`
	MaxOpenConns int    `json:"maxOpenConns"`
	OpenConns    int    `json:"openConns"`
	InUse        int    `json:"inUse"`
	Idle         int    `json:"idle"`
}

// NodeMonitorData wraps monitor data with node info.
type NodeMonitorData struct {
	NodeName  string       `json:"nodeName"`
	NodeIp    string       `json:"nodeIp"`
	Data      *MonitorData `json:"data"`
	CollectAt string       `json:"collectAt"`
}

// serverMonitorRecord reuses the plugin-local generated plugin_monitor_server entity.
type serverMonitorRecord = entitymodel.Server

// Service defines the server-monitor service contract.
type Service interface {
	// CollectAndStore collects metrics and stores them in the database.
	CollectAndStore(ctx context.Context)
	// Collect gathers all server metrics.
	Collect(ctx context.Context) *MonitorData
	// GetDBInfo collects database metrics on demand.
	GetDBInfo(ctx context.Context) *DBInfo
	// GetLatest returns the latest monitor records for each node.
	GetLatest(ctx context.Context, nodeName string) ([]*NodeMonitorData, error)
	// CleanupStale deletes monitor records older than the provided threshold.
	CleanupStale(ctx context.Context, threshold time.Duration) (int64, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	networkMu     sync.Mutex
	startTime     time.Time
	lastNetBytes  *netutil.IOCountersStat
	lastCollectAt time.Time
}

// New creates and returns a new server-monitor service instance.
func New() Service {
	return &serviceImpl{startTime: time.Now()}
}

// CollectAndStore collects metrics and stores them in the database.
func (s *serviceImpl) CollectAndStore(ctx context.Context) {
	data := s.Collect(ctx)
	jsonData, err := gjson.Encode(data)
	if err != nil {
		logger.Errorf(ctx, "encode monitor data failed: %v", err)
		return
	}

	nodeName := ""
	if hostname, hostnameErr := os.Hostname(); hostnameErr == nil {
		nodeName = hostname
	} else {
		logger.Warningf(ctx, "resolve monitor node hostname failed: %v", hostnameErr)
	}

	if err = upsertMonitorSnapshot(ctx, nodeName, getLocalIP(), string(jsonData)); err != nil {
		logger.Errorf(ctx, "store monitor data failed: %v", err)
	}
}

// upsertMonitorSnapshot stores the latest snapshot for one node using the
// table's stable node identity as explicit upsert conflict columns.
func upsertMonitorSnapshot(ctx context.Context, nodeName string, nodeIP string, data string) error {
	_, err := dao.Server.Ctx(ctx).
		Data(do.Server{
			NodeName: nodeName,
			NodeIp:   nodeIP,
			Data:     data,
		}).
		OnConflict(colNodeName, colNodeIp).
		Save()
	return err
}

// Collect gathers all server metrics.
func (s *serviceImpl) Collect(ctx context.Context) *MonitorData {
	return &MonitorData{
		Server:  s.collectServer(ctx),
		CPU:     s.collectCPU(),
		Memory:  s.collectMemory(),
		Disks:   s.collectDisks(),
		Network: s.collectNetwork(),
		GoInfo:  s.collectGoRuntime(),
	}
}

// GetDBInfo collects database metrics on demand.
func (s *serviceImpl) GetDBInfo(ctx context.Context) *DBInfo {
	info := &DBInfo{}
	dbVersion, err := dialect.DatabaseVersion(ctx, g.DB())
	if err != nil {
		logger.Warningf(ctx, "collect database version failed: %v", err)
		info.Version = "unknown"
	} else {
		info.Version = dbVersion
	}
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
	model := dao.Server.Ctx(ctx)
	if nodeName != "" {
		model = model.Where(colNodeName, nodeName)
	}

	records := make([]*serverMonitorRecord, 0)
	err := model.OrderDesc(colUpdatedAt).Scan(&records)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	result := make([]*NodeMonitorData, 0)
	for _, record := range records {
		key := record.NodeName + "|" + record.NodeIp
		if seen[key] {
			continue
		}
		seen[key] = true

		var data MonitorData
		if decodeErr := gjson.DecodeTo([]byte(record.Data), &data); decodeErr != nil {
			continue
		}

		collectAt := record.UpdatedAt
		if collectAt == nil {
			collectAt = record.CreatedAt
		}
		collectAtString := ""
		if collectAt != nil {
			collectAtString = collectAt.Format("Y-m-d H:i:s")
		}

		result = append(result, &NodeMonitorData{
			NodeName:  record.NodeName,
			NodeIp:    record.NodeIp,
			Data:      &data,
			CollectAt: collectAtString,
		})
	}
	return result, nil
}

// CleanupStale deletes monitor records older than the provided threshold.
func (s *serviceImpl) CleanupStale(ctx context.Context, threshold time.Duration) (int64, error) {
	cutoff := time.Now().Add(-threshold)
	result, err := dao.Server.Ctx(ctx).
		WhereLT(colUpdatedAt, cutoff).
		Delete()
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

// collectServer gathers host-level OS, uptime, and service start information.
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

// collectCPU gathers CPU topology and instantaneous utilization metrics.
func (s *serviceImpl) collectCPU() *CPUInfo {
	info := &CPUInfo{Cores: runtime.NumCPU()}
	if cpuInfos, err := cpuutil.Info(); err == nil && len(cpuInfos) > 0 {
		info.ModelName = cpuInfos[0].ModelName
	}
	if percents, err := cpuutil.Percent(time.Second, false); err == nil && len(percents) > 0 {
		info.UsagePercent = percents[0]
	}
	return info
}

// collectMemory gathers virtual-memory totals and usage ratios.
func (s *serviceImpl) collectMemory() *MemoryInfo {
	virtualMemory, err := mem.VirtualMemory()
	if err != nil {
		return &MemoryInfo{}
	}
	return &MemoryInfo{
		Total:        virtualMemory.Total,
		Used:         virtualMemory.Used,
		Available:    virtualMemory.Available,
		UsagePercent: virtualMemory.UsedPercent,
	}
}

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

// collectDisks gathers mounted physical-disk usage while skipping virtual filesystems.
func (s *serviceImpl) collectDisks() []*DiskInfo {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil
	}
	disks := make([]*DiskInfo, 0)
	for _, partition := range partitions {
		if virtualFsTypes[partition.Fstype] {
			continue
		}
		usage, usageErr := disk.Usage(partition.Mountpoint)
		if usageErr != nil || usage.Total == 0 {
			continue
		}
		disks = append(disks, &DiskInfo{
			Path:         partition.Mountpoint,
			FsType:       partition.Fstype,
			Total:        usage.Total,
			Used:         usage.Used,
			Free:         usage.Free,
			UsagePercent: usage.UsedPercent,
		})
	}
	return disks
}

// collectNetwork gathers byte counters and derives rates from the previous sample.
func (s *serviceImpl) collectNetwork() *NetworkInfo {
	counters, err := netutil.IOCounters(false)
	if err != nil || len(counters) == 0 {
		return &NetworkInfo{}
	}
	current := &counters[0]
	info := &NetworkInfo{BytesSent: current.BytesSent, BytesRecv: current.BytesRecv}

	s.networkMu.Lock()
	defer s.networkMu.Unlock()

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

// collectGoRuntime gathers Go runtime, process resource, and service uptime metrics.
func (s *serviceImpl) collectGoRuntime() *GoRuntimeInfo {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	info := &GoRuntimeInfo{
		Version:    runtime.Version(),
		Goroutines: runtime.NumGoroutine(),
		GCPauseNs:  memStats.PauseNs[(memStats.NumGC+255)%256],
		GfVersion:  "v2.10.0",
	}
	if proc, err := process.NewProcess(int32(os.Getpid())); err == nil {
		if cpuPercent, cpuErr := proc.CPUPercent(); cpuErr == nil {
			info.ProcessCPU = cpuPercent
		}
		if memPercent, memErr := proc.MemoryPercent(); memErr == nil {
			info.ProcessMemory = float64(memPercent)
		}
	}
	duration := time.Since(s.startTime)
	// Persist uptime as raw seconds so the frontend can localize the display
	// according to the active language instead of inheriting backend literals.
	info.ServiceUptime = strconv.FormatInt(int64(duration/time.Second), 10)
	return info
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
