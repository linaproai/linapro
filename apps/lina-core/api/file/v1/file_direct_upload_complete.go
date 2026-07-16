package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DirectUploadCompleteReq finishes a client direct upload and creates sys_file metadata.
type DirectUploadCompleteReq struct {
	g.Meta          `path:"/file/direct-upload/complete" method:"post" tags:"File Management" summary:"Complete direct file upload" dc:"Validate that the object written by the client exists at the host-assigned key (or complete cloud multipart with parts), then create the file-center metadata record. Completing an already completed session is idempotent and returns the same file metadata." permission:"system:file:upload"`
	UploadSessionId string              `json:"uploadSessionId" v:"required" dc:"Session id returned by direct-upload init" eg:"a1b2c3d4e5f6789012345678abcdef01"`
	ETag            string              `json:"etag" dc:"Optional ETag returned by the cloud provider after single-object put" eg:"\"abc123\""`
	Parts           []MultipartPartItem `json:"parts" dc:"Required for direct multipart sessions: completed part numbers and etags"`
}

// DirectUploadCompleteRes returns the registered file metadata after complete.
type DirectUploadCompleteRes = UploadRes
