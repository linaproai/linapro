// Package v1 defines shared user API DTOs and compact enum contracts.
package v1

import (
	"lina-core/pkg/statusflag"
)

// UserItem exposes the user fields that are safe for API callers.
type UserItem struct {
	Id        int                `json:"id" dc:"User ID" eg:"1"`
	TenantId  int                `json:"tenantId" dc:"Primary/default tenant ID, 0 means platform" eg:"0"`
	Username  string             `json:"username" dc:"Username" eg:"admin"`
	Nickname  string             `json:"nickname" dc:"User nickname" eg:"Administrator"`
	Email     string             `json:"email" dc:"Email address" eg:"admin@example.com"`
	Phone     string             `json:"phone" dc:"Mobile phone number" eg:"13800138000"`
	Sex       int                `json:"sex" dc:"Gender: 0=unknown 1=male 2=female" eg:"1"`
	Avatar    string             `json:"avatar" dc:"Avatar URL" eg:"/resource/file/avatar.png"`
	Status    statusflag.Enabled `json:"status" dc:"Status: 0=disabled 1=enabled" eg:"1"`
	Remark    string             `json:"remark" dc:"Remark" eg:"System administrator"`
	LoginDate *int64             `json:"loginDate" dc:"Last login time as Unix timestamp in milliseconds" eg:"1778733600000"`
	CreatedAt *int64             `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1778733600000"`
	UpdatedAt *int64             `json:"updatedAt" dc:"Update time as Unix timestamp in milliseconds" eg:"1778733600000"`
}
