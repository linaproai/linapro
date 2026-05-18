// This file implements development service lifecycle commands and readiness checks.

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	services := a.services(backendPort, frontendPort)

	if _, err = fmt.Fprintln(a.stdout, ""); err != nil {
		return fmt.Errorf("write status output: %w", err)
	}
	if _, err = fmt.Fprintln(a.stdout, "LinaPro Framework Status"); err != nil {
		return fmt.Errorf("write status title: %w", err)
	}

	rows := make([]serviceStatusRow, 0, len(services))
	for _, service := range services {
		status := "stopped"
		if isTCPListening(service.Port) || serviceReady(service.URL, 2*time.Second) {
			status = "running"
		}
		pid := readPID(service.PIDPath)
		pidText := "-"
		if pid > 0 {
			pidText = strconv.Itoa(pid)
		}
		rows = append(rows, serviceStatusRow{
			Service: service.Name,
			Status:  status,
			URL:     service.URL,
			PID:     pidText,
			PIDFile: relativePath(a.root, service.PIDPath),
			LogFile: relativePath(a.root, service.LogPath),
		})
	}
	if err = printStatusTable(a.stdout, rows); err != nil {
		return err
	}
	return nil
}

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
	for _, service := range a.services(backendPort, frontendPort) {
		if err = stopService(a.stdout, service); err != nil {
			return err
		}
	}
	return nil
}

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

	backendBinary := filepath.Join(binDir, executableName("lina"))
	if err = os.Remove(backendBinary); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove existing backend binary: %w", err)
	}
	if _, err = fmt.Fprintln(a.stdout, "Building backend..."); err != nil {
		return fmt.Errorf("write build output: %w", err)
	}
	if err = a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "apps", "lina-core"), Env: env}, "go", "build", "-o", backendBinary, "."); err != nil {
		return err
	}

	services := a.services(backendPort, frontendPort)
	services[0].StartName = backendBinary
	services[1].StartName = viteCommand(a.root)
	previousEnv := a.env
	a.env = env
	defer func() {
		a.env = previousEnv
	}()
	for _, service := range services {
		if err = startService(a, service); err != nil {
			return err
		}
	}

	for _, service := range services {
		if err = a.waitHTTP(service.Name, service.URL, service.PIDPath, service.LogPath, defaultWaitTimeout); err != nil {
			return err
		}
		if _, err = fmt.Fprintf(a.stdout, "%s is ready: %s\n", service.Name, service.URL); err != nil {
			return fmt.Errorf("write readiness output: %w", err)
		}
	}

	return runStatus(ctx, a, stopInput)
}

// services returns backend and frontend development service definitions.
func (a *app) services(backendPort int, frontendPort int) []serviceConfig {
	tempDir := filepath.Join(a.root, "temp")
	pidDir := filepath.Join(tempDir, "pids")
	return []serviceConfig{
		{
			Name:    "Backend",
			URL:     fmt.Sprintf("http://127.0.0.1:%d/", backendPort),
			Port:    backendPort,
			PIDPath: filepath.Join(pidDir, "backend.pid"),
			LogPath: filepath.Join(tempDir, "lina-core.log"),
			WorkDir: filepath.Join(a.root, "apps", "lina-core"),
		},
		{
			Name:      "Frontend",
			URL:       fmt.Sprintf("http://127.0.0.1:%d/", frontendPort),
			Port:      frontendPort,
			PIDPath:   filepath.Join(pidDir, "frontend.pid"),
			LogPath:   filepath.Join(tempDir, "lina-vben.log"),
			WorkDir:   filepath.Join(a.root, "apps", "lina-vben", "apps", "web-antd"),
			StartArgs: []string{"--mode", "development", "--host", "127.0.0.1", "--port", strconv.Itoa(frontendPort), "--strictPort"},
		},
	}
}

// printStatusTable renders development service status without terminal-specific dependencies.
func printStatusTable(out io.Writer, rows []serviceStatusRow) error {
	headers := []string{"Service", "Status", "URL", "PID", "PID File", "Log File"}
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	for _, row := range rows {
		values := row.values()
		for i, value := range values {
			if len(value) > widths[i] {
				widths[i] = len(value)
			}
		}
	}

	if err := printTableBorder(out, widths); err != nil {
		return err
	}
	if err := printTableRow(out, widths, headers); err != nil {
		return err
	}
	if err := printTableBorder(out, widths); err != nil {
		return err
	}
	for _, row := range rows {
		if err := printTableRow(out, widths, row.values()); err != nil {
			return err
		}
	}
	if err := printTableBorder(out, widths); err != nil {
		return err
	}
	return nil
}

// values returns the printable table cells for one service status row.
func (r serviceStatusRow) values() []string {
	return []string{r.Service, r.Status, r.URL, r.PID, r.PIDFile, r.LogFile}
}

// printTableBorder prints one ASCII border line for a table.
func printTableBorder(out io.Writer, widths []int) error {
	if _, err := fmt.Fprint(out, "+"); err != nil {
		return fmt.Errorf("write table border: %w", err)
	}
	for _, width := range widths {
		if _, err := fmt.Fprint(out, strings.Repeat("-", width+2)); err != nil {
			return fmt.Errorf("write table border: %w", err)
		}
		if _, err := fmt.Fprint(out, "+"); err != nil {
			return fmt.Errorf("write table border: %w", err)
		}
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return fmt.Errorf("write table border: %w", err)
	}
	return nil
}

// printTableRow prints one padded ASCII table row.
func printTableRow(out io.Writer, widths []int, values []string) error {
	if _, err := fmt.Fprint(out, "|"); err != nil {
		return fmt.Errorf("write table row: %w", err)
	}
	for i, value := range values {
		if _, err := fmt.Fprintf(out, " %-*s |", widths[i], value); err != nil {
			return fmt.Errorf("write table row: %w", err)
		}
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return fmt.Errorf("write table row: %w", err)
	}
	return nil
}

// startService starts a development service and records its PID file.
func startService(a *app, service serviceConfig) error {
	if err := os.MkdirAll(filepath.Dir(service.PIDPath), 0o755); err != nil {
		return fmt.Errorf("create PID directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(service.LogPath), 0o755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}
	logFile, err := os.OpenFile(service.LogPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open %s: %w", service.LogPath, err)
	}

	cmd := a.execCommand(context.Background(), service.StartName, service.StartArgs...)
	cmd.Dir = service.WorkDir
	cmd.Env = a.env
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	configureDetachedProcess(cmd)
	if err = cmd.Start(); err != nil {
		if closeErr := logFile.Close(); closeErr != nil {
			return fmt.Errorf("start %s failed and close log failed: %v: %w", service.Name, closeErr, err)
		}
		return fmt.Errorf("start %s: %w", service.Name, err)
	}
	pid := cmd.Process.Pid
	if err = os.WriteFile(service.PIDPath, []byte(strconv.Itoa(pid)), 0o644); err != nil {
		return fmt.Errorf("write %s PID file: %w", service.Name, err)
	}
	if err = logFile.Close(); err != nil {
		return fmt.Errorf("close %s log file: %w", service.Name, err)
	}
	if err = cmd.Process.Release(); err != nil {
		return fmt.Errorf("release %s process: %w", service.Name, err)
	}
	fmt.Fprintf(a.stdout, "%s started: pid=%d log=%s\n", service.Name, pid, relativePath(a.root, service.LogPath))
	return nil
}

// stopService stops a PID-file-backed service when possible.
func stopService(out io.Writer, service serviceConfig) error {
	pid := readPID(service.PIDPath)
	stopped := false
	if pid > 0 {
		process, err := os.FindProcess(pid)
		if err == nil {
			if killErr := process.Kill(); killErr == nil {
				stopped = true
			}
		}
	}
	if err := os.Remove(service.PIDPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove %s PID file: %w", service.Name, err)
	}
	if stopped {
		fmt.Fprintf(out, "%s stopped\n", service.Name)
		return nil
	}
	fmt.Fprintf(out, "%s is not running\n", service.Name)
	return nil
}

// waitHTTP waits for one service URL to become ready.
func waitHTTP(name string, url string, pidPath string, logPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := newReadinessHTTPClient(2 * time.Second)
	for time.Now().Before(deadline) {
		if readPID(pidPath) == 0 {
			return fmt.Errorf("%s startup failed: PID file does not exist; check log: %s", name, logPath)
		}
		resp, err := client.Get(url)
		if err == nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				return fmt.Errorf("close %s readiness response: %w", name, closeErr)
			}
			if resp.StatusCode < http.StatusInternalServerError {
				return nil
			}
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("%s startup timed out (%s): %s; check log: %s", name, timeout, url, logPath)
}

// serviceReady reports whether an HTTP endpoint responds without server error.
func serviceReady(url string, timeout time.Duration) bool {
	client := newReadinessHTTPClient(timeout)
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	if closeErr := resp.Body.Close(); closeErr != nil {
		return false
	}
	return resp.StatusCode < http.StatusInternalServerError
}

// newReadinessHTTPClient matches curl-style readiness by accepting redirects as responses.
func newReadinessHTTPClient(timeout time.Duration) http.Client {
	return http.Client{
		Timeout: timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// isTCPListening reports whether localhost accepts TCP connections on a port.
func isTCPListening(port int) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), time.Second)
	if err != nil {
		return false
	}
	if closeErr := conn.Close(); closeErr != nil {
		return false
	}
	return true
}

// readPID reads and validates a PID file.
func readPID(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	text := strings.TrimSpace(string(content))
	pid, err := strconv.Atoi(text)
	if err != nil || pid <= 1 {
		return 0
	}
	return pid
}
