package v1

import "github.com/gogf/gf/v2/frame/g"

// InstallReq is the request for installing a plugin.
type InstallReq struct {
	g.Meta        `path:"/plugins/{id}/install" method:"post" tags:"Plugin Management" summary:"Install plugin" permission:"plugin:install" dc:"Execute the plugin's installation life cycle. The source plugin will run its manifest/sql installation SQL, synchronize the menu and management resources, and write the installed status at this stage; the dynamic plugin will continue to execute the runtime installation process. If the target is a dynamic plugin and declares resource-type hostServices (such as storage.resources.paths, network URL pattern, or data.resources.tables), this request will also submit the authorization result confirmed by the host."`
	Id            string                       `json:"id" v:"required|length:1,64" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
	Authorization *HostServiceAuthorizationReq `json:"authorization,omitempty" dc:"The hostServices authorization result after host confirmation; if not passed, the current release will be used by default and the confirmed snapshot will be used. If it has not been confirmed, it will be fully authorized according to the plugin declaration." eg:"{}"`
}

// InstallRes is the response for installing a plugin.
type InstallRes struct {
	Id        string `json:"id" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
	Installed int    `json:"installed" dc:"Installation status: 1=Installed 0=Not installed" eg:"1"`
	Enabled   int    `json:"enabled" dc:"Enabled status: 1=enabled 0=disabled" eg:"0"`
}
