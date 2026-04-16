// Package demo implements public demo services exposed by the plugin-demo-source
// backend.
package demo

import "context"

// Service defines the demo service contract.
type Service interface {
	// Ping returns the public ping payload used by route verification.
	Ping(ctx context.Context) (out *PingOutput, err error)
	// Summary returns the concise backend summary rendered on the plugin page.
	Summary(ctx context.Context) (out *SummaryOutput, err error)
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct{}

// New creates and returns a new demo service instance.
func New() Service {
	return &serviceImpl{}
}
