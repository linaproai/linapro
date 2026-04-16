# component-demo Specification

## Purpose
TBD - created by archiving change v0.5.0. Update Purpose after archive.
## Requirements
### Requirement: 组件演示页面
系统 SHALL 提供一个"组件演示"页面，通过 iframe 嵌入 vben5 官网演示页面（https://www.vben.pro/），展示前端框架的组件能力。

#### Scenario: 正常加载组件演示
- **WHEN** 用户点击"系统信息 > 组件演示"菜单
- **THEN** 页面以 iframe 全屏嵌入 vben5 官网演示页面，iframe 占满内容区域

### Requirement: 加载失败处理
当 iframe 嵌入的外部页面加载失败时，系统 SHALL 展示友好的错误提示页面，告知用户外部资源不可用。

#### Scenario: 外部网站不可访问
- **WHEN** vben5 官网演示页面无法加载（网络错误、网站下线等）
- **THEN** 页面展示错误提示信息，说明外部演示资源暂时不可用，不出现空白页面或浏览器默认错误页

### Requirement: 演示地址可配置
iframe 嵌入的演示地址 SHALL 通过前端配置指定，不硬编码在组件中，方便后续切换演示地址。

#### Scenario: 修改演示地址
- **WHEN** 开发者修改前端配置中的组件演示地址
- **THEN** iframe 加载新配置的演示地址

