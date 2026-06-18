// This file verifies HTTP startup observability and ordering internals.

package httpstartup

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/util/guid"

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

// TestHTTPStartupRegistersSourceRoutesBeforeConsistencyValidation protects the
// startup ordering required by source plugins that register host capability
// providers from HTTP route callbacks.
func TestHTTPStartupRegistersSourceRoutesBeforeConsistencyValidation(t *testing.T) {
	content, err := os.ReadFile("http.go")
	if err != nil {
		t.Fatalf("read HTTP startup source: %v", err)
	}
	text := string(content)
	beforeRoutesIndex := strings.Index(text, "startHTTPRuntimeBeforeSourceRoutes")
	registerRoutesIndex := strings.Index(text, "registerSourcePluginHTTPRoutes")
	finishRuntimeIndex := strings.Index(text, "finishHTTPRuntimeAfterSourceRoutes")
	completeRoutesIndex := strings.Index(text, "completeSourcePluginHTTPRoutes")
	if beforeRoutesIndex < 0 || registerRoutesIndex < 0 || finishRuntimeIndex < 0 || completeRoutesIndex < 0 {
		t.Fatalf("expected split HTTP startup phases to be present")
	}
	if !(beforeRoutesIndex < registerRoutesIndex &&
		registerRoutesIndex < finishRuntimeIndex &&
		finishRuntimeIndex < completeRoutesIndex) {
		t.Fatalf(
			"expected startup order start-before-routes -> register-source-routes -> finish-runtime -> complete-source-routes, got indexes %d %d %d %d",
			beforeRoutesIndex,
			registerRoutesIndex,
			finishRuntimeIndex,
			completeRoutesIndex,
		)
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
