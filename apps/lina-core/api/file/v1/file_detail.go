package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DetailReq defines the request for querying file detail.
type DetailReq struct {
	g.Meta `path:"/file/detail/{id}" method:"get" tags:"文件管理" summary:"获取文件详情" dc:"根据文件ID查询文件完整详细信息，包括文件基本信息、上传者名称和使用场景" permission:"system:file:query"`
	Id     int64 `json:"id" v:"required" dc:"文件ID" eg:"1"`
}

// DetailRes File detail response
type DetailRes struct {
	*entity.SysFile
	CreatedByName string `json:"createdByName" dc:"上传者用户名" eg:"admin"`
	SceneLabel    string `json:"sceneLabel" dc:"使用场景名称" eg:"用户头像"`
}
