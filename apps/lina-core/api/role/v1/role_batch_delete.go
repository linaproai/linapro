// This file defines DTOs for the role batch-delete API.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleBatchDeleteReq is the request structure for deleting multiple roles.
type RoleBatchDeleteReq struct {
	g.Meta `path:"/role" method:"delete" tags:"Role Management" summary:"Batch delete roles" dc:"Delete multiple roles by ID in a single transaction and clear their role-menu and user-role associations. The built-in administrator role cannot be deleted." permission:"system:role:remove"`
	Ids    []int `json:"ids" v:"required|min-length:1" dc:"Role ID list as a query array, e.g. ids[]=2&ids[]=3&ids[]=4" eg:"[2,3,4]"`
}

// RoleBatchDeleteRes is the response structure for deleting multiple roles.
type RoleBatchDeleteRes struct {
	g.Meta `mime:"application/json" example:"{}"`
}
