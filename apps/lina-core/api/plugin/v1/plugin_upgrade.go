// This file defines the plugin runtime-upgrade execution API DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpgradeReq is the request for executing one confirmed plugin runtime upgrade.
type UpgradeReq struct {
	g.Meta        `path:"/plugins/{id}/upgrade" method:"post" tags:"Plugin Management" summary:"Upgrade plugin runtime state" permission:"plugin:install" dc:"Execute a confirmed plugin runtime upgrade for a pending-upgrade plugin. The host re-reads the database-effective manifest and file-discovered target manifest, validates the plugin is still pending_upgrade, and then runs the runtime upgrade flow. This endpoint has side effects: it may execute upgrade SQL, synchronize governance resources, switch the effective release, invalidate plugin runtime caches, and publish lifecycle events."`
	Id            string                       `json:"id" v:"required|length:1,64" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
	Confirmed     bool                         `json:"confirmed" v:"required|boolean" dc:"Explicit operator confirmation. Must be true before upgrade side effects are executed." eg:"true"`
	Authorization *HostServiceAuthorizationReq `json:"authorization,omitempty" dc:"The hostServices authorization result confirmed for the target dynamic release when the target manifest changes resource-scoped hostServices. If omitted, the host keeps the current target release authorization snapshot." eg:"{}"`
}

// UpgradeRes is the response for executing one plugin runtime upgrade.
type UpgradeRes struct {
	PluginId          string `json:"pluginId" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
	RuntimeState      string `json:"runtimeState" dc:"Plugin runtime upgrade state after the request" eg:"normal"`
	EffectiveVersion  string `json:"effectiveVersion" dc:"Database-effective plugin version after the request" eg:"v0.2.0"`
	DiscoveredVersion string `json:"discoveredVersion" dc:"Currently discovered plugin version after the request" eg:"v0.2.0"`
	FromVersion       string `json:"fromVersion" dc:"Database-effective plugin version observed before upgrade" eg:"v0.1.0"`
	ToVersion         string `json:"toVersion" dc:"Target plugin version requested by the operator" eg:"v0.2.0"`
	Executed          bool   `json:"executed" dc:"Whether upgrade side effects were executed by this request" eg:"true"`
}
