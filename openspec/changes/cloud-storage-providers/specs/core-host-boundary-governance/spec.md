## MODIFIED Requirements

### Requirement:宿主稳定目录必须作为真实治理记录存在

系统 SHALL 将默认后台的一级稳定目录作为宿主拥有的稳定菜单记录维护，而非仅在前端投影层临时组装。稳定父级 `menu_key` MUST 至少包含：`dashboard`、`iam`、`org`、`setting`、`content`、`monitor`、`scheduler`、`extension`、`storage`、`developer`。

#### Scenario:初始化宿主稳定目录
- **当** 宿主初始化默认后台菜单骨架时
- **则** 宿主创建并维护包含 `storage` 在内的稳定父级 `menu_key`
- **且** 这些目录记录可被插件 `parent_key` 稳定解析

#### Scenario:某目录下无可见子菜单
- **当** `内容管理`、`组织管理`、`系统监控` 或 `存储管理` 目录当前没有任何可见子菜单时
- **则** 它们在导航投影中被隐藏
- **且** 宿主不删除对应的稳定目录记录

## ADDED Requirements

### Requirement:云对象存储实现由插件扩展且目录由宿主拥有

系统 SHALL 将对象存储领域契约与内置 local provider 保留在宿主，将具体云厂商对象存储后端实现交付为官方源码插件。宿主 MUST 提供 `storage` 稳定目录供云存储配置菜单挂载；MUST NOT 要求单独的 `linapro-storage-core` 插件仅用于创建父目录。

#### Scenario:规划云存储插件边界
- **当** 团队规划 `linapro-storage-cos`、`linapro-storage-oss`、`linapro-storage-obs`、`linapro-storage-qiniu`、`linapro-storage-aws`、`linapro-storage-azure` 或 `linapro-storage-s3` 的能力边界时
- **则** 插件仅承载对应云厂商 `storagecap.Provider`、配置 settings 与连通性探测
- **且** `storagecap.Service`、插件/租户 key 作用域与 local provider 仍保留在宿主
- **且** 「存储管理」父目录由宿主菜单种子拥有
