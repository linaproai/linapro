// This file defines the lifecycle guard contract used by the plugin host to
// protect destructive or topology-changing operations before they mutate plugin
// or tenant state. Source plugins register optional guard implementations under
// their own plugin IDs; the host then invokes only the guards that implement the
// requested hook, aggregates all decisions, and fails closed when a guard
// vetoes, returns an error, times out, or panics.
//
// Long-term maintenance notes:
//   - Guard interfaces are intentionally small and optional so source plugins can
//     protect only the lifecycle actions they own.
//   - Guard reasons are stable i18n keys that the caller can surface through the
//     normal bizerr response path.
//   - The runner executes applicable guards concurrently and applies both
//     per-hook and total timeouts to avoid blocking plugin or tenant lifecycle
//     workflows indefinitely.
//   - Panic recovery and timeout conversion are part of the safety boundary:
//     plugin guard failures must block the protected operation instead of
//     allowing unsafe state changes to continue.
//
// 本文件定义插件宿主使用的生命周期保护契约,用于在卸载插件、禁用插件、
// 删除租户或调整安装模式等会改变系统拓扑或运行状态的操作真正执行前,
// 先让源码插件表达"是否允许继续"。源码插件可按自身 plugin ID 注册可选的
// Guard 实现;宿主只调用实现了当前 Hook 的参与者,聚合所有决策,并在任一
// Guard 否决、返回错误、超时或 panic 时按失败关闭策略阻断后续操作。
//
// 长期维护要点:
//   - Guard 接口刻意保持小而可选,插件只需保护自己关心的生命周期动作。
//   - Guard 返回的 reason 应是稳定的 i18n key,由上层通过 bizerr 响应链路
//     暴露给调用端或审计日志。
//   - 聚合器并发执行所有适用 Guard,同时提供单个 Hook 超时和整体超时,
//     避免插件或租户生命周期流程被某个实现长期阻塞。
//   - panic 恢复与超时转换属于安全边界的一部分:Guard 自身异常必须阻断
//     受保护操作,而不是放任可能破坏业务依赖的数据或运行状态继续变化。

package pluginhost

import (
	"context"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// Lifecycle guard default timeouts.
const (
	// DefaultGuardHookTimeout is the per-hook timeout.
	DefaultGuardHookTimeout = 5 * time.Second
	// DefaultGuardTotalTimeout is the total guard aggregation timeout.
	DefaultGuardTotalTimeout = 10 * time.Second
)

// GuardHook identifies one lifecycle guard hook.
type GuardHook string

// String returns the lifecycle guard hook name.
func (hook GuardHook) String() string {
	return string(hook)
}

// Guard hook constants.
const (
	// GuardHookCanUninstall protects plugin uninstall.
	GuardHookCanUninstall GuardHook = "CanUninstall"
	// GuardHookCanDisable protects global plugin disable.
	GuardHookCanDisable GuardHook = "CanDisable"
	// GuardHookCanTenantDisable protects tenant-scoped plugin disable.
	GuardHookCanTenantDisable GuardHook = "CanTenantDisable"
	// GuardHookCanTenantDelete protects tenant deletion.
	GuardHookCanTenantDelete GuardHook = "CanTenantDelete"
	// GuardHookCanChangeInstallMode protects install-mode changes.
	GuardHookCanChangeInstallMode GuardHook = "CanChangeInstallMode"
)

// CanUninstaller optionally vetoes plugin uninstall.
type CanUninstaller interface {
	// CanUninstall returns whether uninstall may continue and an i18n reason when vetoed.
	CanUninstall(ctx context.Context) (ok bool, reason string, err error)
}

// CanDisabler optionally vetoes global plugin disable.
type CanDisabler interface {
	// CanDisable returns whether disable may continue and an i18n reason when vetoed.
	CanDisable(ctx context.Context) (ok bool, reason string, err error)
}

// CanTenantDisabler optionally vetoes tenant-scoped plugin disable.
type CanTenantDisabler interface {
	// CanTenantDisable returns whether tenant disable may continue and an i18n reason when vetoed.
	CanTenantDisable(ctx context.Context, tenantID int) (ok bool, reason string, err error)
}

// CanTenantDeleter optionally vetoes tenant deletion.
type CanTenantDeleter interface {
	// CanTenantDelete returns whether tenant deletion may continue and an i18n reason when vetoed.
	CanTenantDelete(ctx context.Context, tenantID int) (ok bool, reason string, err error)
}

// CanInstallModeChanger optionally vetoes plugin install-mode changes.
type CanInstallModeChanger interface {
	// CanChangeInstallMode returns whether the install-mode transition may continue.
	CanChangeInstallMode(ctx context.Context, from string, to string) (ok bool, reason string, err error)
}

// GuardParticipant binds a plugin ID to an optional guard implementation.
type GuardParticipant struct {
	PluginID string // PluginID is the hook owner.
	Guard    any    // Guard is the optional hook implementation.
}

// Lifecycle guard registry shared by source plugins linked into the host.
var (
	lifecycleGuardRegistryMu sync.RWMutex
	lifecycleGuardRegistry   = make(map[string]any)
)

// RegisterLifecycleGuard registers one plugin-owned lifecycle guard.
func RegisterLifecycleGuard(pluginID string, guard any) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		panic("pluginhost: lifecycle guard plugin id is empty")
	}
	if guard == nil {
		panic("pluginhost: lifecycle guard is nil")
	}

	lifecycleGuardRegistryMu.Lock()
	defer lifecycleGuardRegistryMu.Unlock()
	lifecycleGuardRegistry[normalizedPluginID] = guard
}

// UnregisterLifecycleGuard removes one plugin-owned lifecycle guard.
func UnregisterLifecycleGuard(pluginID string) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return
	}

	lifecycleGuardRegistryMu.Lock()
	defer lifecycleGuardRegistryMu.Unlock()
	delete(lifecycleGuardRegistry, normalizedPluginID)
}

// ListLifecycleGuardParticipants returns the current lifecycle guard participants.
func ListLifecycleGuardParticipants() []GuardParticipant {
	lifecycleGuardRegistryMu.RLock()
	defer lifecycleGuardRegistryMu.RUnlock()

	items := make([]GuardParticipant, 0, len(lifecycleGuardRegistry))
	for pluginID, guard := range lifecycleGuardRegistry {
		if guard == nil {
			continue
		}
		items = append(items, GuardParticipant{
			PluginID: pluginID,
			Guard:    guard,
		})
	}
	return items
}

// ListLifecycleGuardParticipantsForPlugin returns the lifecycle guard participant
// that owns the requested plugin lifecycle action.
func ListLifecycleGuardParticipantsForPlugin(pluginID string) []GuardParticipant {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return nil
	}

	lifecycleGuardRegistryMu.RLock()
	defer lifecycleGuardRegistryMu.RUnlock()

	guard := lifecycleGuardRegistry[normalizedPluginID]
	if guard == nil {
		return nil
	}
	return []GuardParticipant{
		{
			PluginID: normalizedPluginID,
			Guard:    guard,
		},
	}
}

// GuardRequest describes one lifecycle guard aggregation run.
type GuardRequest struct {
	Hook         GuardHook          // Hook selects which guard interface to invoke.
	TenantID     int                // TenantID is passed to tenant-scoped hooks.
	FromMode     string             // FromMode is passed to install-mode hooks.
	ToMode       string             // ToMode is passed to install-mode hooks.
	Participants []GuardParticipant // Participants are invoked concurrently.
	HookTimeout  time.Duration      // HookTimeout overrides the per-hook timeout.
	TotalTimeout time.Duration      // TotalTimeout overrides the aggregate timeout.
}

// GuardDecision is one plugin hook result.
type GuardDecision struct {
	PluginID  string        // PluginID is the hook owner.
	Hook      GuardHook     // Hook is the invoked hook.
	OK        bool          // OK reports whether this plugin allowed the action.
	Reason    string        // Reason is the i18n key when OK is false.
	Err       error         // Err records a hook error.
	Elapsed   time.Duration // Elapsed is the hook runtime.
	TimedOut  bool          // TimedOut reports per-hook timeout.
	Panicked  bool          // Panicked reports panic recovery.
	PanicText string        // PanicText records the recovered panic value.
	Stack     string        // Stack records the panic stack for logging.
}

// GuardResult is the aggregate lifecycle guard result.
type GuardResult struct {
	OK        bool            // OK reports whether all hooks allowed the action.
	Decisions []GuardDecision // Decisions contains one result per applicable participant.
}

// RunLifecycleGuards invokes all applicable lifecycle guards concurrently.
func RunLifecycleGuards(ctx context.Context, req GuardRequest) GuardResult {
	req = normalizeGuardRequest(req)
	ctx, cancel := context.WithTimeout(ctx, req.TotalTimeout)
	defer cancel()

	results := make(chan GuardDecision, len(req.Participants))
	var wg sync.WaitGroup
	for _, participant := range req.Participants {
		if !participantSupportsHook(participant.Guard, req.Hook) {
			continue
		}
		wg.Add(1)
		go func(item GuardParticipant) {
			defer wg.Done()
			results <- runOneLifecycleGuard(ctx, req, item)
		}(participant)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	aggregate := GuardResult{OK: true}
	for decision := range results {
		if !decision.OK {
			aggregate.OK = false
		}
		aggregate.Decisions = append(aggregate.Decisions, decision)
	}
	return aggregate
}

// normalizeGuardRequest fills lifecycle guard timeout defaults.
func normalizeGuardRequest(req GuardRequest) GuardRequest {
	if req.HookTimeout <= 0 {
		req.HookTimeout = DefaultGuardHookTimeout
	}
	if req.TotalTimeout <= 0 {
		req.TotalTimeout = DefaultGuardTotalTimeout
	}
	return req
}

// participantSupportsHook reports whether one participant implements the requested hook.
func participantSupportsHook(guard any, hook GuardHook) bool {
	switch hook {
	case GuardHookCanUninstall:
		_, ok := guard.(CanUninstaller)
		return ok
	case GuardHookCanDisable:
		_, ok := guard.(CanDisabler)
		return ok
	case GuardHookCanTenantDisable:
		_, ok := guard.(CanTenantDisabler)
		return ok
	case GuardHookCanTenantDelete:
		_, ok := guard.(CanTenantDeleter)
		return ok
	case GuardHookCanChangeInstallMode:
		_, ok := guard.(CanInstallModeChanger)
		return ok
	default:
		return false
	}
}

// runOneLifecycleGuard runs one hook with panic recovery and timeout conversion.
func runOneLifecycleGuard(ctx context.Context, req GuardRequest, participant GuardParticipant) GuardDecision {
	startedAt := time.Now()
	hookCtx, cancel := context.WithTimeout(ctx, req.HookTimeout)
	defer cancel()

	done := make(chan GuardDecision, 1)
	go func() {
		decision := GuardDecision{PluginID: participant.PluginID, Hook: req.Hook, OK: true}
		defer func() {
			if recovered := recover(); recovered != nil {
				decision.OK = false
				decision.Panicked = true
				decision.PanicText = toPanicText(recovered)
				decision.Stack = string(debug.Stack())
				decision.Reason = "plugin." + participant.PluginID + ".guard.panic"
			}
			decision.Elapsed = time.Since(startedAt)
			done <- decision
		}()
		decision.OK, decision.Reason, decision.Err = callLifecycleGuard(hookCtx, req, participant.Guard)
		if decision.Err != nil {
			decision.OK = false
		}
	}()

	select {
	case decision := <-done:
		return decision
	case <-hookCtx.Done():
		return GuardDecision{
			PluginID: participant.PluginID,
			Hook:     req.Hook,
			OK:       false,
			Reason:   "plugin." + participant.PluginID + ".guard.timeout",
			Elapsed:  time.Since(startedAt),
			TimedOut: true,
			Err:      hookCtx.Err(),
		}
	}
}

// callLifecycleGuard dispatches to the selected hook interface.
func callLifecycleGuard(ctx context.Context, req GuardRequest, guard any) (bool, string, error) {
	switch req.Hook {
	case GuardHookCanUninstall:
		return guard.(CanUninstaller).CanUninstall(ctx)
	case GuardHookCanDisable:
		return guard.(CanDisabler).CanDisable(ctx)
	case GuardHookCanTenantDisable:
		return guard.(CanTenantDisabler).CanTenantDisable(ctx, req.TenantID)
	case GuardHookCanTenantDelete:
		return guard.(CanTenantDeleter).CanTenantDelete(ctx, req.TenantID)
	case GuardHookCanChangeInstallMode:
		return guard.(CanInstallModeChanger).CanChangeInstallMode(ctx, req.FromMode, req.ToMode)
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
