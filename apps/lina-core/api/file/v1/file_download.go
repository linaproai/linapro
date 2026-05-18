package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DownloadReq defines the request for downloading a file.
type DownloadReq struct {
	g.Meta `path:"/file/download/{id}" method:"get" tags:"File Management" summary:"Download file" dc:"Download the file based on the file ID and return the binary content of the file" permission:"system:file:download"`
	Id     int64 `json:"id" v:"required" dc:"File ID" eg:"1"`
}

// DownloadRes File download response
type DownloadRes struct {
	g.Meta `mime:"application/octet-stream"`
}
