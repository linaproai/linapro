package v1

import "github.com/gogf/gf/v2/frame/g"

// Server Monitor API

// ServerMonitorReq defines the request for retrieving server metrics.
type ServerMonitorReq struct {
	g.Meta   `path:"/monitor/server" method:"get" tags:"系统监控" summary:"服务监控" dc:"查询服务器监控数据，返回各节点最新的CPU、内存、磁盘、网络、Go运行时等指标信息" permission:"monitor:server:list"`
	NodeName string `json:"nodeName" dc:"按节点名称过滤，不传则返回所有节点" eg:"my-server"`
}

// ServerMonitorRes is the server monitor response.
type ServerMonitorRes struct {
	Nodes  []*ServerNodeInfo `json:"nodes" dc:"各节点监控数据" eg:"[]"`
	DBInfo *DBMetrics        `json:"dbInfo" dc:"数据库指标信息" eg:""`
}

// ServerNodeInfo represents server node information.
type ServerNodeInfo struct {
	NodeName  string         `json:"nodeName" dc:"节点名称（hostname）" eg:"my-server"`
	NodeIp    string         `json:"nodeIp" dc:"节点IP地址" eg:"192.168.1.100"`
	CollectAt string         `json:"collectAt" dc:"数据上报时间" eg:"2025-01-01 12:00:00"`
	Server    *ServerBasic   `json:"server" dc:"服务器基本信息" eg:""`
	CPU       *CPUMetrics    `json:"cpu" dc:"CPU指标" eg:""`
	Memory    *MemoryMetrics `json:"memory" dc:"内存指标" eg:""`
	Disks     []*DiskMetrics `json:"disks" dc:"磁盘使用情况" eg:"[]"`
	Network   *NetMetrics    `json:"network" dc:"网络流量指标" eg:""`
	GoInfo    *GoMetrics     `json:"goInfo" dc:"Go运行时指标" eg:""`
}

// ServerBasic represents basic server information.
type ServerBasic struct {
	Hostname  string `json:"hostname" dc:"主机名" eg:"my-server"`
	OS        string `json:"os" dc:"操作系统" eg:"linux"`
	Arch      string `json:"arch" dc:"系统架构" eg:"amd64"`
	BootTime  string `json:"bootTime" dc:"系统启动时间" eg:"2025-01-01 00:00:00"`
	Uptime    uint64 `json:"uptime" dc:"系统运行时长（秒）" eg:"86400"`
	StartTime string `json:"startTime" dc:"服务启动时间" eg:"2025-01-01 08:00:00"`
}

// CPUMetrics represents CPU metrics.
type CPUMetrics struct {
	Cores        int     `json:"cores" dc:"CPU核心数" eg:"8"`
	ModelName    string  `json:"modelName" dc:"CPU型号" eg:"Intel Core i7-12700"`
	UsagePercent float64 `json:"usagePercent" dc:"CPU使用率（百分比）" eg:"45.5"`
}

// MemoryMetrics represents memory metrics.
type MemoryMetrics struct {
	Total        uint64  `json:"total" dc:"总内存（字节）" eg:"17179869184"`
	Used         uint64  `json:"used" dc:"已用内存（字节）" eg:"8589934592"`
	Available    uint64  `json:"available" dc:"可用内存（字节）" eg:"8589934592"`
	UsagePercent float64 `json:"usagePercent" dc:"内存使用率（百分比）" eg:"50.0"`
}

// DiskMetrics represents disk metrics.
type DiskMetrics struct {
	Path         string  `json:"path" dc:"挂载点路径" eg:"/"`
	FsType       string  `json:"fsType" dc:"文件系统类型" eg:"ext4"`
	Total        uint64  `json:"total" dc:"总容量（字节）" eg:"107374182400"`
	Used         uint64  `json:"used" dc:"已用容量（字节）" eg:"53687091200"`
	Free         uint64  `json:"free" dc:"可用容量（字节）" eg:"53687091200"`
	UsagePercent float64 `json:"usagePercent" dc:"使用率（百分比）" eg:"50.0"`
}

// NetMetrics represents network metrics.
type NetMetrics struct {
	BytesSent uint64  `json:"bytesSent" dc:"总发送字节数" eg:"1073741824"`
	BytesRecv uint64  `json:"bytesRecv" dc:"总接收字节数" eg:"2147483648"`
	SendRate  float64 `json:"sendRate" dc:"发送速率（字节/秒）" eg:"102400"`
	RecvRate  float64 `json:"recvRate" dc:"接收速率（字节/秒）" eg:"204800"`
}

// GoMetrics represents Go runtime metrics.
type GoMetrics struct {
	Version       string  `json:"version" dc:"Go版本" eg:"go1.22.0"`
	Goroutines    int     `json:"goroutines" dc:"Goroutine数量" eg:"42"`
	ProcessCPU    float64 `json:"processCpu" dc:"服务CPU使用率（百分比）" eg:"2.5"`
	ProcessMemory float64 `json:"processMemory" dc:"服务内存使用率（百分比）" eg:"1.8"`
	GCPauseNs     uint64  `json:"gcPauseNs" dc:"最近一次GC暂停时间（纳秒）" eg:"150000"`
	GfVersion     string  `json:"gfVersion" dc:"GoFrame版本" eg:"v2.10.0"`
	ServiceUptime string  `json:"serviceUptime" dc:"服务运行时长" eg:"3天 2小时 15分钟"`
}

// DBMetrics represents database metrics.
type DBMetrics struct {
	Version      string `json:"version" dc:"数据库版本" eg:"8.0.35"`
	MaxOpenConns int    `json:"maxOpenConns" dc:"最大连接数" eg:"100"`
	OpenConns    int    `json:"openConns" dc:"当前打开连接数" eg:"10"`
	InUse        int    `json:"inUse" dc:"使用中连接数" eg:"5"`
	Idle         int    `json:"idle" dc:"空闲连接数" eg:"5"`
}
