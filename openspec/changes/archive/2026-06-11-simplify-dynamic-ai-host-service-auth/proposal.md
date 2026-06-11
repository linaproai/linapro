## Why

当前动态插件调用`AI`能力时，需要在`plugin.yaml`的`hostServices.ai.resources`中声明`purpose`和策略属性，导致插件清单既承担领域方法授权，又承担调用参数策略配置。这个设计增加了动态插件开发复杂度，也让动态插件和源码插件消费`AI`能力的模型不一致。

## What Changes

- **BREAKING**：动态插件`service: ai`不再使用`resources`声明`purpose`、默认档位、输出 token、payload 或资产策略。
- 动态插件`plugin.yaml`只声明`service: ai`和允许调用的`methods`，由该声明推导宿主能力分类和运行时方法授权。
- 动态插件在请求 DTO 中提交`purpose`、`tier`、`maxOutputTokens`、资产引用和其他方法参数，主框架`pluginbridge`不再做`AI`业务参数策略限制。
- 主框架仍负责校验`service + method`授权、DTO 解析、可信`pluginID`来源注入、结构化错误和错误脱敏。
- `AI`请求参数合法性、档位存在性、provider 能力、渠道配置和审计归属由`AI`能力服务及`linapro-ai-core`治理。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `plugin-host-service-extension`：将动态插件`ai` host service 从`purpose`资源授权模型调整为方法授权模型。
- `linapro-ai-core-plugin`：同步动态插件`AI`授权不再限定为授权`purpose`，但仍不得授予渠道、档位管理或调用日志管理权限。

## Impact

- 影响`apps/lina-core/pkg/plugin/pluginbridge`的`hostServices`目录、清单校验、guest SDK 调用 envelope 和`WASM` host service 分发逻辑。
- 影响`apps/lina-core/pkg/plugin/README.md`与`README.zh-CN.md`中的动态插件`hostServices.ai`说明。
- 影响当前依赖`ai.resources`的单元测试；动态插件示例清单经检查未包含旧`ai.resources`声明，无需迁移。
- 不涉及数据库、HTTP API 路由、前端 UI、缓存失效或数据权限查询路径。
