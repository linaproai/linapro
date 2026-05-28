# 插件嵌入快照打包规范

## Purpose
待定 - 由归档变更 dynamic-plugin-embed-snapshot 创建。归档后更新目的。
## Requirements
### Requirement: 动态插件必须支持通过 `go:embed` 声明资源集合

系统 SHALL 允许动态插件像源码插件一样，通过显式的`go:embed`资源声明文件暴露`plugin.yaml`、`frontend`、`manifest`等资源集合，作为作者侧资源打包入口。`manifest`集合 SHALL 保留`manifest/config/`以及插件实际提供的`manifest/metadata.yaml`和其他声明型 YAML 资源的源码路径语义。

#### Scenario: 动态插件声明嵌入资源

- **WHEN** 一个动态插件需要把清单、前端静态资源、安装 SQL、运行期配置模板或声明型 manifest 资源随`.wasm`一起交付
- **THEN** 插件可以通过`go:embed`在代码中声明这些资源
- **AND** 资源声明路径必须保持与插件目录约定一致
- **AND** 构建器必须能够基于该声明读取插件资源

#### Scenario: 动态插件嵌入 manifest config

- **WHEN** 动态插件通过`go:embed`声明`manifest/**`
- **AND** 插件源码包含`manifest/config/config.example.yaml`和`manifest/config/config.yaml`
- **THEN** 构建器在资源集合中保留这两个文件的相对路径
- **AND** 不把它们移动到宿主`manifest/config`或其他全局目录

### Requirement: 构建器必须把嵌入资源转换为宿主快照区段

系统 SHALL 由动态插件构建器把作者侧`go:embed`资源转换为宿主可治理的 Wasm 自定义节快照，而不是要求宿主在运行时直接读取 guest 资源文件系统。

#### Scenario: 从嵌入资源生成快照区段

- **WHEN** 构建器处理一个声明了嵌入资源的动态插件
- **THEN** 构建器必须从该嵌入文件系统中读取`plugin.yaml`、前端资源、SQL 资源、`manifest/config/`资源和声明型 manifest 资源
- **AND** 构建器必须继续生成宿主已识别的 manifest、前端资源、SQL、配置和通用 manifest 资源自定义节
- **AND** 生成后的`.wasm`产物仍可被宿主现有动态插件治理链路直接消费

#### Scenario: 快照区段保留 metadata

- **WHEN** 动态插件源码包含`manifest/metadata.yaml`
- **THEN** 构建器生成的宿主快照区段保留该文件内容和相对路径
- **AND** 宿主运行时可通过动态插件资源视图向`Manifest()`提供该资源

### Requirement: 动态插件构建必须保留目录扫描兼容回退

系统 SHALL 在迁移期继续允许未声明`go:embed`的动态插件通过目录扫描方式完成构建，避免现有插件立即失效。目录扫描回退 MUST 使用与`go:embed`路径一致的资源语义，包含`plugin.yaml`、前端资源、SQL 资源、`manifest/config/`资源和声明型 manifest 资源。

#### Scenario: 旧动态插件未声明嵌入资源

- **WHEN** 构建器处理一个尚未新增`go:embed`资源声明的动态插件
- **THEN** 构建器仍可按现有目录扫描方式读取`plugin.yaml`、前端资源、SQL 资源、`manifest/config/`资源和声明型 manifest 资源
- **AND** 宿主侧无需因为作者侧迁移节奏不同而修改运行时治理逻辑
- **AND** 新旧两类插件生成的快照区段格式必须保持兼容

#### Scenario: 目录扫描保留 manifest config 路径

- **WHEN** 目录扫描回退发现`manifest/config/config.yaml`
- **THEN** 构建产物中的资源路径仍为`manifest/config/config.yaml`
- **AND** 不生成动态插件专属的额外配置路径或额外清单字段

### Requirement: 动态插件构建产物必须收敛到统一输出目录
系统 SHALL 把动态插件的最终 `.wasm` 产物以及 guest runtime 编译阶段产生的中间 `wasm` 文件统一收敛到仓库根 `temp/output/` 或调用方显式指定的输出目录，而不是回写到 `apps/lina-plugins/<plugin-id>/temp/` 源码目录。

#### Scenario: 使用仓库默认输出目录构建动态插件
- **WHEN** 开发者通过仓库标准入口构建一个动态插件
- **THEN** 最终运行时产物写入仓库根 `temp/output/<plugin-id>.wasm`
- **AND** guest runtime 的中间 `wasm` 文件仅允许写入该统一输出目录下的内部工作子目录
- **AND** 构建器不得在插件源码目录下留下编译后的 `wasm` 文件

#### Scenario: 调用方显式指定输出目录
- **WHEN** 调用方通过构建参数显式指定输出目录
- **THEN** 最终运行时产物写入该输出目录
- **AND** guest runtime 的中间 `wasm` 文件也必须收敛在该输出目录或其内部工作子目录下
- **AND** 构建器不得因为显式输出目录为空以外的情况回退到插件源码目录写入编译产物

### Requirement: 动态插件发布产物必须保留插件配置和 manifest 资源路径语义

动态插件发布产物 SHALL 保留与源码插件一致的插件配置和 manifest 资源路径语义。构建工具不得为`manifest/config/`、插件实际提供的`metadata.yaml`或其他声明型 YAML 资源引入不同于源码插件的额外路径约定或额外清单字段，除非先通过 OpenSpec 变更明确设计。

#### Scenario: 发布产物携带配置模板和实际配置

- **WHEN** 动态插件源码包含`manifest/config/config.example.yaml`和`manifest/config/config.yaml`
- **THEN** 发布产物携带这两个资源
- **AND** 宿主运行时仍按`manifest/config/config.example.yaml`识别模板，按`manifest/config/config.yaml`识别 artifact 默认运行配置

#### Scenario: 发布产物携带 metadata

- **WHEN** 动态插件源码包含`manifest/metadata.yaml`
- **THEN** 发布产物携带该资源
- **AND** 宿主运行时可通过当前插件资源视图向`Manifest()`读取提供该文件

#### Scenario: 发布产物不要求 metadata 占位文件

- **WHEN** 动态插件源码未包含`manifest/metadata.yaml`
- **THEN** 构建工具不得要求插件作者提交空白 metadata 文件
- **AND** 发布产物仍可携带`manifest/config/config.example.yaml`、SQL、i18n 和其他实际存在的插件资源

#### Scenario: 构建工具不创建宿主配置副本

- **WHEN** 构建动态插件发布产物
- **THEN** 构建工具不得把插件`manifest/config/config.yaml`复制到`apps/lina-core/manifest/config`
- **AND** 不得要求插件作者维护宿主全局配置键作为插件运行期配置

