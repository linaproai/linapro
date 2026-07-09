// This file implements the guest-side outbound network host-service client.

package domainhostcall

import "lina-core/pkg/plugin/pluginbridge/protocol"

// networkService adapts the network host service to the pluginbridge network
// helper contract.
type networkService struct{ baseService }

// Network creates the outbound network host service guest client.
func Network(invoker HostServiceInvoker) *networkService {
	return &networkService{baseService: newBaseServiceWithHostService(nil, invoker)}
}

// Request executes one governed outbound HTTP request through the host.
func (s *networkService) Request(
	targetURL string,
	request *protocol.HostServiceNetworkRequest,
) (*protocol.HostServiceNetworkResponse, error) {
	payload, err := s.callHostService(
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
