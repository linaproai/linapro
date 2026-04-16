package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeleteReq defines the request for deleting a department.
type DeleteReq struct {
	g.Meta `path:"/dept/{id}" method:"delete" tags:"部门管理" summary:"删除部门" dc:"删除指定部门，如果该部门下存在子部门或关联用户则不允许删除，需先移除子部门和用户后再进行删除操作" permission:"system:dept:remove"`
	Id     int `json:"id" v:"required" dc:"待删除的部门ID" eg:"110"`
}

// DeleteRes defines the response for deleting a department.
type DeleteRes struct{}
