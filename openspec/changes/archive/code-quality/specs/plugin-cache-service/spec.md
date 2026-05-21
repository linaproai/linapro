## MODIFIED Requirements

### Requirement: 动态插件通过基于易失性缓存表的授权命名空间访问宿主分布式缓存

系统 SHALL 为动态插件提供受治理的缓存服务。当前单机默认后端为基于 PostgreSQL SQL 表的实现，集群模式使用 coordination KV backend。所有后端 SHALL 被视为有损缓存，不得作为可靠业务状态的权威来源。

### Requirement: 插件缓存递增在缓存存活期间必须是原子的

系统 SHALL 保证同一插件缓存键的 `incr` 在共享数据库和缓存表存活期间线性递增。`incr` 实现 SHALL 使用方言中性的 CAS 重试模式，不得使用数据库专用的原子自增技巧。

### Requirement: 插件缓存过期清理必须避免热路径全表扫描

读取插件缓存时，系统 SHALL 仅执行只读查询。过期清理必须由后端在读取结果上的过期过滤和后台批量清理处理。`sqltable` 后端必须提供后台批量清理能力。
