package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateReq defines the request for updating a department.
type UpdateReq struct {
	g.Meta   `path:"/dept/{id}" method:"put" tags:"Department Management" summary:"Update department" dc:"Update the specified department. Omitted fields remain unchanged, and a department cannot be moved under itself or one of its descendants." permission:"system:dept:edit"`
	Id       int     `json:"id" v:"required" dc:"Department ID" eg:"100"`
	ParentId *int    `json:"parentId" dc:"Parent department ID, 0 indicates the top-level department and cannot be set to itself or its subordinate departments" eg:"0"`
	Name     *string `json:"name" dc:"Department name, cannot be repeated under the same parent" eg:"R&D center"`
	Code     *string `json:"code" dc:"Department code, unique identifier in the system" eg:"RD"`
	OrderNum *int    `json:"orderNum" dc:"Sorting number, the smaller the value, the higher it is" eg:"2"`
	Leader   *int    `json:"leader" dc:"Owner user ID, associated with the system user table" eg:"1"`
	Phone    *string `json:"phone" dc:"Department contact number" eg:"021-66666666"`
	Email    *string `json:"email" dc:"Department contact email" eg:"rd@company.com"`
	Status   *int    `json:"status" dc:"Department status: 1=normal, 0=disabled" eg:"1"`
	Remark   *string `json:"remark" dc:"Remark" eg:"Responsible for the company's core product research and development"`
}

// UpdateRes defines the response for updating a department.
type UpdateRes struct{}
