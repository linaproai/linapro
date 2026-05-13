// This file defines media plugin enum constants.

package media

// SwitchValue is the numeric on/off enum used by media strategy tables.
type SwitchValue int

// Switch values used by media strategy records.
const (
	// SwitchOn means the option is enabled.
	SwitchOn SwitchValue = 1
	// SwitchOff means the option is disabled.
	SwitchOff SwitchValue = 2
)

// BinaryValue is the numeric yes/no enum used by stream alias records.
type BinaryValue int

// Binary values used by media stream alias records.
const (
	// BinaryNo means the option is false.
	BinaryNo BinaryValue = 0
	// BinaryYes means the option is true.
	BinaryYes BinaryValue = 1
)

// StrategySource is the string enum returned by effective strategy resolution.
type StrategySource string

// Supported media strategy resolution sources.
const (
	// StrategySourceTenantDevice reports a tenant-device binding match.
	StrategySourceTenantDevice StrategySource = "tenantDevice"
	// StrategySourceDevice reports a device binding match.
	StrategySourceDevice StrategySource = "device"
	// StrategySourceTenant reports a tenant binding match.
	StrategySourceTenant StrategySource = "tenant"
	// StrategySourceGlobal reports a global strategy match.
	StrategySourceGlobal StrategySource = "global"
	// StrategySourceNone reports that no strategy matched.
	StrategySourceNone StrategySource = "none"
)

// Pagination defaults used by media list APIs.
const (
	defaultPageNum  = 1
	defaultPageSize = 10
	maxPageSize     = 100
)
