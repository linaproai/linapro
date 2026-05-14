## ADDED Requirements

### Requirement: 前端插件页面发现必须支持空源码插件工作区

宿主前端 SHALL 在源码插件工作区不存在或为空时完成构建和运行。插件页面自动发现必须把缺失或空插件前端目录视为空页面集合，而不是构建错误。

#### Scenario: 前端构建时插件工作区不存在
- **WHEN** 宿主前端构建扫描源码插件页面
- **AND** `apps/lina-plugins` 不存在
- **THEN** 插件页面模块集合为空
- **AND** 前端构建继续执行

#### Scenario: 前端构建时插件工作区为空
- **WHEN** 宿主前端构建扫描源码插件页面
- **AND** `apps/lina-plugins` 是空目录
- **THEN** 插件页面模块集合为空
- **AND** 前端路由注册不创建源码插件页面入口

#### Scenario: Submodule 初始化后插件页面参与构建
- **WHEN** `apps/lina-plugins` submodule 初始化完成
- **AND** 源码插件提供 `frontend/pages/` 页面
- **THEN** 宿主前端构建发现这些页面
- **AND** 插件页面继续参与宿主路由和权限过滤
