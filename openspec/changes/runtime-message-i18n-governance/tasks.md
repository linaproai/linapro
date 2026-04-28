## 1. 基础治理与扫描

- [x] 1.1 根据 `runtime-message-i18n-audit.md` 固化运行时文案分类规则，明确 `UserMessage`、`UserArtifact`、`UserProjection`、`DeveloperDiagnostic`、`OpsLog`、`UserData` 的判定标准和 allowlist 格式
- [x] 1.2 新增后端硬编码运行时文案扫描脚本，覆盖 `gerror.New*`、`gerror.Wrap*`、`Reason/Message/Fallback`、导出表头、状态文本和插件桥接错误构造
- [x] 1.3 新增前端硬编码运行时文案扫描脚本或 ESLint 规则，覆盖 `title`、`label`、`placeholder`、模板文本、`message.*`、`notification.*`、`Modal.confirm`
- [x] 1.4 把扫描命令接入本地验证入口，并确保注释、测试 fixture、用户示例数据、技术单位和英文运维日志不会误阻断
- [x] 1.5 为 `zh-CN`、`en-US`、`zh-TW` 运行时语言包补齐新增 key 的缺失翻译校验，确保宿主和插件资源分别校验

## 2. 后端结构化错误基础设施

- [x] 2.1 新增运行时消息错误模型和构造辅助能力，支持 `errorCode`、`messageKey`、`messageParams`、英文 fallback 和 GoFrame `gcode` 语义
- [x] 2.2 更新统一响应中间件，识别结构化错误并输出本地化 `message`、稳定 `errorCode`、`messageKey` 和 `messageParams`
- [x] 2.3 保留现有 `LocalizeError` 作为旧错误兜底，但禁止新增业务路径继续使用中文自由文本错误
- [x] 2.4 为结构化错误渲染增加单元测试，覆盖 `zh-CN`、`en-US`、`zh-TW`、缺失 key fallback 和参数格式化
- [x] 2.5 确认错误本地化复用现有运行时翻译缓存，不在单个错误热路径构建完整语言包

## 3. 宿主业务错误与导入导出清理

- [x] 3.1 清理用户模块错误、用户导入失败原因、用户导出表头、用户导入模板和性别/状态枚举文案
- [x] 3.2 清理字典类型、字典数据、组合导出和字典导入中的业务错误、sheet 名、表头、状态文本和失败原因
- [x] 3.3 清理系统参数、配置导入导出、文件管理、用户消息、角色、菜单和通知模块中的用户可见错误
- [x] 3.4 清理定时任务、任务处理器、任务日志、任务元数据、缓存、分布式锁和运行时参数相关用户可见错误
- [ ] 3.5 清理插件生命周期、源码插件升级、自动启用、动态插件运行时、前端资源解析和插件治理结果中的固定中文消息
- [ ] 3.6 为用户、字典、配置、插件和任务相关新增运行时语言键补齐 `zh-CN`、`en-US`、`zh-TW` 资源
- [ ] 3.7 为导入导出实现请求级批量本地化上下文，确保行循环内只做缓存查找和参数格式化

## 4. 插件平台与源码插件清理

- [ ] 4.1 清理 `pkg/pluginbridge` 中 bridge codec、host call codec、host service codec 的中英混排错误，改为稳定错误码和英文开发者诊断
- [ ] 4.2 清理 `pkg/pluginfs`、`pkg/plugindb`、插件 data host、WASM host service 和 catalog 校验中的用户可见错误契约
- [ ] 4.3 更新动态插件 guest JSON 错误约定，支持 `errorCode`、`messageKey`、`messageParams` 和本地化 `message`
- [ ] 4.4 清理 `demo-control`、`content-notice`、`org-center`、`monitor-loginlog`、`monitor-operlog`、`plugin-demo-source`、`plugin-demo-dynamic` 后端业务错误和导出文案
- [ ] 4.5 把插件新增运行时文案写入各插件自己的 `manifest/i18n/<locale>/*.json`，禁止把插件运行时 key 集中写入宿主语言包
- [ ] 4.6 更新插件相关单元测试，断言错误码、翻译键和本地化展示，而不是断言固定中文错误文本

## 5. 前端工作台与插件前端清理

- [x] 5.1 更新 `apps/lina-vben/apps/web-antd/src/api/request.ts`，请求错误展示优先使用 `messageKey/messageParams`，再 fallback 到后端 `message`
- [x] 5.2 清理服务器监控页静态标签、tooltip、空状态和时间单位，全部改为 `$t` 或运行时语言包
- [x] 5.3 清理在线用户页面查询表单和表格列硬编码中文
- [x] 5.4 扫描并清理插件前端页面中的运行时可见硬编码中文，保留用户数据和测试 fixture
- [x] 5.5 补齐前端静态语言包和宿主运行时语言包中的 `zh-CN`、`en-US`、`zh-TW` 翻译键
- [x] 5.6 确认 `i18n.enabled=false` 时仍按默认语言展示，且语言切换隐藏逻辑不受本次错误展示改造影响

## 6. 测试与 E2E

- [ ] 6.1 增加后端单元测试，覆盖结构化错误、运行时语言资源缺失、导入失败原因和导出表头本地化
- [ ] 6.2 增加插件平台单元测试，覆盖 bridge/host service 错误码、英文开发者诊断和管理端本地化映射
- [ ] 6.3 增加前端单元测试，覆盖请求拦截器 `messageKey` 优先级和 fallback 行为
- [ ] 6.4 创建 `hack/tests/e2e/i18n/TC0131-structured-error-localization.ts`，验证同一后端业务错误在 `zh-CN`、`en-US`、`zh-TW` 下展示不同语言且错误码稳定
- [ ] 6.5 创建 `hack/tests/e2e/i18n/TC0132-localized-export-artifacts.ts`，验证用户或字典导出表头、状态和导入失败原因按当前语言输出
- [ ] 6.6 创建 `hack/tests/e2e/i18n/TC0133-runtime-hardcoded-copy-regression.ts`，验证服务器监控页和在线用户页切换语言后没有残留硬编码中文
- [ ] 6.7 运行后端相关 `go test`、前端类型检查/构建、硬编码文案扫描和新增 E2E 用例

## 7. 文档、验收与审查

- [x] 7.1 更新宿主和前端 i18n README，说明运行时错误、导入导出、插件文案和硬编码扫描治理规则，并同步维护英文与中文 README
- [x] 7.2 更新 `runtime-message-i18n-audit.md`，记录实施后剩余 allowlist、已清理模块和后续观察项
- [x] 7.3 运行 `openspec status --change runtime-message-i18n-governance`，确认提案、设计、规范和任务状态一致
- [ ] 7.4 调用 `/lina-review` 对实现、规范符合性、i18n 资源完整性和测试覆盖进行审查

## Feedback

- [x] **FB-1**: 去掉 `bizerr` 自定义整型业务错误码，接口 `code` 回归 GoFrame 类型错误码，并按模块命名空间治理业务语义码
- [x] **FB-2**: 增加调用端可见接口错误必须使用 `bizerr` 的项目规范和 `lina-review` 检查，并修正当前实现中的明确违规路径
- [x] **FB-3**: 清理内置运行时参数和公共前端参数注册表中的展示文案硬编码，并统一默认值事实源
- [x] **FB-4**: 为 `bizerr.Code` 增加结构化元数据读取与错误匹配方法，并拆分 `bizerr` 实现文件职责
- [x] **FB-5**: 在设计文档中明确 i18n JSON 按 locale 目录、运行时语义域和 apidoc 子目录分类，并规划后续资源目录重组
- [x] **FB-6**: 落地宿主与插件 i18n JSON 目录重组，更新 loader、动态插件打包、文档和测试
- [x] **FB-7**: 清理 `apps/lina-core/internal/service/plugin/internal/runtime/artifact.go` 中残留的中文硬编码错误文案
- [x] **FB-8**: 将运行时 i18n 检查从临时 Python 脚本迁移为`hack/tools/runtime-i18n`下的 Go 工具
- [x] **FB-9**: 为`hack/tools`下每个工具目录补齐中英文使用说明文档
- [x] **FB-10**: 修正 `org-center` 插件初始化 SQL 中部门编码唯一性、mock 用户关联和关联表反向查询索引问题
- [ ] **FB-11**: 为 `make init` 增加可选重建数据库参数，将默认数据库名改为 `linapro`，并在初始化 SQL 中显式幂等创建数据库
- [x] **FB-12**: 修复 `plugin-demo-dynamic` 独立静态页内置多语言文案的问题，改为复用插件运行时 i18n 资源
- [x] **FB-13**: 清理 `apps/lina-core/internal/cmd` 中 CLI 和数据库初始化诊断错误的中文硬编码，统一改为英文开发者诊断并更新单元测试断言
- [x] **FB-14**: 缩短英文偏好设置抽屉 tab 展示文案，避免 `Appearance` 和 `Shortcut Keys` 超出按钮背景
- [x] **FB-15**: 梳理宿主与源码插件 seed/mock 数据边界，将演示测试数据迁移或补齐到各自 `manifest/sql/mock-data` 目录