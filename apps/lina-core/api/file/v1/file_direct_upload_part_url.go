package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DirectUploadPartURLReq issues short-lived client access for one direct multipart part.
type DirectUploadPartURLReq struct {
	g.Meta          `path:"/file/direct-upload/part-url" method:"post" tags:"File Management" summary:"Issue direct multipart part access" dc:"Create a short-lived signed access description for one multipart part of a direct upload session. Part numbers are 1-based." permission:"system:file:upload"`
	UploadSessionId string `json:"uploadSessionId" v:"required" dc:"Session id returned by direct-upload init" eg:"a1b2c3d4e5f6789012345678abcdef01"`
	PartNumber      int32  `json:"partNumber" v:"required|min:1" dc:"1-based part number" eg:"1"`
	Size            int64  `json:"size" dc:"Optional declared part size in bytes" eg:"8388608"`
}

// DirectUploadPartURLRes returns neutral part access for the client.
type DirectUploadPartURLRes struct {
	Access *DirectUploadAccess `json:"access" dc:"Neutral client transfer description for one part"`
}
