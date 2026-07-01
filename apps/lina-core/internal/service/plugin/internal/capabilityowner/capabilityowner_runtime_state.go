// This file owns dynamic-plugin runtime state persistence for host-call
// adapters without exposing plugin DAO details to the WASM dispatcher.

package capabilityowner

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/datascope"
)

// RuntimeStateStore implements plugin-scoped state persistence with
// sys_plugin_state as the authoritative source.
type RuntimeStateStore struct{}

// NewRuntimeStateStore creates the plugin-owned runtime state persistence store.
func NewRuntimeStateStore() RuntimeStateStore {
	return RuntimeStateStore{}
}

// Get returns one plugin-scoped state value and whether it exists.
func (RuntimeStateStore) Get(ctx context.Context, pluginID string, key string) (string, bool, error) {
	cols := dao.SysPluginState.Columns()
	value, err := dao.SysPluginState.Ctx(ctx).
		Where(pluginStateIdentity(ctx, pluginID, key)).
		Value(cols.StateValue)
	if err != nil {
		return "", false, err
	}
	if value.IsNil() || value.IsEmpty() {
		return "", false, nil
	}
	return value.String(), true, nil
}

// GetMany returns plugin-scoped state values for the requested keys.
func (RuntimeStateStore) GetMany(ctx context.Context, pluginID string, keys []string) (map[string]string, error) {
	values := make(map[string]string, len(keys))
	if len(keys) == 0 {
		return values, nil
	}
	identity := pluginStateIdentity(ctx, pluginID, "")
	cols := dao.SysPluginState.Columns()
	rows, err := dao.SysPluginState.Ctx(ctx).
		Fields(cols.StateKey, cols.StateValue).
		Where(do.SysPluginState{PluginId: identity.PluginId, TenantId: identity.TenantId}).
		WhereIn(cols.StateKey, keys).
		All()
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		key := strings.TrimSpace(row[cols.StateKey].String())
		if key == "" {
			continue
		}
		values[key] = row[cols.StateValue].String()
	}
	return values, nil
}

// Set writes one plugin-scoped state value.
func (s RuntimeStateStore) Set(ctx context.Context, pluginID string, key string, value string) error {
	return s.SetMany(ctx, pluginID, map[string]string{key: value})
}

// SetMany writes plugin-scoped state values inside one transaction.
func (RuntimeStateStore) SetMany(ctx context.Context, pluginID string, values map[string]string) error {
	return dao.SysPluginState.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		for key, value := range values {
			identity := pluginStateIdentity(ctx, pluginID, key)
			_, err := dao.SysPluginState.Ctx(ctx).Data(do.SysPluginState{
				PluginId:   identity.PluginId,
				TenantId:   identity.TenantId,
				StateKey:   identity.StateKey,
				StateValue: value,
			}).InsertIgnore()
			if err != nil {
				return err
			}

			_, err = dao.SysPluginState.Ctx(ctx).
				Where(identity).
				Data(do.SysPluginState{
					StateValue: value,
				}).
				Update()
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete removes one plugin-scoped state value.
func (RuntimeStateStore) Delete(ctx context.Context, pluginID string, key string) error {
	_, err := dao.SysPluginState.Ctx(ctx).
		Where(pluginStateIdentity(ctx, pluginID, key)).
		Delete()
	return err
}

// DeleteMany removes plugin-scoped state values in one bounded query.
func (RuntimeStateStore) DeleteMany(ctx context.Context, pluginID string, keys []string) error {
	identity := pluginStateIdentity(ctx, pluginID, "")
	_, err := dao.SysPluginState.Ctx(ctx).
		Where(do.SysPluginState{PluginId: identity.PluginId, TenantId: identity.TenantId}).
		WhereIn(dao.SysPluginState.Columns().StateKey, keys).
		Delete()
	return err
}

// pluginStateIdentity builds the tenant-scoped plugin state identity used by
// dynamic host state operations.
func pluginStateIdentity(ctx context.Context, pluginID string, key string) do.SysPluginState {
	return do.SysPluginState{
		PluginId: strings.TrimSpace(pluginID),
		TenantId: datascope.CurrentTenantID(ctx),
		StateKey: strings.TrimSpace(key),
	}
}
