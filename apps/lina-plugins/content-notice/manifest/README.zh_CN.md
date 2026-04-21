# Manifest 资源

该目录存放 `content-notice` 的安装与卸载 SQL 资源。

## 内容

- `sql/001-content-notice-schema.sql`：创建通知公告表并初始化相关字典
- `sql/uninstall/001-content-notice-schema.sql`：在卸载且选择清理数据时删除相关字典并移除公告表
