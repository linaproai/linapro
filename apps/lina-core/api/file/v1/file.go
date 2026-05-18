// This file defines shared file response DTOs for the file API.
package v1

import "github.com/gogf/gf/v2/os/gtime"

// FileItem exposes file metadata needed by management UI and file echo flows.
type FileItem struct {
	Id        int64       `json:"id" dc:"File ID" eg:"1"`
	Name      string      `json:"name" dc:"Stored file name" eg:"20260514_avatar.png"`
	Original  string      `json:"original" dc:"Original file name" eg:"avatar.png"`
	Suffix    string      `json:"suffix" dc:"File suffix" eg:"png"`
	Scene     string      `json:"scene" dc:"Usage scene" eg:"avatar"`
	Size      int64       `json:"size" dc:"File size in bytes" eg:"1024"`
	Url       string      `json:"url" dc:"File access URL" eg:"/resource/file/avatar.png"`
	CreatedBy int64       `json:"createdBy" dc:"Uploader user ID" eg:"1"`
	CreatedAt *gtime.Time `json:"createdAt" dc:"Creation time" eg:"2026-05-14 10:00:00"`
	UpdatedAt *gtime.Time `json:"updatedAt" dc:"Update time" eg:"2026-05-14 10:00:00"`
}
