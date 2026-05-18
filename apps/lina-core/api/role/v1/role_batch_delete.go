// This file defines DTOs for the role batch-delete API.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleBatchDeleteReq is the request structure for deleting multiple roles.
type RoleBatchDeleteReq struct {
	g.Meta `path:"/role" method:"delete" tags:"Role Management" summary:"Batch delete roles" dc:"Delete multiple roles by ID in a single transaction and clear their role-menu and user-role associations. The built-in administrator role cannot be deleted." permission:"system:role:remove"`
	Ids    []int `json:"ids" v:"required|min-length:1" dc:"Role ID list supplied as repeated ids query parameters" eg:"2"`
}

// RoleBatchDeleteRes is the response structure for deleting multiple roles.
type RoleBatchDeleteRes struct {
	g.Meta `mime:"application/json" example:"{}"`
}
