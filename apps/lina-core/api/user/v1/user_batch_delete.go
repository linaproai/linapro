// This file defines DTOs for the user batch-delete API.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// BatchDeleteReq defines the request for deleting multiple users.
type BatchDeleteReq struct {
	g.Meta `path:"/user" method:"delete" tags:"User Management" summary:"Batch delete users" dc:"Delete multiple users by ID in a single transaction. The built-in administrator and the current signed-in user cannot be deleted." permission:"system:user:remove"`
	Ids    []int `json:"ids" v:"required|min-length:1" dc:"User ID list supplied as repeated ids query parameters" eg:"1"`
}

// BatchDeleteRes defines the response for deleting multiple users.
type BatchDeleteRes struct{}
