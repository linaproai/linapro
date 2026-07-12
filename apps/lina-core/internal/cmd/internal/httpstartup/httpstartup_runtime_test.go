// This file verifies HTTP runtime construction, backend selection, and lifecycle internals.

package httpstartup

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/startupstats"
	"lina-core/pkg/logger"
)

// TestLogHTTPStartupSummaryEmitsFieldsWithoutSQL verifies startup observability
// uses an aggregate summary instead of ORM SQL text.
func TestLogHTTPStartupSummaryEmitsFieldsWithoutSQL(t *testing.T) {
	ctx := context.Background()
	collector := startupstats.New()
	collector.Add(startupstats.CounterCatalogSnapshotBuilds, 1)
	collector.Add(startupstats.CounterIntegrationSnapshotBuilds, 1)
	collector.Add(startupstats.CounterJobSnapshotBuilds, 1)
	collector.Add(startupstats.CounterPluginScans, 1)
	collector.Add(startupstats.CounterPluginSyncChanged, 2)
	collector.Add(startupstats.CounterPluginSyncNoop, 3)
	collector.RecordPhase(startupstats.PhasePluginBootstrapBuiltin, 9)
	collector.RecordPhase(startupstats.PhasePluginBootstrapAutoEnable, 12)
	collector.RecordPhase(startupstats.PhasePluginStartupConsistency, 4)

	var logs []string
	logger.Logger().SetHandlers(func(ctx context.Context, in *glog.HandlerInput) {
		logs = append(logs, in.ValuesContent())
	})
	t.Cleanup(func() {
		logger.Logger().SetHandlers()
	})

	logHTTPStartupSummary(ctx, collector)

	joined := strings.Join(logs, "\n")
	for _, expected := range []string{
		"startup summary",
		"catalogSnapshots=1",
		"integrationSnapshots=1",
		"jobSnapshots=1",
		"pluginScans=1",
		"pluginChanged=2",
		"pluginNoop=3",
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected startup summary to contain %q, got %q", expected, joined)
		}
	}
	for _, forbidden := range []string{"SHOW FULL COLUMNS", "SELECT ", "INSERT INTO", "UPDATE ", "DELETE "} {
		if strings.Contains(strings.ToUpper(joined), forbidden) {
			t.Fatalf("expected startup summary to omit SQL text %q, got %q", forbidden, joined)
		}
	}
}

// TestResolveRuntimeShutdownTimeoutUsesGoFrameServerConfig verifies host-owned
// cleanup reuses the GoFrame server graceful shutdown timeout.
func TestResolveRuntimeShutdownTimeoutUsesGoFrameServerConfig(t *testing.T) {
	server := g.Server("runtime-shutdown-timeout-" + guid.S())
	server.SetGracefulShutdownTimeout(17)

	timeout := resolveRuntimeShutdownTimeout(server)
	if timeout != 17*time.Second {
		t.Fatalf("expected runtime shutdown timeout to reuse GoFrame server config, got %s", timeout)
	}
}

// TestStartHTTPPluginManagementListPrewarmLogsDebugDuration verifies startup
// prewarming records elapsed time on the debug path for both outcomes.
func TestStartHTTPPluginManagementListPrewarmLogsDebugDuration(t *testing.T) {
	capture := newLogCapture(t)

	testCases := []struct {
		name   string
		err    error
		status string
	}{
		{
			name:   "success",
			status: "succeeded",
		},
		{
			name:   "failure",
			err:    gerror.New("prewarm failed"),
			status: "failed",
		},
	}

	for _, testCase := range testCases {
		capture.Reset()

		startHTTPPluginManagementListPrewarm(
			context.Background(),
			&prewarmLoggingPluginService{managementListErr: testCase.err},
		)

		joined := capture.WaitFor(
			t,
			"prewarm plugin management list finished status="+testCase.status,
		)
		if !strings.Contains(joined, "duration=") {
			t.Fatalf("expected prewarm debug log to include duration, got %q", joined)
		}
	}
}

// TestPrewarmHTTPRuntimeFrontendBundlesLogsDebugDuration verifies synchronous
// startup frontend prewarming records elapsed time on the debug path.
func TestPrewarmHTTPRuntimeFrontendBundlesLogsDebugDuration(t *testing.T) {
	capture := newLogCapture(t)
	testCases := []struct {
		name   string
		err    error
		status string
	}{
		{
			name:   "success",
			status: "succeeded",
		},
		{
			name:   "failure",
			err:    gerror.New("prewarm failed"),
			status: "failed",
		},
	}

	for _, testCase := range testCases {
		capture.Reset()

		prewarmHTTPRuntimeFrontendBundles(
			context.Background(),
			&prewarmLoggingPluginService{frontendBundlesErr: testCase.err},
		)

		joined := capture.Joined()
		expected := "prewarm runtime frontend bundles finished status=" + testCase.status
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected captured log to contain %q, got %q", expected, joined)
		}
		if !strings.Contains(joined, "duration=") {
			t.Fatalf("expected prewarm debug log to include duration, got %q", joined)
		}
	}
}

// TestValidateHTTPStartupPluginConsistencyReturnsErrorOnInvalidState verifies
// startup consistency failures return an error so the top-level startup
// orchestrator can log and exit cleanly instead of crashing the process.
func TestValidateHTTPStartupPluginConsistencyReturnsErrorOnInvalidState(t *testing.T) {
	ctx := startupstats.WithCollector(context.Background(), startupstats.New())
	pluginSvc := &startupConsistencyFailingPluginService{err: gerror.New("invalid startup state")}

	err := validateHTTPStartupPluginConsistency(ctx, pluginSvc)
	if err == nil {
		t.Fatal("expected startup consistency failure to return an error, got nil")
	}
	if !pluginSvc.called {
		t.Fatal("expected startup consistency validator to be called")
	}
	snapshot := startupstats.FromContext(ctx).Snapshot()
	if _, ok := snapshot.Phases[startupstats.PhasePluginStartupConsistency]; !ok {
		t.Fatalf("expected startup consistency phase to be recorded, got %#v", snapshot.Phases)
	}
}

// TestNewHTTPKVCacheProviderSelectsMemoryForSingleNode verifies single-node
// startup explicitly wires the in-process memory provider.
func TestNewHTTPKVCacheProviderSelectsMemoryForSingleNode(t *testing.T) {
	provider, err := newHTTPKVCacheProvider(&config.ClusterConfig{Enabled: false}, nil)
	if err != nil {
		t.Fatalf("select single-node kvcache provider: %v", err)
	}
	service := kvcache.New(kvcache.WithProvider(provider))
	if service.BackendName() != kvcache.BackendMemory {
		t.Fatalf("expected memory backend, got %q", service.BackendName())
	}
	if service.RequiresExpiredCleanup() {
		t.Fatal("expected memory backend to skip expired cleanup")
	}
}

// TestNewHTTPKVCacheProviderSelectsCoordinationForCluster verifies cluster
// startup explicitly wires the coordination KV provider.
func TestNewHTTPKVCacheProviderSelectsCoordinationForCluster(t *testing.T) {
	provider, err := newHTTPKVCacheProvider(&config.ClusterConfig{Enabled: true}, coordination.NewMemory(nil))
	if err != nil {
		t.Fatalf("select cluster kvcache provider: %v", err)
	}
	service := kvcache.New(kvcache.WithProvider(provider))
	if service.BackendName() != kvcache.BackendCoordinationKV {
		t.Fatalf("expected coordination KV backend, got %q", service.BackendName())
	}
	if service.RequiresExpiredCleanup() {
		t.Fatal("expected coordination KV backend to skip expired cleanup")
	}
}

// TestNewHTTPKVCacheProviderRejectsClusterWithoutCoordination verifies cluster
// startup does not silently fall back to memory or process defaults.
func TestNewHTTPKVCacheProviderRejectsClusterWithoutCoordination(t *testing.T) {
	if _, err := newHTTPKVCacheProvider(&config.ClusterConfig{Enabled: true}, nil); err == nil {
		t.Fatal("expected cluster kvcache provider selection to require coordination service")
	}
}

// newLogCapture configures the project logger for one test and captures log
// content while restoring global logger state during cleanup.
func newLogCapture(t *testing.T) *logCapture {
	t.Helper()

	projectLogger := logger.Logger()
	previousLevel := projectLogger.GetLevel()
	projectLogger.SetLevel(glog.LEVEL_ALL)

	capture := &logCapture{}
	projectLogger.SetHandlers(func(ctx context.Context, in *glog.HandlerInput) {
		capture.logsMu.Lock()
		defer capture.logsMu.Unlock()
		capture.logs = append(capture.logs, in.ValuesContent())
	})
	t.Cleanup(func() {
		projectLogger.SetHandlers()
		projectLogger.SetLevel(previousLevel)
	})
	return capture
}

// startupConsistencyFailingPluginService is a narrow fake for startup runtime tests.
type startupConsistencyFailingPluginService struct {
	called bool
	err    error
}

// ValidateStartupConsistency records the startup validation call and returns the configured error.
func (s *startupConsistencyFailingPluginService) ValidateStartupConsistency(context.Context) error {
	s.called = true
	return s.err
}

// prewarmLoggingPluginService is a narrow fake for startup prewarm logging tests.
type prewarmLoggingPluginService struct {
	managementListErr  error
	frontendBundlesErr error
}

// PrewarmManagementList returns the configured result for startup logging tests.
func (s *prewarmLoggingPluginService) PrewarmManagementList(context.Context) error {
	return s.managementListErr
}

// PrewarmRuntimeFrontendBundles returns the configured result for logging tests.
func (s *prewarmLoggingPluginService) PrewarmRuntimeFrontendBundles(context.Context) error {
	return s.frontendBundlesErr
}

// logCapture stores project logger output for one test.
type logCapture struct {
	logs   []string
	logsMu sync.Mutex
}

// Reset clears previously captured log output.
func (c *logCapture) Reset() {
	c.logsMu.Lock()
	defer c.logsMu.Unlock()
	c.logs = nil
}

// Joined returns all currently captured log output.
func (c *logCapture) Joined() string {
	c.logsMu.Lock()
	defer c.logsMu.Unlock()
	return strings.Join(c.logs, "\n")
}

// WaitFor waits until the asynchronous startup prewarm goroutine emits one
// expected log line, then returns all captured log content.
func (c *logCapture) WaitFor(t *testing.T, substring string) string {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for {
		joined := c.Joined()
		if strings.Contains(joined, substring) {
			return joined
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected captured log to contain %q, got %q", substring, joined)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestValidateListenAddressAcceptsCommonForms(t *testing.T) {
	t.Parallel()
	cases := []string{
		":9120",
		"0.0.0.0:9120",
		"127.0.0.1:18080",
		"[::1]:9120",
		":1",
		":65535",
		"127.0.0.1:9120,0.0.0.0:9121",
	}
	for _, address := range cases {
		address := address
		t.Run(address, func(t *testing.T) {
			t.Parallel()
			if err := validateListenAddress(address); err != nil {
				t.Fatalf("expected %q valid, got %v", address, err)
			}
		})
	}
}

func TestValidateListenAddressRejectsInvalidForms(t *testing.T) {
	t.Parallel()
	cases := []string{
		"",
		"   ",
		"not-an-address",
		"9120",
		":0",
		":65536",
		"127.0.0.1:",
		"127.0.0.1:abc",
		",:9120",
		":9120,",
	}
	for _, address := range cases {
		address := address
		t.Run(address, func(t *testing.T) {
			t.Parallel()
			if err := validateListenAddress(address); err == nil {
				t.Fatalf("expected %q invalid", address)
			}
		})
	}
}

func TestResolveServerAddressOverrideEmptyUsesConfig(t *testing.T) {
	t.Setenv(serverAddressEnvName, "")
	address, overridden, err := resolveServerAddressOverride()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if overridden || address != "" {
		t.Fatalf("expected no override, got address=%q overridden=%v", address, overridden)
	}
}

func TestResolveServerAddressOverrideWhitespaceUsesConfig(t *testing.T) {
	t.Setenv(serverAddressEnvName, "   ")
	address, overridden, err := resolveServerAddressOverride()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if overridden || address != "" {
		t.Fatalf("expected no override, got address=%q overridden=%v", address, overridden)
	}
}

func TestResolveServerAddressOverrideValid(t *testing.T) {
	t.Setenv(serverAddressEnvName, "127.0.0.1:18080")
	address, overridden, err := resolveServerAddressOverride()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !overridden || address != "127.0.0.1:18080" {
		t.Fatalf("expected override 127.0.0.1:18080, got address=%q overridden=%v", address, overridden)
	}
}

func TestResolveServerAddressOverrideInvalid(t *testing.T) {
	t.Setenv(serverAddressEnvName, "bad-address")
	_, overridden, err := resolveServerAddressOverride()
	if err == nil {
		t.Fatal("expected error for invalid address")
	}
	if overridden {
		t.Fatal("invalid address must not report overridden")
	}
}
