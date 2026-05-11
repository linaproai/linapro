// This file defines combined dictionary import template DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ImportTemplateReq defines the request for downloading combined import template.
type ImportTemplateReq struct {
	g.Meta `path:"/dict/import-template" method:"get" tags:"Dictionary Management" summary:"Download Dictionary Management Import Template" dc:"Download the dictionary management and import Excel template file, which contains two Sheets of dictionary type and dictionary data. Each Sheet contains sample data and field descriptions." permission:"system:dict:add"`
}

// ImportTemplateRes is the response for template download.
type ImportTemplateRes struct{}
