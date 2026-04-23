## Why

当前宿主虽然已经支持通过`plugin.autoEnable`在启动期自动启用源码插件，但缺少一个面向演示环境的数据保护能力。演示环境通常需要保留完整的查询与浏览体验，同时阻止误操作或恶意操作修改系统数据，因此需要一个默认关闭、按环境显式开启的演示控制开关，并通过源码插件在宿主请求链路中统一落地。

## What Changes

- 在宿主主配置文件中新增演示控制开关配置，默认关闭，用于声明当前实例是否进入只读演示模式。
- 新增官方源码插件`demo-control`，通过宿主发布的全局 HTTP 中间件注册能力接入统一请求治理链路。
- 在演示控制开关开启时，插件基于`RESTful API`的`HTTP Method`拦截系统写操作请求，仅保留查询型请求能力。
- 为保证演示环境仍可正常进入与退出系统，会为登录、登出等必要会话入口保留最小白名单。
- 将`demo-control`加入宿主默认启动自动启用列表，使插件在宿主启动后即处于可治理状态。

## Capabilities

### New Capabilities
- `demo-control-guard`: 定义演示环境只读保护开关、自动启用的演示控制源码插件，以及基于`HTTP Method`的全局写操作拦截规则。

### Modified Capabilities

## Impact

- 受影响代码主要位于`apps/lina-core/internal/service/config/`、`apps/lina-core/manifest/config/`、`apps/lina-core/pkg/pluginservice/config/`与`apps/lina-plugins/`。
- 需要新增源码插件目录`apps/lina-plugins/demo-control/`，并更新插件工作区接线入口与`go.work`。
- 需要补充配置解析与演示控制中间件单元测试，验证默认关闭、显式开启、登录白名单与写操作拦截行为。
