## 1. 骨架与清单

- [x] 1.1 创建 `linapro-auth-ldap` 骨架：plugin.yaml（managed、依赖 extlogin-core）、go.mod、Makefile、plugin_embed、菜单 settings
- [x] 1.2 双语 i18n（plugin/menu/error/apidoc）+ embed 菜单单测
- [x] 1.3 更新 plugins README 与 extlogin-core 协议插件列举

## 2. Settings

- [x] 2.1 settings 服务：host/port/tls_mode/bind/base/filter/template/属性映射/display_name/allow_auto_provision 默认关；bind password 脱敏
- [x] 2.2 settings API GET/PUT + 设置页前端
- [x] 2.3 tls_mode=plain 非本机拒绝

## 3. LDAP 登录

- [x] 3.1 LDAP 客户端：连接、Search/Template 解析 DN、用户 bind、读属性；超时；密码不落日志
- [x] 3.2 `POST /portal/linapro-auth-ldap/login` → LoginByVerifiedIdentity(provider=ldap:default) → handoff JSON
- [x] 3.3 统一失败错误码；未配置 fail-closed

## 4. 装配与前端入口

- [x] 4.1 plugin.go：ProvideExternalIdentity、catalog、路由
- [x] 4.2 login slot 弹层：提交 portal login → completeExternalLoginFromHandoff

## 5. 测试与文档

- [x] 5.1 单元测试：配置校验、DN 模板、subject 抽取、auto-provision 默认 false、plain 非本机拒绝（mock dialer 可选）
- [x] 5.2 go test/vet；openspec validate --strict
- [x] 5.3 E2E TC001 菜单/入口骨架
- [x] 5.4 双语 README + 影响分析记录


## 实施记录

### 影响分析

| 域 | 判断 |
| --- | --- |
| i18n | 有：plugin/menu/error 双语 |
| 缓存 | 无持久缓存；LDAP 连接按请求建立 |
| 数据权限 | 无新增列表数据域；settings 管理权限 |
| 开发工具 | 无跨平台脚本变更 |
| 测试 | 单元：TLS/filter/escape/config；E2E TC001 骨架 |
| DI | 无新增宿主依赖；插件使用 host ExternalLogin + SysConfig；go-ldap 仅插件 go.mod |

### 验证

- `GOWORK=off go test ./...` 通过
- `openspec validate add-ldap-auth-plugin --strict`

## Feedback

- [x] **FB-1**：将目录登录入口与凭证弹层迁移到工作台统一`Vben`按钮、弹层和表单样式，并补齐响应式`E2E`验证

### FB-1 影响分析

| 域 | 判断 |
| --- | --- |
| 架构与插件边界 | 插件内前端展示适配；复用宿主公开的`@vben/common-ui`与主题`token`，不新增宿主契约、共享抽象或跨模块调用 |
| `i18n` | 有运行时展示影响但无资源变更；复用现有`en-US`/`zh-CN`登录键，表单`schema`运行时计算，`make i18n.check`无本插件新增告警 |
| 缓存一致性 | 无缓存读写或失效策略变化 |
| 数据权限 | 无数据查询、写入、存在性暴露或权限边界变化 |
| `API`与性能 | 无接口契约、请求次数或后端装配路径变化；保留原`portal login`与`handoff`流程 |
| 开发工具与跨平台 | 无持久脚本、配置或开发入口变化 |
| 测试 | 更新插件`POM`与`E2E TC001`，覆盖统一全宽按钮、移动弹层、双语文案、必填校验和敏感字段重置；回归宿主认证`TC006` |
| `DI` | 无新增运行期依赖或服务构造变化 |

### FB-1 验证

- `pnpm --filter @lina/web-antd typecheck`通过
- `pnpm --filter @lina/web-antd build`通过
- `make i18n.check`通过
- 插件`E2E TC001`在执行后已移除的临时`ESM`边界下通过（3/3）；宿主认证回归`TC006`通过（5/5）
- 桌面/移动、亮色/暗色截图审查通过，证据位于`temp/20260711/`
- `openspec validate add-ldap-auth-plugin --strict`通过

当前插件测试运行器缺少持久`ESM`包边界，直接加载插件测试会在宿主`fixture`执行前失败；本次测试已真实执行通过，但该既有运行器风险不属于本反馈的持久变更范围。

- [x] **FB-2**：授权登录目录下 LDAP 菜单改名为「LDAP 设置」，设置页布局与表单边距对齐 Google/Discord OIDC（`p-4` + Card + 统一 Ant/VBen 表单样式）
  - **根因**：侧栏菜单 i18n 仍为「LDAP 登录」/「LDAP Login」；设置页未包 `p-4` 外层，表单贴边，样式细节与 Google/Discord 参考页不一致
  - **修复**：`menu.json` zh-CN=`LDAP 设置`、en-US=`LDAP Settings`；`settings.vue` 对齐 `div.p-4` + Card + Form model/name 与统一表单样式；同步 embed 单测、E2E TC-1a2 双语精确断言、plugin.yaml 菜单 fallback 名
  - **验证**：`GOWORK=off go test`；插件 E2E TC001 4/4；`make i18n.check` 无本插件新增失败；`openspec validate --strict` 通过

### FB-2 影响分析

| 域 | 判断 |
| --- | --- |
| 架构与插件边界 | 仅插件内菜单文案与设置页布局；不改宿主契约 |
| `i18n` | 有：`manifest/i18n/{en-US,zh-CN}/menu.json` 菜单标题；页面文案键未改 |
| 缓存一致性 | 无；菜单翻译依赖运行时 i18n，重建宿主后生效 |
| 数据权限 | 无 |
| API 与性能 | 无接口契约变更 |
| 开发工具与跨平台 | 无脚本变更 |
| 测试 | embed 单测 + E2E TC001a/a2；登录入口回归 TC001b/c |
| `DI` | 无 |

