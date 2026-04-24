## Why

当前宿主虽然已经支持通过`plugin.autoEnable`在启动期自动启用源码插件，但缺少一个面向演示环境的数据保护能力。演示环境通常需要保留完整的查询与浏览体验，同时阻止误操作或恶意操作修改系统数据，因此需要一个通过插件启用状态直接生效的演示控制源码插件，在请求链路中统一落地只读保护。

## What Changes

- 新增官方源码插件`demo-control`，通过宿主发布的全局 HTTP 中间件注册能力接入整个系统请求治理链路。
- 当`demo-control`被宿主启用时，插件在`/*`作用域下基于`RESTful API`的`HTTP Method`拦截系统写操作请求，仅保留查询型请求能力。
- 为保证演示环境仍可正常进入与退出系统，会为登录、登出等必要会话入口保留最小白名单。
- 通过宿主主配置文件中的`plugin.autoEnable`控制`demo-control`是否在启动期自动启用，以此作为演示能力的开关入口。

## Capabilities

### New Capabilities
- `demo-control-guard`: 定义通过`plugin.autoEnable`启停的演示控制源码插件，以及基于`HTTP Method`的全局写操作拦截规则。

### Modified Capabilities

## Impact

- 受影响代码主要位于`apps/lina-core/manifest/config/`与`apps/lina-plugins/`。
- 需要新增源码插件目录`apps/lina-plugins/demo-control/`，并更新插件工作区接线入口与`go.work`。
- 需要补充`plugin.autoEnable`配置与演示控制中间件单元测试，验证默认关闭（未启用插件）、显式启用、登录白名单与写操作拦截行为。
