// This file implements the cron registration host service used during
// dynamic-plugin scheduled-job discovery.

package wasm

import (
	"context"

	"lina-core/pkg/pluginbridge"
)

// CronRegistrationCollector captures dynamic-plugin cron declarations during
// one host-driven discovery execution.
type CronRegistrationCollector interface {
	// Register validates and stores one discovered cron contract.
	Register(contract *pluginbridge.CronContract) error
}

// dispatchCronHostService routes cron host service methods to the discovery
// collector bound to the current Wasm execution.
func dispatchCronHostService(
	_ context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	switch method {
	case pluginbridge.HostServiceMethodCronRegister:
		return handleHostCronRegister(hcc, payload)
	default:
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusNotFound,
			"unsupported cron host service method: "+method,
		)
	}
}

// handleHostCronRegister validates one cron registration request and forwards
// it to the current discovery collector.
func handleHostCronRegister(
	hcc *hostCallContext,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	if hcc == nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, "host call context not available")
	}
	if pluginbridge.NormalizeExecutionSource(hcc.executionSource) != pluginbridge.ExecutionSourceCronDiscovery {
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusCapabilityDenied,
			"cron host service only supports cron discovery executions",
		)
	}
	if hcc.cronCollector == nil {
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusInternalError,
			"cron discovery collector not configured",
		)
	}

	request, err := pluginbridge.UnmarshalHostServiceCronRegisterRequest(payload)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}
	if request == nil || request.Contract == nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, "cron registration contract is required")
	}
	contractSnapshot := *request.Contract
	if err = pluginbridge.ValidateCronContracts(hcc.pluginID, []*pluginbridge.CronContract{&contractSnapshot}); err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}
	if err = hcc.cronCollector.Register(&contractSnapshot); err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}
	return pluginbridge.NewHostCallSuccessResponse(nil)
}
