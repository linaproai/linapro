## ADDED Requirements

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
