package v1

import "github.com/gogf/gf/v2/frame/g"

// DetailReq defines the request for querying file detail.
type DetailReq struct {
	g.Meta `path:"/file/detail/{id}" method:"get" tags:"File Management" summary:"Get file details" dc:"Query the complete details of the file based on the file ID, including basic file information, uploader name and usage scenarios" permission:"system:file:query"`
	Id     int64 `json:"id" v:"required" dc:"File ID" eg:"1"`
}

// DetailRes File detail response
type DetailRes struct {
	FileItem
	CreatedByName string `json:"createdByName" dc:"Uploader username" eg:"admin"`
	SceneLabel    string `json:"sceneLabel" dc:"Usage scene name" eg:"User avatar"`
}
