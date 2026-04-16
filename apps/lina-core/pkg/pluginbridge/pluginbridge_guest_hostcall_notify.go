//go:build wasip1

// This file provides guest-side helpers for the governed unified notify host service.

package pluginbridge

// NotifyHostService exposes guest-side helpers for the governed unified notify host service.
type NotifyHostService interface {
	// Send sends one governed notification through the authorized channel.
	Send(channelKey string, request *HostServiceNotifySendRequest) (*HostServiceNotifySendResponse, error)
}

type notifyHostService struct{}

var defaultNotifyHostService NotifyHostService = &notifyHostService{}

// Notify returns the unified notify host service guest client.
func Notify() NotifyHostService {
	return defaultNotifyHostService
}

// Send sends one governed notification through the authorized channel.
func (s *notifyHostService) Send(
	channelKey string,
	request *HostServiceNotifySendRequest,
) (*HostServiceNotifySendResponse, error) {
	payload, err := invokeHostService(
		HostServiceNotify,
		HostServiceMethodNotifySend,
		channelKey,
		"",
		MarshalHostServiceNotifySendRequest(request),
	)
	if err != nil {
		return nil, err
	}
	return UnmarshalHostServiceNotifySendResponse(payload)
}
