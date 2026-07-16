package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ChunkedUploadAbortReq aborts an in-flight chunked upload session.
type ChunkedUploadAbortReq struct {
	g.Meta          `path:"/file/upload/chunked/abort" method:"post" tags:"File Management" summary:"Abort chunked file upload" dc:"Discard an in-flight chunked upload session. Missing sessions are treated as successful no-ops." permission:"system:file:upload"`
	UploadSessionId string `json:"uploadSessionId" v:"required" dc:"Session id returned by chunked init" eg:"a1b2c3d4e5f6789012345678abcdef01"`
}

// ChunkedUploadAbortRes is an empty abort acknowledgement.
type ChunkedUploadAbortRes struct{}
