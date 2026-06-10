// This file contains source-plugin enablement lookup behavior. It keeps reader
// nil handling and plugin ID normalization outside the package entrypoint while
// preserving the published plugin state service contract.

package plugincap

import (
	"context"
	"strings"
)

// IsEnabled reports whether the target plugin is enabled for the current
// request context.
//
// This method is the regular enablement lookup entrypoint exposed to source
// plugins. It is intended for high-frequency checks such as menus, routes,
// permission filtering, and ordinary business branches. It defensively checks
// whether the adapter and underlying service have been injected so source
// plugins do not hit nil-pointer panics when the host is not assembled
// correctly or a test double is incomplete.
//
// The pluginID value is trimmed before forwarding. An empty normalized plugin
// ID returns false so invalid identifiers are never treated as enabled. The
// host plugin service owns the final decision and may answer from the current
// process-local warmed enablement snapshot, so this method is suitable for
// ordinary presentation and business entry checks that can tolerate short cache
// lag.
//
// IsEnabled 判断指定插件在当前请求上下文中是否处于可用状态。
//
// 该方法是发布给源码插件使用的常规启用状态查询入口，主要用于菜单、路由、
// 权限过滤和普通业务分支等高频判断。方法会先保护性检查适配器和底层服务是否
// 已经完成注入，避免源码插件在宿主未正确装配或测试替身缺失时触发空指针异常。
//
// pluginID 会在转发到底层服务前去除首尾空白；如果规整后为空字符串，方法直接
// 返回 false，表示无效插件标识不会被视为启用。实际启用状态由宿主插件服务判定，
// 宿主可能基于当前进程内已经预热的启用状态快照完成读取，因此该方法适合对短暂
// 缓存滞后不敏感的普通展示和业务入口判断。
func (s *stateServiceAdapter) IsEnabled(ctx context.Context, pluginID string) bool {
	if s == nil || s.service == nil {
		return false
	}
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return false
	}
	return s.service.IsEnabled(ctx, normalizedPluginID)
}

// IsProviderEnabled reports platform-level provider availability for the
// target plugin.
//
// This method is published for framework capability consumers that may be host
// services, source plugins, or dynamic plugins. It uses the host-owned provider
// enablement semantics instead of tenant/request business-entry visibility, so
// callers can decide whether a declared provider is usable before applying
// their own fallback behavior.
func (s *stateServiceAdapter) IsProviderEnabled(ctx context.Context, pluginID string) bool {
	if s == nil || s.service == nil {
		return false
	}
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return false
	}
	return s.service.IsProviderEnabled(ctx, normalizedPluginID)
}

// IsEnabledAuthoritative reports the authoritative enablement state of the
// target plugin for the current request context.
//
// This method is intended for source-plugin scenarios that are more sensitive
// to stale state, such as global middleware control, write protection, and
// cross-module switches. Unlike IsEnabled, it asks the host service to bypass
// the current process-local enablement snapshot and read the persisted plugin
// governance state or an equivalent authoritative source, reducing the effect
// of a local snapshot that has not refreshed immediately after an administrator
// changes plugin state.
//
// The method keeps the same safety boundary as the regular lookup: it returns
// false when the adapter or underlying service has not been injected, trims
// pluginID before lookup, returns false for an empty normalized plugin ID, and
// treats unresolved authoritative state as disabled. The underlying service may
// still combine tenant, request scope, runtime upgrade gates, and business
// exposure rules from ctx; true only means the plugin may expose capabilities
// within those boundaries.
//
// IsEnabledAuthoritative 判断指定插件在当前请求上下文中的权威可用状态。
//
// 该方法用于源码插件控制全局中间件、写保护、跨模块开关等对状态陈旧更敏感的
// 场景。与 IsEnabled 不同，它会要求宿主底层服务绕过当前进程内的启用状态快照，
// 改为读取持久化插件治理状态或等价权威来源，从而降低插件状态刚被管理员调整后
// 本地快照尚未刷新的影响。
//
// 方法仍然保留与普通查询一致的安全边界：适配器或底层服务未注入时返回 false；
// pluginID 会先去除首尾空白，规整后为空也返回 false；如果宿主无法解析权威状态，
// 也会按禁用处理。底层服务仍会结合当前 ctx 中的租户、请求范围、运行时升级门禁和
// 业务入口暴露规则共同判断，返回 true 仅表示该插件在这些边界内可以暴露业务能力。
func (s *stateServiceAdapter) IsEnabledAuthoritative(ctx context.Context, pluginID string) bool {
	if s == nil || s.service == nil {
		return false
	}
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return false
	}
	return s.service.IsEnabledAuthoritative(ctx, normalizedPluginID)
}
