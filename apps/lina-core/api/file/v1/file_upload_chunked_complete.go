package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ChunkedUploadCompleteReq finishes a host-mediated chunked upload.
type ChunkedUploadCompleteReq struct {
	g.Meta          `path:"/file/upload/chunked/complete" method:"post" tags:"File Management" summary:"Complete chunked file upload" dc:"Assemble uploaded parts into the final object and create sys_file metadata. Completing an already completed session is idempotent." permission:"system:file:upload"`
	UploadSessionId string              `json:"uploadSessionId" v:"required" dc:"Session id returned by chunked init" eg:"a1b2c3d4e5f6789012345678abcdef01"`
	Parts           []MultipartPartItem `json:"parts" dc:"Optional parts list for cloud multipart complete; host-assembled sessions may omit this field"`
}

// MultipartPartItem is one completed part entry.
type MultipartPartItem struct {
	PartNumber int32  `json:"partNumber" v:"required|min:1" dc:"1-based part number" eg:"1"`
	ETag       string `json:"etag" dc:"Part ETag returned by the storage backend" eg:"\"abc123\""`
}

// ChunkedUploadCompleteRes returns registered file metadata after complete.
type ChunkedUploadCompleteRes = UploadRes
