// This file verifies project-wide GoFrame log handler configuration.

package logger

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

func TestConfigureEnablesJsonHandler(t *testing.T) {
	oldHandler := glog.GetDefaultHandler()
	t.Cleanup(func() {
		glog.SetDefaultHandler(oldHandler)
	})

	Configure(true)

	currentHandler := glog.GetDefaultHandler()
	if currentHandler == nil {
		t.Fatal("expected default handler to be configured")
	}
	if reflect.ValueOf(currentHandler).Pointer() != reflect.ValueOf(glog.HandlerJson).Pointer() {
		t.Fatal("expected default handler to be glog.HandlerJson")
	}
}

func TestConfigureDisablesCustomHandler(t *testing.T) {
	oldHandler := glog.GetDefaultHandler()
	t.Cleanup(func() {
		glog.SetDefaultHandler(oldHandler)
	})

	glog.SetDefaultHandler(glog.HandlerJson)
	Configure(false)

	if glog.GetDefaultHandler() != nil {
		t.Fatal("expected default handler to be cleared")
	}
}

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

func unsafeFieldString(value reflect.Value) string {
	return reflect.NewAt(value.Type(), unsafe.Pointer(value.UnsafeAddr())).Elem().String()
}
