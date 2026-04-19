package v1

import "github.com/gogf/gf/v2/frame/g"

// LogRetentionOption defines one task-level log retention override payload.
type LogRetentionOption struct {
	Mode  string `json:"mode" dc:"日志保留模式：days=按天保留 count=按条数保留 none=不清理" eg:"days"`
	Value int64  `json:"value" dc:"日志保留阈值；mode=days 或 count 时必须大于0，mode=none 时可为0" eg:"30"`
}

// JobPayload defines the shared mutable fields for scheduled job create and update APIs.
type JobPayload struct {
	GroupId              uint64              `json:"groupId" v:"required" dc:"所属分组ID" eg:"1"`
	Name                 string              `json:"name" v:"required|length:1,128" dc:"任务名称，分组内唯一" eg:"任务日志清理"`
	Description          string              `json:"description" dc:"任务描述" eg:"按策略清理执行日志"`
	TaskType             string              `json:"taskType" v:"required|in:handler,shell" dc:"任务类型：handler=Handler 任务 shell=Shell 任务" eg:"handler"`
	HandlerRef           string              `json:"handlerRef" dc:"Handler 唯一引用；taskType=handler 时必填" eg:"host:cleanup-job-logs"`
	Params               map[string]any      `json:"params" dc:"Handler 参数对象；taskType=handler 时按处理器 Schema 校验" eg:"{}"`
	TimeoutSeconds       int                 `json:"timeoutSeconds" d:"300" v:"required|min:1|max:86400" dc:"执行超时时间，单位为秒，范围 1-86400" eg:"300"`
	ShellCmd             string              `json:"shellCmd" dc:"Shell 脚本内容；taskType=shell 时必填" eg:"echo hello"`
	WorkDir              string              `json:"workDir" dc:"Shell 工作目录，不传则使用宿主当前工作目录" eg:"/tmp"`
	Env                  map[string]string   `json:"env" dc:"Shell 环境变量键值对，仅 Shell 任务使用" eg:"{\"FOO\":\"bar\"}"`
	CronExpr             string              `json:"cronExpr" v:"required|length:1,128" dc:"Cron 表达式，支持5段（分 时 日 月 周）与6段（秒 分 时 日 月 周）；5段保存时会自动补 # 秒占位" eg:"17 3 * * *"`
	Timezone             string              `json:"timezone" d:"Asia/Shanghai" dc:"任务时区标识" eg:"Asia/Shanghai"`
	Scope                string              `json:"scope" d:"master_only" v:"required|in:master_only,all_node" dc:"调度范围：master_only=仅主节点执行 all_node=所有节点执行" eg:"master_only"`
	Concurrency          string              `json:"concurrency" d:"singleton" v:"required|in:singleton,parallel" dc:"并发策略：singleton=单例 parallel=并行" eg:"singleton"`
	MaxConcurrency       int                 `json:"maxConcurrency" d:"1" v:"min:1|max:100" dc:"最大并发数；concurrency=parallel 时生效" eg:"1"`
	MaxExecutions        int                 `json:"maxExecutions" d:"0" v:"min:0" dc:"最大执行次数：0=无限制" eg:"0"`
	Status               string              `json:"status" d:"disabled" v:"required|in:enabled,disabled,paused_by_plugin" dc:"任务状态：enabled=启用 disabled=停用 paused_by_plugin=插件处理器不可用" eg:"enabled"`
	LogRetentionOverride *LogRetentionOption `json:"logRetentionOverride" dc:"任务级日志保留策略；不传则跟随系统参数 cron.log.retention" eg:"{\"mode\":\"days\",\"value\":60}"`
}

// CreateReq defines the request for creating one scheduled job.
type CreateReq struct {
	g.Meta `path:"/job" method:"post" tags:"定时任务管理" summary:"创建任务" operLog:"create" dc:"创建一个新的定时任务，支持 Handler 与 Shell 两种任务类型" permission:"system:job:add"`
	JobPayload
}

// CreateRes defines the response for creating one scheduled job.
type CreateRes struct {
	Id uint64 `json:"id" dc:"新建任务ID" eg:"1"`
}
