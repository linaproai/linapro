package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ImportReq defines the request for importing users.
type ImportReq struct {
	g.Meta `path:"/user/import" method:"post" mime:"multipart/form-data" tags:"用户管理" summary:"导入用户数据" dc:"通过Excel文件批量导入用户数据，需使用系统提供的导入模板" permission:"system:user:import"`
}

// ImportRes is the response structure for user import.
type ImportRes struct {
	Success  int              `json:"success" dc:"成功条数" eg:"10"`
	Fail     int              `json:"fail" dc:"失败条数" eg:"2"`
	FailList []ImportFailItem `json:"failList" dc:"失败详情" eg:"[]"`
}

// ImportFailItem represents a failed import record.
type ImportFailItem struct {
	Row    int    `json:"row" dc:"行号" eg:"3"`
	Reason string `json:"reason" dc:"失败原因" eg:"用户名已存在"`
}

// ImportTemplateReq defines the request for downloading the user import template.
type ImportTemplateReq struct {
	g.Meta `path:"/user/import-template" method:"get" tags:"用户管理" summary:"下载导入模板" dc:"下载用户导入Excel模板文件，包含必填字段和数据格式说明" permission:"system:user:import"`
}

// ImportTemplateRes defines the response for downloading the user import template.
type ImportTemplateRes struct{}
