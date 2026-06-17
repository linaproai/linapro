## MODIFIED Requirements

### Requirement:源码插件后端数据库访问在插件内闭环

系统 SHALL 要求官方源码插件在各自插件根目录维护独立的插件本地工具配置，并通过插件本地的`dao/do/entity`完成数据库访问，避免重新依赖宿主`dao/model`包或长期保留散落的`g.DB().Model(...)`直连实现。插件根`hack/config.yaml`MUST 作为插件本地工具配置入口；其中`gfcli.gen.dao`承载 GoFrame ORM 代码生成配置，`build.commands`承载插件自定义构建指令；GoFrame 生成工作目录 MUST 保持为插件`backend/`目录。

#### Scenario:插件后端维护独立的代码生成配置

- **当** 团队创建或维护官方源码插件后端时
- **则** 插件根目录包含`hack/config.yaml`
- **且** 开发者可直接在插件根目录执行`make dao`
- **且** 生成流程在插件`backend/`目录中执行
- **且** 生成结果落入插件本地的`backend/internal/dao`、`backend/internal/model/do`和`backend/internal/model/entity`

#### Scenario:插件服务访问插件自有表或共享读表

- **当** `linapro-org-core`、`linapro-content-notice`、`linapro-monitor-loginlog`、`linapro-monitor-operlog`、`linapro-monitor-server`或`linapro-demo-source`的`backend/internal/service/`访问数据库时
- **则** 插件服务使用插件本地生成的`dao/do/entity`
- **且** 对共享读表的访问也通过插件本地生成的产物完成，除非对应宿主领域能力已经提供稳定公开契约并被当前规范允许使用
- **且** 插件后端不直接依赖宿主`dao/model`包
- **且** 宿主不再并行保留这些插件业务表的 ORM 产物

#### Scenario:当前版本不直接访问数据库的源码插件

- **当** 官方源码插件当前版本仅通过宿主稳定能力完成业务处理时
- **则** 插件不需要创建空的根`hack/config.yaml`
- **且** 未来新增数据库访问时必须先在插件根维护`hack/config.yaml`
- **且** 继续使用插件本地的`make dao`和`dao/do/entity`结构
