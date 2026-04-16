# system-api-docs Specification

## Purpose
TBD - created by archiving change v0.5.0. Update Purpose after archive.
## Requirements
### Requirement: 系统接口文档页面展示
系统 SHALL 提供一个"系统接口"页面，通过集成 Scalar OpenAPI 文档 UI 展示后端 API 文档。文档数据来源于后端 GoFrame 自动生成的 `/api.json` OpenAPI v3 规范文件。

#### Scenario: 正常加载接口文档
- **WHEN** 用户点击"系统信息 > 系统接口"菜单
- **THEN** 页面展示 Scalar UI 渲染的 API 文档，包含所有后端接口的路径、参数、响应定义

### Requirement: 在线接口测试
Scalar UI SHALL 支持用户在文档页面上直接测试接口（Try it 功能），无需跳转到第三方工具。

#### Scenario: 在线测试接口
- **WHEN** 用户在 Scalar 文档页面中选择某个接口并点击"Test Request"/"Try it"
- **THEN** 页面展示请求参数输入区域，用户填写参数后可发送请求并查看响应结果

### Requirement: 接口文档地址可配置
Scalar UI 加载的 OpenAPI 规范文件地址 SHALL 通过前端配置指定，不硬编码在组件中。

#### Scenario: 修改接口文档地址
- **WHEN** 开发者修改前端配置中的 OpenAPI 规范文件地址
- **THEN** Scalar UI 加载新地址的 API 文档

### Requirement: 系统接口自动合并动态插件公开路由文档

系统 SHALL 将当前已启用动态插件的路由合同投影到宿主`OpenAPI`文档中，并展示动态插件对外真实可访问的固定公开路径。

#### Scenario: 动态插件接口以固定公开路径出现在系统接口

- **WHEN** 一个动态插件已经启用且其`active release`成功装载了动态路由合同
- **THEN** 用户访问“系统接口”页面时能够看到该动态插件对应的公开路径
- **AND** 这些路径展示为`/api/v1/extensions/{pluginId}/...`
- **AND** 每个动态路由项至少包含方法、标签、摘要、描述与安全要求

#### Scenario: 可执行动态接口文档展示真实响应语义

- **WHEN** 宿主为已启用且声明可执行 bridge 的动态路由生成`OpenAPI`操作项
- **THEN** 宿主为该路由提供`200`成功响应描述
- **AND** 宿主为该路由提供`500`运行时执行失败描述
- **AND** 使用登录治理的动态路由在文档中声明`BearerAuth`安全要求

#### Scenario: 未接入执行器的动态接口文档带有占位响应说明

- **WHEN** 宿主为未声明可执行 bridge 的动态插件路由生成`OpenAPI`操作项
- **THEN** 宿主为当前尚未接入执行器的动态路由补充`501`占位响应描述

#### Scenario: 动态插件禁用后从系统接口中移除

- **WHEN** 一个已启用的动态插件被禁用、卸载或切换到不再暴露该路由的激活版本
- **THEN** 宿主从主`OpenAPI`文档中移除该插件对应的动态路由投影
- **AND** “系统接口”页面不再展示已失效的公开路径

