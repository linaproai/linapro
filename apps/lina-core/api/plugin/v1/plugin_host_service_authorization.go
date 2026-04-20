// This file defines shared host service authorization DTOs for plugin APIs.

package v1

// HostServiceAuthorizationReq describes the host-confirmed authorization result
// submitted during plugin install or enable flows.
type HostServiceAuthorizationReq struct {
	// Services contains one authorization decision for each resource-scoped host service.
	Services []*HostServiceAuthorizationServiceReq `json:"services" dc:"宿主确认后的服务授权集合；可按 service 收窄 methods 与治理目标集合，storage 使用 paths，network 等使用 resourceRefs，data 使用 tables；空数组表示本次拒绝全部资源申请" eg:"[]"`
}

// HostServiceAuthorizationServiceReq describes one service-level authorization decision.
type HostServiceAuthorizationServiceReq struct {
	// Service is the logical host service identifier.
	Service string `json:"service" dc:"宿主服务标识，如 storage、network、data" eg:"network"`
	// Methods optionally narrows the requested methods; 不传则沿用插件声明的全部 methods.
	Methods []string `json:"methods,omitempty" dc:"宿主确认后的 method 集合，不传则沿用插件声明的全部 methods" eg:"[\"request\"]"`
	// Paths lists the confirmed logical storage paths; 空数组表示拒绝该 service 下全部 path.
	Paths []string `json:"paths,omitempty" dc:"宿主确认后的逻辑路径集合，仅 storage service 使用；可声明单路径或目录前缀路径，空数组表示拒绝该 service 下全部路径申请" eg:"[\"reports/\"]"`
	// ResourceRefs lists the confirmed resource refs; 空数组表示拒绝该 service 下全部 resourceRef.
	ResourceRefs []string `json:"resourceRefs,omitempty" dc:"宿主确认后的治理目标集合；network 使用 URL 模式，低优先级服务继续使用逻辑 resourceRef，空数组表示拒绝该 service 下全部资源申请" eg:"[\"https://*.example.com/api\"]"`
	// Tables lists the confirmed data tables; 空数组表示拒绝该 service 下全部 tables.
	Tables []string `json:"tables,omitempty" dc:"宿主确认后的数据表集合，仅 data service 使用，空数组表示拒绝该 service 下全部表申请" eg:"[\"sys_plugin_node_state\"]"`
}

// HostServicePermissionItem describes one requested or authorized host service block.
type HostServicePermissionItem struct {
	// Service is the logical host service identifier.
	Service string `json:"service" dc:"宿主服务标识，如 runtime、cron、storage、network、data" eg:"storage"`
	// Methods lists the confirmed or requested methods.
	Methods []string `json:"methods" dc:"该宿主服务下允许的方法集合" eg:"[\"put\",\"get\"]"`
	// Paths lists the governed logical storage paths under this service.
	Paths []string `json:"paths,omitempty" dc:"该宿主服务下允许访问的逻辑路径集合，仅 storage service 使用" eg:"[\"reports/\"]"`
	// Tables lists the governed data tables under this service.
	Tables []string `json:"tables,omitempty" dc:"该宿主服务下允许访问的数据表集合，仅 data service 使用" eg:"[\"sys_plugin_node_state\"]"`
	// TableItems lists the governed data tables together with host-resolved display comments.
	TableItems []*HostServicePermissionTableItem `json:"tableItems,omitempty" dc:"该宿主服务下的数据表展示项，仅 data service 使用；当宿主可解析表级说明时会同时返回 comment" eg:"[]"`
	// CronItems lists the discovered cron registrations under this service.
	CronItems []*HostServicePermissionCronItem `json:"cronItems,omitempty" dc:"该宿主服务下发现的定时任务声明，仅 cron service 使用；用于在授权审查页展示任务名称、表达式、调度范围与并发策略" eg:"[]"`
	// Resources lists the governed resource refs under this service.
	Resources []*HostServicePermissionResourceItem `json:"resources,omitempty" dc:"该宿主服务下的治理目标集合；network 使用 URL 模式，低优先级服务继续使用 resourceRef" eg:"[]"`
}

// HostServicePermissionTableItem describes one governed data table descriptor.
type HostServicePermissionTableItem struct {
	// Name is the governed table name.
	Name string `json:"name" dc:"数据表名称" eg:"sys_plugin_node_state"`
	// Comment is the host-resolved table comment when available.
	Comment string `json:"comment,omitempty" dc:"宿主解析到的表说明；解析不到时返回空字符串" eg:"插件节点状态表"`
}

// HostServicePermissionCronItem describes one discovered cron declaration
// exposed through the cron host service.
type HostServicePermissionCronItem struct {
	// Name is the stable plugin-local cron job identifier.
	Name string `json:"name" dc:"定时任务唯一名称" eg:"heartbeat"`
	// DisplayName is the UI-facing cron job title when provided by the plugin.
	DisplayName string `json:"displayName,omitempty" dc:"定时任务显示名称；未提供时使用 name" eg:"动态插件心跳"`
	// Description explains the cron job purpose for operators.
	Description string `json:"description,omitempty" dc:"定时任务说明，便于管理员在授权审查页理解任务用途" eg:"通过 Wasm bridge 执行动态插件内置定时任务，并累计心跳执行次数。"`
	// Pattern stores the raw cron expression declared by the plugin.
	Pattern string `json:"pattern" dc:"插件声明的原始调度表达式" eg:"# */10 * * * *"`
	// Timezone stores the UI display timezone for cron-style patterns.
	Timezone string `json:"timezone,omitempty" dc:"任务时区；cron 表达式解析与展示使用的时区标识" eg:"Asia/Shanghai"`
	// Scope stores the declared dispatch scope.
	Scope string `json:"scope" dc:"调度范围：master_only=仅主节点执行 all_node=所有节点执行" eg:"all_node"`
	// Concurrency stores the declared overlap strategy.
	Concurrency string `json:"concurrency" dc:"并发策略：singleton=单例执行 parallel=允许并行执行" eg:"singleton"`
	// MaxConcurrency stores the parallel overlap ceiling when enabled.
	MaxConcurrency int `json:"maxConcurrency,omitempty" dc:"最大并发数；仅并行策略下有意义，单例策略通常为 1" eg:"1"`
}

// HostServicePermissionResourceItem describes one governed target descriptor.
type HostServicePermissionResourceItem struct {
	// Ref is the governed target identifier or URL pattern.
	Ref string `json:"ref" dc:"治理目标标识；如 network 使用 URL 模式，低优先级服务继续使用 resourceRef；storage 与 data 不使用该字段" eg:"https://*.example.com/api"`
	// AllowMethods optionally narrows nested business methods such as HTTP verbs for services that retain per-resource governance.
	AllowMethods []string `json:"allowMethods,omitempty" dc:"资源级 method 白名单，仅对后续仍保留细粒度治理字段的低优先级服务有意义；storage、network、data 不再使用该字段" eg:"[\"GET\"]"`
	// HeaderAllowList lists request headers the plugin may set for one governed resource when the service supports it.
	HeaderAllowList []string `json:"headerAllowList,omitempty" dc:"插件可设置的请求头白名单，仅对后续仍保留细粒度治理字段的低优先级服务有意义；storage、network、data 不再使用该字段" eg:"[\"x-request-id\"]"`
	// TimeoutMs is the timeout budget in milliseconds for services that retain per-resource governance.
	TimeoutMs int `json:"timeoutMs,omitempty" dc:"宿主治理的超时时间，单位毫秒，仅对后续仍保留细粒度治理字段的低优先级服务有意义；storage、network、data 不再使用该字段" eg:"3000"`
	// MaxBodyBytes is the request/response body size limit in bytes.
	MaxBodyBytes int `json:"maxBodyBytes,omitempty" dc:"宿主治理的最大请求体或响应体大小，单位字节，仅对后续仍保留细粒度治理字段的低优先级服务有意义；storage、network、data 不再使用该字段" eg:"65536"`
	// Attributes carries service-specific governance metadata.
	Attributes map[string]string `json:"attributes,omitempty" dc:"资源级治理参数，仅对后续仍保留细粒度治理字段的低优先级服务有意义；storage、network、data 不再使用该字段" eg:"{}"`
}
