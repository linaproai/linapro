# Tasks

## Summary

- [x] 建立多租户认证主线：`tenantcap`接缝、Pool 隔离模型、`bizctx.TenantId`、固定租户解析链、`pre_token`租户选择、租户切换重签、token revoke、平台管理员 impersonation 和`clientType`会话事实源。
- [x] 建立租户数据治理：1:N membership、平台/租户角色边界、租户配置 fallback、文件/缓存/审计/通知/任务/用户消息隔离、Redis session hot state 和 PostgreSQL 在线会话投影。
- [x] 建立插件租户治理：`scope_nature`、`install_mode`、租户级启用状态、新租户自动启用策略、租户插件自服务、LifecycleGuard 否决聚合、force 通道和启动期一致性校验。
- [x] 收敛历史设计：删除未实现的 resolver 配置表、生命周期 outbox、租户配额占位和租户级 i18n override；`rootDomain`和解析链固定在代码中，跨插件租户清理必须另立可靠编排设计。
- [x] 验证：历史实现记录包含`make db.init`、`make db.mock`、DAO/Controller 生成、后端单元测试、插件 service 测试、前端登录/工作台改造验证、多租户`E2E` TC0178-TC0220、集群一致性验证、关键查询`EXPLAIN`和单租户回归。
- [x] 治理：历史记录覆盖`i18n`中英文资源完整性、`manifest/i18n`和`apidoc`边界、缓存一致性、数据权限、显式 DI、跨平台开发入口、文档双语镜像、OpenSpec 严格校验和`lina-review`审查。
- [x] 交叉影响：组织、日志、通知、调度、插件运行时、插件 host service、配置/字典、工作台、数据库引导、`E2E`和后端一致性等非`user-auth` owner 能力已迁移到`design.md`交叉影响摘要。
- [x] 角色权限树模式切换：独立提交状态与展示勾选分离，仅按钮授权在独立/父子联动切换后不得扩大`menuIds`；覆盖组件单测、E2E TC005 与 issue #82。
- [x] FB-1：Make command smoke 前端夹具跟随`vite.js`入口，避免误触发`pnpm install`；属 CI 夹具修复，无运行时 i18n/缓存/数据权限影响。
