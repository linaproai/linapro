# Manifest 资源

该目录存放 `monitor-operlog` 的安装与卸载 SQL 资源。

## 内容

- `sql/001-monitor-operlog-schema.sql`：创建操作日志表并初始化相关字典
- `sql/uninstall/001-monitor-operlog-schema.sql`：在卸载且选择清理数据时删除相关字典并移除操作日志表
