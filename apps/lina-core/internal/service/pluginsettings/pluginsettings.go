// Package pluginsettings implements the host-side namespaced key-value
// settings store published to source plugins through the
// capability.PluginSettingsService contract.
//
// Implementation notes:
//   - Storage backend is the shared sys_config table at platform scope
//     (tenant_id=0). Plugin settings are intentionally platform-scoped
//     because plugins live at the host level; tenant-scoped overrides can
//     be added later by widening the API surface without changing the
//     storage layout.
//   - Keys are persisted as "<pluginID>.<settingKey>" so every row is
//     traceable to one plugin and the existing unique index
//     (tenant_id, key) prevents accidental cross-plugin collisions.
//   - The service deliberately bypasses sysconfig.Service because that
//     service applies tenant fallback semantics, i18n, audit decoration,
//     and built-in protection that are inappropriate for opaque plugin
//     state. It uses the DAO directly to keep the seam narrow.
package pluginsettings

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// platformTenantID is the platform-scope tenant id used for all plugin
// settings rows. Tenant overrides are deliberately out of scope for the
// initial implementation.
const platformTenantID = 0

// keySeparator joins the pluginID prefix with the per-plugin setting key.
const keySeparator = "."

// Service is the narrow contract exposed to host adapters and plugin
// capability bindings.
type Service interface {
	// GetString returns the persisted value or defaultValue when the row
	// is absent.
	GetString(ctx context.Context, pluginID string, key string, defaultValue string) (string, error)
	// GetBool returns the parsed bool value or defaultValue when the row
	// is absent or fails to parse.
	GetBool(ctx context.Context, pluginID string, key string, defaultValue bool) (bool, error)
	// GetInt returns the parsed int value or defaultValue when the row is
	// absent or fails to parse.
	GetInt(ctx context.Context, pluginID string, key string, defaultValue int) (int, error)
	// SetString writes the setting value. Empty values delete the row.
	SetString(ctx context.Context, pluginID string, key string, value string) error
	// SetSecret writes a secret-style setting. Empty values preserve the
	// previously stored secret instead of clearing it.
	SetSecret(ctx context.Context, pluginID string, key string, value string) error
	// GetMaskedSecret returns a masked projection of the secret value
	// suitable for admin UI display.
	GetMaskedSecret(ctx context.Context, pluginID string, key string) (string, error)
	// Delete removes the setting row; missing rows are not an error.
	Delete(ctx context.Context, pluginID string, key string) error
	// List returns every persisted setting for one plugin namespace as a
	// flat map keyed by the per-plugin setting name (no prefix).
	List(ctx context.Context, pluginID string) (map[string]string, error)
}

// serviceImpl is the host-side implementation backed by sys_config.
type serviceImpl struct{}

// New constructs the default Service implementation.
func New() Service {
	return &serviceImpl{}
}

// GetString returns the persisted value for one plugin setting or the
// supplied default when the row is absent.
func (s *serviceImpl) GetString(ctx context.Context, pluginID string, key string, defaultValue string) (string, error) {
	fullKey, err := buildKey(pluginID, key)
	if err != nil {
		return "", err
	}
	value, found, err := s.fetchValue(ctx, fullKey)
	if err != nil {
		return "", err
	}
	if !found {
		return defaultValue, nil
	}
	return value, nil
}

// GetBool reads a plugin setting and parses it as a bool.
func (s *serviceImpl) GetBool(ctx context.Context, pluginID string, key string, defaultValue bool) (bool, error) {
	raw, err := s.GetString(ctx, pluginID, key, "")
	if err != nil {
		return defaultValue, err
	}
	if raw == "" {
		return defaultValue, nil
	}
	parsed, parseErr := strconv.ParseBool(raw)
	if parseErr != nil {
		return defaultValue, nil
	}
	return parsed, nil
}

// GetInt reads a plugin setting and parses it as an int.
func (s *serviceImpl) GetInt(ctx context.Context, pluginID string, key string, defaultValue int) (int, error) {
	raw, err := s.GetString(ctx, pluginID, key, "")
	if err != nil {
		return defaultValue, err
	}
	if raw == "" {
		return defaultValue, nil
	}
	parsed, parseErr := strconv.Atoi(raw)
	if parseErr != nil {
		return defaultValue, nil
	}
	return parsed, nil
}

// SetString upserts the persisted value. Empty values delete the row.
func (s *serviceImpl) SetString(ctx context.Context, pluginID string, key string, value string) error {
	fullKey, err := buildKey(pluginID, key)
	if err != nil {
		return err
	}
	if value == "" {
		return s.deleteByFullKey(ctx, fullKey)
	}
	return s.upsertValue(ctx, fullKey, value)
}

// SetSecret upserts a secret-style setting. Empty values preserve the
// previously stored secret instead of clearing it.
func (s *serviceImpl) SetSecret(ctx context.Context, pluginID string, key string, value string) error {
	if value == "" {
		return nil
	}
	fullKey, err := buildKey(pluginID, key)
	if err != nil {
		return err
	}
	return s.upsertValue(ctx, fullKey, value)
}

// GetMaskedSecret returns a non-reversible masked projection of the stored
// secret. Missing rows return an empty string.
func (s *serviceImpl) GetMaskedSecret(ctx context.Context, pluginID string, key string) (string, error) {
	raw, err := s.GetString(ctx, pluginID, key, "")
	if err != nil {
		return "", err
	}
	return MaskSecret(raw), nil
}

// Delete removes one plugin setting row.
func (s *serviceImpl) Delete(ctx context.Context, pluginID string, key string) error {
	fullKey, err := buildKey(pluginID, key)
	if err != nil {
		return err
	}
	return s.deleteByFullKey(ctx, fullKey)
}

// List returns every persisted setting for one plugin namespace.
func (s *serviceImpl) List(ctx context.Context, pluginID string) (map[string]string, error) {
	pluginID = strings.TrimSpace(pluginID)
	if pluginID == "" {
		return nil, gerror.New("plugin settings list requires non-empty plugin id")
	}
	prefix := pluginID + keySeparator
	var rows []*entity.SysConfig
	if err := dao.SysConfig.Ctx(ctx).
		Where(dao.SysConfig.Columns().TenantId, platformTenantID).
		WhereLike(dao.SysConfig.Columns().Key, prefix+"%").
		Scan(&rows); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		if !strings.HasPrefix(row.Key, prefix) {
			continue
		}
		out[row.Key[len(prefix):]] = row.Value
	}
	return out, nil
}

// fetchValue reads one sys_config row by full key and reports whether the
// row exists.
func (s *serviceImpl) fetchValue(ctx context.Context, fullKey string) (string, bool, error) {
	var row *entity.SysConfig
	if err := dao.SysConfig.Ctx(ctx).
		Where(dao.SysConfig.Columns().TenantId, platformTenantID).
		Where(dao.SysConfig.Columns().Key, fullKey).
		Scan(&row); err != nil {
		return "", false, err
	}
	if row == nil {
		return "", false, nil
	}
	return row.Value, true, nil
}

// upsertValue inserts or updates one sys_config row.
func (s *serviceImpl) upsertValue(ctx context.Context, fullKey string, value string) error {
	now := time.Now()
	count, err := dao.SysConfig.Ctx(ctx).
		Where(dao.SysConfig.Columns().TenantId, platformTenantID).
		Where(dao.SysConfig.Columns().Key, fullKey).
		Count()
	if err != nil {
		return err
	}
	if count > 0 {
		_, err = dao.SysConfig.Ctx(ctx).
			Where(dao.SysConfig.Columns().TenantId, platformTenantID).
			Where(dao.SysConfig.Columns().Key, fullKey).
			Data(do.SysConfig{
				Value:     value,
				UpdatedAt: &now,
			}).
			Update()
		return err
	}
	_, err = dao.SysConfig.Ctx(ctx).
		Data(do.SysConfig{
			TenantId:  platformTenantID,
			Name:      fullKey,
			Key:       fullKey,
			Value:     value,
			IsBuiltin: 0,
			CreatedAt: &now,
			UpdatedAt: &now,
		}).
		Insert()
	return err
}

// deleteByFullKey removes one sys_config row by full key. Missing rows
// return nil so callers can use this to enforce absence.
func (s *serviceImpl) deleteByFullKey(ctx context.Context, fullKey string) error {
	_, err := dao.SysConfig.Ctx(ctx).
		Where(dao.SysConfig.Columns().TenantId, platformTenantID).
		Where(dao.SysConfig.Columns().Key, fullKey).
		Delete()
	return err
}

// buildKey assembles the namespaced sys_config key from one pluginID and
// per-plugin key name, rejecting empty inputs and accidental embedded
// separators that would let callers escape their namespace.
func buildKey(pluginID string, key string) (string, error) {
	pluginID = strings.TrimSpace(pluginID)
	key = strings.TrimSpace(key)
	if pluginID == "" {
		return "", gerror.New("plugin settings access requires non-empty plugin id")
	}
	if key == "" {
		return "", gerror.New("plugin settings access requires non-empty key")
	}
	if strings.Contains(pluginID, keySeparator) {
		return "", gerror.Newf("plugin id %q must not contain %q", pluginID, keySeparator)
	}
	return pluginID + keySeparator + key, nil
}

// MaskSecret produces the canonical masked projection of one secret value.
// Short secrets fold to a fixed placeholder so the projection itself does
// not leak length; longer secrets preserve the first three and last three
// characters around a "***" infix so operators can tell whether the secret
// changed unexpectedly.
func MaskSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 6 {
		return "***"
	}
	return value[:3] + "***" + value[len(value)-3:]
}
