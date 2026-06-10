// This file adapts infrastructure host-service calls to the shared infra
// capability service.

package wasm

import (
	"context"
	"strings"

	"lina-core/pkg/plugin/capability/infracap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchInfraHostService routes infrastructure-domain host-service calls.
func dispatchInfraHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := infraServiceForHostCall(hcc)
	if service == nil {
		return domainServiceNotScoped("infra")
	}
	if method != bridgehostservice.HostServiceMethodInfraBatchGetStatus {
		return domainMethodNotFound("infra", method)
	}
	var request idsRequest
	if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
		return invalidCapabilityRequest(err)
	}
	capCtx := capabilityContextForHostCall(hcc, bridgehostservice.HostServiceInfra, method)
	result, err := service.BatchGetStatus(ctx, capCtx, componentIDs(request.IDs))
	return domainCapabilityResult(result, err)
}

// infraServiceForHostCall resolves the infrastructure service for one host call.
func infraServiceForHostCall(hcc *hostCallContext) infracap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.Infra()
}

// componentIDs converts transport string identifiers into typed component IDs.
func componentIDs(ids []string) []infracap.ComponentID {
	out := make([]infracap.ComponentID, 0, len(ids))
	for _, id := range ids {
		if value := strings.TrimSpace(id); value != "" {
			out = append(out, infracap.ComponentID(value))
		}
	}
	return out
}
