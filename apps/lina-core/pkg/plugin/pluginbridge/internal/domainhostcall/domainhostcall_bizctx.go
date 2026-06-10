// This file implements the guest-side bizctx capability hostcall client.
// The host remains authoritative for current identity and scope fields.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// bizCtxService adapts current business-context reads to host services.
type bizCtxService struct{ baseService }

// BizCtx creates the current request business-context guest client.
func BizCtx(invoker Invoker) bizctxcap.Service {
	return bizCtxService{baseService: newBaseService(invoker)}
}

// Current returns a read-only snapshot of request context fields.
func (s bizCtxService) Current(_ context.Context) bizctxcap.CurrentContext {
	var out bizctxcap.CurrentContext
	if err := s.callJSONRequest(protocol.HostServiceBizCtx, protocol.HostServiceMethodBizCtxCurrent, nil, &out); err != nil {
		return bizctxcap.CurrentContext{}
	}
	return out
}

var _ bizctxcap.Service = (*bizCtxService)(nil)
