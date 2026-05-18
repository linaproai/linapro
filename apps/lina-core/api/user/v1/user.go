// This file defines shared user response DTOs for the user API.
package v1

import "github.com/gogf/gf/v2/os/gtime"

// UserItem exposes the user fields that are safe for API callers.
type UserItem struct {
	Id        int         `json:"id" dc:"User ID" eg:"1"`
	TenantId  int         `json:"tenantId" dc:"Primary/default tenant ID, 0 means platform" eg:"0"`
	Username  string      `json:"username" dc:"Username" eg:"admin"`
	Nickname  string      `json:"nickname" dc:"User nickname" eg:"Administrator"`
	Email     string      `json:"email" dc:"Email address" eg:"admin@example.com"`
	Phone     string      `json:"phone" dc:"Mobile phone number" eg:"13800138000"`
	Sex       int         `json:"sex" dc:"Gender: 0=unknown 1=male 2=female" eg:"1"`
	Avatar    string      `json:"avatar" dc:"Avatar URL" eg:"/resource/file/avatar.png"`
	Status    int         `json:"status" dc:"Status: 0=disabled 1=enabled" eg:"1"`
	Remark    string      `json:"remark" dc:"Remark" eg:"System administrator"`
	LoginDate *gtime.Time `json:"loginDate" dc:"Last login time" eg:"2026-05-14 10:00:00"`
	CreatedAt *gtime.Time `json:"createdAt" dc:"Creation time" eg:"2026-05-14 10:00:00"`
	UpdatedAt *gtime.Time `json:"updatedAt" dc:"Update time" eg:"2026-05-14 10:00:00"`
}
