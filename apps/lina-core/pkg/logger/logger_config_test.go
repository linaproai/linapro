// This file verifies project-wide GoFrame log handler configuration.

package logger

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

// TestConfigureInstallsCustomHandler verifies Configure always installs the
// LinaPro wrapper handler instead of relying on GoFrame's raw defaults.
func TestConfigureInstallsCustomHandler(t *testing.T) {
	oldHandler := glog.GetDefaultHandler()
	previousEnabled := traceIDEnabledStore.Load()
	t.Cleanup(func() {
		glog.SetDefaultHandler(oldHandler)
		setTraceIDEnabled(previousEnabled)
	})

	Configure(RuntimeConfig{
		Structured:     true,
		TraceIDEnabled: true,
	})

	currentHandler := glog.GetDefaultHandler()
	if currentHandler == nil {
		t.Fatal("expected default handler to be configured")
	}
	if reflect.ValueOf(currentHandler).Pointer() == reflect.ValueOf(glog.HandlerJson).Pointer() {
		t.Fatal("expected Configure to install the LinaPro wrapper handler")
	}
}

// TestStructuredTraceIDAwareHandlerDropsTraceID verifies structured logs omit
// TraceID when the startup switch keeps the feature disabled.
func TestStructuredTraceIDAwareHandlerDropsTraceID(t *testing.T) {
	withTraceIDEnabled(t, false)

	input := &glog.HandlerInput{
		Logger:      glog.New(),
		Buffer:      bytes.NewBuffer(nil),
		TimeFormat:  "2026-04-22 10:00:00",
		LevelFormat: "INFO",
		TraceId:     "trace-1",
		Content:     "hello",
	}

	newDefaultHandler(true)(context.Background(), input)

	output := input.Buffer.String()
	if strings.Contains(output, "trace-1") {
		t.Fatalf("expected structured output without TraceID, got %q", output)
	}
	if !strings.Contains(output, "hello") {
		t.Fatalf("expected structured output to contain content, got %q", output)
	}
}

// TestTextTraceIDAwareHandlerKeepsTraceID verifies plain-text logs retain
// TraceID when the startup switch enables the feature.
func TestTextTraceIDAwareHandlerKeepsTraceID(t *testing.T) {
	withTraceIDEnabled(t, true)

	input := &glog.HandlerInput{
		Logger:      glog.New(),
		Buffer:      bytes.NewBuffer(nil),
		TimeFormat:  "2026-04-22 10:00:00",
		LevelFormat: "INFO",
		TraceId:     "trace-1",
		Content:     "hello",
	}

	newDefaultHandler(false)(context.Background(), input)

	output := input.Buffer.String()
	if !strings.Contains(output, "{trace-1}") {
		t.Fatalf("expected plain-text output to keep TraceID, got %q", output)
	}
	if !strings.Contains(output, "hello") {
		t.Fatalf("expected plain-text output to contain content, got %q", output)
	}
}

// TestTextTraceIDAwareHandlerAvoidsExtraBlankLine verifies plain-text logs keep
// GoFrame's single trailing newline instead of appending a second blank line.
func TestTextTraceIDAwareHandlerAvoidsExtraBlankLine(t *testing.T) {
	withTraceIDEnabled(t, false)

	input := &glog.HandlerInput{
		Logger:      glog.New(),
		Buffer:      bytes.NewBuffer(nil),
		TimeFormat:  "2026-04-22 10:00:00",
		LevelFormat: "INFO",
		Content:     "hello",
	}

	newDefaultHandler(false)(context.Background(), input)

	output := input.Buffer.String()
	if strings.Count(output, "\n") != 1 {
		t.Fatalf("expected plain-text output to end with a single newline, got %q", output)
	}
	if strings.Contains(output, "\n\n") {
		t.Fatalf("expected plain-text output without blank lines, got %q", output)
	}
}

// TestBindServerAlignsSharedOutput verifies the HTTP server inherits the
// shared logger output directory and rolling file pattern.
func TestBindServerAlignsSharedOutput(t *testing.T) {
	server := ghttp.GetServer(fmt.Sprintf("logger-bind-%s", t.Name()))
	tempDir := t.TempDir()

	err := BindServer(server, ServerOutputConfig{
		Path:   tempDir,
		File:   "lina-{Y-m-d}.log",
		Stdout: false,
	})
	if err != nil {
		t.Fatalf("bind server logger: %v", err)
	}

	if server.GetLogPath() != tempDir {
		t.Fatalf("expected server log path %q, got %q", tempDir, server.GetLogPath())
	}
	if server.Logger() == nil {
		t.Fatal("expected server logger to be configured")
	}
	if server.Logger().GetPath() != tempDir {
		t.Fatalf("expected server logger path %q, got %q", tempDir, server.Logger().GetPath())
	}

	configValue := reflect.ValueOf(server).Elem().FieldByName("config")
	accessPattern := unsafeFieldString(configValue.FieldByName("AccessLogPattern"))
	errorPattern := unsafeFieldString(configValue.FieldByName("ErrorLogPattern"))
	if accessPattern != "lina-{Y-m-d}.log" {
		t.Fatalf("expected access log pattern to match shared file, got %q", accessPattern)
	}
	if errorPattern != "lina-{Y-m-d}.log" {
		t.Fatalf("expected error log pattern to match shared file, got %q", errorPattern)
	}
}

// TestBindServerUsesDefaultPatternWhenFileMissing verifies empty file patterns
// fall back to the project default rolling filename.
func TestBindServerUsesDefaultPatternWhenFileMissing(t *testing.T) {
	server := ghttp.GetServer(fmt.Sprintf("logger-bind-default-%s", t.Name()))

	err := BindServer(server, ServerOutputConfig{
		Stdout: true,
	})
	if err != nil {
		t.Fatalf("bind server logger: %v", err)
	}

	configValue := reflect.ValueOf(server).Elem().FieldByName("config")
	accessPattern := unsafeFieldString(configValue.FieldByName("AccessLogPattern"))
	errorPattern := unsafeFieldString(configValue.FieldByName("ErrorLogPattern"))
	if accessPattern != defaultFilePattern {
		t.Fatalf("expected default access log pattern %q, got %q", defaultFilePattern, accessPattern)
	}
	if errorPattern != defaultFilePattern {
		t.Fatalf("expected default error log pattern %q, got %q", defaultFilePattern, errorPattern)
	}
}

// unsafeFieldString reads one unexported string field from a reflected struct
// value for test assertions against GoFrame server config internals.
func unsafeFieldString(value reflect.Value) string {
	return reflect.NewAt(value.Type(), unsafe.Pointer(value.UnsafeAddr())).Elem().String()
}

// withTraceIDEnabled swaps the package-level TraceID switch for one test.
func withTraceIDEnabled(t *testing.T, enabled bool) {
	t.Helper()

	previousEnabled := traceIDEnabledStore.Load()
	setTraceIDEnabled(enabled)
	t.Cleanup(func() {
		setTraceIDEnabled(previousEnabled)
	})
}
