# 实施任务清单

## 1. 数据库支持收敛

- [x] 将运行时数据库支持统一为`PostgreSQL 14+`，移除`SQLite`与`MySQL`驱动注册、方言入口和隐式兼容路径，并确保不支持前缀快速失败。
- [x] 保留`pkg/dialect`公共边界，只围绕`PostgreSQL`实现数据库准备、版本查询、表元数据查询、只读SQL分类和错误分类能力。

## 2. SQL源与初始化流程

- [x] 将宿主与插件SQL资源统一到受治理的`PostgreSQL 14+`语法子集，清理`MySQL`语法残留，并停止依赖`SQLite`运行时转译链路。
- [x] 将`make init`与`make mock`统一纳入方言入口治理：仅`PostgreSQL PrepareDatabase`负责建库或重建，`mock`只加载已初始化数据库。
- [x] 将`sys_online_session`、`sys_locker`、`sys_kv_cache`稳定为持久表，并保持TTL自然过期、锁过期抢占和CAS递增语义。

## 3. 集群、缓存与插件协同

- [x] 将集群部署与coordination配置收敛到`PostgreSQL`路径，不再保留`SQLite`单机锁定、启动告警或coordination降级分支。
- [x] 更新插件缓存与插件数据治理规范，确保`sqltable`缓存、表元数据查询和只读SQL分类仅围绕`PostgreSQL`实现。

## 4. 工具链、CI与发布

- [x] 清理`SQLite`专属E2E、CI smoke、脚本、package scripts和workflow输入，统一main、nightly、release与共享验证模板为`PostgreSQL-only`。
- [x] 更新镜像运行配置、默认配置模板、开发工具说明和发布构建门禁，使默认开发与交付路径只表达`PostgreSQL`依赖。

## 5. 文档与规范同步

- [x] 同步更新根目录与`apps/lina-core`双语README、测试文档和数据库支持说明，统一为`PostgreSQL-only`口径。
- [x] 维护`README.md`与`README.zh-CN.md`双语镜像命名与内容同步规则，消除下划线中文README治理残留。
- [x] 将`project-setup`、`database-dialect-abstraction`、`database-bootstrap-commands`、`cluster-deployment-mode`、`sql-source-syntax`、`volatile-table-bootstrap`、`plugin-cache-service`和`release-image-build`等规范更新到最终状态。

## 6. 验证与审查

- [x] 完成`PostgreSQL`编译、单元测试、工具链验证、OpenSpec校验和静态扫描，确认非归档路径不再保留受支持的`SQLite`入口。
- [x] 记录`i18n`与缓存一致性影响：不新增运行时翻译资源，不新增缓存策略，只删除`SQLite`特殊分支。
- [x] 完成反馈修复与归档前审查，确认聚合文档、规范和原始归档清理结果一致。
