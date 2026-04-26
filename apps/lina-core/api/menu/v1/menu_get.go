package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// GetReq defines the request for querying menu detail.
type GetReq struct {
	g.Meta `path:"/menu/{id}" method:"get" tags:"Menu Management" summary:"Get menu details" dc:"Get menu details based on menu ID, including parent menu name" permission:"system:menu:query"`
	Id     int `json:"id" v:"required|min:1" dc:"Menu ID" eg:"1"`
}

// GetRes defines the response for querying menu detail.
type GetRes struct {
	*MenuItem  `dc:"Menu details" eg:""`
	ParentName string `json:"parentName" dc:"Parent menu name" eg:"System management"`
}
