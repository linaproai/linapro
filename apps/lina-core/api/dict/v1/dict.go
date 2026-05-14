// This file defines shared dictionary response DTOs for the dict API.
package v1

import "github.com/gogf/gf/v2/os/gtime"

// DictDataItem exposes dictionary data fields visible to management callers.
type DictDataItem struct {
	Id        int         `json:"id" dc:"Dictionary data ID" eg:"1"`
	DictType  string      `json:"dictType" dc:"Dictionary type" eg:"sys_user_sex"`
	Label     string      `json:"label" dc:"Dictionary label" eg:"Male"`
	Value     string      `json:"value" dc:"Dictionary value" eg:"1"`
	Sort      int         `json:"sort" dc:"Display order" eg:"1"`
	TagStyle  string      `json:"tagStyle" dc:"Tag style" eg:"primary"`
	CssClass  string      `json:"cssClass" dc:"CSS class name" eg:""`
	Status    int         `json:"status" dc:"Status: 0=disabled 1=enabled" eg:"1"`
	IsBuiltin int         `json:"isBuiltin" dc:"Built-in record flag: 1=yes 0=no" eg:"1"`
	Remark    string      `json:"remark" dc:"Remark" eg:"Default option"`
	CreatedAt *gtime.Time `json:"createdAt" dc:"Creation time" eg:"2026-05-14 10:00:00"`
	UpdatedAt *gtime.Time `json:"updatedAt" dc:"Update time" eg:"2026-05-14 10:00:00"`
}

// DictTypeItem exposes dictionary type fields visible to management callers.
type DictTypeItem struct {
	Id        int         `json:"id" dc:"Dictionary type ID" eg:"1"`
	Name      string      `json:"name" dc:"Dictionary name" eg:"User gender"`
	Type      string      `json:"type" dc:"Dictionary type" eg:"sys_user_sex"`
	Status    int         `json:"status" dc:"Status: 0=disabled 1=enabled" eg:"1"`
	IsBuiltin int         `json:"isBuiltin" dc:"Built-in record flag: 1=yes 0=no" eg:"1"`
	Remark    string      `json:"remark" dc:"Remark" eg:"Default dictionary"`
	CreatedAt *gtime.Time `json:"createdAt" dc:"Creation time" eg:"2026-05-14 10:00:00"`
	UpdatedAt *gtime.Time `json:"updatedAt" dc:"Update time" eg:"2026-05-14 10:00:00"`
}

// DictTypeOptionItem exposes dictionary type option fields for selectors.
type DictTypeOptionItem struct {
	Id   int    `json:"id" dc:"Dictionary type ID" eg:"1"`
	Name string `json:"name" dc:"Dictionary name" eg:"User gender"`
	Type string `json:"type" dc:"Dictionary type" eg:"sys_user_sex"`
}
