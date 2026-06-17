# linactl-build-tool-consolidation Specification

## Purpose
TBD - created by archiving change consolidate-linactl-build-tools. Update Purpose after archive.
## Requirements
### Requirement: `linactl`必须统一承载仓库工具实现

系统 SHALL 将镜像构建、动态插件`Wasm`打包与运行时`i18n`治理扫描实现作为`hack/tools/linactl/internal/`下的内部子组件维护。`hack/tools/image-builder`、`hack/tools/build-wasm`与`hack/tools/runtime-i18n`不得继续作为仓库默认开发路径中的独立`Go`工具模块存在。

#### Scenario: 镜像构建实现通过`linactl`内部组件执行

- **WHEN** 开发者运行`linactl image`或`linactl image.build`
- **THEN** 命令直接调用`hack/tools/linactl/internal/imagebuilder`中的实现
- **AND** 命令不得再通过`go run ./hack/tools/image-builder`调用独立工具

#### Scenario: 动态插件`Wasm`打包实现通过`linactl`内部组件执行

- **WHEN** 开发者运行`linactl wasm`
- **THEN** 命令直接调用`hack/tools/linactl/internal/wasmbuilder`中的实现
- **AND** 命令不得再通过进入`hack/tools/build-wasm`目录执行`go run .`调用独立工具

#### Scenario: 运行时`i18n`治理扫描通过`linactl`内部组件执行

- **WHEN** 开发者运行`linactl i18n.check`
- **THEN** 命令直接调用`hack/tools/linactl/internal/runtimei18n`中的实现
- **AND** 命令不得再通过进入`hack/tools/runtime-i18n`目录执行`go run . scan`或`go run . messages`调用独立工具

### Requirement: 公开开发命令必须保持稳定

系统 SHALL 保持`make image`、`make image.build`、`make wasm`、`make i18n.check`、`make ctrl`、`make dao`、`linactl image`、`linactl image.build`、`linactl wasm`、`linactl i18n.check`、`linactl ctrl`和`linactl dao`的公开入口稳定。工具实现迁移不得要求开发者改用新的命令名称。宿主和插件目录的本地`Makefile` MUST 仅作为薄转发层调用仓库统一工具链，不得重新承载外部`gf`安装、代码生成业务逻辑或与根`Makefile`重复的构建治理逻辑。

#### Scenario: Make 入口继续调用`linactl`

- **WHEN** 开发者运行`make image`、`make image.build`、`make wasm`、`make i18n.check`、`make ctrl`或`make dao`
- **THEN** 根`Makefile`转发到对应`linactl`命令
- **AND** 开发者不需要直接调用旧独立工具目录或外部`gf`

#### Scenario: 宿主本地 Makefile 只保留薄转发层

- **WHEN** 开发者查看`apps/lina-core/Makefile`
- **THEN** 该文件只保留指向仓库统一工具链的宿主相关薄入口
- **AND** 不包含安装外部`gf`、更新外部`gf`或重复维护根构建治理逻辑的目标

#### Scenario: 插件根目录提供一致代码生成入口

- **WHEN** 开发者查看任一官方插件根目录`apps/lina-plugins/<plugin-id>/Makefile`
- **THEN** 该文件提供`ctrl`和`dao`目标
- **AND** 两个目标转发到仓库统一`linactl`并以该插件`backend/`作为生成目标
- **AND** 代码生成目标逻辑集中维护在根目录`hack/makefiles/plugin.codegen.mk`
- **AND** 插件根目录`Makefile`不得硬编码`apps/lina-plugins/<plugin-id>/backend`

### Requirement: 工具整合必须更新仓库引用

系统 SHALL 移除默认开发路径中对`hack/tools/image-builder`、`hack/tools/build-wasm`和`hack/tools/runtime-i18n`的直接引用，包括`go.work`、`CI`夹具、测试辅助和工具文档。历史`OpenSpec`归档内容可以继续保留旧路径作为历史记录。

#### Scenario: 默认工作区不再包含旧工具模块

- **WHEN** 开发者查看根`go.work`
- **THEN** 工作区只包含仍然存在的`Go`模块
- **AND** 不包含已删除的`hack/tools/image-builder`、`hack/tools/build-wasm`或`hack/tools/runtime-i18n`

#### Scenario: 引用扫描不发现默认路径旧入口

- **WHEN** 开发者扫描当前代码、文档和工作流中的`go run ./hack/tools/image-builder`、`go run ./hack/tools/build-wasm`、`go run ./hack/tools/runtime-i18n`、`hack/tools/image-builder`、`hack/tools/build-wasm`和`hack/tools/runtime-i18n`
- **THEN** 除历史归档说明外，不再存在要求使用旧独立工具入口的引用

### Requirement: `linactl`必须内嵌 GoFrame 代码生成入口

系统 SHALL 由`linactl`直接承载宿主和插件后端的 GoFrame controller 与 DAO/DO/Entity 代码生成入口。`linactl ctrl`和`linactl dao`不得要求开发者在本机预先安装`gf`，也不得在默认开发路径中下载、安装或调用`PATH`中的外部`gf`可执行文件。`linactl ctrl`和`linactl dao`默认以`apps/lina-core`作为生成目标；当调用方显式传入插件 ID 或后端目录时，命令 MUST 在对应插件`backend/`目录中执行同一套内嵌 GoFrame 生成流程。

#### Scenario: controller 生成不依赖外部 `gf`

- **WHEN** 开发者运行`linactl ctrl`、根目录`make ctrl`、插件目录`make ctrl`或带插件目标的根目录`make ctrl`
- **THEN** 命令通过`linactl`内嵌的 GoFrame CLI module 执行`gen ctrl`
- **AND** 命令不得调用`gf`、`gf -v`、`gf install`或 GitHub release 下载地址
- **AND** 开发者不需要在`PATH`中提供`gf`可执行文件

#### Scenario: DAO 生成不依赖外部 `gf`

- **WHEN** 开发者运行`linactl dao`、根目录`make dao`、插件目录`make dao`或带插件目标的根目录`make dao`
- **THEN** 命令通过`linactl`内嵌的 GoFrame CLI module 执行`gen dao`
- **AND** 命令不得调用`gf`、`gf -v`、`gf install`或 GitHub release 下载地址
- **AND** 开发者不需要在`PATH`中提供`gf`可执行文件

### Requirement: GoFrame CLI 版本必须由仓库锁定

系统 SHALL 在`hack/tools/linactl`工具模块中显式依赖 GoFrame CLI module，并保持该 module 使用的 GoFrame runtime 版本与宿主`apps/lina-core`使用的 GoFrame runtime 版本一致。默认代码生成路径不得使用 GoFrame CLI 的`latest`发布二进制。

#### Scenario: 依赖版本与宿主 runtime 对齐

- **WHEN** 开发者查看`hack/tools/linactl/go.mod`
- **THEN** 文件显式声明`github.com/gogf/gf/cmd/gf/v2`
- **AND** 该 module 版本与宿主使用的`github.com/gogf/gf/v2`版本保持一致

#### Scenario: 默认生成路径不使用 latest 二进制

- **WHEN** 开发者扫描`linactl ctrl`和`linactl dao`的默认执行路径
- **THEN** 不存在从`github.com/gogf/gf/releases/latest`下载 GoFrame CLI 的逻辑
- **AND** 不存在为了代码生成而安装外部`gf`二进制的逻辑

### Requirement: GoFrame 代码生成必须通过隐藏隔离入口执行

系统 SHALL 通过`linactl`内部隐藏命令执行内嵌 GoFrame CLI。公开`ctrl`和`dao`命令负责分发到隐藏入口，隐藏入口负责调用`gfcmd.GetCommand(ctx)`并以 GoFrame CLI 参数语义执行白名单内的生成命令。GoFrame CLI 的`Fatalf`或`os.Exit`影响范围必须被限制在隐藏执行进程内。

#### Scenario: `ctrl` 分发到隐藏 GoFrame 入口

- **WHEN** 开发者运行`linactl ctrl`
- **THEN** 父命令启动`linactl`隐藏 GoFrame 执行入口
- **AND** 隐藏入口收到等价于`make ctrl`的参数
- **AND** 父命令不在当前进程中直接执行 GoFrame 生成器对象

#### Scenario: `dao` 分发到隐藏 GoFrame 入口

- **WHEN** 开发者运行`linactl dao`
- **THEN** 父命令启动`linactl`隐藏 GoFrame 执行入口
- **AND** 隐藏入口收到等价于`make dao`的参数
- **AND** 父命令不在当前进程中直接执行 GoFrame 生成器对象

#### Scenario: 隐藏入口拒绝非生成命令

- **WHEN** 调用方直接运行隐藏 GoFrame 入口并传入`install`、`build`、`docker`、`run`、`pack`、`env`或未知命令
- **THEN** `linactl`拒绝执行该命令
- **AND** 返回清晰错误说明隐藏入口只支持`gen ctrl`和`gen dao`

### Requirement: GoFrame 代码生成必须保持宿主工作目录语义

系统 SHALL 在执行内嵌 GoFrame CLI 生成命令时使用明确的目标后端目录作为工作目录，使`api/`、`internal/`和`go.mod`解析结果与目标项目一致。系统 MUST 同时维护明确的配置目录，使宿主目标读取`apps/lina-core/hack/config.yaml`，标准插件后端目标读取插件根`apps/lina-plugins/<plugin-id>/hack/config.yaml`，非标准插件目标继续读取目标目录下的`hack/config.yaml`。未指定目标时，默认目标 MUST 为仓库根目录下的`apps/lina-core`。`dao`生成目标缺少配置文件时必须拒绝执行并返回清晰错误；`ctrl`生成目标缺少配置文件时 MAY 使用临时空配置目录继续执行。

#### Scenario: controller 生成默认使用宿主工作目录

- **WHEN** `linactl ctrl`或根目录`make ctrl`未指定目录目标
- **THEN** GoFrame CLI 在仓库根目录下的`apps/lina-core`目录中执行
- **AND** 配置目录为`apps/lina-core/hack`
- **AND** `api/`和`internal/controller`路径按宿主目录解析

#### Scenario: DAO 生成默认使用宿主工作目录

- **WHEN** `linactl dao`或根目录`make dao`未指定插件或目录目标
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
