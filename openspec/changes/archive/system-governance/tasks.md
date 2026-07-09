# Tasks

## Summary

- [x] 建立系统治理能力：操作日志、登录日志、在线用户、服务监控、系统 API 文档、系统信息、组件演示、参数设置、字典导入导出和宿主数据权限治理。
- [x] 建立审计与会话闭环：写操作和标记导出写操作日志，登录/登出写登录日志，登录创建在线会话，登出和强制下线删除会话，认证中间件同时校验 JWT 与会话存在性。
- [x] 建立宿主数据权限闭环：按角色`dataScope`解析有效范围，显式资源策略覆盖用户、文件、用户创建任务和在线会话，用户消息保持当前用户自隔离，全局治理资源明确排除。
- [x] 完成系统自描述和可观测页面：服务监控多节点快照、API 文档 iframe、动态 OpenAPI server URL、版本信息、组件演示和加载失败兜底。
- [x] 新增全局日志最长保留天数运行时参数 `sys.log.retentionDays`，默认 90 天，统一约束操作日志、登录日志、任务执行日志、在线会话和智能中心调用日志的最长存储时间。
- [x] 改进操作日志和登录日志前端删除交互：移除表格复选框和按 ID 勾选删除，改为日期范围选择对话框，支持"删除所有日志"全量清理入口。
- [x] 验证：历史实现覆盖`make db.init`、`make dao`、Go 单元测试、E2E、前端 typecheck、OpenSpec 严格校验和`lina-review`；日志保留验证覆盖默认值、非法值、host config 运行时读取、执行日志全局上限、在线会话边界、登录日志清理和调用日志清理；范围删除验证覆盖复选框移除、范围删除弹窗、全量删除、请求参数和翻译文案。
- [x] 治理：本归档压缩不修改运行时代码、HTTP API、数据库、缓存、数据权限、前端 UI、插件源码、运行时文案、语言包、`manifest/i18n`、`apidoc i18n JSON`、开发工具入口或生产构建；非 owner 能力已迁移为`design.md`交叉影响摘要。
- [x] 移除宿主内建健康检查：删除匿名`GET /api/v1/health`、DTO、控制器、路由装配、`health.timeout`配置、`config.Service.GetHealth`和 apidoc 翻译资源；集群验证改用已认证`GET /api/v1/system/info`。
- [x] 统一宿主对象存储：新增`internal/service/storage`中立组件，将文件中心和插件对象存储的重复本地磁盘读写收敛到统一`storage.Service`；文件中心和插件`Storage()`领域边界保持独立。
- [x] 验证：`openspec validate`通过；Go 测试覆盖 config、httpstartup、middleware、role、cluster、apidoc、file、storage、plugin capabilityhost/wasm 等包；静态检索确认`GetHealth`、`HealthConfig`、`health.timeout`、`file.Storage`、`file.NewLocalStorage`无残留。

## Feedback

- [x] **FB-1**: 合并日志保留参数 SQL，并移除本次新增的缺行默认值兼容分支
- [x] **FB-2**: 合并登录日志插件保留清理索引 SQL
- [x] **FB-3**: 删除弹窗补充日期区域间距和删除所有日志选择
