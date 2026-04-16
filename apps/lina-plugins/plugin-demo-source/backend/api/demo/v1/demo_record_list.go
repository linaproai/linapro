// demo_record_list.go defines the request and response DTOs for querying
// plugin-demo-source records.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListRecordsReq is the request for querying plugin-demo-source records.
type ListRecordsReq struct {
	g.Meta   `path:"/plugins/plugin-demo-source/records" method:"get" tags:"源码插件示例" summary:"分页查询源码插件示例记录" dc:"分页查询 plugin-demo-source 示例页面中的业务记录，支持按标题关键字模糊筛选，用于演示源码插件安装 SQL 创建的数据表在页面中的增删查改操作" permission:"plugin-demo-source:example:view"`
	PageNum  int    `json:"pageNum" dc:"页码，不传时默认第1页" eg:"1"`
	PageSize int    `json:"pageSize" dc:"每页条数，不传时默认10条" eg:"10"`
	Keyword  string `json:"keyword" dc:"按记录标题模糊筛选，不传则查询全部" eg:"示例"`
}

// ListRecordsRes is the response for querying plugin-demo-source records.
type ListRecordsRes struct {
	List  []*RecordItem `json:"list" dc:"当前页记录列表" eg:"[]"`
	Total int           `json:"total" dc:"符合条件的记录总数" eg:"1"`
}

// RecordItem defines one plugin-demo-source record row.
type RecordItem struct {
	Id             int64  `json:"id" dc:"记录ID" eg:"1"`
	Title          string `json:"title" dc:"记录标题" eg:"源码插件 SQL 示例记录"`
	Content        string `json:"content" dc:"记录内容摘要" eg:"该记录用于演示源码插件页面如何操作安装 SQL 创建的数据表。"`
	AttachmentName string `json:"attachmentName" dc:"附件原始文件名，没有附件时返回空字符串" eg:"plugin-demo-source-note.txt"`
	HasAttachment  int    `json:"hasAttachment" dc:"是否存在附件：1=存在 0=不存在" eg:"1"`
	CreatedAt      string `json:"createdAt" dc:"创建时间" eg:"2026-04-16 18:00:00"`
	UpdatedAt      string `json:"updatedAt" dc:"更新时间" eg:"2026-04-16 18:05:00"`
}
