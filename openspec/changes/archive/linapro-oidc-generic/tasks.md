# Tasks

## Summary

- [x] 交付 managed 源码插件 `linapro-oidc-generic`：骨架/依赖/菜单、settings、OIDC Discovery+PKCE+JWKS 登录、handoff 回跳、catalog/ownership、`auth.login.after` 入口、双语文档与 i18n。
- [x] 验证：插件 `go test`/`go vet`、E2E TC001（菜单、未配置 fail-closed、统一按钮样式）、`make i18n.check`、OpenSpec strict。
- [x] FB-1: 企业 OIDC 登录入口迁到工作台统一 Vben 全宽按钮与响应式布局；根因：展示未对齐工作台认证样式；验证 typecheck/build、E2E TC001 4/4、宿主 TC006 5/5。
- [x] FB-2: 菜单改名「OIDC 设置」/「OIDC Settings」，设置页 `p-4`+Card 对齐 Google/Discord；根因：菜单文案与表单边距不一致；验证 go test、E2E TC001、i18n.check。
- [x] 治理：i18n 有（plugin/menu/error 双语）；缓存有（Discovery 15m、JWKS 1h 进程内）；数据权限无新增列表域；DI 无新增宿主运行期依赖；开发工具无跨平台脚本变更。
