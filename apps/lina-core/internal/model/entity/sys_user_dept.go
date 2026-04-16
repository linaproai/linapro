// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysUserDept is the golang structure for table sys_user_dept.
type SysUserDept struct {
	UserId int `json:"userId" orm:"user_id" description:"用户ID"`
	DeptId int `json:"deptId" orm:"dept_id" description:"部门ID"`
}
