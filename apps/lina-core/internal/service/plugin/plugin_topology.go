// This file defines topology abstractions used by the root plugin facade.

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

type singleNodeTopology struct{}

func (singleNodeTopology) IsEnabled() bool {
	return false
}

func (singleNodeTopology) IsPrimary() bool {
	return true
}

func (singleNodeTopology) NodeID() string {
	return "local-node"
}
