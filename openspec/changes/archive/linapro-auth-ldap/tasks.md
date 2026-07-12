# Tasks

## Summary

- [x] 交付 managed 源码插件 `linapro-auth-ldap`：骨架/依赖/菜单、settings、LDAP 连接与 bind、portal login→handoff、catalog/ownership、登录弹层、双语文档与 i18n。
- [x] 验证：插件 `go test`、E2E TC001 菜单/入口/样式、OpenSpec strict、`make i18n.check`。
- [x] FB-1: 目录登录入口与凭证弹层迁到工作台统一 Vben 样式与响应式；验证 typecheck/build、E2E TC001、宿主 TC006。
- [x] FB-2: 菜单改名「LDAP 设置」/「LDAP Settings」，设置页布局对齐 OIDC 参考页；验证 go test、E2E、i18n.check。
- [x] 治理：i18n 有；缓存无持久缓存；数据权限无新增列表域；DI 无新增宿主依赖；密码与 TLS 策略在设计与单测中固化。
