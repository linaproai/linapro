package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteReq defines the request for deleting posts.
type DeleteReq struct {
	g.Meta `path:"/post/{ids}" method:"delete" tags:"Position Management" summary:"Delete position" dc:"Delete one or more positions. If the position has been assigned to the user, deletion is not allowed." permission:"system:post:remove"`
	Ids    string `json:"ids" v:"required" dc:"Position IDs, comma-separated" eg:"1,2,3"`
}

// DeleteRes defines the response for deleting posts.
type DeleteRes struct{}
