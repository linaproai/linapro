// Package v1 defines shared dictionary API DTOs and compact enum contracts.
package v1

import (
	"lina-core/pkg/fallbackoverride"
	"lina-core/pkg/statusflag"
)

// DictDataItem exposes dictionary data fields visible to management callers.
type DictDataItem struct {
	Id             int                   `json:"id" dc:"Dictionary data ID" eg:"1"`
	DictType       string                `json:"dictType" dc:"Dictionary type" eg:"sys_user_sex"`
	Label          string                `json:"label" dc:"Dictionary label" eg:"Male"`
	Value          string                `json:"value" dc:"Dictionary value" eg:"1"`
	Sort           int                   `json:"sort" dc:"Display order" eg:"1"`
	TagStyle       string                `json:"tagStyle" dc:"Tag style" eg:"primary"`
	CssClass       string                `json:"cssClass" dc:"CSS class name" eg:""`
	Status         statusflag.Enabled    `json:"status" dc:"Status: 0=disabled 1=enabled" eg:"1"`
	IsBuiltin      statusflag.YesNo      `json:"isBuiltin" dc:"Built-in record flag: 1=yes 0=no" eg:"1"`
	Remark         string                `json:"remark" dc:"Remark" eg:"Default option"`
	SourceTenantId int                   `json:"sourceTenantId" dc:"Tenant ID that owns the returned effective row; 0 means platform default" eg:"0"`
	IsFallback     bool                  `json:"isFallback" dc:"Whether this row is inherited from the platform default in the current tenant context" eg:"true"`
	CanEdit        bool                  `json:"canEdit" dc:"Whether the current context may directly edit this returned row" eg:"false"`
	CanOverride    bool                  `json:"canOverride" dc:"Whether the current tenant may create its own override for this fallback row" eg:"true"`
	OverrideMode   fallbackoverride.Mode `json:"overrideMode" dc:"Override action mode, such as none or createTenantOverride" eg:"createTenantOverride"`
	CreatedAt      *int64                `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1778733600000"`
	UpdatedAt      *int64                `json:"updatedAt" dc:"Update time as Unix timestamp in milliseconds" eg:"1778733600000"`
}

// DictTypeItem exposes dictionary type fields visible to management callers.
type DictTypeItem struct {
	Id                  int                   `json:"id" dc:"Dictionary type ID" eg:"1"`
	Name                string                `json:"name" dc:"Dictionary name" eg:"User gender"`
	Type                string                `json:"type" dc:"Dictionary type" eg:"sys_user_sex"`
	Status              statusflag.Enabled    `json:"status" dc:"Status: 0=disabled 1=enabled" eg:"1"`
	IsBuiltin           statusflag.YesNo      `json:"isBuiltin" dc:"Built-in record flag: 1=yes 0=no" eg:"1"`
	AllowTenantOverride bool                  `json:"allowTenantOverride" dc:"Whether tenants may create overrides for this platform dictionary type" eg:"true"`
	Remark              string                `json:"remark" dc:"Remark" eg:"Default dictionary"`
	SourceTenantId      int                   `json:"sourceTenantId" dc:"Tenant ID that owns the returned effective row; 0 means platform default" eg:"0"`
	IsFallback          bool                  `json:"isFallback" dc:"Whether this row is inherited from the platform default in the current tenant context" eg:"true"`
	CanEdit             bool                  `json:"canEdit" dc:"Whether the current context may directly edit this returned row" eg:"false"`
	CanOverride         bool                  `json:"canOverride" dc:"Whether the current tenant may create its own override for this fallback row" eg:"true"`
	OverrideMode        fallbackoverride.Mode `json:"overrideMode" dc:"Override action mode, such as none or createTenantOverride" eg:"createTenantOverride"`
	CreatedAt           *int64                `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1778733600000"`
	UpdatedAt           *int64                `json:"updatedAt" dc:"Update time as Unix timestamp in milliseconds" eg:"1778733600000"`
}

// DictTypeOptionItem exposes dictionary type option fields for selectors.
type DictTypeOptionItem struct {
	Id   int    `json:"id" dc:"Dictionary type ID" eg:"1"`
	Name string `json:"name" dc:"Dictionary name" eg:"User gender"`
	Type string `json:"type" dc:"Dictionary type" eg:"sys_user_sex"`
}
