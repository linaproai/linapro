## Context

`linapro-ops-demo-guard`是官方源码插件，通过中间件拦截写请求实现演示环境只读保护。前端租户选择与租户切换通过`linapro-tenant-core`源码插件挂载路径`/x/linapro-tenant-core/api/v1/auth/select-tenant`和`/x/linapro-tenant-core/api/v1/auth/switch-tenant`访问，但演示保护白名单只匹配宿主`/api/v1/auth/*`路径。宿主`/api/v1/auth/select-tenant`和`/api/v1/auth/switch-tenant`已无真实 HTTP 定义，当前 HTTP 入口完全由`linapro-tenant-core`源码插件提供。

## Goals / Non-Goals

**Goals：**

- 使演示环境启用后，租户核心插件挂载路径下的租户选择和租户切换请求能够正常进入认证链。
- 删除已无真实 HTTP 入口的宿主历史白名单，减少演示保护绕过面。
- 保持普通业务写请求、插件治理写请求和其他插件挂载写请求继续被只读保护拦截。

**Non-Goals：**

- 不修改宿主认证 API、`linapro-tenant-core`插件接口定义或前端调用路径。
- 不调整演示保护中间件的整体拦截架构或白名单匹配策略。

## Decisions

**决策：在演示保护插件内补齐租户会话路径白名单**

本次问题来源于`linapro-ops-demo-guard`的请求分类逻辑，而不是宿主认证服务或`linapro-tenant-core`插件契约变化。前端通过`pluginApiPath("linapro-tenant-core", "auth/select-tenant")`和`pluginApiPath("linapro-tenant-core", "auth/switch-tenant")`生成`/x/linapro-tenant-core/api/v1/auth/*`路径，这是源码插件能力挂载路径的正常形态。

修复在`linapro-ops-demo-guard`内收敛最小会话白名单，使其识别租户核心插件挂载路径下的`select-tenant`和`switch-tenant`动作，并删除已无真实 HTTP 入口的宿主历史白名单。白名单只覆盖`POST`方法和当前真实会话路径，不放行其他`/x/{plugin}/api/v1/**`写请求。

**边界判断：**

- 核心宿主边界：不修改`apps/lina-core`核心认证、插件路由或通用 service 契约。
- 插件边界：变更只发生在`apps/lina-plugins/linapro-ops-demo-guard`内，不修改`linapro-tenant-core`业务实现或接口定义。
- 接口性能：请求分类为常量级字符串规范化和精确匹配，不引入数据库访问、远程调用或随数据量增长的装配路径。
- 数据权限：放行后仍由原认证/租户切换链路校验`preToken`、membership、token 和租户边界；中间件本身不读取或暴露业务数据。
- 缓存一致性：不新增缓存或失效路径。
- `i18n`：不新增用户可见文案或语言资源。

**测试策略：**

- 单元测试：扩展中间件测试，证明启用演示保护时租户核心插件挂载路径的`select-tenant`和`switch-tenant`可以通过，同时历史宿主租户会话路径和普通插件挂载写请求仍被拒绝。
- E2E：更新插件自有生命周期 E2E 会话白名单用例，使用前端实际插件挂载路径验证租户选择和租户切换请求不会被只读保护拦截。

## Risks / Trade-offs

- 当前本地服务未启用`linapro-ops-demo-guard`，浏览器端启用态断言未实际命中，需要在启用演示保护的 E2E 环境或 CI 配置中复验。
- 白名单为精确路径匹配，若未来`linapro-tenant-core`变更挂载路径需同步更新白名单。
