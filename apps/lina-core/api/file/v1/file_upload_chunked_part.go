package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ChunkedUploadPartReq uploads one part of a host-mediated chunked session.
type ChunkedUploadPartReq struct {
	g.Meta          `path:"/file/upload/chunked/part" method:"post" mime:"multipart/form-data" tags:"File Management" summary:"Upload one chunked file part" dc:"Append one ordered part to a host-mediated chunked upload session. Part numbers are 1-based." permission:"system:file:upload"`
	UploadSessionId string            `json:"uploadSessionId" v:"required" dc:"Session id returned by chunked init" eg:"a1b2c3d4e5f6789012345678abcdef01"`
	PartNumber      int32             `json:"partNumber" v:"required|min:1" dc:"1-based part number" eg:"1"`
	File            *ghttp.UploadFile `json:"file" type:"file" dc:"Part payload file field"`
}

// ChunkedUploadPartRes acknowledges one uploaded part.
type ChunkedUploadPartRes struct {
	PartNumber    int32  `json:"partNumber" dc:"1-based part number" eg:"1"`
	ETag          string `json:"etag,omitempty" dc:"Part ETag when the backend returns one" eg:"\"abc123\""`
	ReceivedBytes int64  `json:"receivedBytes,omitempty" dc:"Total bytes accepted for the session so far" eg:"8388608"`
}
