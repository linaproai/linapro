## Why

当前插件读取配置仍以宿主任意静态配置为中心，插件运行期配置需要回流到`apps/lina-core/manifest/config`或依赖宿主全局配置键，增加了核心宿主侵入性，也让动态插件授权边界过宽。

插件后续可能维护`metadata.yaml`、`resources/*.yaml`等非运行时配置但具备声明语义的`manifest`资源，需要为源码插件和动态插件提供统一、只读、插件自作用域的读取方式，保持`apps/lina-core`作为通用宿主的稳定边界。

## What Changes

- **BREAKING**：插件配置服务从“读取任意宿主静态配置”调整为“默认读取当前插件自己的运行时配置”，`HostServices.Config()`等价于插件作用域配置读取。
- 新增宿主公开配置读取边界：`HostServices.HostConfig()`只允许读取宿主显式白名单中的公开配置键，不提供完整宿主配置扫描。
- 新增插件`manifest/config/`约定：`config.example.yaml`作为可提交模板，`config.yaml`作为本地或部署环境实际运行配置；开发阶段可直接读取插件源码目录下的`manifest/config/config.yaml`，生产阶段支持部署到宿主配置根下的`plugins/<plugin-id>/config.yaml`。
- 新增插件通用`manifest`资源读取能力：`HostServices.Manifest()`允许插件只读访问当前插件`manifest/`下实际存在的声明资源，例如可选的`metadata.yaml`或`resources/*.yaml`，并阻断绝对路径、路径穿越、URL 和跨插件读取。
- 简化动态插件`plugin.yaml`中的配置读取授权声明：`config`、`hostConfig`和`manifest`等只读服务在清单中只声明`methods: [get]`，SDK 可保留`String`、`Bool`、`Int`、`Duration`、`Scan`等便捷方法，但授权模型不再要求逐个声明这些 typed helper。
- 动态插件打包和运行时加载必须保留`manifest/config/`和`manifest/**/*.yaml`等插件资源，使源码插件和动态插件拥有一致的资源视图。
- 宿主`manifest/config`只维护宿主运行时配置和插件治理配置，不再维护具体插件的业务配置内容；例如`linapro-monitor-server`的监控采集配置由该插件自己的`manifest/config`模板和实际配置维护。
- 不引入插件写配置、保存配置、热更新配置或跨插件共享配置能力；未来如支持运行时可变配置写入，需要另行设计集群一致性、审计和权限模型。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `plugin-config-service`：将插件配置服务从任意宿主静态配置读取改为插件作用域运行时配置读取，并补充`HostConfig()`的宿主公开配置读取边界。
- `plugin-host-service-extension`：调整动态插件`hostServices`声明与授权模型，使只读插件配置、宿主公开配置和清单资源读取使用简化的`get`方法声明，并由宿主按 service 和资源边界授权。
- `plugin-manifest-lifecycle`：扩展插件`manifest`资源语义，支持插件代码通过`Manifest()`读取自身`manifest/`下的声明型资源。
- `plugin-runtime-loading`：要求动态插件运行时资源视图包含配置与通用`manifest`资源，并在插件发布、刷新、启用、禁用和升级时按插件运行时修订失效。
- `plugin-embed-snapshot-packaging`：要求动态插件构建产物保留`manifest/config/`与通用`manifest`资源路径语义，避免源码插件和动态插件出现不同的配置或元数据路径约定。

## Impact

- 影响宿主与插件公共契约：`pkg/pluginservice/config`、源码插件 registrar 注入的 host services、动态插件`pluginbridge` host service 协议和 guest SDK。
- 影响插件目录与发布产物：`apps/lina-plugins/<plugin-id>/manifest/config/config.example.yaml`、本地或部署期`manifest/config/config.yaml`、可选`manifest/metadata.yaml`及动态插件 WASM artifact 资源快照。
- 影响动态插件授权声明：`plugin.yaml`中的`hostServices`校验、规范化、授权快照和审计摘要。
- 影响运行时配置一致性：插件配置仍为只读运行期输入，开发阶段读取源码目录配置，生产阶段读取部署配置根；如实现缓存或快照，必须绑定插件运行时修订、文件来源和集群刷新策略。
- `i18n`影响：本提案不新增运行时用户可见文案、菜单、路由或语言包资源；后续实现若新增调用端可见错误或 API 文档源文本，必须按宿主或插件`i18n`启用状态维护错误码、英文 fallback 和翻译资源。
- 数据权限影响：本提案不新增业务数据读写接口，也不允许插件通过配置或`manifest`读取绕过宿主数据权限；涉及`data` host service 的访问仍受既有授权和数据权限规则约束。
