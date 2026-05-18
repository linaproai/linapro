//go:build wasip1

// This file provides guest-side helpers for the governed outbound HTTP host service.

package guest

// HTTPHostService exposes guest-side helpers for the governed outbound HTTP host service.
type HTTPHostService interface {
	// Request executes one governed outbound HTTP request through the host.
	Request(targetURL string, request *HostServiceNetworkRequest) (*HostServiceNetworkResponse, error)
}

// httpHostService is the default guest-side outbound HTTP host-service client.
type httpHostService struct{}

// defaultHTTPHostService stores the singleton outbound HTTP host-service
// client used by package-level helpers.
var defaultHTTPHostService HTTPHostService = &httpHostService{}

// HTTP returns the outbound HTTP host service guest client.
func HTTP() HTTPHostService {
	return defaultHTTPHostService
}

// Request executes one governed outbound HTTP request through the host.
func (s *httpHostService) Request(
	targetURL string,
	request *HostServiceNetworkRequest,
) (*HostServiceNetworkResponse, error) {
	payload, err := invokeHostService(
		HostServiceNetwork,
		HostServiceMethodNetworkRequest,
		targetURL,
		"",
		MarshalHostServiceNetworkRequest(request),
	)
	if err != nil {
		return nil, err
	}
	return UnmarshalHostServiceNetworkResponse(payload)
}
