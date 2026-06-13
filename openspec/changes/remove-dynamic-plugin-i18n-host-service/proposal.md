## Why

动态插件的多语言资源已经由宿主统一发现、合并、缓存和分发。继续把`i18n`发布为动态插件`hostServices`会让 guest 侧误以为可以主动读取 locale、翻译消息或检索 message key，和“动态插件不直接管理多语言运行时”的边界冲突。

本变更用于收敛动态插件的领域能力边界：动态插件仍可交付`manifest/i18n`资源，但不得声明或调用`service: i18n`；源码插件保留源码内直接注入的`I18n()`能力，由宿主装配期治理。

## What Changes

- 从动态插件 host service catalog、协议常量别名、guest capability directory 和`WASM`dispatcher 中移除`i18n`服务。
- 让`plugin.yaml hostServices`校验拒绝`service: i18n`，并通过测试覆盖该废弃边界。
- 更新`apps/lina-core/pkg/plugin`中英文`README`，说明动态插件多语言由宿主管理，不开放`i18n`host service。
- 更新官网领域能力文档和`i18n`专题文档，去掉动态插件读取或管理`manifest/i18n`的指导。

## Capabilities

### Modified Capabilities

- `plugin-host-service-extension`：调整动态插件普通领域 host service 覆盖范围，明确`i18n`不是动态插件可声明的 host service。

## Impact

- 影响代码：
  - `apps/lina-core/pkg/plugin/pluginbridge/**`
  - `apps/lina-core/internal/service/plugin/internal/wasm/**`
  - 对应 host service catalog、协议别名、guest directory 和单元测试
- 影响文档：
  - `apps/lina-core/pkg/plugin/README.md`
  - `apps/lina-core/pkg/plugin/README.zh-CN.md`
  - 官网领域能力与`i18n`专题文档
- 不影响范围：
  - 不移除`pkg/plugin/capability/i18ncap`。
  - 不移除源码插件`capability.Services.I18n()`或源码插件使用宿主`I18n()`的能力。
  - 不修改 HTTP API、SQL、DAO、DO、Entity 或前端页面。
