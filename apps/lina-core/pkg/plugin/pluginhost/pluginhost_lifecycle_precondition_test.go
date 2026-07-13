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

// TestGlobalLifecycleHookForTargetMapsBeforeHooks verifies target→global mapping.
func TestGlobalLifecycleHookForTargetMapsBeforeHooks(t *testing.T) {
	cases := []struct {
		target LifecycleHook
		global LifecycleHook
	}{
		{LifecycleHookBeforeInstall, LifecycleHookGlobalBeforeInstall},
		{LifecycleHookBeforeEnable, LifecycleHookGlobalBeforeEnable},
		{LifecycleHookBeforeDisable, LifecycleHookGlobalBeforeDisable},
		{LifecycleHookBeforeUninstall, LifecycleHookGlobalBeforeUninstall},
	}
	for _, tc := range cases {
		got, ok := GlobalLifecycleHookForTarget(tc.target)
		if !ok || got != tc.global {
			t.Fatalf("target %s: got %s ok=%v", tc.target, got, ok)
		}
		if !IsGlobalLifecycleHook(got) {
			t.Fatalf("expected %s to be a global hook", got)
		}
	}
	if _, ok := GlobalLifecycleHookForTarget(LifecycleHookAfterInstall); ok {
		t.Fatal("AfterInstall must not map to a global hook")
	}
}

// TestListSourcePluginGlobalLifecycleParticipantsOnlyExplicit verifies only
// plugins that registered a global hook are listed.
func TestListSourcePluginGlobalLifecycleParticipantsOnlyExplicit(t *testing.T) {
	owner := NewDeclarations("test-global-owner")
	if err := owner.Lifecycle().RegisterGlobalBeforeEnableHandler(func(
		_ context.Context,
		_ SourcePluginGlobalLifecycleInput,
	) (bool, string, error) {
		return false, "plugin.mail.transport.kind_conflict", nil
	}); err != nil {
		t.Fatalf("register global handler: %v", err)
	}
	observer := NewDeclarations("test-global-observer")
	cleanupOwner, err := RegisterSourcePluginForTest(owner)
	if err != nil {
		t.Fatalf("register owner: %v", err)
	}
	t.Cleanup(cleanupOwner)
	cleanupObserver, err := RegisterSourcePluginForTest(observer)
	if err != nil {
		t.Fatalf("register observer: %v", err)
	}
	t.Cleanup(cleanupObserver)

	participants := ListSourcePluginGlobalLifecycleParticipants(LifecycleHookGlobalBeforeEnable)
	if len(participants) != 1 || participants[0].PluginID != owner.ID() {
		t.Fatalf("expected only owner participant, got %#v", participants)
	}

	result := RunLifecycleCallbacks(context.Background(), LifecycleRequest{
		Hook: LifecycleHookGlobalBeforeEnable,
		GlobalInput: NewSourcePluginGlobalLifecycleInput(
			"target-smtp-b",
			LifecycleHookGlobalBeforeEnable.String(),
		),
		Participants: participants,
	})
	if result.OK || len(result.Decisions) != 1 {
		t.Fatalf("expected global veto, got %#v", result)
	}
	if result.Decisions[0].Reason != "plugin.mail.transport.kind_conflict" {
		t.Fatalf("unexpected reason: %#v", result.Decisions[0])
	}
}

// TestBeforeEnableAndGlobalInputIsolation verifies self BeforeEnable uses PluginInput
// while global enable uses GlobalInput target identity.
func TestBeforeEnableAndGlobalInputIsolation(t *testing.T) {
	plugin := NewDeclarations("test-enable-isolation")
	var selfPluginID string
	var globalTargetID string
	if err := plugin.Lifecycle().RegisterBeforeEnableHandler(func(
		_ context.Context,
		input SourcePluginLifecycleInput,
	) (bool, string, error) {
		selfPluginID = input.PluginID()
		return true, "", nil
	}); err != nil {
		t.Fatalf("register before enable: %v", err)
	}
	if err := plugin.Lifecycle().RegisterGlobalBeforeEnableHandler(func(
		_ context.Context,
		input SourcePluginGlobalLifecycleInput,
	) (bool, string, error) {
		globalTargetID = input.TargetPluginID()
		return true, "", nil
	}); err != nil {
		t.Fatalf("register global before enable: %v", err)
	}

	targetCallbacks := NewSourcePluginLifecycleCallbackAdapter(plugin.(SourcePluginDefinition))
	targetResult := RunLifecycleCallbacks(context.Background(), LifecycleRequest{
		Hook:        LifecycleHookBeforeEnable,
		PluginInput: NewSourcePluginLifecycleInput(plugin.ID(), LifecycleHookBeforeEnable.String()),
		Participants: []LifecycleParticipant{{
			PluginID:  plugin.ID(),
			Callbacks: targetCallbacks,
		}},
	})
	if !targetResult.OK {
		t.Fatalf("target before enable failed: %#v", targetResult)
	}
	if selfPluginID != plugin.ID() {
		t.Fatalf("self before enable saw plugin id %q", selfPluginID)
	}

	globalCallbacks := NewSourcePluginGlobalLifecycleCallbackAdapter(
		plugin.(SourcePluginDefinition),
		LifecycleHookGlobalBeforeEnable,
	)
	globalResult := RunLifecycleCallbacks(context.Background(), LifecycleRequest{
		Hook: LifecycleHookGlobalBeforeEnable,
		GlobalInput: NewSourcePluginGlobalLifecycleInput(
			"other-plugin",
			LifecycleHookGlobalBeforeEnable.String(),
		),
		Participants: []LifecycleParticipant{{
			PluginID:  plugin.ID(),
			Callbacks: globalCallbacks,
		}},
	})
	if !globalResult.OK {
		t.Fatalf("global before enable failed: %#v", globalResult)
	}
	if globalTargetID != "other-plugin" {
		t.Fatalf("global handler saw target %q", globalTargetID)
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
