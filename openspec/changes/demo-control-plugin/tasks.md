## 1. 宿主配置与接缝扩展

- [x] 1.1 在宿主配置模板中补充`plugin.autoEnable`示例，明确通过是否启用`demo-control`作为演示能力开关
- [x] 1.2 更新插件工作区接线入口，使`demo-control`能够被宿主发现、安装并按`plugin.autoEnable`启用

## 2. 演示控制源码插件实现

- [x] 2.1 新增`apps/lina-plugins/demo-control/`源码插件基础结构、清单、嵌入入口与说明文档
- [x] 2.2 实现基于全局 HTTP 中间件的演示控制服务，在`/*`作用域下按`HTTP Method`阻断写操作并保留最小会话白名单
- [x] 2.3 确保演示控制插件不依赖额外宿主布尔配置，而是仅在插件启用时生效

## 3. 验证与回归保护

- [x] 3.1 补充`plugin.autoEnable`配置测试，覆盖默认关闭（未启用插件）与显式启用`demo-control`
- [x] 3.2 补充演示控制插件中间件测试，覆盖允许查询、拒绝写操作、整个系统作用域、登录白名单与未启用插件直通行为

## Feedback

- [x] **FB-1**: 移除额外的`demo.control.enabled`开关，改为通过`plugin.autoEnable`是否包含`demo-control`作为演示能力开关
- [x] **FB-2**: 将演示控制插件的全局中间件作用域扩展为`/*`，覆盖整个系统请求链路并保留登录白名单
- [x] **FB-3**: 补充`TC0105`端到端用例，覆盖`demo-control`的`plugin.autoEnable`启用态、登录/登出白名单、查询放行与`/*`全局写拦截行为
- [x] **FB-4**: 修正`demo-control`从`plugin.autoEnable`移除后运行态仍持续拦截写请求的问题，确保`plugin.autoEnable`列表是演示能力的唯一开关入口
- [x] **FB-5**: 在演示模式下放行除`demo-control`自身外的插件安装、卸载、启用、禁用操作，并补齐单测与`TC0105`覆盖
