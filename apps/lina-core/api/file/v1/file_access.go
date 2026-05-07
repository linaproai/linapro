// This file declares the uploaded-file URL access API contract.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AccessReq defines the request for accessing an uploaded file by storage path.
type AccessReq struct {
	g.Meta `path:"/uploads/*path" method:"get" tags:"File Management" summary:"Access uploaded file" dc:"Read an uploaded file through the configured file storage backend by its recorded storage path. This public endpoint is intended for direct browser access to uploaded images and files, while still requiring a matching file metadata record and never resolving arbitrary local filesystem paths."`
	Path   string `json:"path" v:"required" dc:"Relative storage object path returned in the uploaded file URL; it must match a file metadata record and is resolved through the configured storage backend." eg:"2026/03/20260319_abc12345.png"`
}

// AccessRes is the binary response for accessing an uploaded file.
type AccessRes struct {
	g.Meta `mime:"application/octet-stream"`
}
