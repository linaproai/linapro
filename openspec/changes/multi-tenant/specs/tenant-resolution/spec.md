## ADDED Requirements

### Requirement: 租户解析责任链
系统 SHALL 提供代码固定的租户解析责任链,在每个 HTTP 请求进入业务处理前依次尝试 `[override, jwt, session, header, subdomain, default]` 解析器,首个返回非空 `TenantID` 的解析器胜出。普通已认证业务请求中,JWT `TenantId` 是权威租户身份;`X-Tenant-Code` 与 subdomain 仅作为登录前或 `pre_token` 阶段的租户 hint,不得覆盖正式 JWT。

#### Scenario: 默认解析顺序
- **WHEN** 一个用户已登录的请求到达,JWT 中包含 `TenantId`
- **AND** 请求头 `X-Tenant-Code` 也存在
- **THEN** `jwt` 解析器先于 `header` 命中,采用 JWT 中的租户
- **AND** 如果 `X-Tenant-Code` 与 JWT 租户不一致,系统忽略该 hint 并记录安全审计事件;普通用户不得用 header 临时切换租户

#### Scenario: 解析链固定在代码中
- **WHEN** 系统启动或处理租户解析请求
- **THEN** 使用代码内置解析链 `[override, jwt, session, header, subdomain, default]`
- **AND** 不读取宿主配置文件或插件数据库表覆盖解析链顺序

#### Scenario: 登录前 header hint
- **WHEN** 未持正式 JWT 的登录请求携带 `X-Tenant-Code: acme`
- **THEN** `header` 解析器可将 `acme` 作为租户 hint
- **AND** 认证成功后仅当用户拥有 acme 的 active membership 时才允许自动选择或返回候选租户

### Requirement: override 解析器
`override` 解析器 SHALL 解析 `X-Tenant-Override` 请求头,但仅当当前用户处于平台上下文、有效数据权限为全部数据权限且具备 `system:tenant:impersonate` 权限时生效;其他情况下 header 被忽略。

#### Scenario: 平台管理员合法 impersonation
- **WHEN** 平台管理员请求带 `X-Tenant-Override: acme`
- **THEN** `bizctx.TenantId` 解析为 `acme.id`,`bizctx.ActingAsTenant = true`
- **AND** 后续查询与写入按租户 acme 的普通租户视图过滤
- **AND** 操作日志携带 `acting_user_id = 平台管理员`,`on_behalf_of_tenant_id = acme.id`

#### Scenario: 普通用户尝试 override
- **WHEN** 普通用户(无 platform 权限)请求带 `X-Tenant-Override: acme`
- **THEN** override 解析器视该 header 为非法并跳过
- **AND** 操作日志记录一条 warn 级安全事件

### Requirement: subdomain 解析器与保留子域
`subdomain` 解析器 SHALL 从请求 host 中按有效 `root_domain` 提取租户子域,且必须排除 `reserved` 列表(默认 `[www, api, admin, static, docs]`)。首版 `root_domain` 由代码默认值固定为空,后台更新接口不得接受非空值;因此子域名解析当前保持禁用,后续迭代开放设置后才可启用。

#### Scenario: 提取租户子域
- **WHEN** 请求 host 为 `acme.app.com` 且 `root_domain = app.com`
- **THEN** 提取 `acme` 作为租户 code
- **AND** 调用 Provider 解析为 TenantID

#### Scenario: 当前版本 rootDomain 未开放
- **WHEN** 平台管理员尝试保存非空 `rootDomain`
- **THEN** 后台拒绝保存
- **AND** 运行时解析策略继续使用空 `rootDomain`,子域名解析器返回空并让链路继续向下

#### Scenario: 命中保留子域
- **WHEN** 请求 host 为 `www.app.com`
- **THEN** subdomain 解析器返回空,链路继续向下
- **AND** 不影响 jwt/session 等后续解析

### Requirement: 未识别请求的处理策略
当所有解析器均未返回有效 TenantID 时,系统 SHALL 按代码固定的 `prompt` 策略返回 `TENANT_REQUIRED` 错误码。该策略只适用于没有正式 JWT 租户身份的登录前、`pre_token` 或内部兜底链路;不得用于覆盖已认证业务请求中的 JWT `TenantId`。

#### Scenario: 固定 prompt 模式下首次登录
- **WHEN** 未持正式 JWT 的 1:N 用户首次登录且未带任何租户 hint
- **THEN** API 返回 `TENANT_REQUIRED` 错误码和用户可见租户列表
- **AND** 前端展示挑选器,用户选定后通过 `/auth/select-tenant` 继续

### Requirement: 解析器配置来源
解析链顺序、保留子域、`root_domain` 与 ambiguous 行为 SHALL 固定在代码中,并通过代码注释说明规则逻辑、优先级和禁用边界。系统 SHALL NOT 创建或读取 `plugin_multi_tenant_resolver_config` 表,宿主 `config.template.yaml` SHALL NOT 暴露 `tenant.resolution.*` 配置项。

#### Scenario: 代码固定解析策略
- **WHEN** 系统解析租户身份或校验租户 code 是否命中保留标签
- **THEN** 直接使用代码内置解析链、保留子域和 `prompt` ambiguous 行为
- **AND** 不执行数据库查询、运行时配置保存或集群配置广播

#### Scenario: rootDomain 当前禁用
- **WHEN** 请求 host 带有子域名
- **THEN** 因代码固定 `rootDomain = ""`,subdomain 解析器返回空并让链路继续向下
- **AND** 后续开放 rootDomain 时必须重新设计配置来源和分布式一致性策略

### Requirement: 解析结果与一致性
解析器结果(子域名/header → TenantID 映射)SHALL 以 `plugin_multi_tenant_tenant` 表为权威数据源;当前实现不引入解析配置缓存或配置变更广播。租户 code 一旦创建不可修改,因此不存在 code 变更导致的解析缓存重映射。

#### Scenario: 租户删除触发失效
- **WHEN** 平台管理员删除租户 `acme`
- **THEN** 后续按 `plugin_multi_tenant_tenant` 权威数据源查询不再命中有效 TenantID
- **AND** 不依赖解析配置缓存失效
