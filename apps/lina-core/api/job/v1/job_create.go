package v1

import "github.com/gogf/gf/v2/frame/g"

// LogRetentionOption defines one task-level log retention override payload.
type LogRetentionOption struct {
	Mode  string `json:"mode" dc:"Log retention mode: days=retain by day count=retain by number of entries none=do not clean" eg:"days"`
	Value int64  `json:"value" dc:"Log retention threshold; must be greater than 0 when mode=days or count, can be 0 when mode=none" eg:"30"`
}

// JobPayload defines the shared mutable fields for scheduled job create and update APIs.
type JobPayload struct {
	GroupId              uint64              `json:"groupId" v:"required" dc:"The group ID to which it belongs" eg:"1"`
	Name                 string              `json:"name" v:"required|length:1,128" dc:"Task name, unique within the group" eg:"Task log cleaning"`
	Description          string              `json:"description" dc:"Task description" eg:"Clean execution logs by policy"`
	TaskType             string              `json:"taskType" v:"required|in:shell" dc:"Task type: The public creation/editing interface only allows shell=Shell tasks; Handler tasks are registered by the source code" eg:"shell"`
	HandlerRef           string              `json:"handlerRef" dc:"Handler is the only reference; the public creation/editing interface is always left blank." eg:""`
	Params               map[string]any      `json:"params" dc:"Handler parameter object; the public creation/editing interface is fixed to an empty object" eg:"{}"`
	TimeoutSeconds       int                 `json:"timeoutSeconds" d:"300" v:"required|min:1|max:86400" dc:"Execution timeout, unit is seconds, range 1-86400" eg:"300"`
	ShellCmd             string              `json:"shellCmd" dc:"Shell script content; required when taskType=shell" eg:"echo hello"`
	WorkDir              string              `json:"workDir" dc:"Shell working directory. If not passed, the current working directory of the host will be used." eg:"/tmp"`
	Env                  map[string]string   `json:"env" dc:"Shell environment variable key-value pairs, only used by Shell tasks" eg:"{\"FOO\":\"bar\"}"`
	CronExpr             string              `json:"cronExpr" v:"required|length:1,128" dc:"Cron expression supports 5 segments (minutes, hours, days, months, weeks) and 6 segments (seconds, minutes, hours, days, months, weeks); # seconds will be automatically added when 5 segments are saved." eg:"17 3 * * *"`
	Timezone             string              `json:"timezone" d:"Asia/Shanghai" dc:"Task time zone identifier" eg:"Asia/Shanghai"`
	Scope                string              `json:"scope" d:"master_only" v:"required|in:master_only,all_node" dc:"Scheduling scope: master_only=Only the master node executes all_node=All nodes execute" eg:"master_only"`
	Concurrency          string              `json:"concurrency" d:"singleton" v:"required|in:singleton,parallel" dc:"Concurrency strategy: singleton=single case parallel=parallel" eg:"singleton"`
	MaxConcurrency       int                 `json:"maxConcurrency" d:"1" v:"min:1|max:100" dc:"Maximum number of concurrencies; effective when concurrency=parallel" eg:"1"`
	MaxExecutions        int                 `json:"maxExecutions" d:"0" v:"min:0" dc:"Maximum number of executions: 0 = no limit" eg:"0"`
	Status               string              `json:"status" d:"disabled" v:"required|in:enabled,disabled" dc:"Task status: enabled=enabled disabled=disabled; paused_by_plugin is a system managed state and does not allow writing when creating or editing" eg:"enabled"`
	LogRetentionOverride *LogRetentionOption `json:"logRetentionOverride" dc:"Task-level log retention policy; if not passed, follow the system parameter cron.log.retention" eg:"{\"mode\":\"days\",\"value\":60}"`
}

// CreateReq defines the request for creating one scheduled job.
type CreateReq struct {
	g.Meta `path:"/job" method:"post" tags:"Job Scheduling / Scheduled Jobs" summary:"Create task" operLog:"create" dc:"Create a new user-built Shell scheduled job; Handler type tasks can only be registered by the host or plugin source code" permission:"system:job:add"`
	JobPayload
}

// CreateRes defines the response for creating one scheduled job.
type CreateRes struct {
	Id uint64 `json:"id" dc:"Create new task ID" eg:"1"`
}
