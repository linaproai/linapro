## ADDED Requirements

### Requirement: 源码插件 HTTP 注册必须接收宿主发布依赖目录
系统 SHALL 在源码插件 HTTP、全局中间件和 Cron 注册回调中向插件暴露宿主发布的依赖目录。插件通过该目录获取稳定宿主能力适配器，并使用显式依赖注入构造插件 Controller 和 Service。

#### Scenario: 插件路由注册构造控制器
- **当** 源码插件在 `http.route.register` 回调中绑定控制器
- **则** 插件从 registrar 获取宿主发布依赖目录
- **且** 插件控制器构造函数接收已构造的插件 service 或其显式依赖
