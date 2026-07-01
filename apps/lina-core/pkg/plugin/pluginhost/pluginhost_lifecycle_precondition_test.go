// This file verifies lifecycle precondition callback aggregation behavior.

package pluginhost

import (
	"context"
	"errors"
	"testing"
	"time"
)

// lifecycleCallbackTestHook provides configurable lifecycle callback behavior.
type lifecycleCallbackTestHook struct {
	ok     bool
	reason string
	err    error
	delay  time.Duration
	panic  bool
}

func (h lifecycleCallbackTestHook) callback(ctx context.Context) (bool, string, error) {
	if h.delay > 0 {
		select {
		case <-time.After(h.delay):
		case <-ctx.Done():
			<-time.After(h.delay)
		}
	}
	if h.panic {
		panic("unit panic")
	}
	return h.ok, h.reason, h.err
}

func lifecycleTestCallbacks(hook LifecycleHook, h lifecycleCallbackTestHook) LifecycleCallbacks {
	var callbacks LifecycleCallbacks
	callbacks.set(hook, func(ctx context.Context, req LifecycleRequest) (bool, string, error) {
		return h.callback(ctx)
	})
	return callbacks
}

// TestRunLifecycleCallbacksAggregatesVetoes verifies all veto reasons are collected.
func TestRunLifecycleCallbacksAggregatesVetoes(t *testing.T) {
	result := RunLifecycleCallbacks(context.Background(), LifecycleRequest{
		Hook: LifecycleHookBeforeUninstall,
		Participants: []LifecycleParticipant{
			{
				PluginID:  "a",
				Callbacks: lifecycleTestCallbacks(LifecycleHookBeforeUninstall, lifecycleCallbackTestHook{ok: false, reason: "plugin.a.reason"}),
			},
			{
				PluginID:  "b",
				Callbacks: lifecycleTestCallbacks(LifecycleHookBeforeUninstall, lifecycleCallbackTestHook{ok: false, reason: "plugin.b.reason"}),
			},
		},
	})
	if result.OK || len(result.Decisions) != 2 {
		t.Fatalf("expected two vetoes, got %#v", result)
	}
}

// TestRunLifecycleCallbacksTreatsTimeoutAsVeto verifies slow hooks fail closed.
func TestRunLifecycleCallbacksTreatsTimeoutAsVeto(t *testing.T) {
	result := RunLifecycleCallbacks(context.Background(), LifecycleRequest{
		Hook:        LifecycleHookBeforeUninstall,
		HookTimeout: time.Millisecond,
		Participants: []LifecycleParticipant{
			{
				PluginID:  "slow",
				Callbacks: lifecycleTestCallbacks(LifecycleHookBeforeUninstall, lifecycleCallbackTestHook{ok: true, delay: 5 * time.Millisecond}),
			},
		},
	})
	if result.OK || len(result.Decisions) != 1 || !result.Decisions[0].TimedOut {
		t.Fatalf("expected timeout veto, got %#v", result)
	}
}

// TestRunLifecycleCallbacksRecoversPanic verifies panics are converted to veto decisions.
func TestRunLifecycleCallbacksRecoversPanic(t *testing.T) {
	result := RunLifecycleCallbacks(context.Background(), LifecycleRequest{
		Hook: LifecycleHookBeforeUninstall,
		Participants: []LifecycleParticipant{
			{
				PluginID:  "panic",
				Callbacks: lifecycleTestCallbacks(LifecycleHookBeforeUninstall, lifecycleCallbackTestHook{ok: true, panic: true}),
			},
		},
	})
	if result.OK || len(result.Decisions) != 1 || !result.Decisions[0].Panicked {
		t.Fatalf("expected panic veto, got %#v", result)
	}
}

// TestRunLifecycleCallbacksTreatsErrorsAsVeto verifies hook errors fail closed.
func TestRunLifecycleCallbacksTreatsErrorsAsVeto(t *testing.T) {
	hookErr := errors.New("hook failed")
	result := RunLifecycleCallbacks(context.Background(), LifecycleRequest{
		Hook: LifecycleHookBeforeUninstall,
		Participants: []LifecycleParticipant{
			{
				PluginID:  "err",
				Callbacks: lifecycleTestCallbacks(LifecycleHookBeforeUninstall, lifecycleCallbackTestHook{ok: true, err: hookErr}),
			},
		},
	})
	if result.OK || len(result.Decisions) != 1 || !errors.Is(result.Decisions[0].Err, hookErr) {
		t.Fatalf("expected error veto, got %#v", result)
	}
}

// TestRunLifecycleCallbacksRunsTenantDelete verifies tenant lifecycle callbacks
// are first-class lifecycle preconditions.
func TestRunLifecycleCallbacksRunsTenantDelete(t *testing.T) {
	result := RunLifecycleCallbacks(context.Background(), LifecycleRequest{
		Hook:        LifecycleHookBeforeTenantDelete,
		TenantInput: NewSourcePluginTenantLifecycleInput(LifecycleHookBeforeTenantDelete.String(), 1001),
		Participants: []LifecycleParticipant{
			{
				PluginID: "tenant-aware",
				Callbacks: lifecycleTestCallbacks(
					LifecycleHookBeforeTenantDelete,
					lifecycleCallbackTestHook{ok: false, reason: "plugin.tenant.delete.blocked"},
				),
			},
		},
	})
	if result.OK || len(result.Decisions) != 1 || result.Decisions[0].Reason != "plugin.tenant.delete.blocked" {
		t.Fatalf("expected tenant delete lifecycle callback to veto, got %#v", result)
	}
}

// TestRunLifecycleCallbacksRunsInstallModeChange verifies install-mode changes
// are first-class lifecycle preconditions.
func TestRunLifecycleCallbacksRunsInstallModeChange(t *testing.T) {
	result := RunLifecycleCallbacks(context.Background(), LifecycleRequest{
		Hook: LifecycleHookBeforeInstallModeChange,
		InstallModeInput: NewSourcePluginInstallModeChangeInput(
			"install-mode-aware",
			LifecycleHookBeforeInstallModeChange.String(),
			"global",
			"tenant_scoped",
		),
		Participants: []LifecycleParticipant{
			{
				PluginID: "install-mode-aware",
				Callbacks: lifecycleTestCallbacks(
					LifecycleHookBeforeInstallModeChange,
					lifecycleCallbackTestHook{ok: false, reason: "plugin.install.mode.blocked"},
				),
			},
		},
	})
	if result.OK || len(result.Decisions) != 1 || result.Decisions[0].Reason != "plugin.install.mode.blocked" {
		t.Fatalf("expected install-mode lifecycle callback to veto, got %#v", result)
	}
}
