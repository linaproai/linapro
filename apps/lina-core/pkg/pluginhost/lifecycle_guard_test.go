// This file verifies lifecycle guard aggregation behavior.

package pluginhost

import (
	"context"
	"errors"
	"testing"
	"time"
)

// lifecycleGuardTestHook provides configurable CanUninstall behavior.
type lifecycleGuardTestHook struct {
	ok     bool
	reason string
	err    error
	delay  time.Duration
	panic  bool
}

// CanUninstall returns the configured test hook result.
func (h lifecycleGuardTestHook) CanUninstall(ctx context.Context) (bool, string, error) {
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

// TestRunLifecycleGuardsAggregatesVetoes verifies all veto reasons are collected.
func TestRunLifecycleGuardsAggregatesVetoes(t *testing.T) {
	result := RunLifecycleGuards(context.Background(), GuardRequest{
		Hook: GuardHookCanUninstall,
		Participants: []GuardParticipant{
			{PluginID: "a", Guard: lifecycleGuardTestHook{ok: false, reason: "plugin.a.reason"}},
			{PluginID: "b", Guard: lifecycleGuardTestHook{ok: false, reason: "plugin.b.reason"}},
		},
	})
	if result.OK || len(result.Decisions) != 2 {
		t.Fatalf("expected two vetoes, got %#v", result)
	}
}

// TestRunLifecycleGuardsTreatsTimeoutAsVeto verifies slow hooks fail closed.
func TestRunLifecycleGuardsTreatsTimeoutAsVeto(t *testing.T) {
	result := RunLifecycleGuards(context.Background(), GuardRequest{
		Hook:        GuardHookCanUninstall,
		HookTimeout: time.Millisecond,
		Participants: []GuardParticipant{
			{PluginID: "slow", Guard: lifecycleGuardTestHook{ok: true, delay: 5 * time.Millisecond}},
		},
	})
	if result.OK || len(result.Decisions) != 1 || !result.Decisions[0].TimedOut {
		t.Fatalf("expected timeout veto, got %#v", result)
	}
}

// TestRunLifecycleGuardsRecoversPanic verifies panics are converted to veto decisions.
func TestRunLifecycleGuardsRecoversPanic(t *testing.T) {
	result := RunLifecycleGuards(context.Background(), GuardRequest{
		Hook: GuardHookCanUninstall,
		Participants: []GuardParticipant{
			{PluginID: "panic", Guard: lifecycleGuardTestHook{ok: true, panic: true}},
		},
	})
	if result.OK || len(result.Decisions) != 1 || !result.Decisions[0].Panicked {
		t.Fatalf("expected panic veto, got %#v", result)
	}
}

// TestRunLifecycleGuardsTreatsErrorsAsVeto verifies hook errors fail closed.
func TestRunLifecycleGuardsTreatsErrorsAsVeto(t *testing.T) {
	hookErr := errors.New("hook failed")
	result := RunLifecycleGuards(context.Background(), GuardRequest{
		Hook: GuardHookCanUninstall,
		Participants: []GuardParticipant{
			{PluginID: "err", Guard: lifecycleGuardTestHook{ok: true, err: hookErr}},
		},
	})
	if result.OK || len(result.Decisions) != 1 || !errors.Is(result.Decisions[0].Err, hookErr) {
		t.Fatalf("expected error veto, got %#v", result)
	}
}
