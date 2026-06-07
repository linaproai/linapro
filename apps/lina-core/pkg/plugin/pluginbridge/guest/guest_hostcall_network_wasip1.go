//go:build wasip1

// This file provides guest-side helpers for the governed outbound network host service.

package guest

import "lina-core/pkg/plugin/pluginbridge/protocol"

// networkHostService is the default guest-side outbound network host-service client.
type networkHostService struct{}

// defaultNetworkHostService stores the singleton outbound network host-service
// client used by package-level helpers.
var defaultNetworkHostService NetworkHostService = &networkHostService{}

// Network returns the outbound network host service guest client.
func Network() NetworkHostService {
	return defaultNetworkHostService
}

// Request executes one governed outbound HTTP request through the host.
func (s *networkHostService) Request(
	targetURL string,
	request *protocol.HostServiceNetworkRequest,
) (*protocol.HostServiceNetworkResponse, error) {
	payload, err := invokeGuestHostService(
		protocol.HostServiceNetwork,
		protocol.HostServiceMethodNetworkRequest,
		targetURL,
		"",
		protocol.MarshalHostServiceNetworkRequest(request),
	)
	if err != nil {
		return nil, err
	}
	return protocol.UnmarshalHostServiceNetworkResponse(payload)
}
