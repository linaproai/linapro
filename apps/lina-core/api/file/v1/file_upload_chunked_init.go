package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ChunkedUploadInitReq starts a host-mediated chunked upload session.
type ChunkedUploadInitReq struct {
	g.Meta      `path:"/file/upload/chunked/init" method:"post" tags:"File Management" summary:"Initialize chunked file upload" dc:"Create a short-lived host-mediated chunked upload session for large files when automatic multipart planning selects proxy multipart. Optional contentHash may enable instant reuse on the direct-upload path; this endpoint focuses on chunked transfer setup." permission:"system:file:upload"`
	Scene       string `json:"scene" v:"required" dc:"Usage scenario identification (required)" eg:"avatar"`
	FileName    string `json:"fileName" v:"required" dc:"Original file name used for suffix and storage name generation" eg:"video.mp4"`
	Size        int64  `json:"size" v:"required|min:1" dc:"Declared file size in bytes" eg:"157286400"`
	ContentType string `json:"contentType" dc:"Optional MIME type constraint for the upload" eg:"video/mp4"`
	ContentHash string `json:"contentHash" dc:"Optional SHA-256 hex digest of file content" eg:""`
}

// ChunkedUploadInitRes is the chunked-upload init response.
type ChunkedUploadInitRes struct {
	UploadSessionId string               `json:"uploadSessionId" dc:"Chunked upload session id for part, complete, or abort" eg:"a1b2c3d4e5f6789012345678abcdef01"`
	Strategy        *UploadStrategy      `json:"strategy,omitempty" dc:"Upload plan; channel is proxy and encoding is multipart"`
	Multipart       *UploadMultipartPlan `json:"multipart,omitempty" dc:"Multipart part size and concurrency hints"`
}
