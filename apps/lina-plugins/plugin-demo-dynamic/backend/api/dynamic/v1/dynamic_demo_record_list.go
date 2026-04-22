// This file defines the demo-record list DTOs for the dynamic plugin sample.

package v1

import "github.com/gogf/gf/v2/util/gmeta"

// DemoRecordListReq is the request for querying dynamic demo records.
type DemoRecordListReq struct {
	gmeta.Meta `path:"/demo-records" method:"get" tags:"动态插件示例" summary:"分页查询动态插件示例记录" dc:"分页查询 plugin-demo-dynamic 示例页面中的业务记录，支持按标题关键字模糊筛选，用于演示动态插件安装 SQL 创建的数据表在页面中的增删查改操作" access:"login" permission:"plugin-demo-dynamic:record:view" operLog:"other"`
	PageNum    int    `json:"pageNum" dc:"页码，不传时默认 1" eg:"1"`
	PageSize   int    `json:"pageSize" dc:"每页条数，不传时默认 20，最大 100" eg:"20"`
	Keyword    string `json:"keyword" dc:"按记录标题关键字模糊筛选，不传则查询全部" eg:"SQL"`
}

// DemoRecordListRes is the response for querying dynamic demo records.
type DemoRecordListRes struct {
	List  []*DemoRecordItem `json:"list" dc:"当前页记录列表" eg:"[{\"id\":\"demo-record-1\",\"title\":\"动态插件 SQL 示例记录\"}]"`
	Total int               `json:"total" dc:"满足筛选条件的记录总数" eg:"1"`
}
