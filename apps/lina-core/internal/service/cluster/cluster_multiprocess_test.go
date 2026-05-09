// This file verifies real multi-process cluster coordination against
// PostgreSQL when integration tests opt in.

package cluster

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"gopkg.in/yaml.v3"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	multiProcessTestTimeout       = 90 * time.Second
	multiProcessHealthWait        = 30 * time.Second
	multiProcessLeaderWait        = 15 * time.Second
	multiProcessElectionLease     = "2s"
	multiProcessElectionRenew     = "300ms"
	multiProcessSessionToken      = "cluster-multiprocess-session"
	multiProcessKVOwnerType       = "module"
	multiProcessKVOwnerKey        = "cluster-multiprocess"
	multiProcessKVNamespace       = "e2e"
	multiProcessKVKey             = "shared"
	multiProcessSentinelLockName  = "cluster-multiprocess-sentinel"
	multiProcessBackendBinaryName = "lina-multiprocess-test"
	multiProcessHealthModeMaster  = "master"
	multiProcessHealthModeSlave   = "slave"
)

// TestClusterTwoHostProcessesSharePostgreSQL verifies two real Lina host
// processes coordinate through one PostgreSQL database instead of process-local
// state. It is skipped unless LINA_TEST_PGSQL_LINK is explicitly provided.
func TestClusterTwoHostProcessesSharePostgreSQL(t *testing.T) {
	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL multi-process cluster test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), multiProcessTestTimeout)
	defer cancel()

	databaseName := fmt.Sprintf("linapro_cluster_mp_%d", time.Now().UnixNano())
	targetLink := postgresLinkWithDatabaseForClusterTest(t, baseLink, databaseName)
	if err := prepareClusterTestDatabase(ctx, targetLink, true); err != nil {
		t.Fatalf("prepare PostgreSQL database for multi-process cluster test failed: %v", err)
	}
	t.Cleanup(func() {
		dropClusterTestDatabase(t, targetLink)
	})

	configDirA := writeClusterProcessConfig(t, targetLink, freeTCPPort(t), "cluster-node-a")
	configDirB := writeClusterProcessConfig(t, targetLink, freeTCPPort(t), "cluster-node-b")
	if err := runHostCommand(ctx, configDirA, "init", "--confirm=init", "--sql-source=local"); err != nil {
		t.Fatalf("initialize multi-process database failed: %v", err)
	}
	if err := seedVolatileSentinels(ctx, targetLink); err != nil {
		t.Fatalf("seed volatile sentinel rows failed: %v", err)
	}

	binaryPath := buildClusterHostBinary(t)
	processA := startClusterHostProcess(t, ctx, binaryPath, configDirA, "cluster-node-a")
	processB := startClusterHostProcess(t, ctx, binaryPath, configDirB, "cluster-node-b")
	t.Cleanup(func() {
		processA.stop(t)
		processB.stop(t)
	})

	waitForProcessMode(t, processA, configDirA, multiProcessHealthModeMaster, multiProcessHealthModeSlave)
	waitForProcessMode(t, processB, configDirB, multiProcessHealthModeMaster, multiProcessHealthModeSlave)
	assertExactlyOneLeaderMode(t, configDirA, configDirB)
	assertVolatileSentinelsExist(t, targetLink)

	leader := processA
	followerConfig := configDirB
	if healthMode(t, configDirB) == multiProcessHealthModeMaster {
		leader = processB
		followerConfig = configDirA
	}
	leader.stop(t)
	waitForSpecificProcessMode(t, followerConfig, multiProcessHealthModeMaster)
	assertVolatileSentinelsExist(t, targetLink)
}

// clusterProcess wraps one started Lina host process and captured output.
type clusterProcess struct {
	cmd     *exec.Cmd
	cancel  context.CancelFunc
	output  *lockedOutputBuffer
	waitErr chan error
}

// lockedOutputBuffer serializes process stdout and stderr capture for race-safe
// failure diagnostics.
type lockedOutputBuffer struct {
	mu     sync.Mutex
	buffer bytes.Buffer
}

// WriteString appends one string to the captured output.
func (b *lockedOutputBuffer) WriteString(value string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buffer.WriteString(value)
}

// WriteByte appends one byte to the captured output.
func (b *lockedOutputBuffer) WriteByte(value byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buffer.WriteByte(value)
}

// String returns the captured output snapshot.
func (b *lockedOutputBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.String()
}

// stop terminates the process and waits for it to exit.
func (p *clusterProcess) stop(t *testing.T) {
	t.Helper()

	if p == nil || p.cmd == nil {
		return
	}
	if p.cancel != nil {
		p.cancel()
	}
	if p.cmd.Process != nil {
		if err := p.cmd.Process.Signal(os.Interrupt); err != nil && !strings.Contains(err.Error(), "process already finished") {
			t.Logf("interrupt cluster host process failed: %v", err)
		}
	}
	done := make(chan error, 1)
	if p.waitErr != nil {
		done = p.waitErr
	} else {
		go func() { done <- p.cmd.Wait() }()
	}
	select {
	case err := <-done:
		if err != nil && !isExpectedProcessStopError(err) {
			t.Logf("cluster host process exited with error: %v\n%s", err, p.output.String())
		}
	case <-time.After(5 * time.Second):
		if p.cmd.Process != nil {
			if err := p.cmd.Process.Kill(); err != nil && !strings.Contains(err.Error(), "process already finished") {
				t.Logf("kill cluster host process failed: %v", err)
			}
		}
		<-done
	}
}

// writeClusterProcessConfig writes a per-process config directory.
func writeClusterProcessConfig(t *testing.T, link string, port int, uploadSuffix string) string {
	t.Helper()

	repoRoot := repositoryRoot(t)
	content, err := os.ReadFile(filepath.Join(repoRoot, "apps", "lina-core", "manifest", "config", "config.yaml"))
	if err != nil {
		t.Fatalf("read base config failed: %v", err)
	}
	var config map[string]any
	if err = yaml.Unmarshal(content, &config); err != nil {
		t.Fatalf("parse base config yaml failed: %v", err)
	}

	section(config, "server")["address"] = fmt.Sprintf("127.0.0.1:%d", port)
	databaseDefault := section(section(config, "database"), "default")
	databaseDefault["link"] = link
	databaseDefault["debug"] = false
	section(config, "logger")["stdout"] = false
	section(config, "logger")["path"] = ""
	clusterCfg := section(config, "cluster")
	clusterCfg["enabled"] = true
	electionCfg := section(clusterCfg, "election")
	electionCfg["lease"] = multiProcessElectionLease
	electionCfg["renewInterval"] = multiProcessElectionRenew
	section(config, "upload")["path"] = filepath.Join(t.TempDir(), "upload-"+uploadSuffix)
	pluginCfg := section(config, "plugin")
	dynamicCfg := section(pluginCfg, "dynamic")
	dynamicCfg["storagePath"] = filepath.Join(t.TempDir(), "output-"+uploadSuffix)
	pluginCfg["autoEnable"] = []any{}

	encoded, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("marshal process config failed: %v", err)
	}
	configDir := t.TempDir()
	if err = os.WriteFile(filepath.Join(configDir, "config.yaml"), encoded, 0o600); err != nil {
		t.Fatalf("write process config failed: %v", err)
	}
	return configDir
}

// section returns a mutable YAML map section, creating it when absent.
func section(config map[string]any, key string) map[string]any {
	value, ok := config[key].(map[string]any)
	if !ok {
		value = map[string]any{}
		config[key] = value
	}
	return value
}

// runHostCommand runs the host binary command with a dedicated config path.
func runHostCommand(ctx context.Context, configDir string, args ...string) error {
	commandArgs := append([]string{"run", "main.go"}, args...)
	command := exec.CommandContext(ctx, "go", commandArgs...)
	command.Dir = repositoryRootFromCWD()
	command.Dir = filepath.Join(command.Dir, "apps", "lina-core")
	command.Env = append(os.Environ(), "GF_GCFG_PATH="+configDir)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go %s failed: %w\n%s", strings.Join(commandArgs, " "), err, string(output))
	}
	return nil
}

// seedVolatileSentinels inserts rows that must survive host process startup.
func seedVolatileSentinels(ctx context.Context, link string) (err error) {
	db, err := gdb.New(gdb.ConfigNode{Link: link})
	if err != nil {
		return fmt.Errorf("open PostgreSQL database failed: %w", err)
	}
	defer func() {
		if closeErr := db.Close(context.Background()); closeErr != nil && err == nil {
			err = fmt.Errorf("close PostgreSQL database failed: %w", closeErr)
		}
	}()

	now := gtime.Now()
	future := now.Add(time.Hour)
	if _, err = db.Exec(
		ctx,
		`INSERT INTO sys_online_session (token_id, user_id, username, dept_name, ip, browser, os, login_time, last_active_time)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT DO NOTHING`,
		multiProcessSessionToken,
		1,
		"admin",
		"Cluster",
		"127.0.0.1",
		"Chrome",
		"macOS",
		now,
		now,
	); err != nil {
		return fmt.Errorf("insert online session sentinel failed: %w", err)
	}
	if _, err = db.Exec(
		ctx,
		`INSERT INTO sys_locker (name, reason, holder, expire_time)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT DO NOTHING`,
		multiProcessSentinelLockName,
		"multi-process sentinel",
		"sentinel",
		future,
	); err != nil {
		return fmt.Errorf("insert locker sentinel failed: %w", err)
	}
	if _, err = db.Exec(
		ctx,
		`INSERT INTO sys_kv_cache (owner_type, owner_key, namespace, cache_key, value_kind, value_bytes, expire_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT DO NOTHING`,
		multiProcessKVOwnerType,
		multiProcessKVOwnerKey,
		multiProcessKVNamespace,
		multiProcessKVKey,
		1,
		[]byte("value"),
		future,
	); err != nil {
		return fmt.Errorf("insert kv cache sentinel failed: %w", err)
	}
	return nil
}

// assertVolatileSentinelsExist verifies startup did not clear volatile tables.
func assertVolatileSentinelsExist(t *testing.T, link string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := gdb.New(gdb.ConfigNode{Link: link})
	if err != nil {
		t.Fatalf("open PostgreSQL database failed: %v", err)
	}
	defer func() {
		if closeErr := db.Close(context.Background()); closeErr != nil {
			t.Errorf("close PostgreSQL database failed: %v", closeErr)
		}
	}()

	assertCount := func(label string, query string, args ...any) {
		t.Helper()
		value, countErr := db.GetValue(ctx, query, args...)
		if countErr != nil {
			t.Fatalf("count %s sentinel failed: %v", label, countErr)
		}
		count := value.Int()
		if count != 1 {
			t.Fatalf("expected %s sentinel to remain once, got %d", label, count)
		}
	}
	assertCount("online session", "SELECT COUNT(*) FROM sys_online_session WHERE token_id=$1", multiProcessSessionToken)
	assertCount("locker", "SELECT COUNT(*) FROM sys_locker WHERE name=$1", multiProcessSentinelLockName)
	assertCount(
		"kv cache",
		"SELECT COUNT(*) FROM sys_kv_cache WHERE owner_type=$1 AND owner_key=$2 AND namespace=$3 AND cache_key=$4",
		multiProcessKVOwnerType,
		multiProcessKVOwnerKey,
		multiProcessKVNamespace,
		multiProcessKVKey,
	)
}

// buildClusterHostBinary builds the host binary once for both process starts.
func buildClusterHostBinary(t *testing.T) string {
	t.Helper()

	binaryPath := filepath.Join(t.TempDir(), multiProcessBackendBinaryName)
	command := exec.Command("go", "build", "-o", binaryPath, ".")
	command.Dir = filepath.Join(repositoryRoot(t), "apps", "lina-core")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build cluster host binary failed: %v\n%s", err, string(output))
	}
	return binaryPath
}

// startClusterHostProcess starts one host process with unique node identity.
func startClusterHostProcess(
	t *testing.T,
	parentCtx context.Context,
	binaryPath string,
	configDir string,
	nodeID string,
) *clusterProcess {
	t.Helper()

	ctx, cancel := context.WithCancel(parentCtx)
	command := exec.CommandContext(ctx, binaryPath)
	command.Dir = filepath.Join(repositoryRoot(t), "apps", "lina-core")
	command.Env = append(os.Environ(), "GF_GCFG_PATH="+configDir, nodeIDEnvName+"="+nodeID)

	output := &lockedOutputBuffer{}
	stdout, err := command.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatalf("capture cluster process stdout failed: %v", err)
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		cancel()
		t.Fatalf("capture cluster process stderr failed: %v", err)
	}
	if err = command.Start(); err != nil {
		cancel()
		t.Fatalf("start cluster host process failed: %v", err)
	}
	go copyProcessOutput(output, stdout)
	go copyProcessOutput(output, stderr)

	process := &clusterProcess{
		cmd:     command,
		cancel:  cancel,
		output:  output,
		waitErr: make(chan error, 1),
	}
	go func() {
		process.waitErr <- command.Wait()
		close(process.waitErr)
	}()
	return process
}

// copyProcessOutput copies process output while preserving it for failure logs.
func copyProcessOutput(buffer *lockedOutputBuffer, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		buffer.WriteString(scanner.Text())
		buffer.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		buffer.WriteString(fmt.Sprintf("[process-output] scanner error: %v\n", err))
	}
}

// waitForProcessMode waits until one process reports any accepted cluster mode.
func waitForProcessMode(t *testing.T, process *clusterProcess, configDir string, acceptedModes ...string) {
	t.Helper()

	deadline := time.Now().Add(multiProcessHealthWait)
	for time.Now().Before(deadline) {
		if err, exited := processExitError(process); exited {
			t.Fatalf("process at %s exited before reporting modes %v: %v\n%s", healthURL(t, configDir), acceptedModes, err, process.output.String())
		}
		mode := healthMode(t, configDir)
		for _, accepted := range acceptedModes {
			if mode == accepted {
				return
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("process at %s did not report modes %v before timeout\n%s", healthURL(t, configDir), acceptedModes, process.output.String())
}

// assertExactlyOneLeaderMode verifies the two processes expose one master and
// one slave health response.
func assertExactlyOneLeaderMode(t *testing.T, configDirA string, configDirB string) {
	t.Helper()

	deadline := time.Now().Add(multiProcessLeaderWait)
	for time.Now().Before(deadline) {
		modeA := healthMode(t, configDirA)
		modeB := healthMode(t, configDirB)
		if (modeA == multiProcessHealthModeMaster && modeB == multiProcessHealthModeSlave) ||
			(modeA == multiProcessHealthModeSlave && modeB == multiProcessHealthModeMaster) {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("expected exactly one master, got %s=%s and %s=%s", healthURL(t, configDirA), healthMode(t, configDirA), healthURL(t, configDirB), healthMode(t, configDirB))
}

// waitForSpecificProcessMode waits until one process reports the expected mode.
func waitForSpecificProcessMode(t *testing.T, configDir string, expectedMode string) {
	t.Helper()

	deadline := time.Now().Add(multiProcessLeaderWait)
	for time.Now().Before(deadline) {
		if mode := healthMode(t, configDir); mode == expectedMode {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("process at %s did not report %s before timeout; last mode=%s", healthURL(t, configDir), expectedMode, healthMode(t, configDir))
}

// processExitError reports whether the started process already exited.
func processExitError(process *clusterProcess) (error, bool) {
	if process == nil || process.waitErr == nil {
		return nil, false
	}
	select {
	case err, ok := <-process.waitErr:
		if !ok {
			return nil, true
		}
		return err, true
	default:
		return nil, false
	}
}

// healthMode returns one process public health mode.
func healthMode(t *testing.T, configDir string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL(t, configDir), nil)
	if err != nil {
		t.Fatalf("build health request failed: %v", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return ""
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			t.Errorf("close health response failed: %v", closeErr)
		}
	}()
	if response.StatusCode != http.StatusOK {
		return ""
	}

	var envelope struct {
		Data struct {
			Mode string `json:"mode"`
		} `json:"data"`
	}
	if err = json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode health response failed: %v", err)
	}
	return envelope.Data.Mode
}

// healthURL builds the anonymous health endpoint URL for a config directory.
func healthURL(t *testing.T, configDir string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(configDir, "config.yaml"))
	if err != nil {
		t.Fatalf("read process config failed: %v", err)
	}
	var config struct {
		Server struct {
			Address string `yaml:"address"`
		} `yaml:"server"`
	}
	if err = yaml.Unmarshal(content, &config); err != nil {
		t.Fatalf("parse process config failed: %v", err)
	}
	return "http://" + strings.TrimPrefix(config.Server.Address, ":") + "/api/v1/health"
}

// freeTCPPort returns an available local TCP port.
func freeTCPPort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("open free tcp port failed: %v", err)
	}
	defer func() {
		if closeErr := listener.Close(); closeErr != nil {
			t.Errorf("close tcp listener failed: %v", closeErr)
		}
	}()
	return listener.Addr().(*net.TCPAddr).Port
}

// repositoryRoot returns the monorepo root.
func repositoryRoot(t *testing.T) string {
	t.Helper()

	root := repositoryRootFromCWD()
	if root == "" {
		t.Fatal("repository root not found")
	}
	return root
}

// repositoryRootFromCWD returns the monorepo root for subprocess helpers.
func repositoryRootFromCWD() string {
	workingDir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for dir := workingDir; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
		if _, err = os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}
	}
	return ""
}

// postgresLinkWithDatabaseForClusterTest returns link with its database name replaced.
func postgresLinkWithDatabaseForClusterTest(t *testing.T, link string, databaseName string) string {
	t.Helper()

	node, err := postgresConfigNodeFromLink(link)
	if err != nil {
		t.Fatalf("parse PostgreSQL link failed: %v", err)
	}
	extra := strings.TrimSpace(node.Extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return fmt.Sprintf("pgsql:%s:%s@%s(%s:%s)/%s%s", node.User, node.Pass, node.Protocol, node.Host, node.Port, databaseName, extra)
}

// dropClusterTestDatabase removes the temporary multi-process test database.
func dropClusterTestDatabase(t *testing.T, targetLink string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	node, err := postgresConfigNodeFromLink(targetLink)
	if err != nil {
		t.Errorf("parse PostgreSQL target link failed: %v", err)
		return
	}
	systemLink := postgresLinkWithDatabaseForClusterTest(t, targetLink, "postgres")
	db, err := gdb.New(gdb.ConfigNode{Link: systemLink})
	if err != nil {
		t.Errorf("open PostgreSQL system database failed: %v", err)
		return
	}
	defer func() {
		if closeErr := db.Close(context.Background()); closeErr != nil {
			t.Errorf("close PostgreSQL system database failed: %v", closeErr)
		}
	}()
	if _, err = db.Exec(ctx, "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid<>pg_backend_pid()", node.Name); err != nil {
		t.Errorf("terminate PostgreSQL target connections failed: %v", err)
		return
	}
	quotedName, err := quotePostgresIdentifierForClusterTest(node.Name)
	if err != nil {
		t.Errorf("quote PostgreSQL database name failed: %v", err)
		return
	}
	if _, err = db.Exec(ctx, "DROP DATABASE IF EXISTS "+quotedName); err != nil {
		t.Errorf("drop PostgreSQL target database failed: %v", err)
	}
}

// isExpectedProcessStopError reports whether a process exited due to test
// shutdown rather than an application startup failure.
func isExpectedProcessStopError(err error) bool {
	if err == nil {
		return true
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return true
	}
	return strings.Contains(err.Error(), "signal: interrupt") ||
		strings.Contains(err.Error(), "signal: killed") ||
		strings.Contains(err.Error(), "context canceled") ||
		strings.Contains(err.Error(), os.Interrupt.String()) ||
		strings.Contains(err.Error(), "exit status 130") ||
		strings.Contains(err.Error(), "exit status 143") ||
		strings.Contains(err.Error(), "use of closed network connection")
}

// prepareClusterTestDatabase creates or rebuilds one PostgreSQL database for
// multi-process testing without depending on a package-internal dialect helper.
func prepareClusterTestDatabase(ctx context.Context, link string, rebuild bool) (err error) {
	configNode, err := postgresConfigNodeFromLink(link)
	if err != nil {
		return err
	}
	quotedName, err := quotePostgresIdentifierForClusterTest(configNode.Name)
	if err != nil {
		return err
	}

	systemNode := *configNode
	systemNode.Link = ""
	systemNode.Name = "postgres"
	systemDB, err := gdb.New(systemNode)
	if err != nil {
		return gerror.Wrap(err, "connect PostgreSQL system database failed")
	}
	defer func() {
		if closeErr := systemDB.Close(ctx); closeErr != nil && err == nil {
			err = gerror.Wrap(closeErr, "close PostgreSQL system database failed")
		}
	}()

	existsValue, err := systemDB.GetValue(ctx, "SELECT 1 FROM pg_database WHERE datname=$1", configNode.Name)
	if err != nil {
		return gerror.Wrap(err, "check PostgreSQL database existence failed")
	}
	exists := !existsValue.IsNil()
	if rebuild && exists {
		if _, err = systemDB.Exec(
			ctx,
			"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid<>pg_backend_pid()",
			configNode.Name,
		); err != nil {
			return gerror.Wrap(err, "terminate PostgreSQL database connections failed")
		}
		if _, err = systemDB.Exec(ctx, "DROP DATABASE IF EXISTS "+quotedName); err != nil {
			return gerror.Wrap(err, "drop PostgreSQL database failed")
		}
		exists = false
	}
	if exists {
		return nil
	}
	if _, err = systemDB.Exec(ctx, "CREATE DATABASE "+quotedName+" ENCODING 'UTF8' LC_COLLATE 'C' LC_CTYPE 'C' TEMPLATE template0"); err != nil {
		return gerror.Wrap(err, "create PostgreSQL database failed")
	}
	return nil
}

// postgresConfigNodeFromLink parses a GoFrame PostgreSQL link and requires a
// target database name.
func postgresConfigNodeFromLink(link string) (*gdb.ConfigNode, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: link})
	if err != nil {
		return nil, gerror.Wrap(err, "parse PostgreSQL database link failed")
	}
	configNode := db.GetConfig()
	if configNode == nil {
		return nil, gerror.New("database link configuration is empty")
	}
	node := *configNode
	if strings.TrimSpace(node.Name) == "" {
		return nil, gerror.New("database name is missing from PostgreSQL database link")
	}
	return &node, nil
}

// quotePostgresIdentifierForClusterTest safely quotes one PostgreSQL
// identifier for CREATE/DROP DATABASE statements.
func quotePostgresIdentifierForClusterTest(identifier string) (string, error) {
	trimmed := strings.TrimSpace(identifier)
	if trimmed == "" {
		return "", gerror.New("PostgreSQL identifier must not be empty")
	}
	if strings.ContainsRune(trimmed, 0) {
		return "", gerror.New("PostgreSQL identifier must not contain NUL bytes")
	}
	return `"` + strings.ReplaceAll(trimmed, `"`, `""`) + `"`, nil
}
