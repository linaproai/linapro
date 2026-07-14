# host-storage-menu-catalog Specification

## Purpose
TBD - created by archiving change cloud-storage-providers. Update Purpose after archive.
## Requirements
### Requirement: 云存储配置菜单挂载到系统设置

系统 SHALL 将云对象存储 provider 插件的配置页挂载到宿主稳定目录 `setting`（系统设置），`parent_key` MUST 为 `setting`。系统 MUST NOT 再维护独立的 `storage`（存储管理）一级宿主稳定目录作为云存储配置挂载点。

#### Scenario: 数据库初始化后不存在 storage 目录

- **WHEN** 宿主执行菜单初始化或幂等菜单种子迁移
- **THEN** `sys_menu` 中 MUST NOT 存在未删除的 `menu_key = storage` 一级目录
- **AND** `menu_key = setting` 的系统设置目录仍由宿主种子维护

#### Scenario: 插件通过 parent_key 挂载到 setting

- **WHEN** 云存储 provider 插件在 `plugin.yaml` 中声明 `parent_key: setting`
- **AND** 该插件已安装并完成菜单同步
- **THEN** 其配置菜单 MUST 作为 `setting`（系统设置）目录的子节点出现
- **AND** MUST NOT 要求 `parent_key` 指向 `storage` 或 `plugin:…` 形式的插件目录键

#### Scenario: 安装任一云存储配置插件后在系统设置下展示

- **WHEN** 至少一个云存储 provider 插件已安装并同步出可见配置子菜单
- **AND** 当前用户具备查看该子菜单的权限
- **THEN** 前端导航 MUST 在「系统设置」下展示对应云存储配置子菜单
- **AND** 前端导航 MUST NOT 展示独立的「存储管理」一级目录

#### Scenario: 卸载全部云存储配置插件后仅无云配置子菜单

- **WHEN** 所有挂载于 `setting` 的云存储 provider 插件均已卸载
- **THEN** 「系统设置」下 MUST NOT 再出现这些云存储配置子菜单
- **AND** 「系统设置」目录本身仍由宿主保留（可因其下仍有字典、参数、文件等宿主菜单而可见）

