## ADDED Requirements

### Requirement: 启动期装配 tenantcap.Provider
框架启动期 SHALL 在插件 enable 阶段检测 `multi-tenant` 是否启用;若启用,等待该插件注册 `tenantcap.Provider`,然后激活 tenancy 中间件;若未启用,装配 no-op Service。

#### Scenario: 启用流程
- **WHEN** 框架启动且 multi-tenant 已 enabled
- **THEN** 启动顺序:
  1. 应用 SQL 迁移
  2. 加载插件 → multi-tenant 注册 Provider
  3. 装配 tenantcap.Service(关联 Provider)
  4. 注册 tenancy 中间件
  5. 启动 HTTP server

#### Scenario: 短路流程
- **WHEN** multi-tenant 未启用
- **THEN** 装配 no-op tenantcap.Service
- **AND** tenancy 中间件存在但仅写入 TenantId=0 即放行

### Requirement: 启动期一致性校验
启动期 SHALL 校验:
1. `sys_plugin.scope_nature` 与 `install_mode` 组合合法。
2. `sys_role` 不包含平台角色布尔字段,且租户角色(`tenant_id>0`)不得配置 `data_scope=1`。
3. `sys_user.tenant_id=0` 必须无 active membership。
4. `multi-tenant` 已 enabled 时必须有 Provider 注册。

校验失败 SHALL 阻止启动并打印明确错误信息。

#### Scenario: 一致性失败阻止启动
- **WHEN** SQL 直接修改导致非法状态
- **THEN** 启动失败,日志明确指出问题
- **AND** 进程退出
