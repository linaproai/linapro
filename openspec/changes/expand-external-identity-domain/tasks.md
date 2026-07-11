## 1. 宿主 handoff 与外部登录契约

- [x] 1.1 宿主实现外部登录 handoff 存储（创建/消费/过期）与交换 API
- [x] 1.2 扩展 `extlogin` 或 auth 装配，使协议插件可登记 LoginOutput→handoff
- [x] 1.3 i18n 错误码与交换失败语义；禁止错误回跳使用内部 err 原文（协议侧配合）
- [x] 1.4 宿主单测：handoff 单次消费、过期、缺码

## 2. 扩展 SPI 与登录编排

- [x] 2.1 扩展 `extidspi` DTO（SubjectKind/AppContext/Secondary 等）保持兼容
- [x] 2.2 SPI 增加 LoginPrepare（或由 auth resolve 路径组合）；fail-closed 不变
- [x] 2.3 更新 auth 外部登录测试

## 3. linapro-extid-core 领域完整化

- [x] 3.1 `distribution: managed`；README/降级说明；去掉 builtin 假设
- [x] 3.2 新增 `backend/cap/extidcap` 完整 Service 接口与 DTO
- [x] 3.3 ticket 存储（签发/peek/consume/invalidate）实现
- [x] 3.4 LoginPrepare、BindByTicket；删除/替换裸 Bind HTTP
- [x] 3.5 链接表 schema：subject_kind、app_context、phone/display/avatar 等扩展字段已并入 `001-linapro-extid-core-identities.sql`（删除冗余 `002` ALTER）
- [x] 3.6 Provider catalog 注册与 ListProviders
- [x] 3.7 未实现方法返回 not-supported；identity 单测覆盖 ticket/bind/list
- [x] 3.8 插件 i18n 错误码与 apidoc

## 4. 协议插件 google/discord

- [x] 4.1 确认 dependencies 指向 linapro-extid-core；README 安装顺序
- [x] 4.2 回调成功走 handoff 回跳；错误安全文案
- [x] 4.3 注册 ProviderDescriptor 到 catalog（若 cap 可用）
- [x] 4.4 构建/vet 通过

## 5. Vben 前端

- [x] 5.1 login 页消费 handoff：调用交换 API 后 completeExternalLogin
- [x] 5.2 去掉对 query accessToken/refreshToken 的依赖
- [x] 5.3 文案 i18n

## 6. 文档与验证

- [x] 6.1 更新 `pkg/plugin` README Auth 行与 oidc 插件 README
- [x] 6.2 `openspec validate expand-external-identity-domain --strict`
- [x] 6.3 宿主 + oidc-core 相关 go test / go build
- [x] 6.4 影响分析记录：i18n/缓存/数据权限/跨平台/测试

## Feedback

- [x] **FB-1**: 插件 API 路由定义去掉冗余 `/plugins/{pluginId}/` 前缀（宿主已挂 `/x/{pluginId}/api/v1`）
- [x] **FB-2**: 安装 `linapro-extid-core` 时 `001-linapro-extid-core-identities.sql` 失败：开发库残留旧表缺少扩展列，`CREATE TABLE IF NOT EXISTS` 空操作后 `COMMENT ON COLUMN` 报错
  - **根因**：开发库残留 pre-expand 表结构；项目无历史兼容负担，安装 SQL 只声明当前完整 schema，不维护升级 ALTER
  - **修复**：`001` 保持 greenfield 完整 `CREATE TABLE` + 注释 + 索引定义（去掉兼容性 `ADD COLUMN`）；本地脏表按卸载语义 `DROP TABLE` 后重装
  - **影响分析**：i18n=无；缓存=无；数据权限=无；开发工具=无；测试策略=DROP 后双次执行安装 SQL 幂等 + API 重装验证
  - **验证**：见反馈执行记录
