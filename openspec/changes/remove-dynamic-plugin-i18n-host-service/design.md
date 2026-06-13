## Context

动态插件运行在`WASM`guest 边界内，运行期只能通过宿主发布的`hostServices`访问外部能力。多语言资源则是插件包的一类静态治理资源：插件通过`manifest/i18n`交付资源，宿主负责发现、合并、缓存、失效和前端运行时分发。

如果动态插件同时拥有`service: i18n`，guest 侧就会获得`locale.get`、`messages.translate`和`messages.keys.find`等运行时翻译入口。这会把多语言决策拆到 guest 内部，导致插件开发者绕过宿主统一语言切换、资源合并和缓存治理。

## Decision

动态插件不发布`i18n`host service。`plugin.yaml hostServices`中的`service: i18n`必须被视为未知服务并拒绝；`pluginbridge.Services`不再提供`I18n()`guest client；`WASM`dispatcher 不再注册或处理`i18n`方法。

源码插件保持现状。源码插件随宿主同进程编译和装配，仍通过`capability.Services.I18n()`接收宿主显式注入的`i18ncap.Service`，这属于源码插件运行期能力目录，不等同于动态插件的 host service 协议。

动态插件需要本地化用户可见内容时，应交付`manifest/i18n`资源，并在业务响应中返回稳定业务键、参数和英文 fallback，或者返回原始业务数据，由宿主前端运行时和宿主资源合并流程完成展示本地化。动态插件后端不得直接读取、合并或翻译自身`manifest/i18n`资源。

## Boundaries

- 宿主通用能力：插件多语言资源发现、合并、缓存、失效、前端运行时语言包分发。
- 源码插件能力：通过源码插件能力目录接收`I18n()`，并遵守宿主注入和`i18n`治理。
- 动态插件能力：交付`manifest/i18n`资源，返回 message key、params、fallback 或业务数据，不声明、不调用`service: i18n`。
- 不新增抽象层；本变更删除一条公开动态协议链路，降低 guest 侧能力面和误用空间。

## Impact Analysis

- `i18n`影响：涉及宿主/插件多语言边界调整。资源交付目录不变，动态插件运行期翻译入口删除；源码插件`I18n()`能力保留。
- 缓存一致性影响：不新增或修改翻译缓存实现、失效策略和权威数据源，仅删除动态 guest 访问入口。
- 数据权限影响：不新增数据读取、写入、下载、聚合或动作接口。
- 开发工具跨平台影响：不修改脚本、构建工具或代码生成入口。
- DI 来源影响：删除`WASM`i18n host service 依赖读取路径，不新增运行期依赖；源码插件能力目录 DI 不变。
- 测试策略：使用 host service catalog/校验测试、dispatcher 覆盖测试和 Go 编译门禁验证动态`i18n`入口不可见且源码能力目录仍可编译。
