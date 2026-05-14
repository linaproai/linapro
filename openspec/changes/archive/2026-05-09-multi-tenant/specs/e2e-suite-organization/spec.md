## ADDED Requirements

### Requirement: 多租户 e2e 模块化分组
`apps/lina-plugins/multi-tenant/hack/tests/e2e/` 下 SHALL 维护 multi-tenant 插件自有 E2E 模块,内部按场景分组:
- `tenant-lifecycle/`:创建、暂停、恢复、删除租户。
- `tenant-isolation/`:跨租户隔离反例矩阵(每张 tenancy-aware 表至少一个用例)。
- `tenant-resolution/`:每种解析器一组用例。
- `tenant-switching/`:1:N 用户切租户、token 失效。
- `platform-admin/`:impersonation、bypass、双轨日志。
- `plugin-governance/`:install_mode 切换、tenant_scoped 启用/禁用、scope_nature 校验。
- `lifecycle-guard/`:钩子否决聚合、超时、panic、--force、reason i18n。
- `tenant-config-override/`:平台默认 + 租户覆盖语义。
- `org-center-tenancy/`:org-center 租户感知、租户事件触发部门初始化。

#### Scenario: 模块化布局
- **WHEN** e2e 套件运行
- **THEN** 各分组内 TC 文件独立,命名 `TC{NNNN}-*.ts`

### Requirement: 跨租户隔离矩阵覆盖率 100%
对每张 tenancy-aware 表(`sys_user / sys_role / sys_dict_data / sys_config / sys_file / sys_online_session / sys_notify_* / sys_user_message / sys_job / sys_job_log / monitor-loginlog / monitor-operlog / plugin_org_center_*`)SHALL 至少一个用例验证"租户 A 不可见租户 B 数据"在 list / get / update / delete 上的反例。

#### Scenario: 矩阵覆盖
- **WHEN** 运行 `apps/lina-plugins/multi-tenant/hack/tests/e2e/tenant-isolation/`
- **THEN** 所有 tenancy-aware 表均有对应反例 TC

### Requirement: 多租户启用/禁用双场景验证
e2e SHALL 验证 multi-tenant 启用与未启用两种状态下的关键路径:
- 未启用:登录页/工作台无租户 UI;所有现有单租户 e2e 用例通过。
- 启用:登录两阶段、切换租户、跨租户隔离全部就位。

#### Scenario: 双场景矩阵
- **WHEN** 运行 fixture `multi-tenant-disabled` 与 `multi-tenant-enabled` 两套 fixture
- **THEN** 各自子集 TC 执行
- **AND** 两个 fixture 间无相互污染
