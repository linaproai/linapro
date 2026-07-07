// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysUserExternalIdentity is the golang structure of table sys_user_external_identity for DAO operations like Where/Data.
type SysUserExternalIdentity struct {
	g.Meta        `orm:"table:sys_user_external_identity, do:true"`
	Id            any        // External identity linkage ID
	UserId        any        // Linked local sys_user ID
	Provider      any        // Stable external provider ID owned by the declaring plugin, e.g. google, discord
	Subject       any        // Immutable provider-issued subject identifier, e.g. OIDC sub
	PluginId      any        // Source-plugin ID that owns the provider and created the linkage
	EmailSnapshot any        // Email captured at link time for audit only, never used as a resolution key
	CreatedAt     *time.Time // Creation time
	UpdatedAt     *time.Time // Update time
}
