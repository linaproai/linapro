// This file implements the guest-side route capability hostcall client.
// The host remains authoritative for dynamic route metadata fields.

package domainhostcall

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/plugin/capability/routecap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// routeService adapts dynamic route metadata reads to host services.
type routeService struct{ baseService }

// Route creates the current dynamic-route metadata guest client.
func Route(invoker Invoker) routecap.Service {
	return routeService{baseService: newBaseService(invoker)}
}

// DynamicRouteMetadata returns the current dynamic-route metadata projection.
func (s routeService) DynamicRouteMetadata(_ *ghttp.Request) *routecap.DynamicRouteMetadata {
	var out routecap.DynamicRouteMetadata
	if err := s.callJSONRequest(protocol.HostServiceRoute, protocol.HostServiceMethodRouteMetadataGet, nil, &out); err != nil {
		return nil
	}
	return &out
}

var _ routecap.Service = (*routeService)(nil)
