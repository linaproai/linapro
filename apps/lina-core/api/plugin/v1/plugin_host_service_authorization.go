// This file defines shared host service authorization DTOs for plugin APIs.

package v1

// HostServiceAuthorizationReq describes the host-confirmed authorization result
// submitted during plugin install or enable flows.
type HostServiceAuthorizationReq struct {
	// Services contains one authorization decision for each resource-scoped host service.
	Services []*HostServiceAuthorizationServiceReq `json:"services" dc:"The service authorization set after confirmation by the host; methods and governance target sets can be narrowed by service. Storage uses paths, network, etc. uses resourceRefs, and data uses tables; an empty array indicates that all resource applications are rejected this time." eg:"[]"`
}

// HostServiceAuthorizationServiceReq describes one service-level authorization decision.
type HostServiceAuthorizationServiceReq struct {
	// Service is the logical host service identifier.
	Service string `json:"service" dc:"Host service identifier, such as storage, network, data" eg:"network"`
	// Methods optionally narrows the requested methods; when omitted, all plugin-declared methods are used.
	Methods []string `json:"methods,omitempty" dc:"The set of methods confirmed by the host. If not passed, all methods declared by the plugin will be used." eg:"[\"request\"]"`
	// Paths lists the confirmed logical storage paths; an empty array rejects all paths under this service.
	Paths []string `json:"paths,omitempty" dc:"A set of logical paths after confirmation by the host, only used by the storage service; a single path or a directory prefix path can be declared. An empty array indicates that all path applications under the service are rejected." eg:"[\"reports/\"]"`
	// ResourceRefs lists the confirmed resource refs; an empty array rejects all resource refs under this service.
	ResourceRefs []string `json:"resourceRefs,omitempty" dc:"The collection of governance targets after host confirmation; network uses URL mode, low-priority services continue to use logical resourceRef, and an empty array means rejecting all resource applications under the service." eg:"[\"https://*.example.com/api\"]"`
	// Tables lists the confirmed data tables; an empty array rejects all tables under this service.
	Tables []string `json:"tables,omitempty" dc:"The collection of data tables after confirmation by the host. It is only used by the data service. An empty array means rejecting all table applications under the service." eg:"[\"sys_plugin_node_state\"]"`
}

// HostServicePermissionItem describes one requested or authorized host service block.
type HostServicePermissionItem struct {
	// Service is the logical host service identifier.
	Service string `json:"service" dc:"Host service identifier, such as runtime, cron, storage, network, data" eg:"storage"`
	// Methods lists the confirmed or requested methods.
	Methods []string `json:"methods" dc:"The set of methods allowed under this host service" eg:"[\"put\",\"get\"]"`
	// Paths lists the governed logical storage paths under this service.
	Paths []string `json:"paths,omitempty" dc:"The set of logical paths allowed to be accessed under this host service, only used by storage service" eg:"[\"reports/\"]"`
	// Tables lists the governed data tables under this service.
	Tables []string `json:"tables,omitempty" dc:"The collection of data tables allowed to be accessed under this host service, only used by data service" eg:"[\"sys_plugin_node_state\"]"`
	// TableItems lists the governed data tables together with host-resolved display comments.
	TableItems []*HostServicePermissionTableItem `json:"tableItems,omitempty" dc:"The data table display items under the host service are only used by the data service; when the host can parse the table-level description, comment will be returned at the same time." eg:"[]"`
	// CronItems lists the discovered cron registrations under this service.
	CronItems []*HostServicePermissionCronItem `json:"cronItems,omitempty" dc:"The scheduled job statement found under this host service is only used by cron service; it is used to display the task name, expression, scheduling range and concurrency strategy on the authorization review page." eg:"[]"`
	// Resources lists the governed resource refs under this service.
	Resources []*HostServicePermissionResourceItem `json:"resources,omitempty" dc:"A collection of governance targets under this host service; network uses URL mode, and low-priority services continue to use resourceRef" eg:"[]"`
}

// HostServicePermissionTableItem describes one governed data table descriptor.
type HostServicePermissionTableItem struct {
	// Name is the governed table name.
	Name string `json:"name" dc:"Data table name" eg:"sys_plugin_node_state"`
	// Comment is the host-resolved table comment when available.
	Comment string `json:"comment,omitempty" dc:"Table description parsed by the host; if it cannot be parsed, an empty string is returned." eg:"Plugin node status table"`
}

// HostServicePermissionCronItem describes one discovered cron declaration
// exposed through the cron host service.
type HostServicePermissionCronItem struct {
	// Name is the stable plugin-local cron job identifier.
	Name string `json:"name" dc:"Unique name of scheduled job" eg:"heartbeat"`
	// DisplayName is the UI-facing cron job title when provided by the plugin.
	DisplayName string `json:"displayName,omitempty" dc:"Scheduled job display name; use name if not provided" eg:"Dynamic plugin heartbeat"`
	// Description explains the cron job purpose for operators.
	Description string `json:"description,omitempty" dc:"Scheduled job descriptions make it easier for administrators to understand the purpose of the task on the authorization review page" eg:"Execute the built-in scheduled jobs of the dynamic plugin through the Wasm bridge, and accumulate the number of heartbeat executions."`
	// Pattern stores the raw cron expression declared by the plugin.
	Pattern string `json:"pattern" dc:"The raw scheduling expression declared by the plugin" eg:"# */10 * * * *"`
	// Timezone stores the UI display timezone for cron-style patterns.
	Timezone string `json:"timezone,omitempty" dc:"Task time zone; time zone identifier used in cron expression parsing and display" eg:"Asia/Shanghai"`
	// Scope stores the declared dispatch scope.
	Scope string `json:"scope" dc:"Scheduling scope: master_only=Only the master node executes all_node=All nodes execute" eg:"all_node"`
	// Concurrency stores the declared overlap strategy.
	Concurrency string `json:"concurrency" dc:"Concurrency strategy: singleton=single case execution parallel=allows parallel execution" eg:"singleton"`
	// MaxConcurrency stores the parallel overlap ceiling when enabled.
	MaxConcurrency int `json:"maxConcurrency,omitempty" dc:"Maximum number of concurrencies; only meaningful under parallel strategy, singleton strategy is usually 1" eg:"1"`
}

// HostServicePermissionResourceItem describes one governed target descriptor.
type HostServicePermissionResourceItem struct {
	// Ref is the governed target identifier or URL pattern.
	Ref string `json:"ref" dc:"Governance target identifier; if network uses URL mode, low-priority services continue to use resourceRef; storage and data do not use this field" eg:"https://*.example.com/api"`
	// AllowMethods optionally narrows nested business methods such as HTTP verbs for services that retain per-resource governance.
	AllowMethods []string `json:"allowMethods,omitempty" dc:"Resource-level method whitelist, which is only meaningful for low-priority services that still retain fine-grained governance fields; storage, network, and data no longer use this field" eg:"[\"GET\"]"`
	// HeaderAllowList lists request headers the plugin may set for one governed resource when the service supports it.
	HeaderAllowList []string `json:"headerAllowList,omitempty" dc:"The request header whitelist that can be set by the plugin is only meaningful for low-priority services that still retain fine-grained governance fields; storage, network, and data no longer use this field." eg:"[\"x-request-id\"]"`
	// TimeoutMs is the timeout budget in milliseconds for services that retain per-resource governance.
	TimeoutMs int `json:"timeoutMs,omitempty" dc:"The timeout of host governance, in milliseconds, is only meaningful for low-priority services that still retain fine-grained governance fields; storage, network, and data no longer use this field." eg:"3000"`
	// MaxBodyBytes is the request/response body size limit in bytes.
	MaxBodyBytes int `json:"maxBodyBytes,omitempty" dc:"The maximum request body or response body size for host management, in bytes. It is only meaningful for low-priority services that still retain fine-grained management fields; storage, network, and data no longer use this field." eg:"65536"`
	// Attributes carries service-specific governance metadata.
	Attributes map[string]string `json:"attributes,omitempty" dc:"Resource-level governance parameters are only meaningful for low-priority services that still retain fine-grained governance fields; storage, network, and data no longer use this field." eg:"{}"`
}
