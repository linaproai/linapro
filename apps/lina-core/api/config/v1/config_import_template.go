// This file defines parameter setting import template DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ConfigImportTemplateReq defines the request for downloading import template.
type ConfigImportTemplateReq struct {
	g.Meta `path:"/config/import-template" method:"get" tags:"Parameter Settings" summary:"Download parameter setting import template" dc:"Download the parameter settings and import the Excel template file, including required fields and data format instructions" permission:"system:config:add"`
}

// ConfigImportTemplateRes is the response for template download.
type ConfigImportTemplateRes struct{}
