// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UserDept is the golang structure for table user_dept.
type UserDept struct {
	TenantId int `json:"tenantId" orm:"tenant_id" description:"Owning tenant ID, 0 means PLATFORM"`
	UserId   int `json:"userId"   orm:"user_id"   description:"User ID"`
	DeptId   int `json:"deptId"   orm:"dept_id"   description:"Department ID"`
}
