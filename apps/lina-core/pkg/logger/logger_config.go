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

// defaultTraceIDEnabledResolver keeps TraceID output disabled unless one
// caller explicitly wires a runtime-effective enable switch.
func defaultTraceIDEnabledResolver(context.Context) bool {
	return false
}

// traceIDEnabledResolverStore keeps the runtime-effective TraceID resolver for
// the package-level default logging handler.
var traceIDEnabledResolverStore atomic.Value

func init() {
	traceIDEnabledResolverStore.Store(TraceIDEnabledResolver(defaultTraceIDEnabledResolver))
}

// ServerOutputConfig defines the shared output target for business and server logs.
type ServerOutputConfig struct {
	Path   string // Path is the target directory for log files.
	File   string // File is the shared rolling file pattern.
	Stdout bool   // Stdout controls whether logs are also written to stdout.
}

// TraceIDEnabledResolver reports whether one log entry should include TraceID.
type TraceIDEnabledResolver func(ctx context.Context) bool

// RuntimeConfig defines project-wide logger handler behavior.
type RuntimeConfig struct {
	Structured             bool                   // Structured controls JSON log formatting.
	TraceIDEnabledResolver TraceIDEnabledResolver // TraceIDEnabledResolver returns the runtime-effective TraceID switch.
}

// Configure applies the project-wide GoFrame log handler with the current
// structured-format and TraceID visibility switches.
func Configure(cfg RuntimeConfig) {
	setTraceIDEnabledResolver(cfg.TraceIDEnabledResolver)
	glog.SetDefaultHandler(newDefaultHandler(cfg.Structured))
}

// setTraceIDEnabledResolver stores one runtime-effective TraceID resolver for
// later handler invocations.
func setTraceIDEnabledResolver(resolver TraceIDEnabledResolver) {
	if resolver == nil {
		resolver = defaultTraceIDEnabledResolver
	}
	traceIDEnabledResolverStore.Store(resolver)
}

// traceIDEnabled reports whether the current log entry should retain its
// TraceID field.
func traceIDEnabled(ctx context.Context) bool {
	resolver, ok := traceIDEnabledResolverStore.Load().(TraceIDEnabledResolver)
	if !ok || resolver == nil {
		return false
	}
	return resolver(ctx)
}

// newDefaultHandler creates the package-level default log handler while still
// allowing LinaPro to suppress TraceID output dynamically.
func newDefaultHandler(structured bool) glog.Handler {
	if structured {
		return structuredTraceIDAwareHandler
	}
	return textTraceIDAwareHandler
}

// structuredTraceIDAwareHandler renders JSON logs after applying the
// runtime-effective TraceID visibility switch.
func structuredTraceIDAwareHandler(ctx context.Context, in *glog.HandlerInput) {
	stripTraceIDIfDisabled(ctx, in)
	glog.HandlerJson(ctx, in)
}

// textTraceIDAwareHandler renders the default text log format after applying
// the runtime-effective TraceID visibility switch.
func textTraceIDAwareHandler(ctx context.Context, in *glog.HandlerInput) {
	stripTraceIDIfDisabled(ctx, in)
	in.Buffer.WriteString(in.String())
	in.Buffer.WriteByte('\n')
	in.Next(ctx)
}

// stripTraceIDIfDisabled clears the handler input TraceID when the current
// runtime-effective logger switch is disabled.
func stripTraceIDIfDisabled(ctx context.Context, in *glog.HandlerInput) {
	if in == nil || traceIDEnabled(ctx) {
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
