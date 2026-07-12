// Package devservice provides development service definitions, process
// lifecycle helpers, readiness checks, and status-table rendering for linactl
// dev, stop, and status commands. It keeps command files focused on option
// parsing and orchestration while preserving platform-neutral process logic.
package devservice

import (
	"bufio"
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

	"linactl/internal/process"
	"linactl/internal/toolutil"
)

// readinessLogTailLines bounds how many trailing log lines are echoed to the
// terminal when a development service fails its readiness probe.
// 当就绪探测失败时，回显日志末尾的最大行数，便于失败时直接定位错误。
const readinessLogTailLines = 30

// Config stores development service paths and ports.
type Config struct {
	Name        string
	DisplayName string
	URL         string
	Port        int
	PIDPath     string
	LogPath     string
	WorkDir     string
	StartName   string
	StartArgs   []string
	Env         []string
}

// RunnerContextServiceEnvKey carries service-level environment additions to
// ProcessRunner implementations that need to inspect them in tests.
const RunnerContextServiceEnvKey = "LINACTL_SERVICE_ENV"

// StatusRow stores one printable development service status row.
type StatusRow struct {
	Entry   string
	Status  string
	URL     string
	PID     string
	PIDFile string
	LogFile string
}

// ProcessRunner creates child process commands for service startup.
type ProcessRunner func(context.Context, string, ...string) *exec.Cmd

// PortInUseFunc reports whether the given TCP port is currently bound. It is
// declared as a type alias so command-level code can inject test doubles.
// 端口占用检测函数类型别名，便于测试注入；生产路径使用 IsTCPListening。
type PortInUseFunc = func(int) bool

// ProcessAliveFunc reports whether a PID currently belongs to a live process.
// It is declared as a type alias so command-level code can inject test doubles.
// 进程存活检测函数类型别名，便于测试注入。
type ProcessAliveFunc = func(int) bool

// ProcessListFunc returns visible operating-system processes.
type ProcessListFunc func() ([]process.Info, error)

// ProcessKillFunc stops one process by PID.
type ProcessKillFunc func(int) error

// Services returns backend and frontend development service definitions.
// Runtime listen and proxy ports are injected through process environment so
// make dev LINA_CORE_PORT=/LINA_VBEN_PORT= (or matching env vars) can start the
// stack without editing config.yaml or vite.config.mts proxy literals.
// 通过进程环境注入监听与代理端口，使 make dev LINA_CORE_PORT=/LINA_VBEN_PORT=
// （或等价环境变量）无需改配置文件即可启动。
func Services(root string, backendPort int, frontendPort int) []Config {
	tempDir := filepath.Join(root, "temp")
	pidDir := filepath.Join(tempDir, "pids")
	backendProxyTarget := fmt.Sprintf("http://127.0.0.1:%d", backendPort)
	return []Config{
		{
			Name:        "Backend",
			DisplayName: "Lina Core",
			URL:         fmt.Sprintf("http://127.0.0.1:%d/", backendPort),
			Port:        backendPort,
			PIDPath:     filepath.Join(pidDir, "lina-core.pid"),
			LogPath:     filepath.Join(tempDir, "lina-core.log"),
			WorkDir:     filepath.Join(root, "apps", "lina-core"),
			Env: []string{
				// Override config server.address so the binary listens on the
				// requested port regardless of manifest defaults.
				// 覆盖 config 中的 server.address，使后端监听请求端口。
				fmt.Sprintf("LINAPRO_SERVER_ADDRESS=:%d", backendPort),
				fmt.Sprintf("LINAPRO_FRONTEND_DEV_SERVER_URL=http://127.0.0.1:%d", frontendPort),
			},
		},
		{
			Name:        "Frontend",
			DisplayName: "Lina Vben",
			URL:         fmt.Sprintf("http://127.0.0.1:%d/", frontendPort),
			Port:        frontendPort,
			PIDPath:     filepath.Join(pidDir, "lina-vben.pid"),
			LogPath:     filepath.Join(tempDir, "lina-vben.log"),
			WorkDir:     filepath.Join(root, "apps", "lina-vben", "apps", "web-antd"),
			StartArgs:   []string{"--mode", "development", "--host", "127.0.0.1", "--port", strconv.Itoa(frontendPort), "--strictPort"},
			Env: []string{
				// Drive vite proxy targets without rewriting vite.config.mts.
				// 通过环境变量驱动 vite 代理目标，无需改写 vite.config.mts。
				fmt.Sprintf("LINAPRO_BACKEND_PROXY_TARGET=%s", backendProxyTarget),
				fmt.Sprintf("VITE_PORT=%d", frontendPort),
			},
		},
	}
}

// EntryName returns the user-facing development entry name for terminal output.
func EntryName(service Config) string {
	if strings.TrimSpace(service.DisplayName) != "" {
		return service.DisplayName
	}
	return service.Name
}

// ReadinessURL returns the URL that WaitHTTP should probe for the given
// service. The backend exposes /api.json (configured via
// server.extensions.apiDocPath) and we require it to answer 2xx, which
// rejects unrelated occupants whose 4xx on / would otherwise be accepted as
// "ready". The frontend Vite dev server returns 200 on /.
// 就绪探测 URL：后端命中 /api.json 才算就绪，避免外部进程的 404 被误判。
func ReadinessURL(service Config) string {
	if service.Name == "Backend" {
		return service.URL + "api.json"
	}
	return service.URL
}

// PrintStatusTable renders development service status without terminal-specific dependencies.
func PrintStatusTable(out io.Writer, rows []StatusRow) error {
	headers := []string{"Entry", "Status", "URL", "PID", "PID File", "Log File"}
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
	return []string{r.Entry, r.Status, r.URL, r.PID, r.PIDFile, r.LogFile}
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

// EnsurePortsAvailable returns an error when any required development port is
// already bound, providing the operator with an actionable message. Tests pass
// a custom probe; production uses IsTCPListening.
// 校验开发端口是否空闲，被占用时直接报错，避免后端启动后 fatal "address already in use"
// 却被外部占用方的 4xx 响应假装为就绪。
func EnsurePortsAvailable(probe PortInUseFunc, backendPort int, frontendPort int) error {
	occupied := occupiedPorts(probe, backendPort, frontendPort)
	if len(occupied) == 0 {
		return nil
	}
	return fmt.Errorf(
		"%s already in use; stop the occupant or choose a different port via the LINA_CORE_PORT/LINA_VBEN_PORT make variables",
		strings.Join(occupied, " and "),
	)
}

// StopCurrentProjectServiceProcessesForOccupiedPorts stops untracked
// current-repository service processes when their development ports are still
// occupied after PID-file-backed stop has run. Unknown or ambiguous matches are
// deliberately left alone so the final EnsurePortsAvailable diagnostic remains
// conservative.
func StopCurrentProjectServiceProcessesForOccupiedPorts(out io.Writer, root string, services []Config, probe PortInUseFunc, list ProcessListFunc, kill ProcessKillFunc) ([]int, error) {
	if probe == nil {
		probe = IsTCPListening
	}
	if list == nil || kill == nil {
		return nil, nil
	}

	occupied := make([]Config, 0, len(services))
	for _, service := range services {
		if probe(service.Port) {
			occupied = append(occupied, service)
		}
	}
	if len(occupied) == 0 {
		return nil, nil
	}

	infos, err := list()
	if err != nil {
		fmt.Fprintf(out, "warning: inspect port occupants failed: %v\n", err)
		return nil, nil
	}

	stoppedPorts := make([]int, 0, len(occupied))
	for _, service := range occupied {
		matches := matchingProjectServiceProcesses(root, service, infos)
		if len(matches) != 1 {
			continue
		}
		match := matches[0]
		if err = kill(match.PID); err != nil {
			return stoppedPorts, fmt.Errorf("stop %s process %d occupying port %d: %w", service.Name, match.PID, service.Port, err)
		}
		if _, err = fmt.Fprintf(out, "%s port %d is occupied by current project process %d; stopped it\n", service.Name, service.Port, match.PID); err != nil {
			return stoppedPorts, fmt.Errorf("write port occupant cleanup output: %w", err)
		}
		stoppedPorts = append(stoppedPorts, service.Port)
	}
	return stoppedPorts, nil
}

// WaitPortsReleased waits briefly for TCP ports to become free after managed
// development services were stopped. It deliberately does not report errors:
// the caller should run EnsurePortsAvailable afterwards so the final diagnostic
// remains the same for external occupants and slow process shutdowns.
func WaitPortsReleased(probe PortInUseFunc, timeout time.Duration, ports ...int) {
	deadline := time.Now().Add(timeout)
	for {
		if len(boundPorts(probe, ports...)) == 0 {
			return
		}
		if timeout <= 0 || time.Now().After(deadline) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// occupiedPorts returns printable labels for all bound development ports.
func occupiedPorts(probe PortInUseFunc, backendPort int, frontendPort int) []string {
	if probe == nil {
		probe = IsTCPListening
	}
	type portCheck struct {
		role string
		port int
	}
	checks := []portCheck{
		{role: "backend", port: backendPort},
		{role: "frontend", port: frontendPort},
	}
	var occupied []string
	for _, check := range checks {
		if probe(check.port) {
			occupied = append(occupied, fmt.Sprintf("%s port %d", check.role, check.port))
		}
	}
	return occupied
}

// boundPorts returns all supplied ports that currently accept TCP connections.
func boundPorts(probe PortInUseFunc, ports ...int) []int {
	if probe == nil {
		probe = IsTCPListening
	}
	occupied := make([]int, 0, len(ports))
	for _, port := range ports {
		if probe(port) {
			occupied = append(occupied, port)
		}
	}
	return occupied
}

func matchingProjectServiceProcesses(root string, service Config, infos []process.Info) []process.Info {
	matches := make([]process.Info, 0, 1)
	for _, info := range infos {
		if info.PID <= 1 || !processBelongsToRoot(root, info) {
			continue
		}
		switch service.Name {
		case "Backend":
			if matchesBackendService(root, info) {
				matches = append(matches, info)
			}
		case "Frontend":
			if matchesFrontendService(service, info) {
				matches = append(matches, info)
			}
		}
	}
	return matches
}

func processBelongsToRoot(root string, info process.Info) bool {
	if pathWithin(root, info.CWD) {
		return true
	}
	return strings.Contains(filepath.ToSlash(info.CommandLine()), filepath.ToSlash(filepath.Clean(root)))
}

func matchesBackendService(root string, info process.Info) bool {
	var (
		command              = filepath.ToSlash(info.CommandLine())
		backendBinary        = filepath.ToSlash(filepath.Join(root, "temp", "bin", "lina"))
		windowsBackendBinary = filepath.ToSlash(filepath.Join(root, "temp", "bin", "lina.exe"))
	)
	return strings.Contains(command, backendBinary) || strings.Contains(command, windowsBackendBinary)
}

func matchesFrontendService(service Config, info process.Info) bool {
	command := strings.ToLower(filepath.ToSlash(info.CommandLine()))
	if !strings.Contains(command, "vite") {
		return false
	}
	if hasExplicitDifferentPort(info.Args, service.Port) {
		return false
	}
	return pathWithin(service.WorkDir, info.CWD) || strings.Contains(command, "apps/lina-vben")
}

func hasExplicitDifferentPort(args []string, expected int) bool {
	expectedText := strconv.Itoa(expected)
	for i, arg := range args {
		if arg == "--port" && i+1 < len(args) {
			return args[i+1] != expectedText
		}
		if strings.HasPrefix(arg, "--port=") {
			return strings.TrimPrefix(arg, "--port=") != expectedText
		}
	}
	return false
}

func pathWithin(root string, path string) bool {
	if root == "" || path == "" {
		return false
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(filepath.Clean(rootAbs), filepath.Clean(pathAbs))
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// StartService starts a development service and records its PID file.
//
// After the child is started we spawn a goroutine that calls cmd.Wait() so
// the kernel can reap the child as soon as it exits. Without this the child
// becomes a zombie when it dies (because Setsid + manual Process.Release
// leaves no one to reap it inside the linactl process), and zombies still
// answer kill(2) signal 0 with success — which would make ProcessAlive lie.
// The detached child still outlives the linactl invocation thanks to Setsid;
// Wait only collects the exit status, it does not block the child from
// running.
// 启动后 spawn goroutine 调用 cmd.Wait() 让内核及时回收子进程，避免 zombie 状态
// 导致 ProcessAlive 误判存活；Setsid 已让子进程脱离会话，Wait 仅回收退出码、
// 不阻止子进程继续运行。
func StartService(root string, stdout io.Writer, stderr io.Writer, env []string, runner ProcessRunner, detach func(*exec.Cmd), service Config) error {
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

	cmd := runner(context.WithValue(context.Background(), RunnerContextServiceEnvKey, append([]string(nil), service.Env...)), service.StartName, service.StartArgs...)
	cmd.Dir = service.WorkDir
	cmd.Env = env
	for _, item := range service.Env {
		key, value, ok := strings.Cut(item, "=")
		if ok && strings.TrimSpace(key) != "" {
			cmd.Env = toolutil.SetEnvValue(cmd.Env, key, value)
		}
	}
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	detach(cmd)
	if err = cmd.Start(); err != nil {
		if closeErr := logFile.Close(); closeErr != nil {
			return fmt.Errorf("start %s failed and close log failed: %w: %w", service.Name, closeErr, err)
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
	go func() {
		// Reap the child to avoid zombies; ignore non-zero exit errors here
		// because StartService is a fire-and-forget launcher and the readiness
		// loop in WaitHTTP will surface the failure as soon as ProcessAlive
		// observes the PID is gone.
		// 后台 Wait 仅用于回收子进程；任何非零退出由 WaitHTTP 探活路径感知。
		if waitErr := cmd.Wait(); waitErr != nil {
			fmt.Fprintf(stderr, "warning: %s wait: %v\n", service.Name, waitErr)
		}
	}()
	fmt.Fprintf(stdout, "%s started: pid=%d log=%s\n", service.Name, pid, toolutil.RelativePath(root, service.LogPath))
	return nil
}

// StopService stops a PID-file-backed service when possible and reports
// whether it sent a kill signal to the recorded process.
func StopService(out io.Writer, service Config) (bool, error) {
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
		return stopped, fmt.Errorf("remove %s PID file: %w", service.Name, err)
	}
	if stopped {
		fmt.Fprintf(out, "%s stopped\n", service.Name)
		return true, nil
	}
	fmt.Fprintf(out, "%s is not running\n", service.Name)
	return false, nil
}

// WaitHTTP waits for one service URL to become ready.
//
// The probe enforces three independent checks per iteration:
//  1. PID file still exists (cleared by StopService or never written).
//  2. The recorded PID still belongs to a live process — caught via the
//     injected ProcessAliveFunc. This catches fatal startup errors (such as
//     "bind: address already in use") that would otherwise be invisible
//     because StartService detaches the child.
//  3. The HTTP endpoint responds with a status code below 500. The caller
//     selects the URL via ReadinessURL so an unrelated occupant cannot
//     accidentally satisfy the probe (e.g. /api.json for the backend).
//
// 探活循环同时校验 PID 文件、子进程存活与 HTTP 响应，避免子进程已 fatal 退出却被
// 端口占用方"假装就绪"的情况。
func WaitHTTP(name string, url string, pidPath string, logPath string, timeout time.Duration, alive ProcessAliveFunc) error {
	if alive == nil {
		alive = func(int) bool { return true }
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	deadline := time.Now().Add(timeout)
	client := newReadinessHTTPClient(2 * time.Second)
	for time.Now().Before(deadline) {
		pid := ReadPID(pidPath)
		if pid == 0 {
			return fmt.Errorf("%s startup failed: PID file does not exist; check log: %s", name, logPath)
		}
		if !alive(pid) {
			return fmt.Errorf("%s startup failed: process %d exited; check log: %s", name, pid, logPath)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("create %s readiness request: %w", name, err)
		}
		resp, err := client.Do(req)
		if err == nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				return fmt.Errorf("close %s readiness response: %w", name, closeErr)
			}
			if resp.StatusCode < http.StatusInternalServerError {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s startup timed out (%s): %s; check log: %s", name, timeout, url, logPath)
		case <-time.After(time.Second):
		}
	}
	return fmt.Errorf("%s startup timed out (%s): %s; check log: %s", name, timeout, url, logPath)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	dialer := net.Dialer{Timeout: time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
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

// TailLogToWriter copies the last N lines of a log file to the given writer
// so operators see the actual failure cause without opening another shell.
// 把日志末尾若干行回显到给定输出，便于失败时快速定位原因。
func TailLogToWriter(out io.Writer, name string, logPath string, lines int) error {
	if lines <= 0 {
		lines = readinessLogTailLines
	}
	file, err := os.Open(logPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("open %s log: %w", name, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(out, "warning: close %s log failed: %v\n", name, closeErr)
		}
	}()

	// Use a ring buffer to keep memory bounded regardless of log size.
	// 使用环形缓冲限制内存，无论日志多大都只保留末尾 N 行。
	buffer := make([]string, 0, lines)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		if len(buffer) < lines {
			buffer = append(buffer, scanner.Text())
			continue
		}
		copy(buffer, buffer[1:])
		buffer[lines-1] = scanner.Text()
	}
	if err = scanner.Err(); err != nil {
		return fmt.Errorf("read %s log: %w", name, err)
	}
	if len(buffer) == 0 {
		return nil
	}
	if _, err = fmt.Fprintf(out, "----- last %d lines of %s log (%s) -----\n", len(buffer), name, logPath); err != nil {
		return fmt.Errorf("write %s log header: %w", name, err)
	}
	for _, line := range buffer {
		if _, err = fmt.Fprintln(out, line); err != nil {
			return fmt.Errorf("write %s log line: %w", name, err)
		}
	}
	if _, err = fmt.Fprintln(out, "----- end of log tail -----"); err != nil {
		return fmt.Errorf("write %s log footer: %w", name, err)
	}
	return nil
}

// DefaultReadinessTailLines exposes the readiness log tail line count for
// callers that prefer to use the package default explicitly rather than
// passing 0 to TailLogToWriter.
// 暴露默认 tail 行数常量，便于调用方明确传值。
const DefaultReadinessTailLines = readinessLogTailLines
