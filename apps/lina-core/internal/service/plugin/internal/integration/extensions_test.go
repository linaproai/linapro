// This file covers declared hook execution helpers owned by the integration service.

package integration_test

import (
	"context"
	"strings"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/pluginhost"
)

func TestRunPluginDeclaredHookHonorsTimeoutAndErrorActions(t *testing.T) {
	services := testutil.NewServices()

	sleepHook := &catalog.HookSpec{
		Event:     pluginhost.ExtensionPointAuthLoginSucceeded,
		Action:    pluginhost.HookActionSleep,
		TimeoutMs: 10,
		SleepMs:   80,
	}
	timeoutCtx, cancel := integration.BuildPluginHookTimeoutContext(context.Background(), sleepHook)
	defer cancel()

	err := services.Integration.RunPluginDeclaredHook(timeoutCtx, "plugin-dynamic-timeout", sleepHook, nil)
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected timeout error for sleep hook, got: %v", err)
	}

	errorHook := &catalog.HookSpec{
		Event:        pluginhost.ExtensionPointAuthLoginSucceeded,
		Action:       pluginhost.HookActionError,
		ErrorMessage: "runtime hook failed on purpose",
	}
	err = services.Integration.RunPluginDeclaredHook(context.Background(), "plugin-dynamic-error", errorHook, nil)
	if err == nil || !strings.Contains(err.Error(), "runtime hook failed on purpose") {
		t.Fatalf("expected declared error hook message, got: %v", err)
	}
}
