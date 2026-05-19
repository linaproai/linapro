// This file defines plugin disablement API DTOs and the typed enabled response
// flag returned after the lifecycle update.

package v1

import (
	"lina-core/pkg/statusflag"

	"github.com/gogf/gf/v2/frame/g"
)

// DisableReq is the request for disabling plugin.
type DisableReq struct {
	g.Meta `path:"/plugins/{id}/disable" method:"put" tags:"Plugin Management" summary:"Disable plugin" permission:"plugin:disable" dc:"Mark the specified plugin as disabled and write the plugin status configuration"`
	Id     string `json:"id" v:"required|length:1,64" dc:"Plugin unique identifier" eg:"linapro-demo-source"`
}

// DisableRes is the response for disabling plugin.
type DisableRes struct {
	Id      string             `json:"id" dc:"Plugin unique identifier" eg:"linapro-demo-source"`
	Enabled statusflag.Enabled `json:"enabled" dc:"Enabled status: 1=enabled 0=disabled" eg:"0"`
}
