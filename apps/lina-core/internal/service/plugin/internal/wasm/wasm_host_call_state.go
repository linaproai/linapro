// This file implements the host:state capability handlers that provide
// plugin-scoped key-value state storage backed by the sys_plugin_state table.

package wasm

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/capabilityowner"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
)

const (
	runtimeStateMaxBatchKeys       = 100
	runtimeStateMaxKeyBytes        = 255
	runtimeStateMaxBatchValueBytes = 1 * 1024 * 1024
)

// defaultRuntimeStateStore is the plugin-owned persistence adapter used by the
// dynamic WASM host-call dispatcher.
var defaultRuntimeStateStore = capabilityowner.NewRuntimeStateStore()

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

	value, found, err := defaultRuntimeStateStore.Get(ctx, hcc.pluginID, key)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}

	resp := &bridgehostcall.HostCallStateGetResponse{}
	if found {
		resp.Value = value
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

	values, err := defaultRuntimeStateStore.GetMany(ctx, hcc.pluginID, keys)
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

	err = defaultRuntimeStateStore.Set(ctx, hcc.pluginID, key, req.Value)
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
	if err = defaultRuntimeStateStore.SetMany(ctx, hcc.pluginID, items); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	return bridgehostcall.NewHostCallEmptySuccessResponse()
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

	err = defaultRuntimeStateStore.Delete(ctx, hcc.pluginID, key)
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
	err = defaultRuntimeStateStore.DeleteMany(ctx, hcc.pluginID, keys)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	return bridgehostcall.NewHostCallEmptySuccessResponse()
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
