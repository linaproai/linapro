// This file verifies project-wide GoFrame log handler configuration.

package logger

import (
	"reflect"
	"testing"

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
