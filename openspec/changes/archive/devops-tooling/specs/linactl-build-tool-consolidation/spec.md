# linactl-build-tool-consolidation Specification

## Purpose

定义`linactl`统一承载仓库工具实现、GoFrame 代码生成入口和目标目录支持的能力边界，覆盖镜像构建、动态插件 Wasm 打包、运行时 i18n 治理扫描、controller/DAO 生成和 Makefile 收敛。

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

系统 SHALL 在执行内嵌 GoFrame CLI 生成命令时使用明确的目标后端目录作为工作目录，使`hack/config.yaml`、`api/`、`internal/`和`go.mod`解析结果与目标项目一致。未指定目标时，默认目标 MUST 为仓库根目录下的`apps/lina-core`；指定插件 ID 时，目标 MUST 为仓库根目录下的`apps/lina-plugins/<plugin-id>/backend`；指定目录时，目标 MUST 为该目录。任何目标目录都 MUST 包含`hack/config.yaml`，否则命令必须拒绝执行并返回清晰错误。

#### Scenario: controller 生成默认使用宿主工作目录

- **WHEN** `linactl ctrl`或根目录`make ctrl`未指定插件或目录目标
- **THEN** GoFrame CLI 在仓库根目录下的`apps/lina-core`目录中执行
- **AND** `api/`和`internal/controller`路径按宿主目录解析

#### Scenario: DAO 生成默认使用宿主工作目录

- **WHEN** `linactl dao`或根目录`make dao`未指定插件或目录目标
- **THEN** GoFrame CLI 在仓库根目录下的`apps/lina-core`目录中执行
- **AND** `gfcli.gen.dao`配置从宿主`hack/config.yaml`解析
- **AND** `internal/dao`、`internal/model/do`和`internal/model/entity`路径按宿主目录解析

#### Scenario: controller 生成使用插件后端工作目录

- **WHEN** 开发者在插件根目录运行`make ctrl`或在根目录运行带插件 ID 的`make ctrl`
- **THEN** GoFrame CLI 在对应`apps/lina-plugins/<plugin-id>/backend`目录中执行
- **AND** `api/`和`internal/controller`路径按插件后端目录解析

#### Scenario: DAO 生成使用插件后端工作目录

- **WHEN** 开发者在插件根目录运行`make dao`或在根目录运行带插件 ID 的`make dao`
- **THEN** GoFrame CLI 在对应`apps/lina-plugins/<plugin-id>/backend`目录中执行
- **AND** `gfcli.gen.dao`配置从插件`backend/hack/config.yaml`解析
- **AND** `internal/dao`、`internal/model/do`和`internal/model/entity`路径按插件后端目录解析

#### Scenario: 缺少 GoFrame 配置的目标目录被拒绝

- **WHEN** 开发者将代码生成目标指向没有`hack/config.yaml`的目录
- **THEN** `linactl`拒绝执行 GoFrame 生成命令
- **AND** 错误消息说明目标目录缺少`hack/config.yaml`

### Requirement: 插件自定义构建指令必须从插件根配置读取

系统 SHALL 允许源码插件和动态插件在插件根`hack/config.yaml`中通过`build.commands`声明自定义构建指令。仓库根目录执行`make build`或`linactl build`且未指定`dir=`时，系统 MUST 遍历`apps/lina-plugins`下所有包含`plugin.yaml`的直接插件目录并执行对应插件的自定义构建指令；指定`dir=apps/lina-plugins/<plugin-id>`时，系统 MUST 只执行该插件的构建。插件自定义构建指令 MUST 在插件根目录中执行，并且变量展开 MUST 只包含`$(TARGET_DIR)`、`$(BUILD_DIR)`和`$(REPO_ROOT)`。非插件或插件定向构建目标 MUST 以`hack/config.yaml`作为唯一自定义命令来源；目标目录缺少`hack/config.yaml`时不得回退读取`package.json`构建脚本。

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

#### Scenario: 仅标准构建变量被展开

- **WHEN** 插件或定向构建目标的`hack/config.yaml`中使用`$(TARGET_DIR)`、`$(BUILD_DIR)`或`$(REPO_ROOT)`
- **THEN** 系统按标准变量集合展开这些值
- **AND** 不包含旧变量名的专门识别或处理分支

#### Scenario: 定向构建目标没有配置文件

- **WHEN** 开发者运行`linactl build dir=<target>`
- **AND** `<target>/hack/config.yaml`不存在
- **THEN** 系统拒绝该定向构建
- **AND** 不读取`<target>/package.json`中的`build`脚本


### Requirement: `linactl`公开参数必须只接受标准契约

系统 SHALL 只接受当前标准化的`linactl`参数和配置输入。显式插件模式参数 MUST 使用标准布尔值，文档示例 SHOULD 使用`plugins=0`或`plugins=1`；未传入`plugins`时，系统 MAY 根据官方插件工作区是否存在插件清单自动选择宿主模式或插件完整模式。系统 MUST reject 显式`plugins=auto`，且不得为该旧值保留专门判断分支。公开命令参数`key`MUST 使用文档中声明的`kebab-case`形式，系统不得将`snake_case key`自动映射为`kebab-case key`，也不得将`kebab-case key`自动映射为`snake_case key`；根`Makefile`包装入口也 MUST 只转发当前标准参数名。动态`WASM`单插件构建的显式源码目录 MUST 使用`dir=<path>`，`wasm`命令不得读取`p`或`plugin-dir`作为插件选择或路径输入。布尔参数 MUST 只接受`true`、`false`、`1`和`0`；动态`WASM`生成派发器的布尔路由值也 MUST 使用同一标准集合。发布标签校验 MUST 通过显式`tag=<version>`参数传入待校验标签，不得读取环境变量作为隐式标签来源。镜像构建 registry MUST 只来自`hack/config.yaml`或显式`registry=<prefix>`参数，不得读取`LINAPRO_IMAGE_REGISTRY`环境变量。镜像构建内部入口不得保留环境打印调试参数或选项，也不得为这些旧调试入口保留专门判断分支。

#### Scenario: 省略插件模式时自动探测

- **WHEN** 开发者运行`linactl build`且未传入`plugins`
- **THEN** 系统根据`apps/lina-plugins`下是否存在插件清单选择宿主模式或插件完整模式
- **AND** 该自动探测不需要公开`plugins=auto`参数值

#### Scenario: 显式`plugins=auto`被拒绝

- **WHEN** 开发者运行`linactl build plugins=auto`
- **THEN** 系统拒绝执行
- **AND** 错误消息说明`plugins`只接受标准布尔值，自动探测应通过省略该参数触发

#### Scenario: 下划线参数 key 不再映射

- **WHEN** 开发者运行下划线形式的基础镜像参数
- **THEN** 系统不将该参数解释为`base-image`
- **AND** 需要覆盖基础镜像时必须使用`base-image=alpine:3.22`

#### Scenario: Make 包装入口不转发旧参数名

- **WHEN** 开发者通过`make release.tag.check`调用发布校验
- **THEN** 根`Makefile`只转发`print-version`参数
- **AND** 不再转发下划线形式的旧参数名

#### Scenario: `wasm`单插件构建使用`dir`

- **WHEN** 开发者运行`linactl wasm dir=apps/lina-plugins/<plugin-id>`或`make wasm dir=apps/lina-plugins/<plugin-id>`
- **THEN** 系统只从该目录构建单个动态插件产物
- **AND** 相对目录按仓库根目录解析

#### Scenario: `wasm`不读取旧插件选择参数

- **WHEN** 开发者运行`linactl wasm p=<plugin-id>`或`linactl wasm plugin-dir=<path>`
- **THEN** 系统不得将`p`解释为单插件选择
- **AND** 不得将`plugin-dir`解释为显式插件源码目录
- **AND** 根`Makefile`的`wasm`目标不得转发`p`参数

#### Scenario: 宽松布尔别名被拒绝

- **WHEN** 开发者运行`linactl image push=on`
- **THEN** 系统拒绝该布尔值
- **AND** 错误消息说明布尔值只接受`true`、`false`、`1`或`0`

#### Scenario: 动态`WASM`布尔路由值只使用标准集合

- **WHEN** 动态`WASM`生成派发器解析布尔路由值
- **THEN** 系统只将`true`和`1`解析为`true`
- **AND** 其它非标准布尔文本不得作为`true`处理

#### Scenario: 发布标签校验只接受显式标签

- **WHEN** 开发者运行`linactl release.tag.check`
- **AND** 未传入`tag=<version>`
- **THEN** 系统拒绝执行
- **AND** 不从环境变量中读取待校验标签

#### Scenario: 镜像 registry 不读取环境变量

- **WHEN** 环境变量`LINAPRO_IMAGE_REGISTRY`存在
- **AND** 开发者运行`linactl image`且未传入`registry`
- **THEN** 系统不使用该环境变量覆盖镜像 registry
- **AND** registry 仅来自`hack/config.yaml`中的`image.registry`

#### Scenario: 镜像构建环境打印入口不可用

- **WHEN** 调用方传入环境打印调试参数或内部选项
- **THEN** 系统不得进入环境打印流程
- **AND** 不输出构建环境兼容信息
- **AND** 不保留旧调试入口的专门判断分支
