# Tasks

## Summary

- [x] 交付宿主边界治理：宿主保留框架核心和管理基座，组织、内容、监控等非核心模块按官方源码插件交付，并通过稳定一层目录挂载和空目录隐藏保持后台可用。
- [x] 交付能力 seam：组织能力接口、登录事件、审计事件、HTTP 路由注册、全局中间件注册、Cron 注册、菜单过滤和权限过滤替代散落的插件状态分支。
- [x] 交付`demo-control`：以`plugin.autoEnable`作为唯一开关，启用后按 Method 拦截`/*`写请求，并保留登录/登出及受控插件治理白名单。
- [x] 交付`pluginbridge`职责拆分：根包收敛为 facade，contract、codec、artifact、hostcall、hostservice、guest 子组件承载唯一协议实现。
- [x] FB-1..48：反馈闭环覆盖能力 seam 收敛、官方源码插件完整迁移、宿主私有常量回收、`orgcap`解耦、插件前后端迁移、局部 E2E 修复、插件本地 ORM、TraceID 静态配置、demo-control 作用域和 bugfix 测试规范。
- [x] 验证：相关 Go 测试、前端构建、E2E、`go test ./pkg/pluginbridge/...`、plugin runtime/wasm/plugindb 测试、动态插件 wasm build、OpenSpec 校验和`lina-review`均已作为归档维护证据保留。
- [x] 治理：记录 i18n、缓存一致性、数据权限、DI、开发工具跨平台和测试策略影响；非 owner 能力完整规范迁移为交叉影响摘要，由`openspec/specs`和对应 owner 分组承载。
