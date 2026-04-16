// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysUserRole is the golang structure for table sys_user_role.
type SysUserRole struct {
	UserId int `json:"userId" orm:"user_id" description:"用户ID"`
	RoleId int `json:"roleId" orm:"role_id" description:"角色ID"`
}
