// Package v1 defines shared file API DTOs and compact enum contracts.
package v1

// FileItem exposes file metadata needed by management UI and file echo flows.
type FileItem struct {
	Id        int64  `json:"id" dc:"File ID" eg:"1"`
	Name      string `json:"name" dc:"Stored file name" eg:"20260514_avatar.png"`
	Original  string `json:"original" dc:"Original file name" eg:"avatar.png"`
	Suffix    string `json:"suffix" dc:"File suffix" eg:"png"`
	Scene     string `json:"scene" dc:"Usage scene" eg:"avatar"`
	Size      int64  `json:"size" dc:"File size in bytes" eg:"1024"`
	Url       string `json:"url" dc:"File access URL" eg:"/resource/file/avatar.png"`
	CreatedBy int64  `json:"createdBy" dc:"Uploader user ID" eg:"1"`
	CreatedAt *int64 `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1778733600000"`
	UpdatedAt *int64 `json:"updatedAt" dc:"Update time as Unix timestamp in milliseconds" eg:"1778733600000"`
}
