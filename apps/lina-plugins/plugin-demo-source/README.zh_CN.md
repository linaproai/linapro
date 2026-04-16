# plugin-demo-source

`plugin-demo-source` 是 Lina 的源码插件样例，用来展示一个在仓库内开发、并随宿主一起编译交付的最小插件闭环。

## 目录结构

```text
plugin-demo-source/
  plugin.yaml
  backend/
  frontend/
  manifest/
```

## 清单范围

`plugin.yaml` 负责保存插件元数据和菜单声明。页面、Slot 和 SQL 资源通过目录约定发现，而不是在元数据中重复维护。

## 后端接入

- 在 `backend/` 中实现插件后端入口
- 将业务逻辑保留在 `backend/internal/service/` 下
- 通过宿主构建使用的源码插件注册入口显式接线

## 前端接入

前端页面会按照仓库约定从插件的 `frontend/` 目录中自动发现。

## SQL 约定

- 安装 SQL 位于 `manifest/sql/`
- 卸载 SQL 位于 `manifest/sql/uninstall/`

## 审查要点

- 元数据保持精简且准确
- 宿主接线关系保持显式
- 页面遵循目录约定
- 插件自有 SQL 与宿主 SQL 分离维护
