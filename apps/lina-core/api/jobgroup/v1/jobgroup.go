// Package v1 defines shared scheduled-job group API DTOs and compact enum contracts.
package v1

import (
	"lina-core/pkg/statusflag"
)

// JobGroupItem exposes scheduled-job group fields visible to management callers.
type JobGroupItem struct {
	Id        int64            `json:"id" dc:"Job group ID" eg:"1"`
	Code      string           `json:"code" dc:"Group code" eg:"default"`
	Name      string           `json:"name" dc:"Group name" eg:"Default group"`
	Remark    string           `json:"remark" dc:"Remark" eg:"Default scheduled-job group"`
	SortOrder int              `json:"sortOrder" dc:"Display order" eg:"0"`
	IsDefault statusflag.YesNo `json:"isDefault" dc:"Default group flag: 1=yes 0=no" eg:"1"`
	CreatedAt *int64           `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1778733600000"`
	UpdatedAt *int64           `json:"updatedAt" dc:"Update time as Unix timestamp in milliseconds" eg:"1778733600000"`
}
