// This file implements the host:state capability handlers that provide
// plugin-scoped key-value state storage backed by the sys_plugin_state table.

package wasm

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
)

const (
	runtimeStateMaxBatchKeys       = 100
	runtimeStateMaxKeyBytes        = 255
	runtimeStateMaxBatchValueBytes = 1 * 1024 * 1024
)

// handleHostStateGet processes OpcodeStateGet requests.
// handleHostStateGet loads one plugin-scoped runtime state value.
func handleHostStateGet(ctx context.Context, hcc *hostCallContext, reqBytes []byte) *bridgehostcall.HostCallResponseEnvelope {
	req, err := bridgehostcall.UnmarshalHostCallStateGetRequest(reqBytes)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	key, err := normalizeRuntimeStateKey(req.Key)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}

	cols := dao.SysPluginState.Columns()
	value, err := dao.SysPluginState.Ctx(ctx).
		Where(pluginStateIdentity(ctx, hcc.pluginID, key)).
		Value(cols.StateValue)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}

	resp := &bridgehostcall.HostCallStateGetResponse{}
	if !value.IsNil() && !value.IsEmpty() {
		resp.Value = value.String()
		resp.Found = true
	}
	return bridgehostcall.NewHostCallSuccessResponse(bridgehostcall.MarshalHostCallStateGetResponse(resp))
}

// handleHostStateGetMany loads plugin-scoped runtime state values in one bounded query.
func handleHostStateGetMany(ctx context.Context, hcc *hostCallContext, reqBytes []byte) *bridgehostcall.HostCallResponseEnvelope {
	var request runtimeStateGetManyRequest
	if err := decodeCapabilityJSONRequest(reqBytes, &request); err != nil {
		return invalidCapabilityRequest(err)
	}
	keys, err := normalizeRuntimeStateKeys(request.Keys)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}

	values, err := getHostStateValues(ctx, hcc.pluginID, keys)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	response := runtimeStateGetManyResponse{
		Values:      values,
		MissingKeys: make([]string, 0),
	}
	for _, key := range keys {
		if _, ok := values[key]; !ok {
			response.MissingKeys = append(response.MissingKeys, key)
		}
	}
	return capabilityJSONResponse(response)
}

// handleHostStateSet processes OpcodeStateSet requests.
// handleHostStateSet upserts one plugin-scoped runtime state value.
func handleHostStateSet(ctx context.Context, hcc *hostCallContext, reqBytes []byte) *bridgehostcall.HostCallResponseEnvelope {
	req, err := bridgehostcall.UnmarshalHostCallStateSetRequest(reqBytes)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	key, err := normalizeRuntimeStateKey(req.Key)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}

	err = upsertHostStateValue(ctx, hcc.pluginID, key, req.Value)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	return bridgehostcall.NewHostCallEmptySuccessResponse()
}

// handleHostStateSetMany upserts plugin-scoped runtime state values in one transaction.
func handleHostStateSetMany(ctx context.Context, hcc *hostCallContext, reqBytes []byte) *bridgehostcall.HostCallResponseEnvelope {
	var request runtimeStateSetManyRequest
	if err := decodeCapabilityJSONRequest(reqBytes, &request); err != nil {
		return invalidCapabilityRequest(err)
	}
	items, err := normalizeRuntimeStateItems(request.Values)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if err = upsertHostStateValues(ctx, hcc.pluginID, items); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	return bridgehostcall.NewHostCallEmptySuccessResponse()
}

// upsertHostStateValue writes one plugin state value using a dialect-neutral
// insert-ignore plus update sequence inside a transaction.
func upsertHostStateValue(ctx context.Context, pluginID string, key string, value string) error {
	return upsertHostStateValues(ctx, pluginID, map[string]string{key: value})
}

// upsertHostStateValues writes plugin state values using a dialect-neutral
// insert-ignore plus update sequence inside one transaction.
func upsertHostStateValues(ctx context.Context, pluginID string, values map[string]string) error {
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

// handleHostStateDelete processes OpcodeStateDelete requests.
// handleHostStateDelete removes one plugin-scoped runtime state value.
func handleHostStateDelete(ctx context.Context, hcc *hostCallContext, reqBytes []byte) *bridgehostcall.HostCallResponseEnvelope {
	req, err := bridgehostcall.UnmarshalHostCallStateDeleteRequest(reqBytes)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	key, err := normalizeRuntimeStateKey(req.Key)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}

	_, err = dao.SysPluginState.Ctx(ctx).
		Where(pluginStateIdentity(ctx, hcc.pluginID, key)).
		Delete()
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	return bridgehostcall.NewHostCallEmptySuccessResponse()
}

// handleHostStateDeleteMany deletes plugin-scoped runtime state values in one query.
func handleHostStateDeleteMany(ctx context.Context, hcc *hostCallContext, reqBytes []byte) *bridgehostcall.HostCallResponseEnvelope {
	var request runtimeStateDeleteManyRequest
	if err := decodeCapabilityJSONRequest(reqBytes, &request); err != nil {
		return invalidCapabilityRequest(err)
	}
	keys, err := normalizeRuntimeStateKeys(request.Keys)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	identity := pluginStateIdentity(ctx, hcc.pluginID, "")
	_, err = dao.SysPluginState.Ctx(ctx).
		Where(do.SysPluginState{PluginId: identity.PluginId, TenantId: identity.TenantId}).
		WhereIn(dao.SysPluginState.Columns().StateKey, keys).
		Delete()
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	return bridgehostcall.NewHostCallEmptySuccessResponse()
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

func getHostStateValues(ctx context.Context, pluginID string, keys []string) (map[string]string, error) {
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

func normalizeRuntimeStateKey(raw string) (string, error) {
	key := strings.TrimSpace(raw)
	if key == "" {
		return "", bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if len([]byte(key)) > runtimeStateMaxKeyBytes {
		return "", bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", runtimeStateMaxKeyBytes))
	}
	return key, nil
}

func normalizeRuntimeStateKeys(rawKeys []string) ([]string, error) {
	keys := make([]string, 0, len(rawKeys))
	seen := make(map[string]struct{}, len(rawKeys))
	for _, rawKey := range rawKeys {
		key, err := normalizeRuntimeStateKey(rawKey)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
		if len(keys) > runtimeStateMaxBatchKeys {
			return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", runtimeStateMaxBatchKeys))
		}
	}
	return keys, nil
}

func normalizeRuntimeStateItems(values map[string]string) (map[string]string, error) {
	if len(values) > runtimeStateMaxBatchKeys {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", runtimeStateMaxBatchKeys))
	}
	items := make(map[string]string, len(values))
	totalBytes := 0
	for rawKey, value := range values {
		key, err := normalizeRuntimeStateKey(rawKey)
		if err != nil {
			return nil, err
		}
		totalBytes += len([]byte(value))
		if totalBytes > runtimeStateMaxBatchValueBytes {
			return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", runtimeStateMaxBatchValueBytes))
		}
		items[key] = value
	}
	return items, nil
}

type runtimeStateGetManyRequest struct {
	Keys []string `json:"keys"`
}

type runtimeStateGetManyResponse struct {
	Values      map[string]string `json:"values"`
	MissingKeys []string          `json:"missingKeys,omitempty"`
}

type runtimeStateSetManyRequest struct {
	Values map[string]string `json:"values"`
}

type runtimeStateDeleteManyRequest struct {
	Keys []string `json:"keys"`
}
