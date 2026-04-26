package v1

import "github.com/gogf/gf/v2/frame/g"

// EnableReq is the request for enabling plugin.
type EnableReq struct {
	g.Meta        `path:"/plugins/{id}/enable" method:"put" tags:"Plugin Management" summary:"Enable plugin" permission:"plugin:enable" dc:"Mark the specified plugin as enabled and write the plugin status configuration; if the plugin declares resource-type hostServices (such as storage.resources.paths, network URL pattern or data.resources.tables), this request will also submit the authorization result confirmed by the host."`
	Id            string                       `json:"id" v:"required|length:1,64" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
	Authorization *HostServiceAuthorizationReq `json:"authorization,omitempty" dc:"The hostServices authorization result after host confirmation; if not passed, the current release will be used by default and the confirmed snapshot will be used. If it has not been confirmed, it will be fully authorized according to the plugin declaration." eg:"{}"`
}

// EnableRes is the response for enabling plugin.
type EnableRes struct {
	Id      string `json:"id" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
	Enabled int    `json:"enabled" dc:"Enabled status: 1=enabled 0=disabled" eg:"1"`
}
