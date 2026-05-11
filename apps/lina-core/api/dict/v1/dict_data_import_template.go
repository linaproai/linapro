// This file defines dictionary data import template DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DataImportTemplateReq defines the request for downloading import template.
type DataImportTemplateReq struct {
	g.Meta `path:"/dict/data/import-template" method:"get" tags:"Dictionary Management" summary:"Download dictionary data import template" dc:"Download the dictionary data import Excel template file, including required fields and data format instructions" permission:"system:dict:add"`
}

// DataImportTemplateRes is the response for template download.
type DataImportTemplateRes struct{}
