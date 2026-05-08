// This file implements the host:state capability handlers that provide
// plugin-scoped key-value state storage backed by the sys_plugin_state table.

package wasm

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	bridgehostcall "lina-core/pkg/pluginbridge/hostcall"
)

// handleHostStateGet processes OpcodeStateGet requests.
// handleHostStateGet loads one plugin-scoped runtime state value.
func handleHostStateGet(ctx context.Context, hcc *hostCallContext, reqBytes []byte) *bridgehostcall.HostCallResponseEnvelope {
	req, err := bridgehostcall.UnmarshalHostCallStateGetRequest(reqBytes)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	key := strings.TrimSpace(req.Key)
	if key == "" {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, "state key must not be empty")
	}

	cols := dao.SysPluginState.Columns()
	value, err := dao.SysPluginState.Ctx(ctx).
		Where(do.SysPluginState{PluginId: hcc.pluginID, StateKey: key}).
		Value(cols.StateValue)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}

	resp := &bridgehostcall.HostCallStateGetResponse{}
	if !value.IsNil() && !value.IsEmpty() {
		resp.Value = value.String()
		resp.Found = true
	}
	return bridgehostcall.NewHostCallSuccessResponse(bridgehostcall.MarshalHostCallStateGetResponse(resp))
}

// handleHostStateSet processes OpcodeStateSet requests.
// handleHostStateSet upserts one plugin-scoped runtime state value.
func handleHostStateSet(ctx context.Context, hcc *hostCallContext, reqBytes []byte) *bridgehostcall.HostCallResponseEnvelope {
	req, err := bridgehostcall.UnmarshalHostCallStateSetRequest(reqBytes)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	key := strings.TrimSpace(req.Key)
	if key == "" {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, "state key must not be empty")
	}

	err = upsertHostStateValue(ctx, hcc.pluginID, key, req.Value)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	return bridgehostcall.NewHostCallEmptySuccessResponse()
}

// upsertHostStateValue writes one plugin state value using a dialect-neutral
// insert-ignore plus update sequence inside a transaction.
func upsertHostStateValue(ctx context.Context, pluginID string, key string, value string) error {
	return dao.SysPluginState.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		_, err := dao.SysPluginState.Ctx(ctx).Data(do.SysPluginState{
			PluginId:   pluginID,
			StateKey:   key,
			StateValue: value,
		}).InsertIgnore()
		if err != nil {
			return err
		}

		_, err = dao.SysPluginState.Ctx(ctx).
			Where(do.SysPluginState{PluginId: pluginID, StateKey: key}).
			Data(do.SysPluginState{
				StateValue: value,
			}).
			Update()
		return err
	})
}

// handleHostStateDelete processes OpcodeStateDelete requests.
// handleHostStateDelete removes one plugin-scoped runtime state value.
func handleHostStateDelete(ctx context.Context, hcc *hostCallContext, reqBytes []byte) *bridgehostcall.HostCallResponseEnvelope {
	req, err := bridgehostcall.UnmarshalHostCallStateDeleteRequest(reqBytes)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	key := strings.TrimSpace(req.Key)
	if key == "" {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, "state key must not be empty")
	}

	_, err = dao.SysPluginState.Ctx(ctx).
		Where(do.SysPluginState{PluginId: hcc.pluginID, StateKey: key}).
		Delete()
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	return bridgehostcall.NewHostCallEmptySuccessResponse()
}
