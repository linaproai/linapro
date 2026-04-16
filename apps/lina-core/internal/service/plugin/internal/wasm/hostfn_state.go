// This file implements the host:state capability handlers that provide
// plugin-scoped key-value state storage backed by the sys_plugin_state table.

package wasm

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/pluginbridge"
)

const pluginStateTable = "sys_plugin_state"

// handleHostStateGet processes OpcodeStateGet requests.
func handleHostStateGet(ctx context.Context, hcc *hostCallContext, reqBytes []byte) *pluginbridge.HostCallResponseEnvelope {
	req, err := pluginbridge.UnmarshalHostCallStateGetRequest(reqBytes)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}
	key := strings.TrimSpace(req.Key)
	if key == "" {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, "state key must not be empty")
	}

	value, err := g.DB().Model(pluginStateTable).Ctx(ctx).
		Where("plugin_id", hcc.pluginID).
		Where("state_key", key).
		Value("state_value")
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}

	resp := &pluginbridge.HostCallStateGetResponse{}
	if !value.IsNil() && !value.IsEmpty() {
		resp.Value = value.String()
		resp.Found = true
	}
	return pluginbridge.NewHostCallSuccessResponse(pluginbridge.MarshalHostCallStateGetResponse(resp))
}

// handleHostStateSet processes OpcodeStateSet requests.
func handleHostStateSet(ctx context.Context, hcc *hostCallContext, reqBytes []byte) *pluginbridge.HostCallResponseEnvelope {
	req, err := pluginbridge.UnmarshalHostCallStateSetRequest(reqBytes)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}
	key := strings.TrimSpace(req.Key)
	if key == "" {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, "state key must not be empty")
	}

	// Upsert: insert or update on duplicate key.
	_, err = g.DB().Ctx(ctx).Exec(ctx,
		"INSERT INTO "+pluginStateTable+" (plugin_id, state_key, state_value, created_at, updated_at) "+
			"VALUES (?, ?, ?, NOW(), NOW()) "+
			"ON DUPLICATE KEY UPDATE state_value = VALUES(state_value), updated_at = NOW()",
		hcc.pluginID, key, req.Value,
	)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	return pluginbridge.NewHostCallEmptySuccessResponse()
}

// handleHostStateDelete processes OpcodeStateDelete requests.
func handleHostStateDelete(ctx context.Context, hcc *hostCallContext, reqBytes []byte) *pluginbridge.HostCallResponseEnvelope {
	req, err := pluginbridge.UnmarshalHostCallStateDeleteRequest(reqBytes)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}
	key := strings.TrimSpace(req.Key)
	if key == "" {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, "state key must not be empty")
	}

	_, err = g.DB().Model(pluginStateTable).Ctx(ctx).
		Where("plugin_id", hcc.pluginID).
		Where("state_key", key).
		Delete()
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	return pluginbridge.NewHostCallEmptySuccessResponse()
}
