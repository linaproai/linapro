package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateReq defines the request for creating a department.
type CreateReq struct {
	g.Meta   `path:"/dept" method:"post" tags:"Department Management" summary:"Create department" dc:"Create a new department. A parent department may be set to form a tree hierarchy, and the department code must be unique." permission:"system:dept:add"`
	ParentId int    `json:"parentId" d:"0" dc:"Parent department ID, 0 indicates top-level department" eg:"100"`
	Name     string `json:"name" v:"required#gf.gvalid.rule.required" dc:"Department name, cannot be repeated under the same parent" eg:"Technology Department"`
	Code     string `json:"code" dc:"Department code, a unique identifier within the system, used to interface with external systems" eg:"TECH"`
	OrderNum *int   `json:"orderNum" d:"0" dc:"Sorting number, the smaller the value, the higher it is. Departments at the same level are sorted in ascending order according to this field." eg:"1"`
	Leader   *int   `json:"leader" dc:"Owner user ID, associated with the system user table" eg:"1"`
	Phone    string `json:"phone" dc:"Department contact number" eg:"021-88888888"`
	Email    string `json:"email" dc:"Department contact email" eg:"tech@company.com"`
	Status   *int   `json:"status" d:"1" dc:"Department status: 1=normal, 0=disabled. After deactivation, users of this department and its sub-departments will not be able to log in." eg:"1"`
	Remark   string `json:"remark" dc:"Remark" eg:"Responsible for the company's technology research and development work"`
}

// CreateRes defines the response for creating a department.
type CreateRes struct {
	Id int `json:"id" dc:"Newly created department ID" eg:"110"`
}
