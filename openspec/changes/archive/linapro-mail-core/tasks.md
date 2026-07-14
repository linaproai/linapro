# Tasks

## Summary

- [x] 宿主生命周期：Target `BeforeEnable`/`AfterEnable`；Global BeforeInstall/Enable/Disable/Uninstall 注册面与仅显式参与者列举；源码 install/enable 写状态前聚合 Target+Global；force 语义对齐；`pkg/plugin` README。
- [x] `linapro-mail-core`：骨架、Connection/Account SQL（`plugin_linapro_mail_core_*`、`tenant_id` 预留）、service/cap/spi、`Resolve(kind)` 0/1/≥2、`GlobalBeforeEnable` kind 冲突、管理 API 与平台唯一账号设置页（测试连接/发送/接收、保存设置）。
- [x] 协议插件 smtp/imap/pop3：硬依赖 mail-core、SPI 注册、Probe（imap/pop3 Fetch staged）、无 Connection 自有表、无冲突样板。
- [x] notify `ChannelTypeEmail`：委托 mail-core；`notifycap.EmailDelivery` 进程内桥接；不可用 fail-closed；delivery 状态可观测。

## Verification

- [x] lifecycle 单测：全局 veto、自管/全局输入隔离、未注册不参与、超时 fail-closed。
- [x] mail SPI/service 单测：kind 冲突、smtp+imap 并存、仅出站入站错误码、settings/probe/test-send/receive。
- [x] notify 邮件通道单测；协议插件编译/消息层单测。
- [x] E2E：TC001 Connection/Account API；TC002 设置页壳与文案/按钮（`plugin:linapro-mail-core`）。
- [x] `make i18n.check`；`go test`/`go build` 相关包；OpenSpec strict；lina-review。

## Feedback

- [x] FB-1~3：四插件 `plugin.json` 展示 key 结构、`i18n.check` 阻断 bare name/description、名称描述产品化。
- [x] FB-4：管理菜单 + `pluginPageMeta.routePath`，「管理」可跳转。
- [x] FB-5~13：设置页对齐系统设置、单账号表单、无账号名称、发件地址默认账号、测试失败弹窗、测试发送/接收、顶部 tip 文案、按钮「保存设置/测试发送/测试接收」、弹窗间距。
- [x] FB-14：Account API 请求/响应字段补齐 `dc`。
- [x] FB-15：表单项 label 字重 500 对齐同类设置页。

## Governance

- [x] i18n：菜单/表单/错误/veto/插件展示元数据/API 源文本；有影响并已维护中英包与 check。
- [x] 数据权限：平台配置控制面；管理 API 遵守平台上下文与权限标签；探测/错误不放宽租户边界。
- [x] 缓存：一期无新增跨实例邮件解析缓存契约；记录无强制缓存影响。
- [x] DI：notify → `EmailDelivery` 桥接；mail-core 提供实现；不直连协议插件 internal；构造函数显式注入。
- [x] 宿主边界：邮件协议/Connection 不进 `pkg/plugin` 领域契约；全局 Hook 通用无 hardcode kind。
- [x] 测试策略：单元（lifecycle/spi/notify/settings）+ 插件 E2E；不依赖真实外网邮箱作为门禁默认路径。
