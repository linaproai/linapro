## Why

系统缺乏自身运行信息的展示能力，管理员无法查看后端API文档、系统运行环境信息，也无法直观了解前端框架的组件能力。v0.5.0 新增"系统信息"顶级菜单，提供API文档在线浏览与测试、系统运行信息展示、前端组件演示三大能力，提升系统的可观测性和开发体验。

## What Changes

- 新增顶级菜单"系统信息"，包含三个子菜单：系统接口、系统信息、组件演示
- **系统接口**：集成 Scalar OpenAPI 文档 UI，前端通过嵌入 Scalar 组件展示后端已有的 `/api.json` OpenAPI v3 接口文档，支持在线测试接口
- **系统信息**：参考 vben5 About 页面样式，展示关于项目、基本信息（Go 版本、OS、数据库等运行时数据）、后端组件、前端组件四个区块；外链第三方地址采用配置对象方便后续修改；后端新增系统信息 API 提供运行时数据
- **组件演示**：通过 iframe 嵌入 vben5 官网演示页面（https://www.vben.pro/），加载失败时展示友好的错误提示页面

## Capabilities

### New Capabilities
- `system-api-docs`: 系统接口文档页面，集成 Scalar UI 展示 OpenAPI 文档并支持在线测试
- `system-info`: 系统信息页面，展示项目信息、运行环境、后端/前端组件信息，后端提供系统信息 API
- `component-demo`: 组件演示页面，iframe 嵌入 vben5 官网演示，含加载失败处理

### Modified Capabilities

无

## Impact

- **前端路由**：新增顶级菜单"系统信息"及三个子路由
- **前端依赖**：引入 `@scalar/api-reference` npm 包用于 OpenAPI 文档渲染
- **后端 API**：新增 `GET /api/v1/system/info` 接口，返回系统运行时信息
- **后端代码**：新增 `api/system/`、`controller/system/`、`service/system/` 模块
- **前端视图**：新增 `views/about/` 目录，包含三个页面组件
