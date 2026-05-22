// This file implements the status command for development services.

package main

import (
	"context"
	"fmt"
	"strconv"

	"linactl/internal/devservice"
	"linactl/internal/process"
	"linactl/internal/toolutil"
)

// runStatus prints development service status using cross-platform checks.
//
// A service only counts as "running" when its PID file points to a live
// process AND that process actually accepts TCP connections. This rejects
// two false-positive scenarios:
//  1. Stale PID file after the previous run crashed (process gone but the
//     port now happens to be listening — possibly an external occupant).
//  2. PID alive but the port is not bound yet (still booting).
//
// 仅当 PID 文件指向存活进程、且对应端口在监听时才算运行中，避免把外部占用方
// 或残留 PID 误判为本项目运行中。
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

	aliveProbe := a.processAlive
	if aliveProbe == nil {
		aliveProbe = process.Alive
	}
	listenProbe := a.portInUse
	if listenProbe == nil {
		listenProbe = devservice.IsTCPListening
	}

	rows := make([]devservice.StatusRow, 0, len(services))
	for _, service := range services {
		pid := devservice.ReadPID(service.PIDPath)
		alive := pid > 0 && aliveProbe(pid)
		status := "stopped"
		if alive && listenProbe(service.Port) {
			status = "running"
		}
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
