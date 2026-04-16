// This file defines the typed authorization review states exposed by plugin
// list projections.

package runtime

// AuthorizationStatus identifies the current host-service authorization state
// for one plugin release.
type AuthorizationStatus string

const (
	// AuthorizationStatusNotRequired means the plugin does not request governed host-service targets.
	AuthorizationStatusNotRequired AuthorizationStatus = "not_required"
	// AuthorizationStatusPending means the plugin still awaits host confirmation for governed targets.
	AuthorizationStatusPending AuthorizationStatus = "pending"
	// AuthorizationStatusConfirmed means the host has persisted the final authorization snapshot.
	AuthorizationStatusConfirmed AuthorizationStatus = "confirmed"
)
