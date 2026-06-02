# Tasks

## Summary

- [x] 交付分布式基础设施两阶段演进：先以`sys_locker`、`sys_cache_revision`和`sys_kv_cache`补齐基础锁、leader election、cachecoord 和 host KV cache，再将真实集群模式收敛到 Redis coordination provider。
- [x] 建立 Redis coordination 配置、启动探活、provider 抽象、key namespace、lock/KV/revision/event/health、fake provider、真实 Redis 可选集成测试和 cluster smoke。
- [x] 完成 leader election、Master-Only 任务门禁、插件 per-plugin 锁、插件 host lock、coordination KV backend、auth revoke、`pre_token`、session hot state、permission/runtime-config/plugin-runtime revision/event 和 system info 诊断接入。
- [x] FB-1: Redis provider 包边界收敛到`coordination/internal/redis`；处理：`kvcache`只保留 coordination KV 适配层，业务缓存层不直接表达 Redis backend；验证：coordination、cachecoord、kvcache、Wasm host service 测试通过。
- [x] FB-2: 项目介绍文档误将`OpenSpec`描述为内置必需工作流；处理：调整为可选但推荐并说明框架支持；验证：README 中英文同步和 OpenSpec 校验通过。
- [x] FB-3: `lina-openspec-archive-consolidate`默认读取边界会重复聚合既有聚合目录；处理：默认只读取日期前缀原始归档，显式指定时才处理非日期目录；验证：技能文档扫描、OpenSpec 校验和 diff 检查通过。
- [x] FB-4: Main CI 缺少活跃 OpenSpec 完成状态检查；处理：新增固定版本 OpenSpec list 判定 job；验证：YAML 解析、脚本失败路径和严格校验通过。
- [x] FB-5/FB-6/FB-7: 路由注册控制器初始化、sysinfo 构造入口和 plugin runtime 服务接口过宽；处理：变量化路由注册，统一`sysinfo.NewV1`诊断依赖注入，拆分 runtime 窄接口；验证：相关 Go 测试、OpenSpec 校验、diff 检查和审查通过。
- [x] FB-8/FB-9/FB-10: Redis CI、cluster smoke、维护注释和单机测试边界不足；处理：新增真实 Redis service 集成测试、nightly cluster smoke、脚本/YAML 注释，并隔离缺少 Redis 时的常规测试；验证：YAML 解析、shell 语法、真实 Redis smoke、OpenSpec 校验和审查通过。
- [x] 关键验证：`go test`覆盖 config、coordination、cluster、locker、cachecoord、kvcache、auth、session、role、cron、plugin runtime、i18n、sysinfo、middleware、monitor-online 等相关包；真实 Redis 验证通过`LINA_TEST_REDIS_ADDR`显式启用；cluster smoke 验证 Redis 启动、primary、登录和 system info coordination 诊断。
- [x] 治理：实现阶段同步过新增错误码和 sysinfo apidoc i18n；前端未展示新增 coordination 字段时无需运行时 UI 语言包；数据权限边界保持在 PostgreSQL 投影和既有 tenantcap/datascope；缓存一致性以 Redis KV/revision/event/lock 为集群事实源，单机保持本地或 SQL 分支；本次归档压缩不修改运行时代码、API、数据库、前端、插件源码或 i18n 资源。
