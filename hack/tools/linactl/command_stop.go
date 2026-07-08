// This file implements the stop command for development services.

package main

import (
	"context"
	"fmt"

	"linactl/internal/devservice"
)

// runStop stops services that were started by linactl.
func runStop(ctx context.Context, a *app, input commandInput) error {
	if dir := input.Get("dir"); dir != "" {
		return runConfiguredCommandDir(ctx, a, dir, "stop")
	}
	_, err := stopServices(a, input)
	return err
}

// stopServices stops PID-file-backed development services and returns ports
// whose recorded processes received a stop signal.
func stopServices(a *app, input commandInput) ([]int, error) {
	backendPort, err := input.Int("backend_port", defaultBackendPort)
	if err != nil {
		return nil, err
	}
	frontendPort, err := input.Int("frontend_port", defaultFrontendPort)
	if err != nil {
		return nil, err
	}

	if _, err = fmt.Fprintln(a.stdout, "Stopping services..."); err != nil {
		return nil, fmt.Errorf("write stop output: %w", err)
	}
	stoppedPorts := make([]int, 0, 2)
	for _, service := range devservice.Services(a.root, backendPort, frontendPort) {
		stopped, stopErr := devservice.StopService(a.stdout, service)
		if stopErr != nil {
			return stoppedPorts, stopErr
		}
		if stopped {
			stoppedPorts = append(stoppedPorts, service.Port)
		}
	}
	return stoppedPorts, nil
}
