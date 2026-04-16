package demo

import "context"

const pingMessage = "pong"

// PingOutput defines one public plugin ping payload.
type PingOutput struct {
	// Message is the public ping response returned from the plugin API.
	Message string
}

// Ping returns one public plugin ping payload.
func (s *Service) Ping(ctx context.Context) (out *PingOutput, err error) {
	return &PingOutput{
		Message: pingMessage,
	}, nil
}
