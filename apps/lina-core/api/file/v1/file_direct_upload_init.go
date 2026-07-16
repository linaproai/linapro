package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DirectUploadInitReq starts a client direct upload session for the file center.
type DirectUploadInitReq struct {
	g.Meta      `path:"/file/direct-upload/init" method:"post" tags:"File Management" summary:"Initialize direct file upload" dc:"Create a short-lived direct upload session. When the active object backend supports client direct access, the response includes a neutral transfer description (presigned URL or form post). When only local or unsupported backends are available, mode is proxy and the client must use the standard multipart upload API. Optional contentHash enables instant reuse of an existing identical file." permission:"system:file:upload"`
	Scene       string `json:"scene" v:"required" dc:"Usage scenario identification (required)" eg:"avatar"`
	FileName    string `json:"fileName" v:"required" dc:"Original file name used for suffix and storage name generation" eg:"avatar.png"`
	Size        int64  `json:"size" v:"required|min:1" dc:"Declared file size in bytes" eg:"102400"`
	ContentType string `json:"contentType" dc:"Optional MIME type constraint for the upload" eg:"image/png"`
	ContentHash string `json:"contentHash" dc:"Optional SHA-256 hex digest of file content for instant reuse" eg:"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"`
}

// DirectUploadInitRes is the direct-upload init response.
type DirectUploadInitRes struct {
	// InstantReuse reports that contentHash matched an existing file and no upload is required.
	InstantReuse bool `json:"instantReuse" dc:"True when an existing identical file was reused without upload" eg:"false"`
	// UploadSessionId identifies the session for complete/abort when InstantReuse is false and mode is not only proxy without session.
	UploadSessionId string `json:"uploadSessionId,omitempty" dc:"Direct upload session id for complete or abort" eg:"a1b2c3d4e5f6789012345678abcdef01"`
	// Access is the neutral transfer description. Mode proxy means use multipart upload.
	Access *DirectUploadAccess `json:"access,omitempty" dc:"Client transfer description when upload is required"`
	// File is populated when InstantReuse is true.
	File *UploadRes `json:"file,omitempty" dc:"Existing file metadata when instant reuse succeeds"`
}

// DirectUploadAccess describes a vendor-neutral client transfer payload returned
// by direct-upload init and reused by direct-download. Callers branch on Mode
// and never on cloud provider IDs.
type DirectUploadAccess struct {
	Mode            string            `json:"mode" dc:"Transfer mode: presigned_url, form_post, temporary_credentials, or proxy" eg:"presigned_url"`
	Operation       string            `json:"operation" dc:"Transfer operation: put or get" eg:"put"`
	Method          string            `json:"method,omitempty" dc:"HTTP method for presigned_url mode" eg:"PUT"`
	URL             string            `json:"url,omitempty" dc:"Target URL for presigned_url or form_post modes" eg:"https://bucket.example.com/object?X-Amz-Signature=..."`
	Headers         map[string]string `json:"headers,omitempty" dc:"Required request headers for presigned_url mode"`
	FormFields      map[string]string `json:"formFields,omitempty" dc:"Required form fields for form_post mode"`
	AccessKeyID     string            `json:"accessKeyId,omitempty" dc:"Temporary access key id when mode is temporary_credentials" eg:""`
	SecretAccessKey string            `json:"secretAccessKey,omitempty" dc:"Temporary secret when mode is temporary_credentials" eg:""`
	SessionToken    string            `json:"sessionToken,omitempty" dc:"Temporary session token when mode is temporary_credentials" eg:""`
	ExpiresAt       int64             `json:"expiresAt,omitempty" dc:"Access expiry as Unix timestamp in milliseconds" eg:"1710000000000"`
}
