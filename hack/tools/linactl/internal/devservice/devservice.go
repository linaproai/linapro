// Package devservice provides development service definitions, process
// lifecycle helpers, readiness checks, and status-table rendering for linactl
// dev, stop, and status commands. It keeps command files focused on option
// parsing and orchestration while preserving platform-neutral process logic.
package devservice

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"linactl/internal/toolutil"
)

// Config stores development service paths and ports.
type Config struct {
	Name      string
	URL       string
	Port      int
	PIDPath   string
	LogPath   string
	WorkDir   string
	StartName string
	StartArgs []string
}

// StatusRow stores one printable development service status row.
type StatusRow struct {
	Service string
	Status  string
	URL     string
	PID     string
	PIDFile string
	LogFile string
}

// ProcessRunner creates child process commands for service startup.
type ProcessRunner func(context.Context, string, ...string) *exec.Cmd

// Services returns backend and frontend development service definitions.
func Services(root string, backendPort int, frontendPort int) []Config {
	tempDir := filepath.Join(root, "temp")
	pidDir := filepath.Join(tempDir, "pids")
	return []Config{
		{
			Name:    "Backend",
			URL:     fmt.Sprintf("http://127.0.0.1:%d/", backendPort),
			Port:    backendPort,
			PIDPath: filepath.Join(pidDir, "backend.pid"),
			LogPath: filepath.Join(tempDir, "lina-core.log"),
			WorkDir: filepath.Join(root, "apps", "lina-core"),
		},
		{
			Name:      "Frontend",
			URL:       fmt.Sprintf("http://127.0.0.1:%d/", frontendPort),
			Port:      frontendPort,
			PIDPath:   filepath.Join(pidDir, "frontend.pid"),
			LogPath:   filepath.Join(tempDir, "lina-vben.log"),
			WorkDir:   filepath.Join(root, "apps", "lina-vben", "apps", "web-antd"),
			StartArgs: []string{"--mode", "development", "--host", "127.0.0.1", "--port", strconv.Itoa(frontendPort), "--strictPort"},
		},
	}
}

// PrintStatusTable renders development service status without terminal-specific dependencies.
func PrintStatusTable(out io.Writer, rows []StatusRow) error {
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
func (r StatusRow) values() []string {
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

// StartService starts a development service and records its PID file.
func StartService(root string, stdout io.Writer, env []string, runner ProcessRunner, detach func(*exec.Cmd), service Config) error {
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

	cmd := runner(context.Background(), service.StartName, service.StartArgs...)
	cmd.Dir = service.WorkDir
	cmd.Env = env
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	detach(cmd)
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
	fmt.Fprintf(stdout, "%s started: pid=%d log=%s\n", service.Name, pid, toolutil.RelativePath(root, service.LogPath))
	return nil
}

// StopService stops a PID-file-backed service when possible.
func StopService(out io.Writer, service Config) error {
	pid := ReadPID(service.PIDPath)
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

// WaitHTTP waits for one service URL to become ready.
func WaitHTTP(name string, url string, pidPath string, logPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := newReadinessHTTPClient(2 * time.Second)
	for time.Now().Before(deadline) {
		if ReadPID(pidPath) == 0 {
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

// ServiceReady reports whether an HTTP endpoint responds without server error.
func ServiceReady(url string, timeout time.Duration) bool {
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

// IsTCPListening reports whether localhost accepts TCP connections on a port.
func IsTCPListening(port int) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), time.Second)
	if err != nil {
		return false
	}
	if closeErr := conn.Close(); closeErr != nil {
		return false
	}
	return true
}

// ReadPID reads and validates a PID file.
func ReadPID(path string) int {
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
