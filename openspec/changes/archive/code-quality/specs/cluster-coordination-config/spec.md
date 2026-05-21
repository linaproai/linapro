## MODIFIED Requirements

### Requirement: 非 PostgreSQL 数据库链接必须在 coordination 启动前失败

系统仅支持 PostgreSQL 运行时数据库。`sqlite:`、`mysql:` 或未知数据库链接 MUST 在方言解析阶段失败，不得进入 Redis coordination 探活或业务启动流程。
