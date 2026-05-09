## ADDED Requirements

### Requirement: 租户解析责任链
系统 SHALL 提供可配置的租户解析责任链,在每个 HTTP 请求进入业务处理前依次尝试 `[override, jwt, session, header, subdomain, default]` 解析器,首个返回非空 `TenantID` 的解析器胜出。普通已认证业务请求中,JWT `TenantId` 是权威租户身份;`X-Tenant-Code` 与 subdomain 仅作为登录前或 `pre_token` 阶段的租户 hint,不得覆盖正式 JWT。

#### Scenario: 默认解析顺序
- **WHEN** 一个用户已登录的请求到达,JWT 中包含 `TenantId`
- **AND** 请求头 `X-Tenant-Code` 也存在
- **THEN** `jwt` 解析器先于 `header` 命中,采用 JWT 中的租户
- **AND** 如果 `X-Tenant-Code` 与 JWT 租户不一致,系统忽略该 hint 并记录安全审计事件;普通用户不得用 header 临时切换租户

#### Scenario: 配置的解析链可调整顺序
- **WHEN** 平台管理员将解析链改为 `[override, jwt, default]`
- **THEN** `header` 与 `subdomain` 解析器不参与解析
- **AND** 仅 JWT claim 与默认解析器生效

#### Scenario: 登录前 header hint
- **WHEN** 未持正式 JWT 的登录请求携带 `X-Tenant-Code: acme`
- **THEN** `header` 解析器可将 `acme` 作为租户 hint
- **AND** 认证成功后仅当用户拥有 acme 的 active membership 时才允许自动选择或返回候选租户

### Requirement: override 解析器
`override` 解析器 SHALL 解析 `X-Tenant-Override` 请求头,但仅当当前用户为平台管理员且具备 `platform:tenant:impersonate` 权限时生效;其他情况下 header 被忽略。

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
`subdomain` 解析器 SHALL 从请求 host 中按配置的 `root_domain` 提取租户子域,且必须排除 `reserved` 列表(默认 `[www, api, admin, static, docs]`);命中保留子域或未配置 root_domain 时返回空。

#### Scenario: 提取租户子域
- **WHEN** 请求 host 为 `acme.app.com` 且 `root_domain = app.com`
- **THEN** 提取 `acme` 作为租户 code
- **AND** 调用 Provider 解析为 TenantID

#### Scenario: 命中保留子域
- **WHEN** 请求 host 为 `www.app.com`
- **THEN** subdomain 解析器返回空,链路继续向下
- **AND** 不影响 jwt/session 等后续解析

### Requirement: 未识别请求的处理策略
当所有解析器均未返回有效 TenantID 时,系统 SHALL 按 `tenant.resolution.on_ambiguous` 配置处理:`prompt`(默认)返回 `TENANT_REQUIRED` 错误码;`reject` 返回 401;`first_owned` 自动选用户 membership 第一条。该策略只适用于没有正式 JWT 租户身份的登录前、`pre_token` 或内部兜底链路;不得用于覆盖已认证业务请求中的 JWT `TenantId`。

#### Scenario: prompt 模式下首次登录
- **WHEN** 1:N 用户首次登录,未带任何租户 hint
- **THEN** API 返回 `TENANT_REQUIRED` 错误码 + 用户可见租户列表
- **AND** 前端展示挑选器,用户选定后通过 `/auth/select-tenant` 继续

#### Scenario: reject 模式下严格拒绝
- **WHEN** 配置 `on_ambiguous: reject` 且 jwt 无 TenantId,header/subdomain 也未命中
- **THEN** 返回 401 `TENANT_REQUIRED`
- **AND** 前端不展示挑选器(由严格策略决定)

### Requirement: 解析器配置双源
解析链顺序、保留子域、ambiguous 行为等 SHALL 同时支持配置文件(`config.yaml` 启动期初值)与管理后台(`plugin_multi_tenant_resolver_config` 表运行时配置);运行时优先读 DB,DB 缺失时 fallback 到配置文件。

#### Scenario: 平台管理员后台改配置
- **WHEN** 平台管理员在管理后台改 ambiguous 行为为 `reject`
- **THEN** 写入 `plugin_multi_tenant_resolver_config`,触发集群广播失效
- **AND** 所有节点立即生效,无需重启

#### Scenario: 配置变更前置校验
- **WHEN** 平台管理员尝试把解析链中所有解析器都禁用
- **THEN** 后台拒绝保存,提示"至少保留 default 解析器"
- **AND** 配置不变

### Requirement: 解析结果缓存与一致性
解析器结果(子域名/header → TenantID 映射)SHALL 使用进程内 LRU 缓存(TTL 30s)+ 集群失效广播;租户被删除时立即广播失效。租户 code 一旦创建不可修改,因此不存在 code 变更导致的解析缓存重映射。

#### Scenario: 租户删除触发失效
- **WHEN** 平台管理员删除租户 `acme`
- **THEN** 解析缓存全集群失效
- **AND** 后续请求以 `acme.app.com` 不再命中有效 TenantID
