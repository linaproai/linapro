// This file implements the dev command that rebuilds and restarts local services.

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"linactl/internal/devservice"
	"linactl/internal/frontend"
	"linactl/internal/portcheck"
	"linactl/internal/process"
	"linactl/internal/toolutil"
)

// runDev builds and starts backend and frontend development services.
func runDev(ctx context.Context, a *app, input commandInput) error {
	backendPort, err := input.Int("backend_port", defaultBackendPort)
	if err != nil {
		return err
	}
	frontendPort, err := input.Int("frontend_port", defaultFrontendPort)
	if err != nil {
		return err
	}
	// 在重建/启动任何子进程之前，先校验三处端口一致性（命令传入的
	// backend_port、后端 manifest config 的 server.address、前端 vite proxy
	// target）；任一不一致都直接 fail-fast，避免错配端口启动后再以"探活
	// 超时"或"接口 404"等间接形式暴露问题。
	// Verify backend_port, backend manifest server.address, and frontend vite
	// proxy targets all agree before launching any subprocess. Any mismatch is
	// surfaced immediately instead of letting downstream services start with
	// inconsistent ports.
	if err = portcheck.Verify(a.root, backendPort); err != nil {
		return err
	}
	if err = ensureFrontendDeps(ctx, a); err != nil {
		return err
	}
	pluginsEnabled, env, err := prepareOfficialPluginBuildEnv(ctx, a, input)
	if err != nil {
		return err
	}
	skipWasm, err := input.Bool("skip_wasm", !pluginsEnabled)
	if err != nil {
		return err
	}

	stopInput := commandInput{Params: map[string]string{
		"backend_port":  strconv.Itoa(backendPort),
		"frontend_port": strconv.Itoa(frontendPort),
	}}
	if err = runStop(ctx, a, stopInput); err != nil {
		return err
	}

	// After stopping our own services, verify the ports are actually free.
	// If something else is still bound, fail fast with an actionable hint
	// instead of letting the backend silently die with "address already in use".
	// 停掉自身服务后再校验端口空闲，杜绝外部进程占用导致的"假成功"。
	if err = devservice.EnsurePortsAvailable(a.portInUse, backendPort, frontendPort); err != nil {
		return err
	}

	tempDir := filepath.Join(a.root, "temp")
	binDir := filepath.Join(tempDir, "bin")
	if err = os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("create temp bin directory: %w", err)
	}

	if !skipWasm {
		if err = runWasm(ctx, a, commandInput{Params: map[string]string{"out": filepath.Join(a.root, "temp", "output")}}); err != nil {
			return err
		}
	}
	if err = runPreparePackedAssets(ctx, a, commandInput{}); err != nil {
		return err
	}

	backendBinary := filepath.Join(binDir, toolutil.ExecutableName("lina"))
	if err = os.Remove(backendBinary); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove existing backend binary: %w", err)
	}
	if _, err = fmt.Fprintln(a.stdout, "Building backend..."); err != nil {
		return fmt.Errorf("write build output: %w", err)
	}
	if err = a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "apps", "lina-core"), Env: env}, "go", "build", "-o", backendBinary, "."); err != nil {
		return err
	}

	services := devservice.Services(a.root, backendPort, frontendPort)
	services[0].StartName = backendBinary
	services[1].StartName = toolutil.ViteCommand(a.root)
	previousEnv := a.env
	a.env = env
	defer func() {
		a.env = previousEnv
	}()
	for _, service := range services {
		if err = devservice.StartService(a.root, a.stdout, a.stderr, a.env, a.execCommand, process.ConfigureDetached, service); err != nil {
			return err
		}
	}

	for _, service := range services {
		// readinessURL distinguishes the backend from the frontend so we do
		// not accept arbitrary 4xx responses from an unrelated process that
		// happens to be bound to the port (the backend exposes /api.json
		// configured via server.extensions.apiDocPath).
		// 后端就绪探测打 /api.json 并要求 2xx，避免外部进程的 404 被误判为就绪。
		readinessURL := devservice.ReadinessURL(service)
		if err = a.waitHTTP(service.Name, readinessURL, service.PIDPath, service.LogPath, defaultWaitTimeout); err != nil {
			tailErr := devservice.TailLogToWriter(a.stderr, service.Name, service.LogPath, devservice.DefaultReadinessTailLines)
			if tailErr != nil {
				fmt.Fprintf(a.stderr, "warning: tail %s log failed: %v\n", service.Name, tailErr)
			}
			return err
		}
		if _, err = fmt.Fprintf(a.stdout, "%s is ready: %s\n", service.Name, service.URL); err != nil {
			return fmt.Errorf("write readiness output: %w", err)
		}
	}

	return runStatus(ctx, a, stopInput)
}

// ensureFrontendDeps delegates frontend dependency checks to the frontend
// subcomponent while preserving the command-level app dependency shape.
func ensureFrontendDeps(ctx context.Context, a *app) error {
	return frontend.EnsureDeps(ctx, a.root, a.stdout, func(runCtx context.Context, dir string, name string, args ...string) error {
		return a.runCommand(runCtx, commandOptions{Dir: dir}, name, args...)
	})
}
