## ADDED Requirements

### Requirement: 文件存储路径租户前缀
本地文件存储与对象存储 SHALL 在路径中包含租户前缀:本地路径 `/storage/t/{tenant_id}/yyyy/mm/dd/{file_id}`;对象存储 key `t/{tenant_id}/yyyy/mm/dd/{file_id}`。`tenant_id=0` 表示平台共享文件。

#### Scenario: 租户用户上传文件
- **WHEN** 租户 A 用户上传文件
- **THEN** 文件落在 `/storage/t/A/yyyy/mm/dd/{file_id}`
- **AND** `sys_file.tenant_id = A`

#### Scenario: 平台共享文件
- **WHEN** 平台管理员上传平台 logo
- **THEN** 文件落在 `/storage/t/0/yyyy/mm/dd/{file_id}`
- **AND** 所有租户都可读

### Requirement: 文件读取的租户校验
文件下载/预览接口 SHALL 校验请求 `bizctx.TenantId` 与 `sys_file.tenant_id` 匹配;不匹配时返回 403。平台管理员管理平台模式(`TenantId=0`)只能通过显式 `/platform/*` 只读接口跨租户访问文件;impersonation 模式按目标租户执行同样的租户匹配校验。

#### Scenario: 跨租户文件访问被拒
- **WHEN** 租户 B 用户尝试下载租户 A 的文件
- **THEN** 返回 403 `bizerr.CodeFileTenantForbidden`
- **AND** 不返回文件内容

#### Scenario: 平台管理员跨租户只读访问
- **WHEN** 平台管理员在管理平台模式下通过 `/platform/files/{id}` 请求租户 A 的文件
- **THEN** 校验平台权限后返回文件内容
- **AND** 操作日志记录跨租户只读访问事件

#### Scenario: impersonation 文件访问
- **WHEN** 平台管理员以 impersonation 租户 A 的 token 请求租户 B 的文件
- **THEN** 返回 403 `bizerr.CodeFileTenantForbidden`
- **AND** 不因平台管理员身份绕过目标租户过滤

### Requirement: 缓存 key 必须携带租户维度
所有运行时缓存的 key SHALL 在租户敏感场景中携带 `tenant_id` 维度;cache key 构造统一通过 helper `CacheKey(tenant, scope, key)`,禁止散点拼接。

#### Scenario: 字典缓存 key
- **WHEN** 租户 A 读取字典 `user_status`
- **THEN** 缓存 key = `dict:tenant=A:type=user_status`
- **AND** 租户 B 的同一字典查询使用 key `dict:tenant=B:type=user_status`,互不影响

### Requirement: 集群失效广播携带租户范围
`distributed-cache-coordination` 失效消息 SHALL 携带 `tenant_id` 字段;invalidation 必须显式指定租户作用域之一:
- 单租户:`tenant_id = T`(仅失效该租户该 scope 的缓存)。
- 平台默认变更:`tenant_id = 0` + `cascade_to_tenants = true`(失效所有租户的对应 scope)。
- 全局兜底(罕见):`tenant_id = -1`(显式标志"全租户清空",必须有审计日志)。

#### Scenario: 单租户失效不影响其他租户
- **WHEN** 租户 A 修改了配置 `upload.max_size`
- **THEN** 广播失效消息 `{scope: config, key: upload.max_size, tenant_id: A}`
- **AND** 仅 `(tenant_id=A, key=upload.max_size)` 缓存失效
- **AND** 其他租户的相同 key 缓存不受影响

#### Scenario: 平台默认变更级联失效
- **WHEN** 平台管理员修改 `tenant_id=0` 的平台默认
- **THEN** 广播 `{tenant_id: 0, cascade_to_tenants: true}`
- **AND** 所有节点失效 `(tenant_id=*, scope=config, key=upload.max_size)` 的缓存

### Requirement: 审计日志的租户与 impersonation 双轨
`monitor-operlog` 与 `monitor-loginlog` SHALL 包含字段:
- `tenant_id`:操作发生的租户上下文(=`bizctx.TenantId`)。
- `acting_user_id`:实际操作人的全局 user_id(`bizctx.UserId`)。
- `on_behalf_of_tenant_id`:仅 impersonation 时有值,= 平台管理员代为操作的目标租户。
- `is_impersonation`:布尔标识。

#### Scenario: 普通用户操作
- **WHEN** 租户 A 用户 U 修改自己角色
- **THEN** operlog 写 `tenant_id=A, acting_user_id=U, on_behalf_of_tenant_id=NULL, is_impersonation=false`

#### Scenario: 平台管理员 impersonation 操作
- **WHEN** 平台管理员 P 以 impersonation 形态在租户 A 中操作
- **THEN** operlog 写 `tenant_id=A, acting_user_id=P, on_behalf_of_tenant_id=A, is_impersonation=true`
- **AND** 租户管理员查看此日志时附带"平台管理员代操作"标记

### Requirement: 跨租户操作必须显式
任何"跨租户读取/写入"路径 SHALL 仅供平台管理员使用,且必须经过显式 API(以 `/platform/*` 前缀)或专用平台 service;严禁在普通业务 API 中存在隐式跨租户访问。平台管理员 impersonation 某租户不属于全量跨租户 bypass,而是按目标租户普通视图过滤。

#### Scenario: 普通业务 API 不允许跨租户
- **WHEN** 任何 `/api/*` 业务接口尝试跨租户查询
- **THEN** 必须经过显式 `/platform/*` API 或专用平台 service
- **AND** 校验失败时返回 `bizerr.CodeCrossTenantNotAllowed`

### Requirement: 业务表的隔离测试覆盖
e2e 测试套件 SHALL 包含针对每张 tenancy-aware `sys_*` 表与插件业务表的"跨租户隔离"反例用例;每张表至少一个用例验证"租户 A 不可见租户 B 数据"。平台控制面表不纳入跨租户隔离矩阵,应通过平台权限与治理 service 访问控制覆盖。

#### Scenario: 跨租户隔离矩阵
- **WHEN** e2e 套件运行
- **THEN** 每张 tenancy-aware 表至少一个 TC 验证 list / get / update / delete 在跨租户场景下被拒
- **AND** 隔离矩阵覆盖率 100%(矩阵清单维护在 e2e fixtures 中)
