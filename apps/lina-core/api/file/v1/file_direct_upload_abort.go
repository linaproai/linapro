package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DirectUploadAbortReq aborts an in-flight direct upload session.
type DirectUploadAbortReq struct {
	g.Meta          `path:"/file/direct-upload/abort" method:"post" tags:"File Management" summary:"Abort direct file upload" dc:"Discard an in-flight direct upload session. Missing sessions are treated as successful no-ops." permission:"system:file:upload"`
	UploadSessionId string `json:"uploadSessionId" v:"required" dc:"Session id returned by direct-upload init" eg:"a1b2c3d4e5f6789012345678abcdef01"`
}

// DirectUploadAbortRes is an empty abort acknowledgement.
type DirectUploadAbortRes struct{}
