## ADDED Requirements

### Requirement:插件 manifest 资源必须支持插件自作用域只读读取

系统 SHALL 为源码插件和动态插件提供`HostServices.Manifest()`能力，使插件代码能够只读读取当前插件`manifest/`目录下的声明型资源。读取范围 MUST 绑定当前插件 ID，不得允许插件读取宿主 manifest、其他插件 manifest、任意文件系统路径或 URL。`metadata.yaml` SHALL 在插件实际提供该文件时作为可通过该能力读取的插件声明资源，但系统不得要求所有插件都提交`metadata.yaml`。

#### Scenario:源码插件读取自身 metadata

- **WHEN** 源码插件`plugin-a`调用`HostServices.Manifest().Get(ctx, "metadata.yaml")`
- **AND** `apps/lina-plugins/plugin-a/manifest/metadata.yaml`存在
- **THEN** 系统返回该文件内容
- **AND** 读取作用域限定为`plugin-a`的`manifest/`目录

#### Scenario:插件未提供 metadata

- **WHEN** 插件未维护`manifest/metadata.yaml`
- **THEN** 插件无需为了目录规范提交空白或占位 metadata 文件
- **AND** 插件清单不得申请读取不存在的`metadata.yaml`

#### Scenario:动态插件读取 artifact 中的 metadata

- **WHEN** 动态插件`plugin-a`调用`manifest.get`读取`metadata.yaml`
- **AND** 当前 active release artifact 携带`manifest/metadata.yaml`
- **THEN** 系统从该 active release 的资源快照返回文件内容
- **AND** 该内容绑定当前 active release 的 checksum 或 generation

#### Scenario:插件扫描 metadata 到结构体

- **WHEN** 插件调用`HostServices.Manifest().Scan(ctx, "metadata.yaml", "", &metadata)`
- **AND** `metadata.yaml`是合法 YAML 文档
- **THEN** 系统将文件内容绑定到插件提供的结构体
- **AND** 结构体业务语义和验证逻辑由插件内部维护

### Requirement:Manifest 资源读取必须执行路径安全治理

系统 SHALL 将`Manifest()`的路径参数解释为相对当前插件`manifest/`根目录的 slash 路径。系统 MUST 拒绝空根读取、绝对路径、路径穿越、Windows drive path、URL、跨插件路径和未授权路径；动态插件还 MUST 按`plugin.yaml`中的`manifest.resources.paths`与宿主确认后的授权快照校验。

#### Scenario:合法相对路径被允许

- **WHEN** 插件读取`metadata.yaml`或`resources/policy.yaml`
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

### Requirement:插件 manifest 专用生命周期资源不得被通用 Manifest 读取混用

系统 SHALL 保持插件`manifest/sql/`、`manifest/i18n/`和`manifest/config/`等专用目录的既有治理边界。`HostServices.Manifest()`用于读取声明型元数据资源，不得绕过 SQL 生命周期管线、i18n 资源管线或插件配置服务。插件运行期配置 MUST 通过`HostServices.Config()`读取。

#### Scenario:配置目录通过 Config 读取

- **WHEN** 插件需要读取`manifest/config/config.yaml`中的运行期配置
- **THEN** 插件使用`HostServices.Config()`读取配置键或扫描配置段
- **AND** 不通过`HostServices.Manifest()`把配置文件作为普通 manifest 资源读取

#### Scenario:SQL 和 i18n 资源继续由专用管线处理

- **WHEN** 插件安装 SQL 或 i18n 资源需要被宿主加载
- **THEN** 系统继续使用插件生命周期、数据库和 i18n 管线扫描`manifest/sql/`和`manifest/i18n/`
- **AND** `HostServices.Manifest()`不得成为执行 SQL 或加载翻译包的替代入口
