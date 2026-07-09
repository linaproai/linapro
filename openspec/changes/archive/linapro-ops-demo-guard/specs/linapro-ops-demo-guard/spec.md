## MODIFIED Requirements

### Requirement:linapro-ops-demo-guard 插件必须保留最小会话白名单

系统 SHALL 在 linapro-ops-demo-guard 启用时保留登录、令牌刷新、租户选择、租户切换和退出行为，使演示环境保持可用。租户选择与租户切换 SHALL 通过`linapro-tenant-core`源码插件挂载路径进入认证链；linapro-ops-demo-guard SHALL 仅对白名单中的当前会话动作放行，不得保留历史宿主租户选择或租户切换 HTTP 路径白名单，也不得放行其他插件挂载写请求。

#### Scenario:登录保持允许
- **当** `linapro-ops-demo-guard` 已启用
- **且** 请求为 `POST /api/v1/auth/login` 时
- **则** linapro-ops-demo-guard 允许请求继续

#### Scenario:刷新令牌保持允许
- **当** `linapro-ops-demo-guard` 已启用
- **且** 请求为 `POST /api/v1/auth/refresh` 时
- **则** linapro-ops-demo-guard 允许请求继续

#### Scenario:租户核心插件挂载路径租户选择保持允许
- **当** `linapro-ops-demo-guard` 已启用
- **且** 请求为 `POST /x/linapro-tenant-core/api/v1/auth/select-tenant` 时
- **则** linapro-ops-demo-guard 允许请求继续

#### Scenario:租户核心插件挂载路径租户切换保持允许
- **当** `linapro-ops-demo-guard` 已启用
- **且** 请求为 `POST /x/linapro-tenant-core/api/v1/auth/switch-tenant` 时
- **则** linapro-ops-demo-guard 允许请求继续

#### Scenario:退出保持允许
- **当** `linapro-ops-demo-guard` 已启用
- **且** 请求为 `POST /api/v1/auth/logout` 时
- **则** linapro-ops-demo-guard 允许请求继续

#### Scenario:其他插件挂载写请求仍被拒绝
- **当** `linapro-ops-demo-guard` 已启用
- **且** 请求为 `POST /x/linapro-tenant-core/api/v1/platform/tenants` 时
- **则** linapro-ops-demo-guard 以清晰的只读演示消息拒绝请求
- **且** 请求不继续进入业务处理

#### Scenario:历史宿主租户会话路径不再放行
- **当** `linapro-ops-demo-guard` 已启用
- **且** 请求为 `POST /api/v1/auth/select-tenant` 或 `POST /api/v1/auth/switch-tenant` 时
- **则** linapro-ops-demo-guard 以清晰的只读演示消息拒绝请求
- **且** 请求不继续进入业务处理
