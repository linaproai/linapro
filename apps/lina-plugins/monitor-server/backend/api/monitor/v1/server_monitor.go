package v1

import "github.com/gogf/gf/v2/frame/g"

// Server Monitor API

// ServerMonitorReq defines the request for retrieving server metrics.
type ServerMonitorReq struct {
	g.Meta   `path:"/monitor/server" method:"get" tags:"System Monitoring" summary:"Service monitoring" dc:"Query server monitoring data and return the latest CPU, memory, disk, network, Go runtime, and other metrics for each node." permission:"monitor:server:list"`
	NodeName string `json:"nodeName" dc:"Filter by node name, when omitted, all nodes will be returned" eg:"my-server"`
}

// ServerMonitorRes is the server monitor response.
type ServerMonitorRes struct {
	Nodes  []*ServerNodeInfo `json:"nodes" dc:"Monitoring data of each node" eg:"[]"`
	DBInfo *DBMetrics        `json:"dbInfo" dc:"Database indicator information" eg:""`
}

// ServerNodeInfo represents server node information.
type ServerNodeInfo struct {
	NodeName  string         `json:"nodeName" dc:"Node name (Hostname)" eg:"my-server"`
	NodeIp    string         `json:"nodeIp" dc:"Node IP address" eg:"192.168.1.100"`
	CollectAt string         `json:"collectAt" dc:"Data reporting time" eg:"2025-01-01 12:00:00"`
	Server    *ServerBasic   `json:"server" dc:"Server basic information" eg:""`
	CPU       *CPUMetrics    `json:"cpu" dc:"CPU indicators" eg:""`
	Memory    *MemoryMetrics `json:"memory" dc:"Memory metrics" eg:""`
	Disks     []*DiskMetrics `json:"disks" dc:"Disk usage" eg:"[]"`
	Network   *NetMetrics    `json:"network" dc:"Network traffic metrics" eg:""`
	GoInfo    *GoMetrics     `json:"goInfo" dc:"Go runtime metrics" eg:""`
}

// ServerBasic represents basic server information.
type ServerBasic struct {
	Hostname  string `json:"hostname" dc:"Hostname" eg:"my-server"`
	OS        string `json:"os" dc:"Operating system" eg:"linux"`
	Arch      string `json:"arch" dc:"System architecture" eg:"amd64"`
	BootTime  string `json:"bootTime" dc:"System startup time" eg:"2025-01-01 00:00:00"`
	Uptime    uint64 `json:"uptime" dc:"System running time (seconds)" eg:"86400"`
	StartTime string `json:"startTime" dc:"Service start time" eg:"2025-01-01 08:00:00"`
}

// CPUMetrics represents CPU metrics.
type CPUMetrics struct {
	Cores        int     `json:"cores" dc:"Number of CPU cores" eg:"8"`
	ModelName    string  `json:"modelName" dc:"CPU model" eg:"Intel Core i7-12700"`
	UsagePercent float64 `json:"usagePercent" dc:"CPU usage (percent)" eg:"45.5"`
}

// MemoryMetrics represents memory metrics.
type MemoryMetrics struct {
	Total        uint64  `json:"total" dc:"Total memory (bytes)" eg:"17179869184"`
	Used         uint64  `json:"used" dc:"Memory used (bytes)" eg:"8589934592"`
	Available    uint64  `json:"available" dc:"Available memory (bytes)" eg:"8589934592"`
	UsagePercent float64 `json:"usagePercent" dc:"Memory usage (percentage)" eg:"50.0"`
}

// DiskMetrics represents disk metrics.
type DiskMetrics struct {
	Path         string  `json:"path" dc:"Mount point path" eg:"/"`
	FsType       string  `json:"fsType" dc:"File system type" eg:"ext4"`
	Total        uint64  `json:"total" dc:"Total capacity (bytes)" eg:"107374182400"`
	Used         uint64  `json:"used" dc:"Used capacity (bytes)" eg:"53687091200"`
	Free         uint64  `json:"free" dc:"Available capacity (bytes)" eg:"53687091200"`
	UsagePercent float64 `json:"usagePercent" dc:"Usage (percentage)" eg:"50.0"`
}

// NetMetrics represents network metrics.
type NetMetrics struct {
	BytesSent uint64  `json:"bytesSent" dc:"Total number of bytes sent" eg:"1073741824"`
	BytesRecv uint64  `json:"bytesRecv" dc:"Total number of bytes received" eg:"2147483648"`
	SendRate  float64 `json:"sendRate" dc:"Send rate (bytes/second)" eg:"102400"`
	RecvRate  float64 `json:"recvRate" dc:"Receive rate (bytes/second)" eg:"204800"`
}

// GoMetrics represents Go runtime metrics.
type GoMetrics struct {
	Version       string  `json:"version" dc:"Go version" eg:"go1.22.0"`
	Goroutines    int     `json:"goroutines" dc:"Number of Goroutines" eg:"42"`
	ProcessCPU    float64 `json:"processCpu" dc:"Service CPU usage (percent)" eg:"2.5"`
	ProcessMemory float64 `json:"processMemory" dc:"Service memory usage (percent)" eg:"1.8"`
	GCPauseNs     uint64  `json:"gcPauseNs" dc:"Last GC pause time (nanoseconds)" eg:"150000"`
	GfVersion     string  `json:"gfVersion" dc:"GoFrame version" eg:"v2.10.0"`
	ServiceUptime string  `json:"serviceUptime" dc:"Service running time" eg:"3 days 2 hours 15 minutes"`
}

// DBMetrics represents database metrics.
type DBMetrics struct {
	Version      string `json:"version" dc:"Database version" eg:"8.0.35"`
	MaxOpenConns int    `json:"maxOpenConns" dc:"Maximum number of connections" eg:"100"`
	OpenConns    int    `json:"openConns" dc:"Number of currently open connections" eg:"10"`
	InUse        int    `json:"inUse" dc:"Number of connections in use" eg:"5"`
	Idle         int    `json:"idle" dc:"Number of idle connections" eg:"5"`
}
