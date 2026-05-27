## Context

当前`plugin-config-service`基线允许源码插件和动态插件读取任意宿主 GoFrame 静态配置，动态插件还需要在`plugin.yaml`中逐项声明`get`、`exists`、`string`、`bool`、`int`、`duration`等读取方法。这种模型把插件运行期配置放在宿主全局配置树中，既扩大了动态插件读取范围，也让插件配置、插件元数据和宿主核心配置的职责边界不清晰。

GoFrame 的`gcfg`支持按`config.yaml`等文件名自动发现配置，也支持当前工作目录、`config/`、`manifest/config/`、可执行文件目录及源码`main`目录等搜索路径，并可通过`GF_GCFG_PATH`和`GF_GCFG_FILE`调整默认配置目录与文件名。本设计复用 GoFrame 配置能力，但不让插件直接读取宿主全局`g.Cfg()`；宿主需要为每个插件构造受作用域约束的配置视图。

插件可能维护`manifest/metadata.yaml`、`manifest/resources/*.yaml`等声明型资源。这些资源不属于运行时配置，也不应进入宿主`apps/lina-core/manifest/config`。源码插件和动态插件需要共享同一套资源路径语义，并通过宿主服务读取，而不是让插件代码自行拼接文件系统路径。

## Goals / Non-Goals

**Goals:**

- 让插件运行期配置归属于插件自身，目录统一为`apps/lina-plugins/<plugin-id>/manifest/config/`。
- 让`config.example.yaml`只承担模板职责，`config.yaml`承担实际运行配置职责；生产部署可把实际配置统一放在宿主配置根的`plugins/<plugin-id>/config.yaml`。
- 通过`HostServices.Config()`、`HostServices.HostConfig()`和`HostServices.Manifest()`区分插件配置、宿主公开配置和插件清单资源。
- 让动态插件`plugin.yaml`中的只读读取方法声明收敛为`methods: [get]`，typed helper 由 SDK 或源码适配器提供。
- 为源码插件和动态插件提供一致的`manifest`资源读取视图，支持实际存在的`metadata.yaml`、`resources/*.yaml`等声明型 YAML 文件。
- 让宿主配置只维护宿主运行时配置和插件治理配置，不再维护具体插件的业务配置内容。
- 明确运行期配置与发布产物资源的缓存一致性、失效触发和集群边界。

**Non-Goals:**

- 不提供插件写配置、保存配置、在线编辑插件配置或配置热重载管理接口。
- 不允许插件通过配置服务读取完整宿主配置树、宿主私有密钥或其他插件配置。
- 不把`metadata.yaml`等声明资源纳入宿主配置管理页面或`sys_config`参数管理。
- 不改变`data`、`storage`、`network`等既有 host service 的授权和数据权限模型。
- 不把插件`manifest/sql/`、`manifest/i18n/`等生命周期专用资源改造成通用业务配置读取入口。

## Decisions

### 1. 拆分三类读取接口

`HostServices.Config()`返回当前插件作用域的运行期配置视图。插件侧读取键时不需要带`plugin-id`前缀，宿主在构造服务时已经注入当前插件 ID。源码插件适配器可以提供`Get`、`Exists`、`Scan`、`String`、`Bool`、`Int`和`Duration`等便捷方法；动态插件授权层只识别`config.get`，guest SDK 的 typed helper 统一映射到`get`结果后解析。

`HostServices.HostConfig()`返回宿主公开配置读取视图。该接口不等价于`Config()`，必须拒绝完整宿主配置扫描、根节点读取和任意 key 读取，只允许宿主显式白名单中的公开配置键，例如工作台基础路径、公开语言设置、公开运行模式等。源码插件可以拥有 typed helper，但不得拥有`Scan`或`Data`这类完整树读取能力；动态插件必须在`hostServices`中声明`resources.keys`或等价白名单。

`HostServices.Manifest()`返回当前插件`manifest/`下声明型资源读取视图，用于`metadata.yaml`或`resources/*.yaml`等版本化只读资源。它不承担运行期配置职责，`manifest/config/**`仍通过`Config()`读取；`manifest/sql/**`和`manifest/i18n/**`仍由插件生命周期、数据库和 i18n 管线管理。

备选方案是继续在`Config()`中同时读取宿主配置、插件配置和元数据文件。该方案接口少，但会让权限、审计、缓存和部署策略混在一起，因此不采用。

### 2. 插件配置文件解析使用宿主构造的 scoped resolver

宿主为每个插件构造独立配置 resolver，并按运行模式组织候选来源：

| 场景 | 候选路径 | 说明 |
| --- | --- | --- |
| 开发阶段 | `apps/lina-plugins/<plugin-id>/manifest/config/config.yaml` | 直接读取插件源码目录，便于插件独立调试。 |
| 生产阶段 | `<gcfg配置根>/plugins/<plugin-id>/config.yaml` | 推荐部署路径，例如`/opt/linapro/config/plugins/<plugin-id>/config.yaml`。 |
| 动态发布默认 | artifact 中的`manifest/config/config.yaml` | 仅作为随发布携带的插件默认运行配置，不应包含环境密钥。 |
| 模板 | `manifest/config/config.example.yaml` | 只作为用户复制和理解配置结构的模板，不作为运行期自动默认值。 |

生产配置根优先跟随 GoFrame 配置目录语义：当运维通过`GF_GCFG_PATH`或命令行设置配置根时，宿主从同一配置根解析`plugins/<plugin-id>/config.yaml`；否则使用可执行文件同级`config/`等既有部署约定。插件配置 resolver 可以使用独立`gcfg.Config`或等价解析器读取文件，但不得把插件键注入宿主全局`g.Cfg()`。

配置优先级为：生产外部配置高于开发源码配置，高于动态 artifact 默认配置。缺失配置键不读取`config.example.yaml`兜底，插件业务默认值由插件内部结构体默认值或读取方法默认值承担。

宿主`manifest/config`中的`plugin.autoEnable`、`plugin.dynamic.storagePath`、`plugin.allowForceUninstall`等仍属于宿主插件治理配置；`linapro-monitor-server`的`monitor.interval`、`retentionMultiplier`等具体插件业务配置不得继续在宿主`manifest/config`中维护。

### 3. 动态插件授权声明只描述动作和资源边界

动态插件`plugin.yaml`示例：

```yaml
hostServices:
  - service: config
    methods: [get]
  - service: hostConfig
    methods: [get]
    resources:
      keys:
        - workspace.basePath
        - i18n.default
  - service: manifest
    methods: [get]
    resources:
      paths:
        - resources/*.yaml
```

`methods`只表达实际 host call 动作。`String`、`Bool`、`Int`、`Duration`、`Scan`属于调用端 SDK 或源码适配器 helper，不进入授权方法集合。宿主审计仍记录`service`、`method=get`、插件 ID、资源键或资源路径摘要、结果状态和耗时。

备选方案是在`plugin.yaml`继续声明所有 typed helper。该方案看似精细，但实际读取动作仍是只读获取，声明过细会让插件清单频繁随 SDK helper 变化，不利于稳定授权模型，因此不采用。

### 4. `Manifest()`读取版本化声明资源并执行路径治理

`Manifest()`的路径参数必须是相对`manifest/`根目录的 slash 路径。宿主必须拒绝空路径根读取、绝对路径、`..`路径穿越、Windows drive path、URL、跨插件路径和未声明的动态插件资源路径。`metadata.yaml`仅在插件实际提供该文件时属于可读取资源；动态插件不得在`plugin.yaml`中声明不存在的 manifest 资源路径，更深层资源可通过`resources/*.yaml`等白名单声明。

源码插件开发阶段可从源码目录读取，生产阶段从编译嵌入资源或 registrar 发布的资源视图读取。动态插件只能从 artifact 快照资源视图读取，资源内容绑定 active release 的 checksum 或 generation。

备选方案是让插件代码直接读取本地文件系统。该方案在源码插件开发时简单，但动态插件、生产二进制和集群节点会出现不同资源视图，因此不采用。

### 5. 缓存与一致性按资源性质区分

插件配置是只读运行期输入。开发阶段可以依赖本地文件读取和 GoFrame 文件缓存能力；生产阶段配置变更应通过重新部署、重启或未来显式 reload 流程生效。当前变更不承诺跨节点自动热更新生产配置文件。若实现进程内配置缓存，缓存键必须至少包含插件 ID、配置来源、文件 mtime 或 checksum；动态 artifact 默认配置还必须绑定 active release checksum 或 generation。

插件`manifest`资源是发布版本的一部分。源码插件资源随宿主二进制或源码工作区变化；动态插件资源随 active release 变化。动态插件安装、启用、禁用、卸载、升级或同版本刷新后，既有`plugin-runtime`修订号必须触发 manifest 资源视图、配置默认视图和相关派生缓存失效。集群模式不得仅依赖本地进程内缓存暴露旧插件资源。

### 6. 宿主边界和 i18n 边界保持专门治理

本变更属于`apps/lina-core`提供的通用插件扩展能力，不绑定默认管理工作台页面结构。插件 i18n 资源仍由`manifest/i18n/`和 i18n 管线治理，`Manifest()`不作为运行时 UI 翻译包读取入口。实现阶段如新增用户或调用端可见错误，需要使用宿主`bizerr`结构化错误码、英文 fallback 和必要的翻译资源；本提案文档本身不新增运行时文案资源。

## Risks / Trade-offs

- 插件从任意宿主配置读取迁移到作用域配置后，旧插件若依赖宿主私有 key 会失败。缓解方式：实施阶段同步调整官方插件，并把真正需要公开的宿主配置键纳入`HostConfig()`白名单。
- 生产外部配置和 artifact 默认配置同时存在时可能让开发者误判生效来源。缓解方式：配置 resolver 暴露来源元数据用于日志和诊断，文档明确外部配置优先。
- 允许`Manifest()`读取任意`manifest`文件可能绕过生命周期管线。缓解方式：默认读取范围限定为声明型资源，保留`config/`、`sql/`、`i18n/`等专用目录的服务边界。
- 开发阶段直接读取插件源码目录配置会带来环境差异。缓解方式：仅在开发工作区或源码插件本地模式启用，生产以 GoFrame 配置根下的`plugins/<plugin-id>/config.yaml`为推荐入口。
- 当前不支持跨节点热更新生产插件配置。缓解方式：明确生产配置变更通过部署流程生效；未来若新增 reload，需要单独设计运行时配置修订号、事件广播、审计和失败回滚。

## Migration Plan

1. 新增或调整公共 host service 契约，提供`Config()`、`HostConfig()`和`Manifest()`三个读取入口，并让源码插件 registrar 与动态插件 host service registry 共用同一运行期服务实例。
2. 调整动态插件`plugin.yaml`校验和授权快照，将`config`、`hostConfig`、`manifest`的读取方法规范化为`get`，并拒绝写入、保存、reload 和未知方法。
3. 实现插件配置 resolver，支持开发源码目录、生产 GoFrame 配置根和动态 artifact 默认配置来源，并确保`config.example.yaml`不参与运行期读取。
4. 实现 manifest resource resolver，统一源码插件嵌入资源、源码开发目录和动态 artifact 快照资源视图，并补齐路径治理。
5. 更新官方源码插件、动态示例插件和构建产物资源打包逻辑，确保`manifest/config/`和实际存在的声明型 YAML 资源被发现、打包和测试覆盖，同时删除宿主配置中已经迁移到插件自有配置的业务配置项。
6. 运行 OpenSpec 校验、Go 后端测试、动态插件构建 smoke 和必要的静态路径治理测试。

回滚时可恢复旧`plugin-config-service`读取宿主任意静态配置的适配层，但不建议继续保留完整宿主配置读取能力；如必须短期回退，应仅作为内部迁移分支并保持动态插件默认拒绝完整配置快照。

## Open Questions

无需要继续向用户确认的阻塞问题。实施时需要根据现有代码结构决定具体包名和接口命名，但不影响本变更的行为边界。
