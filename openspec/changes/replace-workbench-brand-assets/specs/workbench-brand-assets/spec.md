## ADDED Requirements

### Requirement: 默认管理工作台使用 LinaPro 本地 Logo
系统 SHALL 使用项目内随管理工作台构建交付的本地图标 Logo 资源作为默认管理工作台 Logo，不得继续使用 Vben 远程默认 Logo。

#### Scenario: 渲染工作台壳层 Logo
- **WHEN** 用户登录并进入默认管理工作台
- **THEN** 基础布局中的 Logo 使用项目本地 `linapro-mark.png` 资源
- **AND** Logo 右侧正常渲染应用名文本
- **AND** 页面不再请求 Vben 默认 Logo 远程地址

#### Scenario: 渲染认证页 Logo
- **WHEN** 未登录用户访问认证页
- **THEN** 认证页中的 Logo 使用与工作台壳层一致的项目本地 `linapro-mark.png` 资源
- **AND** Logo 右侧正常渲染应用名文本

#### Scenario: 加载宿主公开前端配置
- **WHEN** 默认管理工作台启动并成功读取宿主公开前端配置
- **THEN** 配置返回的默认 Logo 地址使用项目本地 `linapro-mark.png` 资源
- **AND** 配置返回值不得将 Logo 覆盖为 Vben 默认远程地址

### Requirement: 默认管理工作台使用 LinaPro favicon
系统 SHALL 使用项目提供的 `favicon.ico` 作为默认管理工作台浏览器标签页图标。

#### Scenario: 加载浏览器标签页图标
- **WHEN** 浏览器加载默认管理工作台入口页面
- **THEN** `/favicon.ico` 返回项目提供的 favicon 文件
- **AND** 该 favicon 随前端构建产物一起交付
