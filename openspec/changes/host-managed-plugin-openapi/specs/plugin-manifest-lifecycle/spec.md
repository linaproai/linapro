## MODIFIED Requirements

### Requirement: `plugin.yaml` 保持精简且可声明插件菜单

系统 SHALL 继续让 `plugin.yaml` 只承载插件基础身份、菜单等静态清单，不要求源码插件在 `plugin.yaml` 中显式声明后端路由。

#### Scenario: `plugin.yaml` 不重复声明源码插件后端路由

- **WHEN** 宿主解析一个源码插件的 `plugin.yaml`
- **THEN** 清单不需要声明源码插件后端路由列表
- **AND** 源码插件后端路由以注册代码和 DTO `g.Meta` 为唯一事实来源
- **AND** 宿主在路由注册时自动采集路由归属与文档元数据，而不是要求开发者维护第二份路由清单
