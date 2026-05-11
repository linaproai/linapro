// This file defines dictionary type import template DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// TypeImportTemplateReq defines the request for downloading import template.
type TypeImportTemplateReq struct {
	g.Meta `path:"/dict/type/import-template" method:"get" tags:"Dictionary Management" summary:"Download dictionary type import template" dc:"Download the dictionary type import Excel template file, including required fields and data format instructions" permission:"system:dict:add"`
}

// TypeImportTemplateRes is the response for template download.
type TypeImportTemplateRes struct{}
