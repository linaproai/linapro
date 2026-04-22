// demo_ping.go implements the public ping behavior for the demo service.

package demo

import "context"

// pingMessage is the static response returned by the public ping endpoint.
const pingMessage = "pong"

// PingOutput defines one public plugin ping payload.
type PingOutput struct {
	// Message is the public ping response returned from the plugin API.
	Message string
}

// Ping returns one public plugin ping payload.
func (s *serviceImpl) Ping(_ context.Context) (out *PingOutput, err error) {
	return &PingOutput{
		Message: pingMessage,
	}, nil
}
