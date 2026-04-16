## MODIFIED Requirements

### Requirement: 插件目录与清单契约统一

系统 SHALL 为所有插件提供统一的目录结构与清单契约。源码插件 MUST 放置在 `apps/lina-plugins/<plugin-id>/` 目录下；当前 `dynamic` 动态 `wasm` 插件 MUST 能从 `plugin.dynamic.storagePath` 中被发现，并解析出与源码插件等价的 manifest 信息。

#### Scenario: 发现源码插件目录

- **WHEN** 宿主扫描 `apps/lina-plugins/` 下的插件目录
- **THEN** 仅将包含合法清单文件的目录识别为插件
- **AND** 每个插件的 `plugin-id` 在宿主范围内唯一
- **AND** 清单仅需包含插件基础信息与一级插件类型

#### Scenario: `plugin.yaml` 保持精简且可声明插件菜单

- **WHEN** 宿主解析 `plugin.yaml`
- **THEN** 清单只强制要求 `id`、`name`、`version`、`type` 等基础字段
- **AND** 宿主不再要求 `schemaVersion`、`compatibility`、`entry` 等扩展元数据
- **AND** 插件若需要向宿主注册菜单或按钮权限，必须在清单 `menus` 元数据中声明
- **AND** 前端页面、`Slot` 与 SQL 文件位置仍优先按照目录与代码约定推导，而不是在清单中重复配置

#### Scenario: 清单一级类型只保留源码与动态两类

- **WHEN** 宿主解析 `plugin.yaml` 中的 `type`
- **THEN** `type` 仅允许 `source` 或 `dynamic`
- **AND** 当前仅 `wasm` 作为动态插件的产物语义，不再作为一级插件类型
- **AND** 对历史上的 `wasm` 一级类型值，宿主在治理视角下统一按 `dynamic` 处理

#### Scenario: 安装动态插件产物

- **WHEN** 管理员上传一个 `wasm` 文件安装动态插件
- **THEN** 宿主能够解析出与源码模式一致的插件标识、名称、版本与一级插件类型
- **AND** 对缺少这些基础字段的动态插件拒绝安装
- **AND** 宿主将上传产物写入 `plugin.dynamic.storagePath/<plugin-id>.wasm`

#### Scenario: 动态插件通过嵌入资源声明生成清单与 SQL 快照

- **WHEN** 动态插件作者使用 `go:embed` 声明 `plugin.yaml`、`manifest/sql` 与 `manifest/sql/uninstall`
- **THEN** 构建器必须从该嵌入文件系统中读取这些资源
- **AND** 运行时产物中嵌入的 manifest 与 SQL 快照必须继续作为宿主安装、上传和生命周期治理的真相源
- **AND** 宿主不得改为通过 guest 运行时方法动态获取这些治理资源

#### Scenario: 动态插件产物使用独立存储目录

- **WHEN** 宿主发现、上传或同步一个 `dynamic` `wasm` 动态插件产物
- **THEN** 运行时产物 MUST 使用 `plugin.dynamic.storagePath/<plugin-id>.wasm` 作为宿主侧规范落盘路径
- **AND** 宿主不得再依赖 `apps/lina-plugins/<plugin-id>/plugin.yaml` 作为 runtime 发现入口
- **AND** 运行时样例插件的可读源码目录 SHOULD 与源码插件一样继续收敛在 `backend/`、`frontend/` 与 `manifest/` 下维护

#### Scenario: 当前生效 release 从稳定归档重载

- **WHEN** 一个动态插件已经存在当前生效的 release，且宿主需要再次加载其 active manifest
- **THEN** 宿主从 `plugin.dynamic.storagePath/releases/<plugin-id>/<version>/<plugin-id>.wasm` 这类稳定归档路径重载该 release
- **AND** 宿主不会因为 staging 目录里出现更新的 `wasm` 文件就立即替换当前服务中的 active release
- **AND** active manifest 在重载后仍然包含该 release 内嵌声明的 Hook、通用资源契约与菜单元数据
