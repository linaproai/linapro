## Why

启用`linapro-ops-demo-guard`后，演示环境应保持登录、刷新、租户选择、租户切换和退出等最小会话流程可用。前端通过`pluginApiPath("linapro-tenant-core", "auth/select-tenant")`和`pluginApiPath("linapro-tenant-core", "auth/switch-tenant")`生成`/x/linapro-tenant-core/api/v1/auth/*`路径访问租户核心插件，但演示保护白名单只匹配宿主`/api/v1/auth/*`路径，导致这两个`POST`请求被误判为普通写操作并返回只读拒绝。同时，宿主`/api/v1/auth/select-tenant`和`/api/v1/auth/switch-tenant`已无真实 HTTP 入口，保留历史白名单会扩大演示保护绕过面。

## What Changes

- 在`linapro-ops-demo-guard`中间件中新增两个精确会话白名单路径：`/x/linapro-tenant-core/api/v1/auth/select-tenant`和`/x/linapro-tenant-core/api/v1/auth/switch-tenant`，白名单仅接受`POST`方法。
- 删除已无真实 HTTP 入口的宿主`/api/v1/auth/select-tenant`与`/api/v1/auth/switch-tenant`历史白名单常量。
- 补充回归测试，覆盖宿主路径和插件挂载路径下的最小会话白名单行为。

## Capabilities

- 演示环境启用后，租户用户可通过插件挂载路径选择租户登录，admin 可通过插件挂载路径切换租户。
- 登录、刷新、退出等既有会话流程不受影响。
- 普通业务写请求、插件治理写请求和其他插件挂载写请求继续被只读保护拦截。

## Impact

- 修改范围限定在官方源码插件`linapro-ops-demo-guard`的请求分类逻辑、测试和本次 OpenSpec 记录。
- 不修改宿主认证 API、租户核心插件 API、数据库、前端页面、路由生成或插件治理契约。
- 数据权限影响：不新增业务数据读取、写入、列表、详情、导出、聚合、下载或租户/组织可见性逻辑；仅恢复既有认证会话动作进入原认证链。
- 缓存一致性影响：不新增缓存、快照、失效或跨节点同步逻辑。
- `i18n`影响：不新增或修改运行时文案、菜单、API 文档源文本、插件清单或语言包资源。
- 开发工具跨平台影响：不修改脚本、`Makefile`、`linactl`、CI 或构建入口。
