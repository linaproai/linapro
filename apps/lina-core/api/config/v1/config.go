// This file defines shared system-configuration response DTOs for the config API.
package v1

import "github.com/gogf/gf/v2/os/gtime"

// ConfigItem exposes configuration fields visible to management callers.
type ConfigItem struct {
	Id        int64       `json:"id" dc:"Config parameter ID" eg:"1"`
	Name      string      `json:"name" dc:"Config parameter name" eg:"Main frame page"`
	Key       string      `json:"key" dc:"Config parameter key" eg:"sys.index"`
	Value     string      `json:"value" dc:"Config parameter value" eg:"/dashboard"`
	IsBuiltin int         `json:"isBuiltin" dc:"Built-in record flag: 1=yes 0=no" eg:"1"`
	Remark    string      `json:"remark" dc:"Remark" eg:"Default route"`
	CreatedAt *gtime.Time `json:"createdAt" dc:"Creation time" eg:"2026-05-14 10:00:00"`
	UpdatedAt *gtime.Time `json:"updatedAt" dc:"Update time" eg:"2026-05-14 10:00:00"`
}
