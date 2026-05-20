package v1

// Status identifies the anonymous health-probe result.
type Status string

const (
	// StatusOK means the host can serve traffic.
	StatusOK Status = "ok"
	// StatusUnavailable means the health probe failed.
	StatusUnavailable Status = "unavailable"
)

// Mode identifies the current host deployment role.
type Mode string

const (
	// ModeSingle identifies a standalone host.
	ModeSingle Mode = "single"
	// ModeMaster identifies the primary node in a cluster.
	ModeMaster Mode = "master"
	// ModeSlave identifies a secondary node in a cluster.
	ModeSlave Mode = "slave"
)
