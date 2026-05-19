## ADDED Requirements

### Requirement: 插件工作区更新必须限定为离线文件覆盖

系统 SHALL 将 `plugins.install`、`plugins.update` 和直接插件目录覆盖定义为开发阶段离线文件更新能力。该能力只能写入或替换 `apps/lina-plugins/<plugin-id>` 中的插件文件和工具锁定状态，不得修改运行时数据库中的有效插件版本、插件安装状态、插件启用状态、插件治理资源或插件业务数据。

#### Scenario: 更新源码插件文件后不修改运行时数据库
- **WHEN** 用户运行 `make plugins.update`
- **AND** 工具用新版本源码插件目录覆盖 `apps/lina-plugins/plugin-demo`
- **THEN** 命令只更新本地插件文件和插件锁定状态
- **AND** 命令不连接宿主运行时数据库
- **AND** 命令不修改 `sys_plugin.version`、`release_id`、安装状态或启用状态

#### Scenario: 直接覆盖源码插件目录
- **WHEN** 用户手动覆盖 `apps/lina-plugins/plugin-demo` 目录中的插件文件
- **THEN** 该操作仅代表开发阶段文件变化
- **AND** 运行时数据库状态仍保持原有效版本
- **AND** 宿主启动后通过插件运行时升级状态标记处理文件和数据库元数据差异

#### Scenario: 离线更新不执行插件升级 SQL
- **WHEN** 用户通过插件工作区工具安装或更新插件文件
- **THEN** 工具不得执行插件 `manifest/sql` 中的升级 SQL
- **AND** 工具不得调用插件自定义升级回调
- **AND** 工具不得同步菜单、权限、i18n、apidoc、路由或 cron 等运行时治理资源
