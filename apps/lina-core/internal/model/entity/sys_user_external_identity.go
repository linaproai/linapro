// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysUserExternalIdentity is the golang structure for table sys_user_external_identity.
type SysUserExternalIdentity struct {
	Id            int        `json:"id"            orm:"id"             description:"External identity linkage ID"`
	UserId        int        `json:"userId"        orm:"user_id"        description:"Linked local sys_user ID"`
	Provider      string     `json:"provider"      orm:"provider"       description:"Stable external provider ID owned by the declaring plugin, e.g. google, discord"`
	Subject       string     `json:"subject"       orm:"subject"        description:"Immutable provider-issued subject identifier, e.g. OIDC sub"`
	PluginId      string     `json:"pluginId"      orm:"plugin_id"      description:"Source-plugin ID that owns the provider and created the linkage"`
	EmailSnapshot string     `json:"emailSnapshot" orm:"email_snapshot" description:"Email captured at link time for audit only, never used as a resolution key"`
	CreatedAt     *time.Time `json:"createdAt"     orm:"created_at"     description:"Creation time"`
	UpdatedAt     *time.Time `json:"updatedAt"     orm:"updated_at"     description:"Update time"`
}
