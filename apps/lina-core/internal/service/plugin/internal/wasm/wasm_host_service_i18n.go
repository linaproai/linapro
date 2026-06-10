// This file adapts runtime-translation host-service calls to the shared i18n
// capability service.

package wasm

import (
	"context"

	"lina-core/pkg/plugin/capability/i18ncap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchI18nHostService routes runtime translation host-service calls.
func dispatchI18nHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := i18nServiceForHostCall(hcc)
	if service == nil {
		return domainServiceNotScoped("i18n")
	}
	switch method {
	case bridgehostservice.HostServiceMethodI18nGetLocale:
		return capabilityJSONResponse(service.GetLocale(ctx))
	case bridgehostservice.HostServiceMethodI18nTranslate:
		var request translateRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		return capabilityJSONResponse(service.Translate(ctx, request.Key, request.Fallback))
	case bridgehostservice.HostServiceMethodI18nFindMessageKeys:
		var request findMessagesRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		return capabilityJSONResponse(service.FindMessageKeys(ctx, request.Prefix, request.Keyword))
	default:
		return domainMethodNotFound("i18n", method)
	}
}

// i18nServiceForHostCall resolves the runtime translation service for one host call.
func i18nServiceForHostCall(hcc *hostCallContext) i18ncap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.I18n()
}

// translateRequest carries one translation request.
type translateRequest struct {
	Key      string `json:"key"`
	Fallback string `json:"fallback"`
}

// findMessagesRequest carries message-key lookup filters.
type findMessagesRequest struct {
	Prefix  string `json:"prefix"`
	Keyword string `json:"keyword"`
}
