package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DirectDownloadReq issues optional client direct download access for one file.
type DirectDownloadReq struct {
	g.Meta `path:"/file/{id}/direct-download" method:"get" tags:"File Management" summary:"Issue direct file download access" dc:"Return a short-lived direct download description when the active backend supports client get access. When unsupported, mode is proxy and the client must use the standard download or public access URL." permission:"system:file:download"`
	Id     int64 `json:"id" in:"path" v:"required|min:1" dc:"File ID" eg:"1"`
}

// DirectDownloadRes returns neutral get access for one file.
type DirectDownloadRes struct {
	Access   *DirectUploadAccess `json:"access" dc:"Client transfer description for get; proxy means host-mediated download"`
	ProxyUrl string              `json:"proxyUrl,omitempty" dc:"Host-mediated download URL when mode is proxy" eg:"/api/v1/file/1/download"`
}
