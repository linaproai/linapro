# Manifest Resources

该目录用于存放源码插件样例自有的 `SQL` 生命周期资源。

## 目录内容

- `sql/001-plugin-demo-source-records.sql`：安装时创建插件自有示例表，并初始化一条种子记录
- `sql/uninstall/001-plugin-demo-source-records.sql`：卸载时在用户选择清理存储数据后删除该插件自有示例表

菜单声明保存在 `plugin.yaml` 中；`SQL` 资源仅用于插件自有数据的生命周期变更。
