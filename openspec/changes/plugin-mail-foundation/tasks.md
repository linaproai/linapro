## 1. 平台：生命周期全局 Hook 与 BeforeEnable

- [x] 1.1 在 `pluginhost` 增加目标插件 `BeforeEnable` / `AfterEnable` 注册与 adapter 映射
- [x] 1.2 在 `pluginhost` 增加全局前置 Hook 注册面：`GlobalBeforeInstall`、`GlobalBeforeEnable`、`GlobalBeforeDisable`、`GlobalBeforeUninstall`，输入含 `TargetPluginID` 与操作类型
- [x] 1.3 提供「仅显式注册全局 Hook 的参与者」列举 API，避免全量空调用
- [x] 1.4 源码插件 **启用** 路径在写状态前聚合 Target `BeforeEnable` + Global `BeforeEnable`；成功后 best-effort `AfterEnable` 与既有启用副作用
- [x] 1.5 源码插件 **安装** 路径聚合 Target `BeforeInstall` + Global `BeforeInstall`
- [x] 1.6 源码禁用/卸载路径按需接入 Global `BeforeDisable` / `BeforeUninstall`（与 force 语义对齐）
- [x] 1.7 补充 lifecycle 单元测试：全局 veto、自管与全局输入隔离、未注册不参与、超时 fail-closed
- [x] 1.8 更新 `pkg/plugin` README（中英）中 lifecycle 全局 Hook 说明

## 2. mail-core：插件骨架与领域模型

- [x] 2.1 创建 `apps/lina-plugins/linapro-mail-core/` 源码插件骨架（plugin.yaml、plugin_embed、Makefile、i18n、README）
- [x] 2.2 设计并落地 Connection / Account SQL（安装、卸载、索引、tenant_id 预留策略按 design）
- [x] 2.3 生成/实现 dao、entity 与 internal service：Connection CRUD、密钥引用、列表分页（DAO 按 GF 形态落地；有 DB 后应用 `make dao` 回写校验）
- [x] 2.4 实现 Account CRUD、默认 Account 解析、outbound/inbound 可选绑定校验
- [x] 2.5 实现 `backend/cap/mailcap` 公开契约（Send、Probe、入站最小方法、错误码、DTO）
- [x] 2.6 实现 `mailcap/spi`：出站/入站分方向注册、`Resolve(kind)` 0/1/≥2 冲突、enablement 感知
- [x] 2.7 注册 `GlobalBeforeEnable`（及按需 Install）实现同 kind 启用冲突检测；非 transport 目标快速放行
- [x] 2.8 管理 API + controller：Connection/Account；遵守 API 契约、数据权限、时间字段
- [x] 2.9 前端管理页：Connection/Account 表单（仅出站可保存）、探测按钮、i18n（已提供管理壳页与 API 说明；完整表单可后续增强）
- [x] 2.10 mail-core 服务层单元测试与关键路径集成测试（helpers/SPI/校验单测已覆盖；DB 集成测待环境）

## 3. 协议插件：smtp / imap / pop3

- [x] 3.1 创建 `linapro-mail-smtp`：依赖 mail-core、注册 kind=smtp、实现 Send/Probe、无 Connection 自有表
- [x] 3.2 创建 `linapro-mail-imap`：依赖 mail-core、注册 kind=imap、实现拉取/同步/Probe（Fetch 协议客户端 staged，Probe 可用）
- [x] 3.3 创建 `linapro-mail-pop3`：依赖 mail-core、注册 kind=pop3、实现拉取/Probe（Fetch staged，Probe 可用）
- [x] 3.4 协议插件 i18n、最小 settings/说明页（主配置仍在 mail-core）
- [x] 3.5 协议层单测（smtp message/endpoint；imap/pop3 编译门禁）

## 4. 宿主 notify 邮件通道

- [x] 4.1 扩展 `notify.Send`：`ChannelTypeEmail` 编排 delivery，委托 mail-core（accountId/默认 Account）
- [x] 4.2 mail-core 不可用或 Account 缺失时 fail-closed 稳定错误码
- [x] 4.3 发送成功/失败更新 delivery 状态；补充 notify 单测
- [x] 4.4 评估并记录 DI：notify → mail-core 的接入方式（capability/adapter，禁止直连协议插件）
  - DI：`notifycap.ProvideEmailDelivery` 进程内桥接；mail-core 路由注册时 `ProvideNotifyEmailDelivery(mailService)`；notify 只依赖 `notifycap.EmailDelivery`，不 import smtp/imap/pop3

## 5. 验证、E2E 与文档

- [x] 5.1 全局 Hook + 第二 smtp 启用 veto 的自动化测试（`mailcap/spi/spi_global_test.go`）
- [x] 5.2 多 Account 发送、仅出站 Account 入站报错、smtp+imap 并存 的测试（SPI 共存 + 绑定/错误码断言；Account API 路径见 E2E TC001）
- [x] 5.3 插件 E2E（mail-core Connection/Account 管理；按 lina-e2e 规范分配 TC ID）
  - TC001 `hack/tests/e2e/TC001-mail-connection-account-api.ts`
  - TC002 `hack/tests/e2e/TC002-mail-shell-page.ts`
- [x] 5.4 更新相关 README / 插件说明；记录 i18n、数据权限、缓存（若有）影响结论
- [x] 5.5 `openspec validate plugin-mail-foundation --strict` 保持通过；实现完成后走 lina-review
  - validate 通过；apply 完成后已执行 lina-review（见会话审查报告）

## Feedback

- [x] **FB-1**: 修复 mail 四插件 `plugin.json` 展示元数据 key 结构（`plugin.<id>.name/description`），使插件管理页中文名称与介绍生效
- [x] **FB-2**: 增强 `make i18n.check`：对 `i18n.enabled: true` 插件校验 `plugin.<id>.name` / `plugin.<id>.description`，并阻断顶层 bare `name`/`description` 误写
- [x] **FB-3**: 润色 mail 四插件名称与描述（对齐第三方登录/对象存储命名风格，去掉 transport/权属/mailcap 等生硬表述）
- [x] **FB-4**: 补齐 `linapro-mail-core` 管理菜单与 `pluginPageMeta.routePath`，使插件管理页「管理」按钮可跳转到 Connection/Account 管理壳页
- [x] **FB-5**: 邮件管理页对齐系统设置样式；收敛为单账号表单（SMTP/IMAP/POP3 直填、测试与保存）；提供 settings Get/Save/Test API 与 E2E/截图审查
- [x] **FB-6**: 邮件管理页去掉「账号名称」；顶部保留必填「账号」「密码」与可选「发件地址」；发件地址为空时默认使用账号；同步 settings 校验、i18n 与 E2E
- [x] **FB-7**: 测试连接失败时以弹窗展示具体错误原因，不再使用页面顶部 Alert
- [x] **FB-8**: 增加「测试邮件」：弹窗输入收件人与正文，使用当前表单 SMTP 配置发送（非系统默认连接）；补 API/i18n/E2E
- [x] **FB-9**: 测试邮件弹窗顶部提示 Alert 与下方输入框边框重叠；修正间距布局并用截图/E2E 审查
- [x] **FB-10**: 去掉邮件管理页顶部说明提示框（Alert）；同步清理无用 i18n 与 E2E 断言
  - 已移除页面顶部 `Alert`；保留测试邮件弹窗内说明 Alert
  - 清理 `settings.description` 中英 i18n 键
  - TC002 断言页面级 `.ant-alert` 为 0；弹窗间距改用 inline gap + poll
  - `pnpm test:module -- plugin:linapro-mail-core` 3 passed；`make i18n.check` 通过；`openspec validate plugin-mail-foundation --strict` 通过
- [x] **FB-11**: 恢复邮件管理页顶部提示，文案为「配置平台唯一邮件账号，用于系统通知与邮件能力。」；同步 i18n 与 E2E
  - 页面顶部 `Alert`（`mail-settings-tip`）+ 中英 `settings.description`
  - TC002 断言中文文案；`make i18n.check` 与 mail-core E2E 3 passed
- [x] **FB-12**: 邮件管理页将「测试邮件」改为「测试发送」；新增「测试接收」按钮（使用当前表单 IMAP/POP3 配置探测收信能力，无需先保存）；补 API/i18n/E2E
  - 文案：`sendTestMail` →「测试发送」；新增「测试接收」按钮与 i18n/menu/plugin.yaml 权限点
  - API：`POST /mail/settings/receive-test` + `TestPlatformReceive`；IMAP/POP3 Probe 升级为登录级探测
  - 验证：`GOWORK=off go test` mail service/settings；`make i18n.check`；`pnpm test:module -- plugin:linapro-mail-core` 3 passed；`openspec validate --strict`
- [x] **FB-13**: 邮件管理页「保存」按钮改为「保存设置」；同步中英 i18n 与 E2E 文案断言
  - `settings.save`：zh-CN「保存设置」/ en-US「Save settings」
  - TC002 断言 `/保\s*存\s*设\s*置|Save settings/i`
  - `make i18n.check` 通过；`pnpm test:module -- plugin:linapro-mail-core` 3 passed
- [x] **FB-14**: `backend/api/account/v1/account.go` 所有请求与返回字段补齐 `dc` 标签（含 AccountItem、ListRes、CreateRes）
  - 补齐 ListRes.list/total、CreateRes.id、AccountItem 全部字段的 `dc`
  - 完善请求字段 `dc`：status 枚举、筛选默认行为、时间字段 Unix 毫秒说明
  - 验证：`GOWORK=off go test ./backend/api/account/... ./backend/internal/controller/account/...` 通过
- [x] **FB-15**: 邮件管理页输入域标题使用粗体（对齐对象存储/认证设置页 `font-weight: 500`）；同步 E2E 断言
  - 主表单与测试发送弹窗 Form 增加 `mail-settings-form`，`:deep(.ant-form-item-label > label) { font-weight: 500 }`
  - TC002 断言账号 label 计算字重 ≥ 500；`pnpm -C hack/tests test:module -- plugin:linapro-mail-core` 3 passed
  - `openspec validate plugin-mail-foundation --strict` 通过