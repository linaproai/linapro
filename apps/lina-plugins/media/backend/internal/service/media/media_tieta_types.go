// This file defines Tieta authentication data structures used by media strategy resolution.

package media

// TietaUser describes the user identity returned by Tieta token validation.
type TietaUser struct {
	Id           int64  // Id is the Tieta user ID.
	DeptId       int64  // DeptId is the Tieta department ID.
	Username     string // Username is the Tieta login name.
	RealName     string // RealName is the Tieta display name.
	Mobile       string // Mobile is the user mobile number.
	UserType     string // UserType identifies internal or tenant users.
	CustomerCode string // CustomerCode is the Tieta customer code.
	TenantId     string // TenantId is the Tieta customer/tenant ID.
	DeptName     string // DeptName is the Tieta department name.
	RegionCode   int64  // RegionCode is the Tieta region code.
	OrgId        int64  // OrgId is the Tieta organization ID.
	Enable       bool   // Enable reports whether the Tieta user is enabled.
}

// tietaUserResponse models Tieta's user-info response envelope.
type tietaUserResponse struct {
	Msg  string         `json:"msg"`
	Code int            `json:"code"`
	Data *tietaUserInfo `json:"data"`
}

// tietaUserInfo models Tieta's user-info response body.
type tietaUserInfo struct {
	ID           int64  `json:"id"`
	DeptId       int64  `json:"deptId"`
	UserName     string `json:"username"`
	NickName     string `json:"nickName"`
	Phone        string `json:"phone"`
	UserType     string `json:"userType"`
	CustomerCode string `json:"customerCode"`
	CustomerId   string `json:"customerId"`
	DeptName     string `json:"deptName"`
	RegionCode   int64  `json:"regionCode"`
	OrgID        int64  `json:"orgId"`
	Enable       bool   `json:"enable"`
}

// tietaTenantDeviceResponse models Tieta's device-permission response envelope.
type tietaTenantDeviceResponse struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data bool   `json:"data"`
}
