// This file defines shared scheduled-job group response DTOs for the jobgroup API.
package v1

import "github.com/gogf/gf/v2/os/gtime"

// JobGroupItem exposes scheduled-job group fields visible to management callers.
type JobGroupItem struct {
	Id        int64       `json:"id" dc:"Job group ID" eg:"1"`
	Code      string      `json:"code" dc:"Group code" eg:"default"`
	Name      string      `json:"name" dc:"Group name" eg:"Default group"`
	Remark    string      `json:"remark" dc:"Remark" eg:"Default scheduled-job group"`
	SortOrder int         `json:"sortOrder" dc:"Display order" eg:"0"`
	IsDefault int         `json:"isDefault" dc:"Default group flag: 1=yes 0=no" eg:"1"`
	CreatedAt *gtime.Time `json:"createdAt" dc:"Creation time" eg:"2026-05-14 10:00:00"`
	UpdatedAt *gtime.Time `json:"updatedAt" dc:"Update time" eg:"2026-05-14 10:00:00"`
}
