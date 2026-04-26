// This file declares the list-notice request/response DTOs used by the
// content-notice source plugin.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Notice List API

// ListReq defines the request for listing notices.
type ListReq struct {
	g.Meta    `path:"/notice" method:"get" tags:"Notices" summary:"Get notification and announcement list" dc:"Query notifications and announcements by page, with filtering by title, type, and creator." permission:"system:notice:query"`
	PageNum   int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize  int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	Title     string `json:"title" dc:"Filter by title (fuzzy match)" eg:"System maintenance"`
	Type      int    `json:"type" dc:"Filter by type: 1=Notice 2=Announcement" eg:"1"`
	CreatedBy string `json:"createdBy" dc:"Filter by creator username" eg:"admin"`
}

// ListRes Notice list response
type ListRes struct {
	List  []*ListItem `json:"list" dc:"Notification and announcement list" eg:"[]"`
	Total int         `json:"total" dc:"Total number of items" eg:"20"`
}

// ListItem Notice list item
type ListItem struct {
	*NoticeEntity
	CreatedByName string `json:"createdByName" dc:"Creator username" eg:"admin"`
}
