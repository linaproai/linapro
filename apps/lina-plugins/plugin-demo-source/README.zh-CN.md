# plugin-demo-source

`plugin-demo-source` 是 `LinaPro` 的源码插件样例，用来展示一个在仓库内开发、由宿主发现、并通过插件管理页显式安装后挂载到默认工作台的完整插件闭环。

## 目录结构

```text
plugin-demo-source/
  plugin.yaml
  plugin_embed.go
  backend/
    api/
    internal/
      controller/
      service/
      dao/
      model/do/
      model/entity/
    hack/config.yaml
    plugin.go
  frontend/
    pages/
  manifest/
    sql/
    sql/uninstall/
```

## 样例覆盖点

- `manifest/sql/` 下的安装 SQL 会创建插件自有表`plugin_demo_source_record`
- `manifest/sql/mock-data/`下的`mock`SQL会提供本地样例数据使用的可选演示记录
- `manifest/sql/uninstall/` 下的卸载 SQL 会在用户确认清理存储数据时删除该插件自有表
- `frontend/pages/sidebar-entry.vue` 中的示例页面可以对插件自有表执行增删查改，并支持附件上传与下载
- 插件自有附件文件存放在宿主上传目录下的 `plugin-demo-source/` 命名空间中
- 禁用插件时仅隐藏菜单和路由，不清理数据表数据和已存储文件
- 卸载插件时宿主会弹窗，让用户选择是否同时清理插件自有数据表数据和存储文件

## 清单范围

`plugin.yaml` 负责保存插件元数据、菜单声明和按钮权限。页面与 `SQL` 资源仍然通过目录约定发现，而不是在元数据中重复维护。

## 后端接入

- 在 `backend/` 中实现插件后端入口
- 将业务逻辑保留在 `backend/internal/service/` 下
- 插件访问数据库时，将本地 ORM 生成工件维护在 `backend/internal/dao` 与 `backend/internal/model/{do,entity}` 下
- 通过宿主构建使用的源码插件注册入口显式接线安装、启用、禁用和卸载生命周期
- 将插件自有清理逻辑保留在插件服务中，便于在卸载 `SQL` 删除表之前按需清理附件文件

## 前端接入

- 前端页面会按照仓库约定从插件的 `frontend/` 目录中自动发现
- 示例页保留原有摘要卡片，并新增 `VXE` 表格与弹窗表单来维护示例记录
- 卸载时是否清理数据的选择由宿主插件管理页提供，而不是插件页面自行实现

## SQL 约定

- 安装 SQL 位于 `manifest/sql/`
- 卸载 SQL 位于 `manifest/sql/uninstall/`
- 安装 SQL 需要具备幂等性，以便在“卸载但保留数据”后重新安装时继续复用原有数据
- 当插件存在自有文件存储时，卸载 SQL 应与插件清理钩子协同工作，确保表数据和文件可一起清理

## 审查要点

- 元数据保持精简且准确
- 宿主接线关系保持显式
- 页面遵循目录约定
- 插件自有 SQL 与宿主 SQL 分离维护
- 禁用仅隐藏能力，不清理插件自有数据
- 卸载同时覆盖“保留数据”和“清理数据”两条生命周期路径
