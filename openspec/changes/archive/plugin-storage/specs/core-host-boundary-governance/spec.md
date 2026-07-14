## Requirements

### Requirement: 宿主稳定目录必须作为真实治理记录存在

系统 SHALL 将默认后台的一级稳定目录作为宿主拥有的稳定菜单记录维护，而非仅在前端投影层临时组装。稳定父级 `menu_key` MUST 至少包含：`dashboard`、`iam`、`org`、`setting`、`content`、`monitor`、`scheduler`、`extension`、`developer`。系统 MUST NOT 将 `storage` 作为一级宿主稳定目录。

#### Scenario: 初始化宿主稳定目录

- **WHEN** 宿主初始化默认后台菜单骨架时
- **THEN** 宿主创建并维护上述稳定父级 `menu_key`（不含 `storage`）
- **AND** 这些目录记录可被插件 `parent_key` 稳定解析

#### Scenario: 某目录下无可见子菜单

- **WHEN** `内容管理`、`组织管理` 或 `系统监控` 目录当前没有任何可见子菜单时
- **THEN** 它们在导航投影中被隐藏
- **AND** 宿主不删除对应的稳定目录记录

### Requirement: 云对象存储实现由插件扩展且配置挂载到系统设置

系统 SHALL 将对象存储领域契约与内置 local provider 保留在宿主，将具体云厂商对象存储后端实现交付为官方源码插件。云存储配置菜单 MUST 挂载到宿主已有 `setting`（系统设置）稳定目录；MUST NOT 要求单独的 `storage` 一级目录或 `linapro-storage-core` 类壳插件。

#### Scenario: 规划云存储插件边界

- **WHEN** 团队规划 `linapro-storage-cos`、`linapro-storage-oss`、`linapro-storage-obs`、`linapro-storage-qiniu`、`linapro-storage-aws`、`linapro-storage-azure` 或 `linapro-storage-s3` 的能力边界时
- **THEN** 插件仅承载对应云厂商 `storagecap.Provider`、配置 settings 与连通性探测
- **AND** `storagecap.Service`、插件/租户 key 作用域与 local provider 仍保留在宿主
- **AND** 配置入口挂载在宿主「系统设置」目录下

#### Scenario: 禁止仅为菜单存在的 storage-core 壳插件

- **WHEN** 评估是否新增 `linapro-storage-core` 或等价领域壳插件
- **THEN** 方案 MUST NOT 以壳插件仅用于提供一级目录或聚合配置入口
- **AND** Storage 领域契约与解析逻辑 MUST 继续归属宿主
