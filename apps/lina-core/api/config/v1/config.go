// Package v1 defines shared system-configuration API DTOs and compact enum contracts.
package v1

import (
	"lina-core/pkg/configvaluetype"
	"lina-core/pkg/fallbackoverride"
	"lina-core/pkg/statusflag"
)

// ConfigValueOption is one selectable label/value pair for enum-like parameter types.
type ConfigValueOption struct {
	Label string `json:"label" dc:"Display label for the option" eg:"Right"`
	Value string `json:"value" dc:"Stored option value" eg:"panel-right"`
}

// ConfigItem exposes configuration fields visible to management callers.
type ConfigItem struct {
	Id             int64                 `json:"id" dc:"Config parameter ID" eg:"1"`
	Name           string                `json:"name" dc:"Config parameter name" eg:"Main frame page"`
	Key            string                `json:"key" dc:"Config parameter key" eg:"sys.index"`
	Value          string                `json:"value" dc:"Config parameter value" eg:"/dashboard"`
	ValueType      configvaluetype.Code  `json:"valueType" dc:"Value input type: text, textarea, number, boolean, select, radio, multi_select, richtext" eg:"select"`
	Options        []ConfigValueOption   `json:"options" dc:"Selectable options for select/radio/multi_select; empty for other types" eg:"[{\"label\":\"Right\",\"value\":\"panel-right\"}]"`
	IsBuiltin      statusflag.YesNo      `json:"isBuiltin" dc:"Built-in record flag: 1=yes 0=no" eg:"1"`
	Remark         string                `json:"remark" dc:"Remark" eg:"Default route"`
	SourceTenantId int                   `json:"sourceTenantId" dc:"Tenant ID that owns the returned effective row; 0 means platform default" eg:"0"`
	IsFallback     bool                  `json:"isFallback" dc:"Whether this row is inherited from the platform default in the current tenant context" eg:"true"`
	CanEdit        bool                  `json:"canEdit" dc:"Whether the current context may directly edit this returned row" eg:"false"`
	CanOverride    bool                  `json:"canOverride" dc:"Whether the current tenant may create its own override for this fallback row" eg:"true"`
	OverrideMode   fallbackoverride.Mode `json:"overrideMode" dc:"Override action mode, such as none or createTenantOverride" eg:"createTenantOverride"`
	CreatedAt      *int64                `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1778733600000"`
	UpdatedAt      *int64                `json:"updatedAt" dc:"Update time as Unix timestamp in milliseconds" eg:"1778733600000"`
}
