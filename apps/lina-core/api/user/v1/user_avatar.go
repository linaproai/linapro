package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UpdateAvatarReq defines the request for updating the current user avatar.
type UpdateAvatarReq struct {
	g.Meta `path:"/user/profile/avatar" method:"put" tags:"用户管理" summary:"更新用户头像" dc:"更新当前用户头像URL，需先通过文件上传接口上传头像文件获取URL"`
	Avatar string `json:"avatar" v:"required" dc:"头像URL地址" eg:"/api/v1/uploads/2026/03/20260319_abc12345.png"`
}

// UpdateAvatarRes defines the response for updating the current user avatar.
type UpdateAvatarRes struct{}
