// This file configures the project-wide GoFrame log handler used by the host.

package logger

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

// defaultFilePattern is the fallback rolling file pattern used when the
// project config omits an explicit shared log filename.
const defaultFilePattern = "{Y-m-d}.log"

// traceIDEnabledStore keeps the process-wide TraceID output switch loaded from
// the static logger configuration file at startup.
var traceIDEnabledStore atomic.Bool

// ServerOutputConfig defines the shared output target for business and server logs.
type ServerOutputConfig struct {
	Path   string // Path is the target directory for log files.
	File   string // File is the shared rolling file pattern.
	Stdout bool   // Stdout controls whether logs are also written to stdout.
}

// RuntimeConfig defines project-wide logger handler behavior.
type RuntimeConfig struct {
	Structured     bool // Structured controls JSON log formatting.
	TraceIDEnabled bool // TraceIDEnabled controls whether log output keeps TraceID fields.
}

// Configure applies the project-wide GoFrame log handler with the current
// structured-format and TraceID visibility switches loaded from config.yaml.
func Configure(cfg RuntimeConfig) {
	setTraceIDEnabled(cfg.TraceIDEnabled)
	glog.SetDefaultHandler(newDefaultHandler(cfg.Structured))
}

// setTraceIDEnabled stores the process-wide TraceID output switch for later
// handler invocations.
func setTraceIDEnabled(enabled bool) {
	traceIDEnabledStore.Store(enabled)
}

// traceIDEnabled reports whether log output should retain TraceID fields.
func traceIDEnabled() bool {
	return traceIDEnabledStore.Load()
}

// newDefaultHandler creates the package-level default log handler while still
// allowing LinaPro to suppress TraceID output based on static startup config.
func newDefaultHandler(structured bool) glog.Handler {
	if structured {
		return structuredTraceIDAwareHandler
	}
	return textTraceIDAwareHandler
}

// structuredTraceIDAwareHandler renders JSON logs after applying the
// configured TraceID visibility switch.
func structuredTraceIDAwareHandler(ctx context.Context, in *glog.HandlerInput) {
	stripTraceIDIfDisabled(in)
	glog.HandlerJson(ctx, in)
}

// textTraceIDAwareHandler renders the default text log format after applying
// the configured TraceID visibility switch.
func textTraceIDAwareHandler(ctx context.Context, in *glog.HandlerInput) {
	stripTraceIDIfDisabled(in)
	// HandlerInput.String already appends the trailing newline for text logs.
	in.Buffer.WriteString(in.String())
	in.Next(ctx)
}

// stripTraceIDIfDisabled clears the handler input TraceID when the configured
// logger switch is disabled.
func stripTraceIDIfDisabled(in *glog.HandlerInput) {
	if in == nil || traceIDEnabled() {
		return
	}
	in.TraceId = ""
}

// BindServer configures the HTTP server to reuse the shared business logger
// output target so both server and application logs end up in the same sink.
func BindServer(server *ghttp.Server, cfg ServerOutputConfig) error {
	if server == nil {
		return gerror.New("http server is nil")
	}

	server.SetLogger(Logger().Clone())
	return server.SetConfigWithMap(map[string]any{
		"logPath":          strings.TrimSpace(cfg.Path),
		"logStdout":        cfg.Stdout,
		"accessLogPattern": normalizedFilePattern(cfg.File),
		"errorLogPattern":  normalizedFilePattern(cfg.File),
	})
}

// normalizedFilePattern trims one configured rolling pattern and falls back to
// the project default when the input is empty.
func normalizedFilePattern(pattern string) string {
	if strings.TrimSpace(pattern) == "" {
		return defaultFilePattern
	}
	return strings.TrimSpace(pattern)
}
