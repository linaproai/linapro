## MODIFIED Requirements

### Requirement:插件 manifest 资源必须支持插件自作用域只读读取

系统 SHALL 为源码插件和动态插件提供`HostServices.Manifest()`能力，使插件代码能够只读读取当前插件`manifest/`目录下的原始资源。读取范围 MUST 绑定当前插件 ID，不得允许插件读取宿主 manifest、其他插件 manifest、任意文件系统路径或 URL。`metadata.yaml` SHALL 在插件实际提供该文件时作为可通过该能力读取的普通可选资源，但系统不得要求所有插件都提交`metadata.yaml`，也不得为 `metadata.yaml` 保留独立的 `Metadata` 服务、`metadata` host service、`metadata.get` 或等价插件可见读取入口。

#### Scenario:源码插件读取自身 metadata 普通资源

- **WHEN** 源码插件`plugin-a`调用`HostServices.Manifest().Get(ctx, "metadata.yaml")`
- **AND** `apps/lina-plugins/plugin-a/manifest/metadata.yaml`存在
- **THEN** 系统返回该文件内容
- **AND** 读取作用域限定为`plugin-a`的`manifest/`目录
- **AND** 该读取不经过独立的 `Metadata` 服务

#### Scenario:源码插件读取自身 config 资源原文

- **WHEN** 源码插件`plugin-a`调用`HostServices.Manifest().Get(ctx, "config/config.example.yaml")`
- **AND** `apps/lina-plugins/plugin-a/manifest/config/config.example.yaml`存在
- **THEN** 系统返回该文件原始内容
- **AND** 不把该读取结果作为插件运行期有效配置自动生效

#### Scenario:源码插件读取自身 SQL 资源原文

- **WHEN** 源码插件`plugin-a`调用`HostServices.Manifest().Get(ctx, "sql/001-schema.sql")`
- **AND** `apps/lina-plugins/plugin-a/manifest/sql/001-schema.sql`存在
- **THEN** 系统返回该 SQL 文件原始内容
- **AND** 不执行该 SQL

#### Scenario:源码插件读取自身 i18n 资源原文

- **WHEN** 源码插件`plugin-a`调用`HostServices.Manifest().Get(ctx, "i18n/zh-CN/plugin.json")`
- **AND** `apps/lina-plugins/plugin-a/manifest/i18n/zh-CN/plugin.json`存在
- **THEN** 系统返回该 JSON 文件原始内容
- **AND** 不把该读取作为翻译资源注册或缓存失效动作

#### Scenario:插件未提供 metadata

- **WHEN** 插件未维护`manifest/metadata.yaml`
- **THEN** 插件无需为了目录规范提交空白或占位 metadata 文件
- **AND** 插件清单不得申请读取不存在的`metadata.yaml`

#### Scenario:动态插件读取 artifact 中的 metadata 普通资源

- **WHEN** 动态插件`plugin-a`调用`manifest.get`读取`metadata.yaml`
- **AND** 当前 active release artifact 携带`manifest/metadata.yaml`
- **AND** 当前授权快照允许读取`metadata.yaml`
- **THEN** 系统从该 active release 的资源快照返回文件内容
- **AND** 该内容绑定当前 active release 的 checksum 或 generation

#### Scenario:动态插件读取已授权专用目录资源原文

- **WHEN** 动态插件`plugin-a`调用`manifest.get`读取`config/config.example.yaml`、`sql/001-schema.sql`或`i18n/zh-CN/plugin.json`
- **AND** 当前 active release artifact 携带对应资源
- **AND** 当前授权快照允许读取对应路径
- **THEN** 系统从该 active release 的资源快照返回文件原始内容
- **AND** 不触发配置生效、SQL 执行或翻译资源注册

#### Scenario:插件扫描 YAML manifest 资源到结构体

- **WHEN** 插件调用`HostServices.Manifest().Scan(ctx, "metadata.yaml", "", &metadata)`
- **AND** `metadata.yaml`是合法 YAML 文档
- **THEN** 系统将文件内容绑定到插件提供的结构体
- **AND** 结构体业务语义和验证逻辑由插件内部维护

### Requirement:Manifest 资源读取必须执行路径安全治理

系统 SHALL 将`Manifest()`的路径参数解释为相对当前插件`manifest/`根目录的 slash 路径。系统 MUST 拒绝空根读取、绝对路径、路径穿越、Windows drive path、URL、跨插件路径和未授权路径；动态插件还 MUST 按`plugin.yaml`中 `service: manifest` 的 `resources.paths` 与宿主确认后的授权快照校验。系统 MUST 允许合法的 `config/`、`sql/` 和 `i18n/` manifest 相对路径参与同一套路径安全和授权校验。

#### Scenario:合法相对路径被允许

- **WHEN** 插件读取`metadata.yaml`、`resources/policy.yaml`、`config/config.example.yaml`、`sql/001-schema.sql`或`i18n/zh-CN/plugin.json`
- **AND** 该路径位于当前插件`manifest/`目录下且满足授权策略
- **THEN** 系统允许读取该资源
- **AND** 返回内容不包含其他插件资源

#### Scenario:路径穿越被拒绝

- **WHEN** 插件读取`../other-plugin/manifest/metadata.yaml`或`../../apps/lina-core/manifest/config/config.yaml`
- **THEN** 系统拒绝该请求
- **AND** 不访问目标文件系统路径

#### Scenario:绝对路径和 URL 被拒绝

- **WHEN** 插件读取`/etc/passwd`、`C:\\secret.yaml`或`http://example.com/config.yaml`
- **THEN** 系统拒绝该请求
- **AND** 不发起本地文件或网络读取

#### Scenario:动态插件未授权 manifest 路径被拒绝

- **WHEN** 动态插件当前授权快照只允许读取`metadata.yaml`
- **AND** 插件调用`manifest.get`读取`config/config.example.yaml`
- **THEN** 系统拒绝该请求
- **AND** 不返回 artifact 中该资源内容

### Requirement:插件 manifest 专用生命周期资源可被 Manifest 读取但不得被混用为生效管线

系统 SHALL 保持插件`manifest/sql/`、`manifest/i18n/`和`manifest/config/`等专用目录的既有治理边界。`HostServices.Manifest()`用于读取插件自身打包的 manifest 原始资源，可以读取这些专用目录中的文件原文，但不得绕过 SQL 生命周期管线、i18n 资源管线或插件配置服务。插件运行期有效配置 MUST 通过`HostServices.Config()`读取。

#### Scenario:配置目录可读取原文但有效配置通过 Config 读取

- **WHEN** 插件需要读取打包的`manifest/config/config.yaml`原文
- **THEN** 插件 MAY 使用`HostServices.Manifest().Get(ctx, "config/config.yaml")`
- **AND** 该读取只返回打包原文
- **AND** 插件需要读取运行期有效配置时 MUST 使用`HostServices.Config()`

#### Scenario:SQL 和 i18n 资源继续由专用管线处理

- **WHEN** 插件安装 SQL 或 i18n 资源需要被宿主加载
- **THEN** 系统继续使用插件生命周期、数据库和 i18n 管线扫描`manifest/sql/`和`manifest/i18n/`
- **AND** `HostServices.Manifest()`不得成为执行 SQL 或加载翻译包的替代入口

#### Scenario:旧 Metadata 服务语义被移除

- **WHEN** 插件代码或动态插件清单需要读取`manifest/metadata.yaml`
- **THEN** 系统只提供`HostServices.Manifest()`或`manifest.get`作为读取入口
- **AND** 系统不得继续发布`Metadata()`、`metadata.get`或`service: metadata`读取入口
