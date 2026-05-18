// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysRoleMenu is the golang structure for table sys_role_menu.
type SysRoleMenu struct {
	TenantId int `json:"tenantId" orm:"tenant_id" description:"Role-menu relation tenant ID, 0 means PLATFORM"`
	RoleId   int `json:"roleId"   orm:"role_id"   description:"Role ID"`
	MenuId   int `json:"menuId"   orm:"menu_id"   description:"Menu ID"`
}
