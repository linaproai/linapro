## E2E 冲突治理记录

本记录用于沉淀完整 E2E 回归中发现的测试间共享状态冲突，并作为后续新增用例的检查清单。本变更不新增产品运行时文案，不修改业务 API 或数据库结构；i18n 相关调整仅约束测试断言方式，继续验证运行时语言包、接口文档和页面文案的语义正确性。

## 已识别冲突类型

| 冲突类型 | 代表用例 | 触发原因 | 治理方式 |
| --- | --- | --- | --- |
| 插件生命周期共享状态 | `TC0066-source-plugin-lifecycle.ts`、`TC0128-traditional-chinese-plugin-pages.ts`、`TC0129-traditional-chinese-apidoc.ts` | 安装、启用、卸载、同步插件会改变同一组插件表、菜单投影、运行时资源与插件包产物 | 通过 `pluginLifecycle`、`filesystemArtifact`、`runtimeI18nCache` 分类进入串行池；依赖源码插件的只读用例统一走 `fixtures/plugin.ts` 幂等准备 |
| 运行时 i18n 缓存版本刷新 | `TC0124-runtime-i18n-etag.ts`、`TC0107-runtime-i18n-switch.ts` | 其他插件或语言包操作可能刷新运行时 bundle version，导致条件请求返回合法的 `200 + 新 ETag` | ETag 用例改为协议语义断言，接受 `304` 或合法刷新响应，并将缓存相关用例声明为 `runtimeI18nCache` |
| 系统参数全局变更 | `TC0102-login-page-presentation.ts`、`TC0080-public-frontend-config.ts`、调度 shell 开关相关用例 | 测试会修改登录页布局、公共前端配置、shell 执行开关等全局参数 | 使用 `systemConfig` 分类进入串行池，测试结束后恢复原值 |
| 字典共享数据变更 | `TC0012-dict-type-crud.ts`、`TC0054-dict-type-import-upload.ts`、`TC0059-dict-type-cascade-delete.ts` | 字典类型和字典数据会影响全局下拉、标签和导入导出结果 | `settings/dict` 目录声明 `dictionaryData` 分类，新增高风险字典修改用例必须串行或提供有理由并行例外 |
| 菜单与权限矩阵变更 | `TC0060-menu-crud.ts`、`TC0063-auth-menu.ts`、插件权限治理用例 | 菜单、角色菜单、按钮权限和插件生成菜单会改变共享权限矩阵 | `iam/menu` 与插件目录声明 `permissionMatrix`，且插件菜单同步额外声明 `pluginLifecycle` |
| 共享日志数据增长 | `TC0116-english-built-in-governance-data-localization.ts` | 审计日志、登录日志和调度日志会被其它用例持续写入，固定读取首页文本容易被并行写入影响 | 将用例声明为 `sharedDatabaseSeed` 串行；业务状态断言改用 API 中的 `status`、`operType`、`title`、`operSummary` 稳定字段 |
| 依赖本地化 UI 文本计算业务状态 | `TC0021-user-dept-tree-count.ts` | 从树节点展示文本反推部门统计，语言或文案变化会影响业务断言 | 改为通过部门树 API 的稳定 `id=0` 节点和数值 counter 断言业务状态，UI 文案仅作为展示行为验证 |
| 下拉/弹层交互抢占 | `TC0024-dept-leader-select.ts` | 直接点击选择器时可能被遮罩、焦点或上一次弹层状态影响 | 改为聚焦输入并等待确定性下拉状态，减少跨用例残留弹层影响 |
| 插件页面前置条件隐式依赖 | `TC0108-english-runtime-page-audit.ts`、内容通知相关用例 | 测试假设其他文件已安装或启用 `content-notice` 等源码插件 | 统一调用 `ensureSourcePluginEnabled`，由 fixture 幂等安装、启用、刷新投影并按需加载 mock SQL |

## 新增用例检查清单

- 新用例是否安装、启用、禁用、卸载、上传、同步或升级插件？如果是，必须进入串行池并声明 `pluginLifecycle`，涉及插件包或运行时插件产物时同步声明 `filesystemArtifact`。
- 新用例是否修改系统参数、公共前端配置、字典、菜单、角色权限或按钮权限？如果是，必须声明对应隔离类别，并确保测试结束后恢复原值或清理自建数据。
- 新用例是否依赖源码插件页面、API、菜单或 mock 数据？如果是，应通过 `fixtures/plugin.ts` 或同类共享 helper 准备前置条件，不允许依赖其他测试先运行。
- 新用例是否验证缓存、ETag 或运行时语言包？如果是，必须验证条件请求确实带上前置条件，并接受 `304` 或合法 `200 + 新 ETag + body`。
- 新用例是否在业务断言中读取本地化 UI 文本？如果是，应改用稳定 ID、code、labelKey、permission key、API counter 或响应字段完成业务判断，展示文案单独断言。
- 新用例是否需要保留并行执行但命中高风险启发式？如果确认安全，必须添加 `parallelIsolationAllowlist`，写明分类和原因；否则校验器应阻断。

## 验收证据

- `pnpm test:validate` 校验命名、目录归属、manifest 引用、隔离分类和高风险模式。
- `node scripts/run-suite.mjs module i18n --list` 验证 module 模式仍按同一并行/串行边界拆分，并输出隔离类别摘要。
- 受影响 E2E 子集应至少覆盖运行时 i18n ETag、插件生命周期、组织部门树计数和内容通知依赖场景。
- 完整回归应通过 `pnpm test` 记录并行阶段、串行阶段和失败项；若环境未启动或外部依赖不可用，应在验收材料中明确记录。

## 回归观察

- 第二轮完整回归中，并行阶段 `102/102` 通过，证明新增串行边界能将已识别的共享日志、运行时缓存和插件生命周期用例移出并行池。
- 串行阶段继续执行到设置/字典模块前，插件生命周期、运行时 i18n、菜单权限、监控日志、组织树计数和调度大部分高风险路径均已通过。
- 回归后段出现 `ERR_CONNECTION_REFUSED` 级联失败，经 `make status` 确认为前后端 dev server 中途退出；重启服务后抽样复验设置/字典尾段通过。该问题属于长回归环境稳定性风险，建议后续 CI 使用受监督的服务进程或生产构建服务运行完整回归，避免 dev server 退出造成大量噪声失败。
