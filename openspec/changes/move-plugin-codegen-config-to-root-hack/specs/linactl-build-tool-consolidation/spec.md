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

### Requirement: 插件自定义构建指令必须从插件根配置读取

系统 SHALL 允许源码插件和动态插件在插件根`hack/config.yaml`中通过`build.commands`声明自定义构建指令。仓库根目录执行`make build`或`linactl build`且未指定`dir=`时，系统 MUST 遍历`apps/lina-plugins`下所有包含`plugin.yaml`的直接插件目录并执行对应插件的自定义构建指令；指定`dir=apps/lina-plugins/<plugin-id>`时，系统 MUST 只执行该插件的构建。插件自定义构建指令 MUST 在插件根目录中执行，并支持`$(PLUGIN_ROOT)`和`$(REPO_ROOT)`变量展开。

#### Scenario: 根目录构建遍历所有插件

- **WHEN** 开发者在仓库根目录运行`make build`且未指定`dir=`
- **THEN** 系统构建宿主前端、宿主打包资源和宿主后端
- **AND** 遍历`apps/lina-plugins`下所有直接插件目录
- **AND** 对每个包含`plugin.yaml`的插件执行其`hack/config.yaml`中的`build.commands`

#### Scenario: 定向构建只执行指定插件

- **WHEN** 开发者运行`make build dir=apps/lina-plugins/<plugin-id>`
- **THEN** 系统只执行该插件`hack/config.yaml`中的`build.commands`
- **AND** 不执行其他插件的构建指令

#### Scenario: 插件未声明自定义构建指令

- **WHEN** 插件根`hack/config.yaml`不存在或未声明`build.commands`
- **THEN** 系统跳过该插件的自定义构建指令
- **AND** 不将缺少自定义构建指令视为错误

### Requirement: 官方源码插件聚合模块必须自动生成

系统 SHALL 在插件完整构建、开发和动态插件构建准备过程中根据`apps/lina-plugins`下的插件清单和插件 Go module 自动生成临时源码插件聚合模块，并基于宿主根`go.work`生成临时`temp/go.work.plugins`。官方插件工作区根目录 MUST NOT 维护`go.mod`、`go.sum`或手写源码插件聚合 Go 文件作为构建事实源。自动聚合模块 MUST 使用固定 module path `lina-plugins`以满足宿主`official_plugins`构建标签下的桥接导入。

#### Scenario: 插件完整构建自动生成源码插件聚合模块

- **WHEN** 开发者运行`make build`或`make dev`且官方插件工作区存在插件清单
- **THEN** 系统扫描`apps/lina-plugins`下包含`plugin.yaml`、插件`go.mod`和`backend/plugin.go`的源码插件
- **AND** 生成`temp/official-plugins`聚合模块，自动 blank import 这些源码插件的`backend`包
- **AND** 生成`temp/go.work.plugins`，包含宿主模块、聚合模块和各插件自身 Go module

#### Scenario: 动态插件不进入源码插件聚合

- **WHEN** 官方插件工作区包含`type: dynamic`插件
- **THEN** 自动聚合模块不导入该动态插件的`backend`包
- **AND** 动态插件仍通过自身 Go module 参与`WASM`构建

#### Scenario: 官方插件工作区根模块文件不作为事实源

- **WHEN** 官方插件工作区根目录存在历史遗留的`go.mod`或`go.sum`
- **THEN** 临时`go.work`生成过程忽略该根模块
- **AND** 源码插件接入仍只由自动生成的`temp/official-plugins`聚合模块承担
