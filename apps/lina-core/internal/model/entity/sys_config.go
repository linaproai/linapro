// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysConfig is the golang structure for table sys_config.
type SysConfig struct {
	Id               int64      `json:"id"               orm:"id"                description:"Config parameter ID"`
	TenantId         int        `json:"tenantId"         orm:"tenant_id"         description:"Owning tenant ID, 0 means PLATFORM default"`
	Name             string     `json:"name"             orm:"name"              description:"Config parameter name"`
	Key              string     `json:"key"              orm:"key"               description:"Config parameter key"`
	Value            string     `json:"value"            orm:"value"             description:"Config parameter value"`
	IsBuiltin        int        `json:"isBuiltin"        orm:"is_builtin"        description:"Built-in record flag: 1=yes, 0=no"`
	Remark           string     `json:"remark"           orm:"remark"            description:"Remark"`
	CreatedAt        *time.Time `json:"createdAt"        orm:"created_at"        description:"Creation time"`
	UpdatedAt        *time.Time `json:"updatedAt"        orm:"updated_at"        description:"Modification time"`
	DeletedAt        *time.Time `json:"deletedAt"        orm:"deleted_at"        description:"Deletion time"`
	ValueType        string     `json:"valueType"        orm:"value_type"        description:"Parameter value input type: text, textarea, number, boolean, select, radio, multi_select, richtext"`
	Options          string     `json:"options"          orm:"options"           description:"JSON array of {label,value} options for select/radio/multi_select; empty for other types"`
	SystemManageable int        `json:"systemManageable" orm:"system_manageable" description:"Whether the parameter may be governed on the system parameter admin surface: 1=yes, 0=no"`
}
