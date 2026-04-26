// This file declares the delete-notice request/response DTOs used by the
// content-notice source plugin.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Notice Delete API

// DeleteReq defines the request for deleting notices.
type DeleteReq struct {
	g.Meta `path:"/notice/{ids}" method:"delete" tags:"Notices" summary:"Delete notification or announcement" dc:"Delete one or more notifications or announcements" permission:"system:notice:remove"`
	Ids    string `json:"ids" v:"required" dc:"Announcement IDs, comma-separated" eg:"1,2,3"`
}

// DeleteRes Notice delete response
type DeleteRes struct{}
