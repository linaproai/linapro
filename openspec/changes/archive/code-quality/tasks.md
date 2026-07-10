# Tasks

## Summary

- [x] 建立代码质量治理主线：API 合同、响应 DTO、毫秒时间戳、GoFrame ORM/软删除、事务、`panic`、源码可读性、显式依赖注入、宿主运行能力、启动 SQL 效率、前端运行性能和 Go 单测效率。
- [x] 完成关键运行质量整改：用户/角色/菜单删除和角色授权事务化，用户/角色批量删除，健康检查，GoFrame 优雅停机复用，受保护上传访问，调度默认时区配置，关键索引补齐，菜单子树内存遍历和 SQL debug 默认关闭。
- [x] 完成运行期依赖治理：Controller、Middleware、Service、源码插件、host service adapter 和 WASM host service 通过构造函数逐项接收依赖；缓存敏感组件复用启动期共享实例或共享后端；初始化和注册 API 对可预期失败返回`error`。
- [x] 完成启动效率治理：同一启动编排复用 catalog、integration 和 job snapshots；插件无差异同步不写库、不打开空事务、不重复 post-read；内置任务使用声明投影注册；启动摘要替代默认 SQL 明细。
- [x] 完成测试效率治理：Go 主单测路径保留`-race`；真实 dynamic Wasm 执行收敛到 smoke；普通插件测试使用 synthetic artifact、fake executor 和轻量 fixture；`linactl test.go`输出测试计划和耗时摘要。
- [x] 完成 OpenSpec 归档文档治理：建立信息分层、能力 owner 映射、分阶段压缩和验证报告规则；以`plugin-framework`为样板验证后批量压缩全部高体量分组和低体量重复清理；归档体量从约`3.5M`降至约`1.5M`，归档`spec.md`从`277`降至`116`，跨分组重复能力从`61`降至`0`，完全重复主规范副本从`16`降至`0`。
- [x] 完成 E2E 质量审查增强：在`.agents/rules/testing.md`中新增 E2E 质量审查结果级要求，覆盖触发范围、覆盖有效性、断言有效性、稳定性、严重性分级和审查输出证据；不绑定具体定位器、封装模式或编码实现方式。
- [x] 反馈`FB-1`~`FB-3`：API 输入 DTO 标签、实现范围和路由地址漂移；根因：接口治理整改扩大到不必要范围且参数标签不统一；处理：统一`json`参数标签，保持既有路由地址，撤回越界模块开关；验证：API/前端调用与 OpenSpec 校验通过。
- [x] 反馈`FB-4`~`FB-8`：运行时配置错误、关闭错误日志、ctx 传递和`panic` allowlist 检查边界；根因：可恢复错误与诊断路径混用；处理：显式返回错误、补充调用上下文、迁移 allowlist 检查并降低扫描耦合；验证：静态检查和相关 Go 测试通过。
- [x] 反馈`FB-9`~`FB-17`：SQL 注释、上传路由、shutdown、HTTP 文件拆分、健康超时、配置/中间件接口、插件安装 SQL 和字典 E2E；根因：治理细节分散在启动、SQL、接口和测试边界；处理：按 owner 模块收口并补充保护；验证：`make db.init`、E2E、OpenSpec 和审查通过。
- [x] 反馈`FB-18`~`FB-24`：启动重复 SQL、内置任务幂等、startup snapshot、时间戳响应和标准库时间类型；根因：启动阶段反复扫描、无差异仍写库、API 时间边界未落地；处理：共享快照、差异同步、投影复用、`pkg/apitime`和 DAO 标准时间配置；验证：启动测试、Go 测试、前端适配和 apidoc i18n 资源同步通过。
- [x] 反馈`FB-25`~`FB-27`：文件顶部注释、`linactl`命名和根目录共享实现堆积；根因：文件职责和开发工具子组件边界不清；处理：补充职责注释、建立`command_<command>.go`约束、迁移共享实现到`internal/<component>/`；验证：静态检索、命令 smoke 和审查通过。
- [x] 反馈`FB-28`~`FB-45`：服务构造、WASM host service、插件 host service、源码插件 controller、tenant/bizctx/sourceupgrade 和聚合依赖结构体；根因：运行期依赖通过隐式`New()`、包级默认实例或聚合结构体隐藏；处理：启动期显式注入、host service 目录发布、能力契约收敛和规则文件/审查项更新；验证：静态扫描、host/source-plugin/WASM 相关测试和`lina-review`通过。
- [x] 反馈`FB-46`~`FB-52`：重复 controller 声明、运行时初始化`panic`、源码插件注册`panic`、过长 pluginhost 文件、半初始化 tenant service、panic allowlist 路径和 API 包兼容别名；根因：拆分与 DI 迁移后遗留旧入口；处理：删除重复声明，错误返回上移，拆分职责文件，移除兼容别名和常量转发；验证：编译、静态扫描、panic allowlist 测试和 OpenSpec 校验通过。
- [x] 反馈`FB-53`~`FB-55`：API 响应暴露实体、源码插件 DTO 命名和聚合测试归属；根因：响应合同受数据库实体驱动且插件测试 ownership 不闭环；处理：独立响应 DTO、逐字段映射、插件 API DTO 重命名、测试迁回各插件目录；验证：API 静态验证、插件测试和审查通过。
- [x] 验证：历史实现运行过`make db.init`、`make dao`、`make ctrl`、Go 单元测试、E2E、前端适配验证、静态扫描、OpenSpec 校验和`lina-review`；本次压缩仅需 OpenSpec 严格校验、重复能力扫描和 Markdown 空白检查。
- [x] 治理：本归档压缩不修改运行时代码、HTTP API、数据库、缓存、数据权限、前端 UI、插件源码、运行时文案、语言包、`manifest/i18n`、`apidoc i18n JSON`、开发工具入口或生产构建；`i18n`影响仅限中文 OpenSpec 历史文档。
- [x] 官方插件 panic 治理解耦：宿主精确白名单 + `backend/plugin.go` `init` fail-fast AST 自动放行；`PLUGIN_TENANT_PROVISIONING_POLICY_INVALID`文案改为 multi-tenant governance 语义并同步`en-US/error.json`；错误码不变。验证：`cmd` panic 治理测试、plugin 服务测试、静态检索无`linapro-tenant-core`品牌文案、OpenSpec 校验。
- [x] 治理补充：`i18n`有影响（错误源文案与语言包）；缓存/数据权限/跨平台工具/DI 无新增运行期依赖；无用户可观察页面变更，无需 E2E。
