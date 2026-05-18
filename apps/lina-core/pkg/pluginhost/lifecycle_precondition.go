// This file defines lifecycle callback contracts used by the plugin host
// before and after it mutates plugin or tenant state.

package pluginhost

import (
	"context"
	"runtime/debug"
	"sync"
	"time"
)

// Lifecycle precondition callback default timeouts.
const (
	// DefaultLifecycleHookTimeout is the per-callback timeout.
	DefaultLifecycleHookTimeout = 5 * time.Second
	// DefaultLifecycleTotalTimeout is the total callback aggregation timeout.
	DefaultLifecycleTotalTimeout = 10 * time.Second
)

// LifecycleHook identifies one lifecycle callback.
type LifecycleHook string

// String returns the lifecycle hook name.
func (hook LifecycleHook) String() string {
	return string(hook)
}

// Lifecycle hook constants.
const (
	// LifecycleHookBeforeInstall protects plugin install.
	LifecycleHookBeforeInstall LifecycleHook = "BeforeInstall"
	// LifecycleHookAfterInstall observes successful plugin install.
	LifecycleHookAfterInstall LifecycleHook = "AfterInstall"
	// LifecycleHookBeforeUpgrade protects plugin runtime upgrade.
	LifecycleHookBeforeUpgrade LifecycleHook = "BeforeUpgrade"
	// LifecycleHookUpgrade performs plugin-owned runtime upgrade work.
	LifecycleHookUpgrade LifecycleHook = "Upgrade"
	// LifecycleHookAfterUpgrade observes successful plugin runtime upgrade.
	LifecycleHookAfterUpgrade LifecycleHook = "AfterUpgrade"
	// LifecycleHookBeforeDisable protects global plugin disable.
	LifecycleHookBeforeDisable LifecycleHook = "BeforeDisable"
	// LifecycleHookAfterDisable observes successful global plugin disable.
	LifecycleHookAfterDisable LifecycleHook = "AfterDisable"
	// LifecycleHookBeforeUninstall protects plugin uninstall.
	LifecycleHookBeforeUninstall LifecycleHook = "BeforeUninstall"
	// LifecycleHookUninstall performs plugin-owned uninstall cleanup work.
	LifecycleHookUninstall LifecycleHook = "Uninstall"
	// LifecycleHookAfterUninstall observes successful plugin uninstall.
	LifecycleHookAfterUninstall LifecycleHook = "AfterUninstall"
	// LifecycleHookBeforeTenantDisable protects tenant-scoped plugin disable.
	LifecycleHookBeforeTenantDisable LifecycleHook = "BeforeTenantDisable"
	// LifecycleHookAfterTenantDisable observes successful tenant-scoped plugin disable.
	LifecycleHookAfterTenantDisable LifecycleHook = "AfterTenantDisable"
	// LifecycleHookBeforeTenantDelete protects tenant deletion.
	LifecycleHookBeforeTenantDelete LifecycleHook = "BeforeTenantDelete"
	// LifecycleHookAfterTenantDelete observes successful tenant deletion.
	LifecycleHookAfterTenantDelete LifecycleHook = "AfterTenantDelete"
	// LifecycleHookBeforeInstallModeChange protects install-mode changes.
	LifecycleHookBeforeInstallModeChange LifecycleHook = "BeforeInstallModeChange"
	// LifecycleHookAfterInstallModeChange observes successful install-mode changes.
	LifecycleHookAfterInstallModeChange LifecycleHook = "AfterInstallModeChange"
)

// BeforeInstaller optionally vetoes plugin install.
type BeforeInstaller interface {
	// BeforeInstall returns whether install may continue and an i18n reason when vetoed.
	BeforeInstall(ctx context.Context, input SourcePluginLifecycleInput) (ok bool, reason string, err error)
}

// AfterInstaller optionally observes successful plugin install.
type AfterInstaller interface {
	// AfterInstall performs best-effort follow-up after install succeeds.
	AfterInstall(ctx context.Context, input SourcePluginLifecycleInput) error
}

// BeforeUpgrader optionally vetoes plugin runtime upgrade.
type BeforeUpgrader interface {
	// BeforeUpgrade returns whether upgrade may continue and an i18n reason when vetoed.
	BeforeUpgrade(ctx context.Context, input SourcePluginUpgradeInput) (ok bool, reason string, err error)
}

// AfterUpgrader optionally observes successful plugin runtime upgrade.
type AfterUpgrader interface {
	// AfterUpgrade performs best-effort follow-up after upgrade succeeds.
	AfterUpgrade(ctx context.Context, input SourcePluginUpgradeInput) error
}

// Upgrader optionally performs plugin-owned upgrade work.
type Upgrader interface {
	// Upgrade performs plugin-owned upgrade work before host upgrade SQL runs.
	Upgrade(ctx context.Context, input SourcePluginUpgradeInput) error
}

// BeforeDisabler optionally vetoes global plugin disable.
type BeforeDisabler interface {
	// BeforeDisable returns whether disable may continue and an i18n reason when vetoed.
	BeforeDisable(ctx context.Context, input SourcePluginLifecycleInput) (ok bool, reason string, err error)
}

// AfterDisabler optionally observes successful global plugin disable.
type AfterDisabler interface {
	// AfterDisable performs best-effort follow-up after disable succeeds.
	AfterDisable(ctx context.Context, input SourcePluginLifecycleInput) error
}

// BeforeUninstaller optionally vetoes plugin uninstall.
type BeforeUninstaller interface {
	// BeforeUninstall returns whether uninstall may continue and an i18n reason when vetoed.
	BeforeUninstall(ctx context.Context, input SourcePluginLifecycleInput) (ok bool, reason string, err error)
}

// AfterUninstaller optionally observes successful plugin uninstall.
type AfterUninstaller interface {
	// AfterUninstall performs best-effort follow-up after uninstall succeeds.
	AfterUninstall(ctx context.Context, input SourcePluginLifecycleInput) error
}

// Uninstaller optionally performs plugin-owned uninstall cleanup work.
type Uninstaller interface {
	// Uninstall performs plugin-owned cleanup before host uninstall SQL runs.
	Uninstall(ctx context.Context, input SourcePluginUninstallInput) error
}

// BeforeTenantDisabler optionally vetoes tenant-scoped plugin disable.
type BeforeTenantDisabler interface {
	// BeforeTenantDisable returns whether tenant disable may continue and an i18n reason when vetoed.
	BeforeTenantDisable(ctx context.Context, input SourcePluginTenantLifecycleInput) (ok bool, reason string, err error)
}

// AfterTenantDisabler optionally observes successful tenant-scoped plugin disable.
type AfterTenantDisabler interface {
	// AfterTenantDisable performs best-effort follow-up after tenant disable succeeds.
	AfterTenantDisable(ctx context.Context, input SourcePluginTenantLifecycleInput) error
}

// BeforeTenantDeleter optionally vetoes tenant deletion.
type BeforeTenantDeleter interface {
	// BeforeTenantDelete returns whether tenant deletion may continue and an i18n reason when vetoed.
	BeforeTenantDelete(ctx context.Context, input SourcePluginTenantLifecycleInput) (ok bool, reason string, err error)
}

// AfterTenantDeleter optionally observes successful tenant deletion.
type AfterTenantDeleter interface {
	// AfterTenantDelete performs best-effort follow-up after tenant deletion succeeds.
	AfterTenantDelete(ctx context.Context, input SourcePluginTenantLifecycleInput) error
}

// BeforeInstallModeChanger optionally vetoes plugin install-mode changes.
type BeforeInstallModeChanger interface {
	// BeforeInstallModeChange returns whether the install-mode transition may continue.
	BeforeInstallModeChange(
		ctx context.Context,
		input SourcePluginInstallModeChangeInput,
	) (ok bool, reason string, err error)
}

// AfterInstallModeChanger optionally observes successful plugin install-mode changes.
type AfterInstallModeChanger interface {
	// AfterInstallModeChange performs best-effort follow-up after the install-mode transition succeeds.
	AfterInstallModeChange(
		ctx context.Context,
		input SourcePluginInstallModeChangeInput,
	) error
}

// LifecycleParticipant binds a plugin ID to optional lifecycle callbacks.
type LifecycleParticipant struct {
	PluginID string // PluginID is the callback owner.
	Callback any    // Callback is the optional lifecycle callback implementation.
}

// ListSourcePluginLifecycleParticipants returns callback participants for all
// registered source plugins.
func ListSourcePluginLifecycleParticipants() []LifecycleParticipant {
	plugins := ListSourcePlugins()
	items := make([]LifecycleParticipant, 0, len(plugins))
	for _, plugin := range plugins {
		if plugin == nil || plugin.ID() == "" {
			continue
		}
		callback := NewSourcePluginLifecycleCallbackAdapter(plugin)
		if callback == nil {
			continue
		}
		items = append(items, LifecycleParticipant{
			PluginID: plugin.ID(),
			Callback: callback,
		})
	}
	return items
}

// ListSourcePluginLifecycleParticipantsForPlugin returns callback participants
// for the source plugin that owns the requested lifecycle action.
func ListSourcePluginLifecycleParticipantsForPlugin(pluginID string) []LifecycleParticipant {
	plugin, ok := GetSourcePlugin(pluginID)
	if !ok || plugin == nil {
		return nil
	}
	callback := NewSourcePluginLifecycleCallbackAdapter(plugin)
	if callback == nil {
		return nil
	}
	return []LifecycleParticipant{{
		PluginID: plugin.ID(),
		Callback: callback,
	}}
}

// LifecycleRequest describes one lifecycle callback aggregation run.
type LifecycleRequest struct {
	Hook             LifecycleHook // Hook selects which callback interface to invoke.
	PluginInput      SourcePluginLifecycleInput
	UninstallInput   SourcePluginUninstallInput
	UpgradeInput     SourcePluginUpgradeInput
	TenantInput      SourcePluginTenantLifecycleInput
	InstallModeInput SourcePluginInstallModeChangeInput
	Participants     []LifecycleParticipant // Participants are invoked concurrently.
	HookTimeout      time.Duration          // HookTimeout overrides the per-callback timeout.
	TotalTimeout     time.Duration          // TotalTimeout overrides the aggregate timeout.
}

// LifecycleDecision is one plugin callback result.
type LifecycleDecision struct {
	PluginID  string        // PluginID is the callback owner.
	Hook      LifecycleHook // Hook is the invoked hook.
	OK        bool          // OK reports whether this plugin allowed the action.
	Reason    string        // Reason is the i18n key when OK is false.
	Err       error         // Err records a hook error.
	Elapsed   time.Duration // Elapsed is the hook runtime.
	TimedOut  bool          // TimedOut reports per-hook timeout.
	Panicked  bool          // Panicked reports panic recovery.
	PanicText string        // PanicText records the recovered panic value.
	Stack     string        // Stack records the panic stack for logging.
}

// LifecycleResult is the aggregate lifecycle callback result.
type LifecycleResult struct {
	OK        bool                // OK reports whether all callbacks allowed the action.
	Decisions []LifecycleDecision // Decisions contains one result per applicable participant.
}

// RunLifecycleCallbacks invokes all applicable lifecycle callbacks concurrently.
func RunLifecycleCallbacks(ctx context.Context, req LifecycleRequest) LifecycleResult {
	req = normalizeLifecycleRequest(req)
	ctx, cancel := context.WithTimeout(ctx, req.TotalTimeout)
	defer cancel()

	results := make(chan LifecycleDecision, len(req.Participants))
	var wg sync.WaitGroup
	for _, participant := range req.Participants {
		if !participantSupportsLifecycleHook(participant.Callback, req.Hook) {
			continue
		}
		wg.Add(1)
		go func(item LifecycleParticipant) {
			defer wg.Done()
			results <- runOneLifecycleCallback(ctx, req, item)
		}(participant)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	aggregate := LifecycleResult{OK: true}
	for decision := range results {
		if !decision.OK {
			aggregate.OK = false
		}
		aggregate.Decisions = append(aggregate.Decisions, decision)
	}
	return aggregate
}

// normalizeLifecycleRequest fills lifecycle callback timeout defaults.
func normalizeLifecycleRequest(req LifecycleRequest) LifecycleRequest {
	if req.HookTimeout <= 0 {
		req.HookTimeout = DefaultLifecycleHookTimeout
	}
	if req.TotalTimeout <= 0 {
		req.TotalTimeout = DefaultLifecycleTotalTimeout
	}
	return req
}

// participantSupportsLifecycleHook reports whether one participant implements
// the requested hook.
func participantSupportsLifecycleHook(callback any, hook LifecycleHook) bool {
	switch hook {
	case LifecycleHookBeforeInstall:
		_, ok := callback.(BeforeInstaller)
		return ok
	case LifecycleHookAfterInstall:
		_, ok := callback.(AfterInstaller)
		return ok
	case LifecycleHookBeforeUpgrade:
		_, ok := callback.(BeforeUpgrader)
		return ok
	case LifecycleHookUpgrade:
		_, ok := callback.(Upgrader)
		return ok
	case LifecycleHookAfterUpgrade:
		_, ok := callback.(AfterUpgrader)
		return ok
	case LifecycleHookBeforeDisable:
		_, ok := callback.(BeforeDisabler)
		return ok
	case LifecycleHookAfterDisable:
		_, ok := callback.(AfterDisabler)
		return ok
	case LifecycleHookBeforeUninstall:
		_, ok := callback.(BeforeUninstaller)
		return ok
	case LifecycleHookUninstall:
		_, ok := callback.(Uninstaller)
		return ok
	case LifecycleHookAfterUninstall:
		_, ok := callback.(AfterUninstaller)
		return ok
	case LifecycleHookBeforeTenantDisable:
		_, ok := callback.(BeforeTenantDisabler)
		return ok
	case LifecycleHookAfterTenantDisable:
		_, ok := callback.(AfterTenantDisabler)
		return ok
	case LifecycleHookBeforeTenantDelete:
		_, ok := callback.(BeforeTenantDeleter)
		return ok
	case LifecycleHookAfterTenantDelete:
		_, ok := callback.(AfterTenantDeleter)
		return ok
	case LifecycleHookBeforeInstallModeChange:
		_, ok := callback.(BeforeInstallModeChanger)
		return ok
	case LifecycleHookAfterInstallModeChange:
		_, ok := callback.(AfterInstallModeChanger)
		return ok
	default:
		return false
	}
}

// runOneLifecycleCallback runs one hook with panic recovery and timeout conversion.
func runOneLifecycleCallback(
	ctx context.Context,
	req LifecycleRequest,
	participant LifecycleParticipant,
) LifecycleDecision {
	startedAt := time.Now()
	hookCtx, cancel := context.WithTimeout(ctx, req.HookTimeout)
	defer cancel()

	done := make(chan LifecycleDecision, 1)
	go func() {
		decision := LifecycleDecision{PluginID: participant.PluginID, Hook: req.Hook, OK: true}
		defer func() {
			if recovered := recover(); recovered != nil {
				decision.OK = false
				decision.Panicked = true
				decision.PanicText = toPanicText(recovered)
				decision.Stack = string(debug.Stack())
				decision.Reason = "plugin." + participant.PluginID + ".lifecycle.panic"
			}
			decision.Elapsed = time.Since(startedAt)
			done <- decision
		}()
		decision.OK, decision.Reason, decision.Err = callLifecycleCallback(hookCtx, req, participant.Callback)
		if decision.Err != nil {
			decision.OK = false
		}
	}()

	select {
	case decision := <-done:
		return decision
	case <-hookCtx.Done():
		return LifecycleDecision{
			PluginID: participant.PluginID,
			Hook:     req.Hook,
			OK:       false,
			Reason:   "plugin." + participant.PluginID + ".lifecycle.timeout",
			Elapsed:  time.Since(startedAt),
			TimedOut: true,
			Err:      hookCtx.Err(),
		}
	}
}

// callLifecycleCallback dispatches to the selected hook interface.
func callLifecycleCallback(ctx context.Context, req LifecycleRequest, callback any) (bool, string, error) {
	switch req.Hook {
	case LifecycleHookBeforeInstall:
		return callback.(BeforeInstaller).BeforeInstall(ctx, req.PluginInput)
	case LifecycleHookAfterInstall:
		return true, "", callback.(AfterInstaller).AfterInstall(ctx, req.PluginInput)
	case LifecycleHookBeforeUpgrade:
		return callback.(BeforeUpgrader).BeforeUpgrade(ctx, req.UpgradeInput)
	case LifecycleHookUpgrade:
		return true, "", callback.(Upgrader).Upgrade(ctx, req.UpgradeInput)
	case LifecycleHookAfterUpgrade:
		return true, "", callback.(AfterUpgrader).AfterUpgrade(ctx, req.UpgradeInput)
	case LifecycleHookBeforeDisable:
		return callback.(BeforeDisabler).BeforeDisable(ctx, req.PluginInput)
	case LifecycleHookAfterDisable:
		return true, "", callback.(AfterDisabler).AfterDisable(ctx, req.PluginInput)
	case LifecycleHookBeforeUninstall:
		return callback.(BeforeUninstaller).BeforeUninstall(ctx, req.PluginInput)
	case LifecycleHookUninstall:
		return true, "", callback.(Uninstaller).Uninstall(ctx, req.UninstallInput)
	case LifecycleHookAfterUninstall:
		return true, "", callback.(AfterUninstaller).AfterUninstall(ctx, req.PluginInput)
	case LifecycleHookBeforeTenantDisable:
		return callback.(BeforeTenantDisabler).BeforeTenantDisable(ctx, req.TenantInput)
	case LifecycleHookAfterTenantDisable:
		return true, "", callback.(AfterTenantDisabler).AfterTenantDisable(ctx, req.TenantInput)
	case LifecycleHookBeforeTenantDelete:
		return callback.(BeforeTenantDeleter).BeforeTenantDelete(ctx, req.TenantInput)
	case LifecycleHookAfterTenantDelete:
		return true, "", callback.(AfterTenantDeleter).AfterTenantDelete(ctx, req.TenantInput)
	case LifecycleHookBeforeInstallModeChange:
		return callback.(BeforeInstallModeChanger).BeforeInstallModeChange(ctx, req.InstallModeInput)
	case LifecycleHookAfterInstallModeChange:
		return true, "", callback.(AfterInstallModeChanger).AfterInstallModeChange(ctx, req.InstallModeInput)
	default:
		return true, "", nil
	}
}

// toPanicText converts a recovered panic value into a loggable string.
func toPanicText(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return "panic"
}
