## ADDED Requirements

### Requirement: 语言切换 MUST NOT 触发整套权限/菜单/路由重载
前端在用户切换语言时 SHALL 仅刷新与语言强相关的本地状态，包括公共配置同步与字典缓存重置。语言切换 MUST NOT 触发 `refreshAccessibleState` 等"重新拉取菜单 + 重新生成路由"的全量权限重载流程；菜单与路由标题 MUST 通过响应式 `$t(...)` 自动更新，禁止在路由生成阶段把当前语言文案"烘焙"成静态字符串。

#### Scenario: 语言切换只更新公共配置与字典缓存
- **WHEN** 用户在 UI 中切换 `preferences.app.locale`
- **THEN** 前端 MUST 调用 `syncPublicFrontendSettings(locale)` 同步公共配置
- **AND** MUST 调用 `useDictStore().resetCache()` 重置字典缓存
- **AND** MUST NOT 调用 `refreshAccessibleState(router)` 重新生成路由

#### Scenario: 菜单标题随语言响应式更新
- **WHEN** 用户切换语言后停留在当前页面
- **THEN** 菜单与面包屑 MUST 自动以新语言文案展示
- **AND** 此过程 MUST NOT 重新请求 `/api/v1/user/info` 或菜单接口

#### Scenario: 路由 meta.title MUST 通过 i18n key 引用
- **WHEN** 任意路由配置定义 `meta.title`
- **THEN** 该字段 MUST 是 i18n key 或 `() => $t(...)` 形式
- **AND** MUST NOT 在路由初始化阶段一次性求值为某个语言下的字符串
