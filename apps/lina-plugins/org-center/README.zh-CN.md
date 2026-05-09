# org-center

`org-center` 是 LinaPro 官方提供的组织管理源码插件。

## 能力范围

该插件负责：

- 部门管理
- 岗位管理
- 宿主用户管理所消费的可选组织能力

## 宿主边界

宿主保留用户管理、认证和菜单治理能力；本插件补充组织相关 API、菜单和 `manifest/sql/` 下的插件自有表结构。

## 目录结构

```text
org-center/
  plugin.yaml
  backend/
  frontend/
  manifest/
```
