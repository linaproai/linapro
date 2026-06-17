## MODIFIED Requirements

### Requirement: GoFrame 代码生成必须保持宿主工作目录语义

系统 SHALL 在执行内嵌 GoFrame CLI 生成命令时使用明确的目标后端目录作为工作目录，使`api/`、`internal/`和`go.mod`解析结果与目标项目一致。系统 MUST 同时维护明确的配置目录，使宿主目标读取`apps/lina-core/hack/config.yaml`，标准插件后端目标读取插件根`apps/lina-plugins/<plugin-id>/hack/config.yaml`，非标准插件目标继续读取目标目录下的`hack/config.yaml`。未指定目标时，默认目标 MUST 为仓库根目录下的`apps/lina-core`。`dao`生成目标缺少配置文件时必须拒绝执行并返回清晰错误；`ctrl`生成目标缺少配置文件时 MAY 使用临时空配置目录继续执行。

#### Scenario: controller 生成默认使用宿主工作目录

- **WHEN** `linactl ctrl`或根目录`make ctrl`未指定目录目标
- **THEN** GoFrame CLI 在仓库根目录下的`apps/lina-core`目录中执行
- **AND** 配置目录为`apps/lina-core/hack`
- **AND** `api/`和`internal/controller`路径按宿主目录解析

#### Scenario: DAO 生成默认使用宿主工作目录

- **WHEN** `linactl dao`或根目录`make dao`未指定目录目标
- **THEN** GoFrame CLI 在仓库根目录下的`apps/lina-core`目录中执行
- **AND** `gfcli.gen.dao`配置从宿主`apps/lina-core/hack/config.yaml`解析
- **AND** `internal/dao`、`internal/model/do`和`internal/model/entity`路径按宿主目录解析

#### Scenario: controller 生成使用插件后端工作目录和插件根配置目录

- **WHEN** 开发者在插件根目录运行`make ctrl`或在根目录运行`linactl ctrl dir=apps/lina-plugins/<plugin-id>/backend`
- **THEN** GoFrame CLI 在对应`apps/lina-plugins/<plugin-id>/backend`目录中执行
- **AND** 配置目录为对应插件根`apps/lina-plugins/<plugin-id>/hack`
- **AND** `api/`和`internal/controller`路径按插件后端目录解析

#### Scenario: DAO 生成使用插件后端工作目录和插件根配置目录

- **WHEN** 开发者在插件根目录运行`make dao`或在根目录运行`linactl dao dir=apps/lina-plugins/<plugin-id>/backend`
- **THEN** GoFrame CLI 在对应`apps/lina-plugins/<plugin-id>/backend`目录中执行
- **AND** `gfcli.gen.dao`配置从插件根`apps/lina-plugins/<plugin-id>/hack/config.yaml`解析
- **AND** `internal/dao`、`internal/model/do`和`internal/model/entity`路径按插件后端目录解析

#### Scenario: 非插件目标继续读取目标目录配置

- **WHEN** 开发者运行`linactl dao dir=<non-plugin-target>`或`linactl ctrl dir=<non-plugin-target>`
- **THEN** GoFrame CLI 在`<non-plugin-target>`目录中执行
- **AND** 配置目录为`<non-plugin-target>/hack`
- **AND** 目标不被强制解释为官方插件目录

#### Scenario: DAO 目标缺少配置文件时被拒绝

- **WHEN** 开发者将`linactl dao`目标指向缺少有效`config.yaml`的配置目录
- **THEN** `linactl`拒绝执行 GoFrame DAO 生成命令
- **AND** 错误消息说明缺少的配置文件路径

#### Scenario: controller 目标缺少配置文件时使用空配置目录

- **WHEN** 开发者将`linactl ctrl`目标指向缺少有效`config.yaml`的配置目录
- **THEN** `linactl`使用临时空配置目录执行 GoFrame controller 生成
- **AND** 生成工作目录仍为目标后端目录

## ADDED Requirements

### Requirement: GoFrame 代码生成目标选择参数必须收敛到`dir=`

系统 SHALL 要求`linactl ctrl`和`linactl dao`只使用`dir=`作为显式目标选择参数。`p=`、`plugin=`、`target=`和其他未知参数 MUST 被拒绝，并返回清晰错误说明仅支持`dir=`。未传入`dir=`时，系统 MUST 使用宿主默认目标。

#### Scenario: 使用`dir=`选择插件后端目标

- **WHEN** 开发者运行`linactl dao dir=apps/lina-plugins/<plugin-id>/backend`
- **THEN** 命令解析该目录为代码生成工作目录
- **AND** 不需要额外传入插件 ID 或目标名称

#### Scenario: 旧插件 ID 参数被拒绝

- **WHEN** 开发者运行`linactl dao p=<plugin-id>`或`linactl ctrl plugin=<plugin-id>`
- **THEN** 命令失败
- **AND** 错误消息说明`p=`和`plugin=`不再受支持，必须使用`dir=`

#### Scenario: 旧目标参数被拒绝

- **WHEN** 开发者运行`linactl dao target=<path>`
- **THEN** 命令失败
- **AND** 错误消息说明`target=`不再受支持，必须使用`dir=`
