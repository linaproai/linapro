# Tasks

## Summary

- [x] 宿主：`extidspi` manager-backed SPI、`LoginByExternalIdentity` 委托 provider、移除 `sys_user_external_identity`、`usercap.ProvisionExternal`+UsernameAnchor、登录页 `auth.login.after`/`auth.login.social` 分槽。
- [x] 插件 `linapro-extlogin-core`：链接表与 SPI 实现、managed 分发、`extidcap`（ticket/LoginPrepare/BindByTicket/catalog/handoff）、schema 扩展字段、协议插件 handoff 与依赖、Vben handoff 交换。
- [x] 验证：宿主 auth/extidspi/usercap 与插件 identity 单测（fail-closed、并发唯一索引、自隔离、ticket/bind）；`go build`/`go vet`；E2E 目录/菜单/入口/未配置 fail-closed；OpenSpec strict。
- [x] FB-1~3（core 介绍 i18n/文案、插件详情 Descriptions nowrap）：根因分别为缺 plugin.json、技术化文案、标签换行；验证 JSON/i18n.check/E2E TC-13。
- [x] FB-4: 占位凭证导致 IdP invalid_client；修复 fail-closed 回登录页、Vite `/portal` 代理、returnTo；验证 go test + E2E。
- [x] FB-5: 授权登录目录由 core 安装创建，协议设置挂其下；验证 E2E TC001。
- [x] FB-6~8: OIDC 菜单 i18n/去「设置」后缀/落地页 hint；验证 embed 单测 + E2E。
- [x] FB-9/12: 登录页社交图标行与协议全宽按钮分槽；验证 TC006 与插件 E2E。
- [x] FB-10: Google/Discord 介绍产品化；验证 JSON 对齐 plugin.yaml。
- [x] FB-11: 私有表名归一为 `plugin_linapro_extlogin_core_*`；验证 identity 单测与静态检索。
- [x] expand FB-1: API 路径去掉冗余 `/plugins/{pluginId}/` 前缀。
- [x] expand FB-2: 安装 SQL greenfield 完整 schema，脏库 DROP 后重装；验证幂等执行。
- [x] 治理：i18n 有（错误码/菜单/登录文案）；缓存 handoff/ticket 按会话同类；数据权限自隔离例外已记录；DI 无跨插件 internal 依赖；跨平台脚本无强制变更。
