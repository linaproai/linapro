// This file defines the source-plugin visible plugin-state contract.

package plugincap

import "context"

// StateService defines plugin enablement lookups published to source plugins.
type StateService interface {
	// IsEnabled reports whether the plugin is currently installed, enabled, and
	// allowed to expose business entries for the current request scope. Hosts may
	// satisfy this read from a process-local snapshot, so callers that gate
	// whole-system controls should use IsEnabledAuthoritative instead.
	IsEnabled(ctx context.Context, pluginID string) bool
	// IsProviderEnabled reports whether pluginID is platform-enabled and may
	// serve framework capability provider calls. This check is intentionally
	// independent from tenant/request business-entry visibility: provider
	// availability is owned by the platform plugin enabled snapshot, and
	// capability consumers apply their own request-level fallback semantics.
	IsProviderEnabled(ctx context.Context, pluginID string) bool
	// IsEnabledAuthoritative reports whether pluginID is currently installed,
	// enabled, and allowed to expose business entries for the current request
	// scope after bypassing process-local enablement snapshots. The host still
	// preserves tenant/request visibility, runtime upgrade gates, and the normal
	// false-on-read-failure behavior; the only intentional difference from
	// IsEnabled is that the read is forced through persisted plugin governance
	// state instead of a warmed platform snapshot. Use this method when a source
	// plugin gates whole-system behavior, such as global middleware or write
	// protection, where stale process-local state would keep controls active or
	// inactive after an operator changes plugin status.
	//
	// 该接口用于源码插件读取插件启用状态的“权威结果”。它会绕过当前进程内的启用状态快照，
	// 改为读取持久化插件治理状态，同时仍保留租户/请求范围和运行时升级门禁。适用于全局中间件、
	// 演示模式写保护等不能接受本地快照滞后的控制逻辑；普通菜单、路由、权限过滤等高频判断优先使用
	// IsEnabled。
	IsEnabledAuthoritative(ctx context.Context, pluginID string) bool
}
