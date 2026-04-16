## ADDED Requirements

### Requirement: 接口文档元数据与项目定位一致
系统 SHALL 使系统接口文档页面所使用的 `OpenAPI` 标题、说明和页面介绍与 `Lina` 的统一项目定位保持一致。

#### Scenario: 生成 OpenAPI 元数据
- **WHEN** 宿主生成或读取 `OpenAPI` 文档标题与说明
- **THEN** 标题和说明使用与 `Lina` 项目定位一致的语义
- **AND** 不再使用“`Lina Admin API`”“后台管理系统接口文档”或等价表述

#### Scenario: 展示系统接口页面
- **WHEN** 用户打开“系统接口”页面
- **THEN** 页面展示的接口文档标题或介绍信息与后端 `OpenAPI` 元数据保持一致
- **AND** 页面不会把宿主 API 整体定义为仅服务后台管理系统的接口集合
