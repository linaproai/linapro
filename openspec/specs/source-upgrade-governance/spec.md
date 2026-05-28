# 源码升级治理规范

## Purpose

定义框架元数据展示的来源，并明确旧的开发期源码升级命令不再作为当前规范能力。源码插件文件更新属于离线文件覆盖，运行时状态和数据升级由插件运行时升级治理负责。
## Requirements
### Requirement:框架元数据必须集中维护并直接在系统信息中展示

框架 SHALL 将其名称、版本、描述、主页、仓库 URL 和许可证保存在 `apps/lina-core/manifest/config/metadata.yaml` 中。系统信息 API 必须直接返回该元数据，使系统信息页面无需前端硬编码值即可渲染项目卡片。

#### Scenario: 系统信息 API 返回框架元数据
- **WHEN** 管理工作台请求系统信息
- **THEN** 响应包含框架名称、版本、描述、主页、仓库 URL 和许可证
- **AND** 每个值来自宿主 `metadata.yaml`

### Requirement: 源码插件升级实现不得作为公共 pkg 能力暴露

系统 SHALL 将源码插件升级、发现版本对比、发布快照同步和升级执行器视为插件运行时升级治理的内部实现。除稳定的插件管理 API、运行时状态 DTO 和必要治理契约外，旧`sourceupgrade`实现不得作为插件或外部组件可直接依赖的公共`pkg`能力；相关实现 MUST 收敛到职责明确的宿主`internal`组件。

#### Scenario: 源码插件升级通过运行时治理入口执行

- **WHEN** 管理员升级源码插件
- **THEN** 操作通过插件运行时升级治理 API 执行
- **AND** 业务代码不得直接 import 旧`pkg/sourceupgrade`执行升级逻辑

#### Scenario: 内部升级执行器不被插件依赖

- **WHEN** 源码插件需要声明升级回调、SQL 或治理资源
- **THEN** 插件通过`pluginhost`生命周期和插件 manifest 资源声明
- **AND** 插件不得依赖宿主内部 source upgrade scanner、executor 或 state reconciler

