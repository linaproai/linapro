// This file defines the dynamic-plugin declaration contract used by guest
// plugin startup code. It keeps build/discovery declarations separate from
// runtime domain capability services.

package pluginbridge

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// declarations is the default guest-side declaration facade.
type declarations struct {
	routes RouteDeclarations
	jobs   JobDeclarations
}

// NewDeclarations creates one dynamic-plugin declaration facade. The default
// route facade is a no-op so the same RegisterPlugin function can run during
// host-driven Jobs discovery, while the default Jobs facade uses the governed
// jobs.register host-call.
func NewDeclarations(options ...DeclarationOption) Declarations {
	plugin := &declarations{
		routes: noopRouteDeclarations{},
		jobs:   jobDeclarations{},
	}
	for _, option := range options {
		if option != nil {
			option(plugin)
		}
	}
	return plugin
}

// WithDeclarationRoutes replaces the route declaration facade.
func WithDeclarationRoutes(routes RouteDeclarations) DeclarationOption {
	return func(plugin *declarations) {
		if plugin != nil && routes != nil {
			plugin.routes = routes
		}
	}
}

// WithDeclarationJobs replaces the Jobs declaration facade.
func WithDeclarationJobs(jobs JobDeclarations) DeclarationOption {
	return func(plugin *declarations) {
		if plugin != nil && jobs != nil {
			plugin.jobs = jobs
		}
	}
}

// Routes returns the dynamic route declaration facade.
func (p *declarations) Routes() RouteDeclarations {
	if p == nil || p.routes == nil {
		return noopRouteDeclarations{}
	}
	return p.routes
}

// Jobs returns the built-in Jobs declaration facade.
func (p *declarations) Jobs() JobDeclarations {
	if p == nil || p.jobs == nil {
		return jobDeclarations{}
	}
	return p.jobs
}

// noopRouteDeclarations accepts route declarations when the dynamic plugin
// declaration facade is executed at runtime for non-route discovery.
type noopRouteDeclarations struct{}

// Group ignores route declarations during runtime discovery executions.
func (noopRouteDeclarations) Group(_ string, _ string) error {
	return nil
}

// jobDeclarations submits Jobs declarations through governed host services.
type jobDeclarations struct{}

var _ JobDeclarations = jobDeclarations{}

// Register submits one dynamic-plugin job declaration to the current host-side
// Jobs discovery collector.
func (jobDeclarations) Register(contract *protocol.JobContract) error {
	if contract == nil {
		return gerror.New("job contract cannot be nil")
	}
	contractSnapshot := *contract
	_, err := invokeGuestHostService(
		protocol.HostServiceJobs,
		protocol.HostServiceMethodJobsRegister,
		"",
		"",
		protocol.MarshalHostServiceJobsRegisterRequest(&protocol.HostServiceJobsRegisterRequest{Contract: &contractSnapshot}),
	)
	return err
}
