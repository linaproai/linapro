// Package logger provides the project-wide logging wrapper used by the Lina core host service and host-side extensions.
package logger

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

var (
	logger   *glog.Logger
	initOnce = sync.Once{}
)

// Logger creates and returns a logger for logging.
func Logger() *glog.Logger {
	initOnce.Do(func() {
		logger = g.Log()
		logger.SetStackSkip(1)
		logger.SetFlags(glog.F_TIME_STD | glog.F_FILE_SHORT)
	})
	return logger
}

// Print prints <v> with newline using fmt.Sprintln.
// The parameter <v> can be multiple variables.
func Print(ctx context.Context, v ...interface{}) {
	Logger().Print(ctx, v...)
}

// Printf prints <v> with format <format> using fmt.Sprintf.
// The parameter <v> can be multiple variables.
func Printf(ctx context.Context, format string, v ...interface{}) {
	Logger().Printf(ctx, format, v...)
}

// Fatal prints the logging content with [FATA] header and newline, then exit the current process.
func Fatal(ctx context.Context, v ...interface{}) {
	Logger().Fatal(ctx, v...)
}

// Fatalf prints the logging content with [FATA] header, custom format and newline, then exit the current process.
func Fatalf(ctx context.Context, format string, v ...interface{}) {
	Logger().Fatalf(ctx, format, v...)
}

// Panic prints the logging content with [PANI] header and newline, then panics.
func Panic(ctx context.Context, v ...interface{}) {
	Logger().Panic(ctx, v...)
}

// Panicf prints the logging content with [PANI] header, custom format and newline, then panics.
func Panicf(ctx context.Context, format string, v ...interface{}) {
	Logger().Panicf(ctx, format, v...)
}

// Info prints the logging content with [INFO] header and newline.
func Info(ctx context.Context, v ...interface{}) {
	Logger().Info(ctx, v...)
}

// Infof prints the logging content with [INFO] header, custom format and newline.
func Infof(ctx context.Context, format string, v ...interface{}) {
	Logger().Infof(ctx, format, v...)
}

// Debug prints the logging content with [DEBU] header and newline.
func Debug(ctx context.Context, v ...interface{}) {
	Logger().Debug(ctx, v...)
}

// Debugf prints the logging content with [DEBU] header, custom format and newline.
func Debugf(ctx context.Context, format string, v ...interface{}) {
	Logger().Debugf(ctx, format, v...)
}

// Notice prints the logging content with [NOTI] header and newline.
// It also prints caller stack info if stack feature is enabled.
func Notice(ctx context.Context, v ...interface{}) {
	Logger().Notice(ctx, v...)
}

// Noticef prints the logging content with [NOTI] header, custom format and newline.
// It also prints caller stack info if stack feature is enabled.
func Noticef(ctx context.Context, format string, v ...interface{}) {
	Logger().Noticef(ctx, format, v...)
}

// Warning prints the logging content with [WARN] header and newline.
// It also prints caller stack info if stack feature is enabled.
func Warning(ctx context.Context, v ...interface{}) {
	Logger().Warning(ctx, v...)
}

// Warningf prints the logging content with [WARN] header, custom format and newline.
// It also prints caller stack info if stack feature is enabled.
func Warningf(ctx context.Context, format string, v ...interface{}) {
	Logger().Warningf(ctx, format, v...)
}

// Error prints the logging content with [ERRO] header and newline.
// It also prints caller stack info if stack feature is enabled.
func Error(ctx context.Context, v ...interface{}) {
	Logger().Error(ctx, v...)
}

// Errorf prints the logging content with [ERRO] header, custom format and newline.
// It also prints caller stack info if stack feature is enabled.
func Errorf(ctx context.Context, format string, v ...interface{}) {
	Logger().Errorf(ctx, format, v...)
}

// Critical prints the logging content with [CRIT] header and newline.
// It also prints caller stack info if stack feature is enabled.
func Critical(ctx context.Context, v ...interface{}) {
	Logger().Critical(ctx, v...)
}

// Criticalf prints the logging content with [CRIT] header, custom format and newline.
// It also prints caller stack info if stack feature is enabled.
func Criticalf(ctx context.Context, format string, v ...interface{}) {
	Logger().Criticalf(ctx, format, v...)
}
