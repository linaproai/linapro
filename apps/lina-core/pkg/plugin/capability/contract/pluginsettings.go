// This file defines the source-plugin visible namespaced settings contract.
//
// Source plugins (auth providers, integrations, etc.) often need to persist
// small amounts of per-plugin configuration: client ID/secret pairs, feature
// flags, JSON rule maps. Rather than have every plugin own a private SQL
// table with its own schema, migration, controller, and seed file, the host
// publishes one unified key-value seam backed by sys_config. Each plugin is
// scoped to its own namespace by pluginID so plugins cannot read or write
// each other's settings, the operator gets a single auditable table, and
// schema upgrades stop requiring per-plugin ALTER TABLE statements.

package contract

import "context"

// PluginSettingsService is the host-published namespaced settings store
// exposed to source plugins. Implementations MUST scope every read/write to
// the supplied pluginID so a misbehaving plugin cannot touch another
// plugin's namespace. Empty pluginID or empty key inputs MUST return an
// error.
type PluginSettingsService interface {
	// GetString returns the persisted setting value or defaultValue when the
	// key is missing. The setting is returned verbatim with no decoding so
	// the caller is responsible for parsing structured payloads (JSON,
	// comma-separated lists, etc.).
	GetString(ctx context.Context, pluginID string, key string, defaultValue string) (string, error)
	// GetBool returns the parsed boolean value or defaultValue when the key
	// is missing. Values that fail to parse as bool also return
	// defaultValue. Accepted truthy values follow strconv.ParseBool ("true",
	// "1", "t", "T", "TRUE", "True"); other values resolve to false.
	GetBool(ctx context.Context, pluginID string, key string, defaultValue bool) (bool, error)
	// GetInt returns the parsed integer value or defaultValue when the key
	// is missing or fails to parse.
	GetInt(ctx context.Context, pluginID string, key string, defaultValue int) (int, error)
	// SetString writes the setting value. An empty value clears the row so
	// subsequent reads return the configured default. Use SetSecret for
	// fields that the operator might submit empty to mean "keep current".
	SetString(ctx context.Context, pluginID string, key string, value string) error
	// SetSecret writes a secret-style setting. An empty value is treated as
	// "preserve the existing stored secret" rather than "clear the secret"
	// so admin UIs can mask secrets in GET responses and submit empty on
	// PUT requests without accidentally wiping the secret.
	SetSecret(ctx context.Context, pluginID string, key string, value string) error
	// GetMaskedSecret reads a secret value and returns a masked projection
	// suitable for displaying in admin UIs. Empty rows return an empty
	// string. Short secrets resolve to a fixed "***" placeholder; longer
	// secrets preserve the first three and last three characters around
	// a "***" infix so operators can verify they have not rotated to an
	// unexpected value.
	GetMaskedSecret(ctx context.Context, pluginID string, key string) (string, error)
	// Delete removes one setting row. Missing rows return nil so callers
	// can use Delete to enforce "absent" without first checking existence.
	Delete(ctx context.Context, pluginID string, key string) error
	// List returns every persisted setting for one plugin namespace as a
	// flat map keyed by the per-plugin setting key (without the namespace
	// prefix). Empty namespaces return an empty map without error.
	List(ctx context.Context, pluginID string) (map[string]string, error)
}
