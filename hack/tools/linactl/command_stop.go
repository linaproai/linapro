// This file implements the stop command for development services.

package main

import (
	"context"
	"fmt"

	"linactl/internal/devservice"
)

// runStop stops services that were started by linactl.
func runStop(_ context.Context, a *app, input commandInput) error {
	backendPort, err := input.Int("backend_port", defaultBackendPort)
	if err != nil {
		return err
	}
	frontendPort, err := input.Int("frontend_port", defaultFrontendPort)
	if err != nil {
		return err
	}

	if _, err = fmt.Fprintln(a.stdout, "Stopping services..."); err != nil {
		return fmt.Errorf("write stop output: %w", err)
	}
	for _, service := range devservice.Services(a.root, backendPort, frontendPort) {
		if err = devservice.StopService(a.stdout, service); err != nil {
			return err
		}
	}
	return nil
}
