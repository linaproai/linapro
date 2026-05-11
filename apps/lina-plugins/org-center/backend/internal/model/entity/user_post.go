// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UserPost is the golang structure for table user_post.
type UserPost struct {
	TenantId int `json:"tenantId" orm:"tenant_id" description:"Owning tenant ID, 0 means PLATFORM"`
	UserId   int `json:"userId"   orm:"user_id"   description:"User ID"`
	PostId   int `json:"postId"   orm:"post_id"   description:"Post ID"`
}
