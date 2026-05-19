// This file implements the status command for development services.

package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"linactl/internal/devservice"
	"linactl/internal/toolutil"
)

// runStatus prints development service status using cross-platform checks.
func runStatus(_ context.Context, a *app, input commandInput) error {
	backendPort, err := input.Int("backend_port", defaultBackendPort)
	if err != nil {
		return err
	}
	frontendPort, err := input.Int("frontend_port", defaultFrontendPort)
	if err != nil {
		return err
	}
	services := devservice.Services(a.root, backendPort, frontendPort)

	if _, err = fmt.Fprintln(a.stdout, ""); err != nil {
		return fmt.Errorf("write status output: %w", err)
	}
	if _, err = fmt.Fprintln(a.stdout, "LinaPro Framework Status"); err != nil {
		return fmt.Errorf("write status title: %w", err)
	}

	rows := make([]devservice.StatusRow, 0, len(services))
	for _, service := range services {
		status := "stopped"
		if devservice.IsTCPListening(service.Port) || devservice.ServiceReady(service.URL, 2*time.Second) {
			status = "running"
		}
		pid := devservice.ReadPID(service.PIDPath)
		pidText := "-"
		if pid > 0 {
			pidText = strconv.Itoa(pid)
		}
		rows = append(rows, devservice.StatusRow{
			Service: service.Name,
			Status:  status,
			URL:     service.URL,
			PID:     pidText,
			PIDFile: toolutil.RelativePath(a.root, service.PIDPath),
			LogFile: toolutil.RelativePath(a.root, service.LogPath),
		})
	}
	if err = devservice.PrintStatusTable(a.stdout, rows); err != nil {
		return err
	}
	return nil
}
