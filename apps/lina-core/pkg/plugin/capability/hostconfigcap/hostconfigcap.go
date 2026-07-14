// Package hostconfigcap defines host configuration capabilities published to
// plugins. Service covers static host configuration reads and governed
// sys_config values through explicit method-level authorization and cache
// governance.
package hostconfigcap

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/container/gvar"

	"lina-core/pkg/plugin/capability/capmodel"
)

// Service defines governed host config values that source plugins may read.
// Static reads are read-only; mutable sys_config rows are exposed through the
// SysConfig subresource so callers do not confuse persisted system settings
// with static GoFrame host configuration.
type Service interface {
	// Get returns the raw host config value for the requested key or root
	// snapshot. When the key is absent or nil, Get returns defaultValue wrapped
	// as *gvar.Var. Passing nil preserves absent-key nil return semantics. Empty
	// strings remain present values; typed helpers own blank-value fallback
	// semantics.
	Get(ctx context.Context, key string, defaultValue any) (*gvar.Var, error)
	// Exists reports whether a host config key is available.
	Exists(ctx context.Context, key string) (bool, error)
	// String reads a host config string value or returns defaultValue when
	// the key is absent or blank.
	String(ctx context.Context, key string, defaultValue string) (string, error)
	// Bool reads a host config bool value or returns defaultValue when the key is absent.
	Bool(ctx context.Context, key string, defaultValue bool) (bool, error)
	// Int reads a host config int value or returns defaultValue when the key is absent.
	Int(ctx context.Context, key string, defaultValue int) (int, error)
	// Duration reads a host config duration value or returns defaultValue when
	// the key is absent or blank.
	Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error)
	// SysConfig returns governed sys_config row operations.
	SysConfig() SysConfigService
}

// SysConfigService defines governed sys_config methods. Reads and writes must
// validate key visibility, tenant fallback semantics, audit source, transaction
// boundaries, and cross-node sys_config cache revision impact.
type SysConfigService interface {
	// Get returns one visible sys_config value or a denied error when the
	// key is absent or outside caller scope.
	Get(ctx context.Context, key SysConfigKey) (*SysConfigInfo, error)
	// BatchGet returns visible sys_config values and opaque missing keys
	// with bounded batch semantics and key-level authorization.
	BatchGet(ctx context.Context, keys []SysConfigKey) (*capmodel.BatchResult[*SysConfigInfo, SysConfigKey], error)
	// List returns one bounded page of visible sys_config values. Risk:
	// read; resource: config keys; context:
	// actor and tenant; data permission: key-level visibility; performance:
	// bounded page; audit/cache: read-only.
	List(ctx context.Context, input ListSysConfigInput) (*capmodel.PageResult[*SysConfigInfo], error)
	// SetValue writes one governed sys_config value after key authorization,
	// tenant fallback, audit, transaction, and sys_config cache revision
	// checks. The optional options argument controls write metadata such as
	// SystemManageable; nil options keep host defaults (plugin closed-loop on
	// first insert). Prefer BatchSetValue when writing multiple keys so one
	// transaction and one runtime-config revision cover the whole batch.
	SetValue(ctx context.Context, key SysConfigKey, value string, options *SetSysConfigValueOptions) error
	// BatchSetValue writes multiple governed sys_config values in one
	// transaction and one runtime-config revision bump. Shared options apply
	// to every item. An empty items slice is a successful no-op.
	BatchSetValue(ctx context.Context, items []SetSysConfigValueItem, options *SetSysConfigValueOptions) error
	// Reset resets one governed sys_config value to its owner default after key
	// authorization and transaction-after cache revision.
	Reset(ctx context.Context, key SysConfigKey) error
	// EnsureVisible rejects when any sys_config key is outside caller scope.
	EnsureVisible(ctx context.Context, keys []SysConfigKey) error
}

// SysConfigKey identifies one governed sys_config key.
type SysConfigKey string

// ListSysConfigInput constrains sys_config listing.
type ListSysConfigInput struct {
	// Keyword filters by key or label.
	Keyword string
	// Page constrains result size and stable sorting.
	Page capmodel.PageRequest
}

// SetSysConfigValueOptions controls optional SetValue / BatchSetValue write
// metadata shared by every written key.
type SetSysConfigValueOptions struct {
	// SystemManageable controls whether the row may be listed and mutated on
	// the system config management page.
	//
	//   - nil on first insert: defaults to false (plugin closed-loop)
	//   - nil on update: leaves the existing flag unchanged
	//   - non-nil: writes the flag on both insert and update
	//
	// Plugin settings that are maintained only at the plugin entry point MUST
	// pass false (use gconv.PtrBool(false)).
	SystemManageable *bool
}

// SetSysConfigValueItem is one key/value pair for BatchSetValue.
type SetSysConfigValueItem struct {
	// Key is the sys_config key to write.
	Key SysConfigKey
	// Value is the raw sys_config.value string.
	Value string
}

// SysConfigInfo describes one sys_config value visible to a plugin
// management caller.
type SysConfigInfo struct {
	// Key is the sys_config key.
	Key SysConfigKey
	// Value is the raw sys_config.value string.
	Value string
	// SystemManageable reports whether the row may be governed on the system
	// config management page.
	SystemManageable bool
	// LabelKey is the optional i18n label key.
	LabelKey string
	// Label is the optional locale-resolved label.
	Label string
}
