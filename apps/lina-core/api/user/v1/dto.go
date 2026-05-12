package v1

// UserDTO is a data transfer object for user information.
// It excludes sensitive fields like password, tenantId, etc.
type UserDTO struct {
	Id          int      `json:"id"        dc:"User ID"`
	Username    string   `json:"username"  dc:"Username"`
	Nickname    string   `json:"nickname"  dc:"User nickname"`
	Email       string   `json:"email"     dc:"Email address"`
	Phone       string   `json:"phone"     dc:"Mobile phone number"`
	Sex         int      `json:"sex"       dc:"Gender: 0=unknown, 1=male, 2=female"`
	Avatar      string   `json:"avatar"    dc:"Avatar URL"`
	Status      int      `json:"status"    dc:"Status: 0=disabled, 1=enabled"`
	Remark      string   `json:"remark"    dc:"Remark"`
	LoginDate   *string  `json:"loginDate" dc:"Last login time"`
	CreatedAt   *string  `json:"createdAt" dc:"Creation time"`
	UpdatedAt   *string  `json:"updatedAt" dc:"Update time"`
	DeptId      int      `json:"deptId"    dc:"Department ID"`
	DeptName    string   `json:"deptName"  dc:"Department name"`
	PostIds     []int    `json:"postIds"   dc:"Position ID list"`
	RoleIds     []int    `json:"roleIds"   dc:"Role ID list"`
	TenantIds   []int    `json:"tenantIds" dc:"Tenant ID list when multi-tenancy is enabled"`
	TenantNames []string `json:"tenantNames" dc:"Tenant name list when multi-tenancy is enabled"`
}

// UserListDTO is a data transfer object for user list.
type UserListDTO struct {
	Id        int     `json:"id"        dc:"User ID"`
	Username  string  `json:"username"  dc:"Username"`
	Nickname  string  `json:"nickname"  dc:"User nickname"`
	Email     string  `json:"email"     dc:"Email address"`
	Phone     string  `json:"phone"     dc:"Mobile phone number"`
	Sex       int     `json:"sex"       dc:"Gender: 0=unknown, 1=male, 2=female"`
	Avatar    string  `json:"avatar"    dc:"Avatar URL"`
	Status    int     `json:"status"    dc:"Status: 0=disabled, 1=enabled"`
	Remark    string  `json:"remark"    dc:"Remark"`
	LoginDate *string `json:"loginDate" dc:"Last login time"`
	CreatedAt *string `json:"createdAt" dc:"Creation time"`
	DeptName  string  `json:"deptName"  dc:"Department name"`
}
