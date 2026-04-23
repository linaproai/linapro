## 1. 宿主配置与接缝扩展

- [x] 1.1 在宿主配置模板中补充`plugin.autoEnable`示例，明确通过是否启用`demo-control`作为演示能力开关
- [x] 1.2 更新插件工作区接线入口，使`demo-control`能够被宿主发现、安装并按`plugin.autoEnable`启用

## 2. 演示控制源码插件实现

- [x] 2.1 新增`apps/lina-plugins/demo-control/`源码插件基础结构、清单、嵌入入口与说明文档
- [x] 2.2 实现基于全局 HTTP 中间件的演示控制服务，按`HTTP Method`阻断写操作并保留最小会话白名单
- [x] 2.3 确保演示控制插件不依赖额外宿主布尔配置，而是仅在插件启用时生效

## 3. 验证与回归保护

- [x] 3.1 补充`plugin.autoEnable`配置测试，覆盖默认关闭（未启用插件）与显式启用`demo-control`
- [x] 3.2 补充演示控制插件中间件测试，覆盖允许查询、拒绝写操作、登录白名单与未启用插件直通行为

## Feedback

- [x] **FB-1**: 移除额外的`demo.control.enabled`开关，改为通过`plugin.autoEnable`是否包含`demo-control`作为演示能力开关
