// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UserPost is the golang structure for table user_post.
type UserPost struct {
	UserId int `json:"userId" orm:"user_id" description:"用户ID"`
	PostId int `json:"postId" orm:"post_id" description:"岗位ID"`
}
