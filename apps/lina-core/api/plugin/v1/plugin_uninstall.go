package v1

import "github.com/gogf/gf/v2/frame/g"

// UninstallReq is the request for uninstalling a plugin.
type UninstallReq struct {
	g.Meta           `path:"/plugins/{id}" method:"delete" tags:"Plugin Management" summary:"Uninstall plugin" permission:"plugin:uninstall" dc:"Execute the plugin's uninstall life cycle. Both source plugins and dynamic plugins will deactivate the plugin at this stage, and you can click the confirmation option to decide whether to clean up the plugin's own storage data at the same time; after checking, the host will execute the uninstall SQL under manifest/sql/uninstall, and the dynamic plugin will also clean up the plugin's own storage files according to authorized storage paths."`
	Id               string `json:"id" v:"required|length:1,64" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
	PurgeStorageData *int   `json:"purgeStorageData" dc:"Whether to clear the plugin's own storage data at the same time when uninstalling the plugin: 1=Clear the data table data and associated files 0=Keep; clear by default if not uploaded" eg:"1"`
	Force            bool   `json:"force" dc:"Whether to force uninstall after lifecycle guard vetoes; requires plugin.allowForceUninstall=true" eg:"false"`
}

// UninstallRes is the response for uninstalling a plugin.
type UninstallRes struct {
	Id        string `json:"id" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
	Installed int    `json:"installed" dc:"Installation status: 1=Installed 0=Not installed" eg:"0"`
	Enabled   int    `json:"enabled" dc:"Enabled status: 1=enabled 0=disabled" eg:"0"`
}
