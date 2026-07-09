# Tasks

## Summary

- [x] 将运行时数据库支持统一为`PostgreSQL 14+`，移除`SQLite`与`MySQL`驱动注册、方言入口、DDL转译、启动降级和隐式兼容路径。
- [x] 保留`pkg/dialect`公共边界，围绕`PostgreSQL`提供数据库准备、版本查询、表元数据查询、只读SQL分类和错误分类能力。
- [x] 将宿主与插件SQL源统一为受治理的`PostgreSQL 14+`语法子集，并将`init`、`mock`纳入方言入口和显式确认治理。
- [x] 将`sys_online_session`、`sys_locker`、`sys_kv_cache`稳定为`PostgreSQL`持久表，通过读取时过期判断、TTL清理、锁过期抢占和CAS递增自然收敛。
- [x] 交叉影响：集群、插件缓存、插件manifest、发布镜像、README/i18n、项目初始化、角色授权和字典复用读取已迁移为`design.md`摘要，完整规范由对应owner分组或`openspec/specs`承载。
- [x] 验证：完成`PostgreSQL`编译、单元测试、工具链验证、OpenSpec校验和静态扫描，确认非归档路径不再保留受支持的`SQLite`入口。
- [x] 治理：不新增运行时翻译资源、插件`manifest/i18n`或`apidoc i18n JSON`；不新增缓存策略，只删除`SQLite`特殊分支；数据库压缩不影响运行时代码、HTTP API、数据权限、前端UI、插件源码、运行期依赖或生产构建。
