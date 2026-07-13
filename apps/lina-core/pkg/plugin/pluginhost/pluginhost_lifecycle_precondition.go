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
	// LifecycleHookBeforeEnable protects global plugin enable for the target plugin.
	LifecycleHookBeforeEnable LifecycleHook = "BeforeEnable"
	// LifecycleHookAfterEnable observes successful global plugin enable for the target plugin.
	LifecycleHookAfterEnable LifecycleHook = "AfterEnable"
	// LifecycleHookGlobalBeforeInstall lets owner plugins veto another plugin's install.
	LifecycleHookGlobalBeforeInstall LifecycleHook = "GlobalBeforeInstall"
	// LifecycleHookGlobalBeforeEnable lets owner plugins veto another plugin's enable.
	LifecycleHookGlobalBeforeEnable LifecycleHook = "GlobalBeforeEnable"
	// LifecycleHookGlobalBeforeDisable lets owner plugins veto another plugin's disable.
	LifecycleHookGlobalBeforeDisable LifecycleHook = "GlobalBeforeDisable"
	// LifecycleHookGlobalBeforeUninstall lets owner plugins veto another plugin's uninstall.
	LifecycleHookGlobalBeforeUninstall LifecycleHook = "GlobalBeforeUninstall"
)

type lifecycleCallback func(ctx context.Context, req LifecycleRequest) (ok bool, reason string, err error)

// LifecycleCallbacks groups the lifecycle hooks implemented by one plugin.
type LifecycleCallbacks struct {
	hooks map[LifecycleHook]lifecycleCallback
}

func (callbacks *LifecycleCallbacks) set(hook LifecycleHook, callback lifecycleCallback) {
	if callback == nil {
		return
	}
	if callbacks.hooks == nil {
		callbacks.hooks = make(map[LifecycleHook]lifecycleCallback)
	}
	callbacks.hooks[hook] = callback
}

func (callbacks LifecycleCallbacks) has(hook LifecycleHook) bool {
	if callbacks.hooks == nil {
		return false
	}
	return callbacks.hooks[hook] != nil
}

func (callbacks LifecycleCallbacks) call(ctx context.Context, req LifecycleRequest) (bool, string, error) {
	if callbacks.hooks == nil || callbacks.hooks[req.Hook] == nil {
		return true, "", nil
	}
	return callbacks.hooks[req.Hook](ctx, req)
}

func (callbacks LifecycleCallbacks) empty() bool {
	return len(callbacks.hooks) == 0
}

// LifecycleParticipant binds a plugin ID to optional lifecycle callbacks.
type LifecycleParticipant struct {
	PluginID  string             // PluginID is the callback owner.
	Callbacks LifecycleCallbacks // Callbacks contains the participant hook functions.
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
		callbacks := NewSourcePluginLifecycleCallbackAdapter(plugin)
		if callbacks.empty() {
			continue
		}
		items = append(items, LifecycleParticipant{
			PluginID:  plugin.ID(),
			Callbacks: callbacks,
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
	callbacks := NewSourcePluginLifecycleCallbackAdapter(plugin)
	if callbacks.empty() {
		return nil
	}
	return []LifecycleParticipant{{
		PluginID:  plugin.ID(),
		Callbacks: callbacks,
	}}
}

// ListSourcePluginGlobalLifecycleParticipants returns participants that
// explicitly registered the given global Before* hook. Plugins without that
// registration are omitted so install/enable paths stay cheap.
func ListSourcePluginGlobalLifecycleParticipants(hook LifecycleHook) []LifecycleParticipant {
	if !IsGlobalLifecycleHook(hook) {
		return nil
	}
	plugins := ListSourcePlugins()
	items := make([]LifecycleParticipant, 0, len(plugins))
	for _, plugin := range plugins {
		if plugin == nil || plugin.ID() == "" {
			continue
		}
		callbacks := NewSourcePluginGlobalLifecycleCallbackAdapter(plugin, hook)
		if callbacks.empty() {
			continue
		}
		items = append(items, LifecycleParticipant{
			PluginID:  plugin.ID(),
			Callbacks: callbacks,
		})
	}
	return items
}

// IsGlobalLifecycleHook reports whether hook is a GlobalBefore* operation.
func IsGlobalLifecycleHook(hook LifecycleHook) bool {
	switch hook {
	case LifecycleHookGlobalBeforeInstall,
		LifecycleHookGlobalBeforeEnable,
		LifecycleHookGlobalBeforeDisable,
		LifecycleHookGlobalBeforeUninstall:
		return true
	default:
		return false
	}
}

// GlobalLifecycleHookForTarget maps a target-scoped Before* hook to its global twin.
// ok is false when the target hook has no global counterpart.
func GlobalLifecycleHookForTarget(hook LifecycleHook) (LifecycleHook, bool) {
	switch hook {
	case LifecycleHookBeforeInstall:
		return LifecycleHookGlobalBeforeInstall, true
	case LifecycleHookBeforeEnable:
		return LifecycleHookGlobalBeforeEnable, true
	case LifecycleHookBeforeDisable:
		return LifecycleHookGlobalBeforeDisable, true
	case LifecycleHookBeforeUninstall:
		return LifecycleHookGlobalBeforeUninstall, true
	default:
		return "", false
	}
}

// LifecycleRequest describes one lifecycle callback aggregation run.
type LifecycleRequest struct {
	Hook             LifecycleHook // Hook selects which callback function to invoke.
	PluginInput      SourcePluginLifecycleInput
	UninstallInput   SourcePluginUninstallInput
	UpgradeInput     SourcePluginUpgradeInput
	TenantInput      SourcePluginTenantLifecycleInput
	InstallModeInput SourcePluginInstallModeChangeInput
	// GlobalInput carries target identity for GlobalBefore* participants.
	GlobalInput  SourcePluginGlobalLifecycleInput
	Participants []LifecycleParticipant // Participants are invoked concurrently.
	HookTimeout  time.Duration          // HookTimeout overrides the per-callback timeout.
	TotalTimeout time.Duration          // TotalTimeout overrides the aggregate timeout.
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
		if !participant.Callbacks.has(req.Hook) {
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
		decision.OK, decision.Reason, decision.Err = participant.Callbacks.call(hookCtx, req)
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

// toPanicText converts a recovered panic value into a loggable string.
func toPanicText(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return "panic"
}
