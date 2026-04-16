package v1

import "github.com/gogf/gf/v2/frame/g"

// HostCallDemoReq is the request for invoking the host call demo endpoint.
type HostCallDemoReq struct {
	g.Meta `path:"/host-call-demo" method:"get" tags:"动态插件示例" summary:"宿主调用能力演示" dc:"演示动态插件通过统一宿主服务模型调用 runtime、storage、network 和 data 四类核心能力。接口会写入运行时日志、读写插件隔离存储、访问受治理上游并对宿主确认授权的数据表执行结构化 CRUD。若 query 中传 skipNetwork=1，则跳过外部网络请求，便于离线验证其他能力。" access:"login" permission:"plugin-demo-dynamic:backend:view" operLog:"other"`
}

// HostCallDemoRes is the response for the host call demo endpoint.
type HostCallDemoRes struct {
	VisitCount int                     `json:"visitCount" dc:"当前累计访问次数，通过 runtime.state 宿主服务实现持久化计数" eg:"1"`
	PluginID   string                  `json:"pluginId" dc:"当前插件唯一标识" eg:"plugin-demo-dynamic"`
	Runtime    *HostCallDemoRuntimeRes `json:"runtime" dc:"runtime 宿主服务返回的基础信息摘要" eg:"{\"now\":\"2026-04-14T10:00:00+08:00\",\"uuid\":\"0d63c6a3-ec9d-4e39-a14f-d9b165a21ef9\",\"node\":\"node-1\"}"`
	Storage    *HostCallDemoStorageRes `json:"storage" dc:"storage 宿主服务执行摘要" eg:"{\"pathPrefix\":\"host-call-demo/\",\"objectPath\":\"host-call-demo/demo.json\",\"stored\":true,\"listedCount\":1,\"deleted\":true}"`
	Network    *HostCallDemoNetworkRes `json:"network" dc:"network 宿主服务执行摘要" eg:"{\"url\":\"https://example.com\",\"skipped\":false,\"statusCode\":200,\"contentType\":\"text/html\"}"`
	Data       *HostCallDemoDataRes    `json:"data" dc:"data 宿主服务执行摘要" eg:"{\"table\":\"sys_plugin_node_state\",\"recordKey\":\"101\",\"listTotal\":1,\"countTotal\":1,\"updated\":true,\"deleted\":true}"`
	Message    string                  `json:"message" dc:"宿主调用演示说明信息" eg:"Host service demo executed through runtime, storage, network, and data services."`
}

// HostCallDemoRuntimeRes describes runtime service results.
type HostCallDemoRuntimeRes struct {
	Now  string `json:"now" dc:"宿主当前时间字符串" eg:"2026-04-14T10:00:00+08:00"`
	UUID string `json:"uuid" dc:"宿主生成的唯一标识，用于本次演示资源隔离" eg:"0d63c6a3-ec9d-4e39-a14f-d9b165a21ef9"`
	Node string `json:"node" dc:"当前宿主节点标识" eg:"node-1"`
}

// HostCallDemoStorageRes describes storage service results.
type HostCallDemoStorageRes struct {
	PathPrefix  string `json:"pathPrefix" dc:"本次使用的已授权逻辑路径前缀" eg:"host-call-demo/"`
	ObjectPath  string `json:"objectPath" dc:"本次写入的逻辑对象路径" eg:"host-call-demo/demo.json"`
	Stored      bool   `json:"stored" dc:"是否成功写入并读取回对象" eg:"true"`
	ListedCount int    `json:"listedCount" dc:"按前缀列出的对象数量" eg:"1"`
	Deleted     bool   `json:"deleted" dc:"是否成功删除临时对象" eg:"true"`
}

// HostCallDemoNetworkRes describes network service results.
type HostCallDemoNetworkRes struct {
	URL         string `json:"url" dc:"本次申请并访问的目标 URL 或 URL 模式" eg:"https://example.com"`
	Skipped     bool   `json:"skipped" dc:"是否通过 skipNetwork=1 跳过了网络请求" eg:"false"`
	StatusCode  int    `json:"statusCode" dc:"上游 HTTP 状态码，跳过或失败时为 0" eg:"200"`
	ContentType string `json:"contentType" dc:"上游响应内容类型" eg:"text/html"`
	BodyPreview string `json:"bodyPreview" dc:"上游响应体预览，最多返回前 120 个字符" eg:"<!doctype html>"`
	Error       string `json:"error" dc:"网络请求失败时的错误摘要；成功或跳过时为空" eg:"Get https://example.com: context deadline exceeded"`
}

// HostCallDemoDataRes describes data service results.
type HostCallDemoDataRes struct {
	Table      string `json:"table" dc:"本次使用的宿主确认授权数据表名" eg:"sys_plugin_node_state"`
	RecordKey  string `json:"recordKey" dc:"宿主返回的记录主键值" eg:"101"`
	ListTotal  int    `json:"listTotal" dc:"按过滤条件分页查询到的记录总数" eg:"1"`
	CountTotal int    `json:"countTotal" dc:"按相同过滤条件执行 count 查询得到的总数" eg:"1"`
	Updated    bool   `json:"updated" dc:"是否成功完成更新并读取回更新后的记录" eg:"true"`
	Deleted    bool   `json:"deleted" dc:"是否成功删除临时记录" eg:"true"`
}
