package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserMsg Count API

// CountReq defines the request for retrieving unread message counts.
type CountReq struct {
	g.Meta `path:"/user/message/count" method:"get" tags:"User Messages" summary:"Get the number of unread messages" dc:"Get the number of unread messages of the currently logged in user, which is used to display the message icon corner on the front end."`
}

// CountRes Unread message count response
type CountRes struct {
	Count int `json:"count" dc:"Number of unread messages" eg:"5"`
}
