// This file implements the cron registration host service used during
// dynamic-plugin scheduled-job discovery.

package wasm

import (
	"context"

	bridgecontract "lina-core/pkg/pluginbridge/contract"
	bridgehostcall "lina-core/pkg/pluginbridge/hostcall"
	bridgehostservice "lina-core/pkg/pluginbridge/hostservice"
)

// CronRegistrationCollector captures dynamic-plugin cron declarations during
// one host-driven discovery execution.
type CronRegistrationCollector interface {
	// Register validates and stores one discovered cron contract.
	Register(contract *bridgecontract.CronContract) error
}

// dispatchCronHostService routes cron host service methods to the discovery
// collector bound to the current Wasm execution.
func dispatchCronHostService(
	_ context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	switch method {
	case bridgehostservice.HostServiceMethodCronRegister:
		return handleHostCronRegister(hcc, payload)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported cron host service method: "+method,
		)
	}
}

// handleHostCronRegister validates one cron registration request and forwards
// it to the current discovery collector.
func handleHostCronRegister(
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	if hcc == nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, "host call context not available")
	}
	if bridgecontract.NormalizeExecutionSource(hcc.executionSource) != bridgecontract.ExecutionSourceCronDiscovery {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			"cron host service only supports cron discovery executions",
		)
	}
	if hcc.cronCollector == nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusInternalError,
			"cron discovery collector not configured",
		)
	}

	request, err := bridgehostservice.UnmarshalHostServiceCronRegisterRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	if request == nil || request.Contract == nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, "cron registration contract is required")
	}
	contractSnapshot := *request.Contract
	if err = bridgecontract.ValidateCronContracts(hcc.pluginID, []*bridgecontract.CronContract{&contractSnapshot}); err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	if err = hcc.cronCollector.Register(&contractSnapshot); err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	return bridgehostcall.NewHostCallSuccessResponse(nil)
}
