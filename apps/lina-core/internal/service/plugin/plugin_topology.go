// This file defines the topology contract used by root plugin runtime
// orchestration and the single-node fallback used when clustering is disabled.

package plugin

// Topology defines the cluster semantics required by plugin runtime behavior.
type Topology interface {
	// IsEnabled reports whether the host is running in clustered mode.
	IsEnabled() bool
	// IsPrimary reports whether the current node is the primary node.
	IsPrimary() bool
	// NodeID returns the stable identifier of the current node.
	NodeID() string
}

// singleNodeTopology provides the default topology used when clustering is disabled.
type singleNodeTopology struct{}

// IsEnabled reports false because the default topology is always single-node.
func (singleNodeTopology) IsEnabled() bool {
	return false
}

// IsPrimary reports true because the only node is also the primary node.
func (singleNodeTopology) IsPrimary() bool {
	return true
}

// NodeID returns the stable placeholder node identifier for single-node mode.
func (singleNodeTopology) NodeID() string {
	return "local-node"
}
