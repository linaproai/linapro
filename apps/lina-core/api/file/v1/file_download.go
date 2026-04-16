package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DownloadReq defines the request for downloading a file.
type DownloadReq struct {
	g.Meta `path:"/file/download/{id}" method:"get" tags:"文件管理" summary:"下载文件" dc:"根据文件ID下载文件，返回文件二进制内容" permission:"system:file:download"`
	Id     int64 `json:"id" v:"required" dc:"文件ID" eg:"1"`
}

// DownloadRes File download response
type DownloadRes struct {
	g.Meta `mime:"application/octet-stream"`
}
