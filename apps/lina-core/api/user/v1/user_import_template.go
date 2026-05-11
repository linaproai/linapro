// This file defines user import template DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ImportTemplateReq defines the request for downloading the user import template.
type ImportTemplateReq struct {
	g.Meta `path:"/user/import-template" method:"get" tags:"User Management" summary:"Download import template" dc:"Download the user import Excel template file, including required fields and data format instructions" permission:"system:user:import"`
}

// ImportTemplateRes defines the response for downloading the user import template.
type ImportTemplateRes struct{}
