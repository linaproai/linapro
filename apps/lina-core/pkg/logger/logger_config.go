// This file configures the project-wide GoFrame log handler used by the host.

package logger

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

// defaultFilePattern is the fallback rolling file pattern used when the
// project config omits an explicit shared log filename.
const defaultFilePattern = "{Y-m-d}.log"

// ServerOutputConfig defines the shared output target for business and server logs.
type ServerOutputConfig struct {
	Path   string // Path is the target directory for log files.
	File   string // File is the shared rolling file pattern.
	Stdout bool   // Stdout controls whether logs are also written to stdout.
}

// Configure applies the project-wide default GoFrame log handler.
func Configure(structured bool) {
	if structured {
		glog.SetDefaultHandler(glog.HandlerJson)
		return
	}
	glog.SetDefaultHandler(nil)
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
