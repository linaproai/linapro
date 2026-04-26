package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UploadReq defines the request for uploading a file.
type UploadReq struct {
	g.Meta `path:"/file/upload" method:"post" mime:"multipart/form-data" tags:"File Management" summary:"Upload files" dc:"Upload a single file to the server, supporting common file formats, and the file information is automatically recorded in the file management table. scene is a required parameter, and the system will automatically record the relationship between the file and the usage scene." permission:"system:file:upload"`
	Scene  string `json:"scene" v:"required" dc:"Usage scenario identification (required): avatar=user avatar notice_image=notification announcement image notice_attachment=notification announcement attachment other=other" eg:"avatar"`
}

// UploadRes File upload response
type UploadRes struct {
	Id       int64  `json:"id" dc:"File ID" eg:"1"`
	Name     string `json:"name" dc:"Store file name" eg:"20260319_abc12345.png"`
	Original string `json:"original" dc:"original file name" eg:"avatar.png"`
	Url      string `json:"url" dc:"File access URL" eg:"/api/v1/uploads/2026/03/20260319_abc12345.png"`
	Suffix   string `json:"suffix" dc:"file suffix" eg:"png"`
	Size     int64  `json:"size" dc:"File size (bytes)" eg:"102400"`
}
