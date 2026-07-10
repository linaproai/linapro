## MODIFIED Requirements

### Requirement: 初始化与注册 API 必须返回错误给调用方决策
系统 SHALL 要求宿主和源码插件的运行时初始化、源码插件注册、registrar、回调注册、路由注册、Cron 注册和中间件注册 API 在依赖缺失、注册参数非法、配置来源缺失、后端创建失败或校验失败时返回 `error`。这些 API MUST NOT 在内部直接 `panic` 处理可预期错误；是否中止进程、忽略或降级必须由调用栈最上层入口显式决定。

#### Scenario: 源码插件注册 API 返回错误
- **WHEN** 源码插件声明无效 extension point、无效执行模式、nil callback 或重复注册
- **THEN** `pluginhost` 注册 API 返回 `error`
- **AND** API 内部不得直接 `panic`

#### Scenario: 顶层静态注册入口选择失败退出
- **WHEN** 源码插件包级 `init` 调用注册 API 收到错误
- **THEN** 该顶层静态注册入口可以显式 `panic`
- **AND** panic 治理扫描 MUST 将该调用识别为顶层入口收到错误后的失败退出
- **AND** 识别方式可以是宿主精确 allowlist 条目，也可以是官方插件工作区对 `backend/plugin.go` `init` 注册 fail-fast 模式的自动归类
- **AND** 官方插件集合变化时，不得要求维护按插件 ID 枚举的宿主 allowlist 清单

#### Scenario: 运行期回调缺少宿主依赖
- **WHEN** HTTP、Cron、Hook 或中间件注册回调在执行期发现宿主发布依赖缺失
- **THEN** 回调返回 `error`
- **AND** 宿主调用方决定阻断启动、记录失败或执行其他降级策略
