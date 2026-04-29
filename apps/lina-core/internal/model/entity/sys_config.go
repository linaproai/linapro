// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysConfig is the golang structure for table sys_config.
type SysConfig struct {
	Id        uint64      `json:"id"        orm:"id"         description:"Config parameter ID"`
	Name      string      `json:"name"      orm:"name"       description:"Config parameter name"`
	Key       string      `json:"key"       orm:"key"        description:"Config parameter key"`
	Value     string      `json:"value"     orm:"value"      description:"Config parameter value"`
	Remark    string      `json:"remark"    orm:"remark"     description:"Remark"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Modification time"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"Deletion time"`
}
