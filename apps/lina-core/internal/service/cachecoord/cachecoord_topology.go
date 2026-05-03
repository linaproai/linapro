// This file provides topology adapters for cache coordination construction.

package cachecoord

// staticTopology is a minimal topology used by services that only know the
// cluster switch but cannot import the full cluster service due package cycles.
type staticTopology struct {
	enabled bool
	primary bool
	nodeID  string
}

// NewStaticTopology creates one static topology view for service-level cache
// coordination.
func NewStaticTopology(enabled bool) Topology {
	return staticTopology{
		enabled: enabled,
		primary: !enabled,
		nodeID:  "local-node",
	}
}

// IsEnabled reports the configured cluster switch.
func (t staticTopology) IsEnabled() bool {
	return t.enabled
}

// IsPrimary reports the configured primary flag.
func (t staticTopology) IsPrimary() bool {
	return t.primary
}

// NodeID returns the configured node identifier.
func (t staticTopology) NodeID() string {
	if t.nodeID == "" {
		return "local-node"
	}
	return t.nodeID
}
