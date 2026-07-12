## ADDED Requirements

### Requirement: 宿主必须提供存储管理稳定一级目录

系统 SHALL 在宿主菜单种子中提供稳定一级目录，其稳定业务键 MUST 为 `storage`，展示名源文案为「存储管理」（英文运行时 i18n 为 `Storage`），`type` MUST 为目录 `D`。该目录 MUST 作为云对象存储 provider 插件配置页的语义挂载点，MUST NOT 由某一云厂商插件或 `linapro-storage-core` 类壳插件创建。

#### Scenario: 数据库初始化后存在 storage 目录

- **WHEN** 宿主执行菜单初始化或幂等菜单种子迁移
- **THEN** `sys_menu` 中存在 `menu_key = storage` 且 `type = D` 的一级目录
- **AND** 该目录不依赖任何 `lina-plugins` 插件安装

#### Scenario: 插件通过 parent_key 挂载到 storage

- **WHEN** 云存储 provider 插件在 `plugin.yaml` 中声明 `parent_key: storage`
- **AND** 该插件已安装并完成菜单同步
- **THEN** 其配置菜单 MUST 作为 `storage` 目录的子节点出现
- **AND** MUST NOT 要求 `parent_key` 指向 `plugin:…` 形式的插件目录键

### Requirement: 存储管理空目录必须隐藏

系统 SHALL 在导航投影中隐藏没有任何当前用户可见子菜单的 `storage` 目录。数据库中的宿主稳定目录记录 MUST 在插件全部卸载后仍保留。

#### Scenario: 未安装云存储配置插件时隐藏

- **WHEN** 当前没有任何已安装且产生可见子菜单的云存储 provider 插件
- **THEN** 前端导航 MUST NOT 展示「存储管理」目录入口

#### Scenario: 安装任一云存储配置插件后展示

- **WHEN** 至少一个云存储 provider 插件已安装并同步出可见配置子菜单
- **AND** 当前用户具备查看该子菜单的权限
- **THEN** 前端导航 MUST 展示「存储管理」目录及其可见子菜单

#### Scenario: 卸载全部云存储配置插件后再次隐藏

- **WHEN** 所有挂载于 `storage` 的云存储 provider 插件均已卸载
- **THEN** 导航 MUST 再次隐藏「存储管理」
- **AND** `sys_menu` 中 `menu_key = storage` 的宿主记录 MUST 仍然存在

### Requirement: 存储管理目录排序必须稳定

系统 SHALL 将 `storage` 目录排序置于扩展中心之前、开发中心之前的稳定整数位置。推荐种子值为 `sort = 10`，并将 `extension` 顺延为 `sort = 11`、`developer` 顺延为 `sort = 12`，除非迁移策略另有明确等价顺序。

#### Scenario: 侧栏相对顺序

- **WHEN** 扩展中心、存储管理、开发中心在导航中均可见
- **THEN** 顺序 MUST 为存储管理 → 扩展中心 → 开发中心