## ADDED Requirements

### Requirement: `linactl`必须内嵌 GoFrame 代码生成入口

系统 SHALL 由`linactl`直接承载宿主 GoFrame controller 和 DAO/DO/Entity 代码生成入口。`linactl ctrl`和`linactl dao`不得要求开发者在本机预先安装`gf`，也不得在默认开发路径中下载、安装或调用`PATH`中的外部`gf`可执行文件。

#### Scenario: controller 生成不依赖外部 `gf`

- **WHEN** 开发者运行`linactl ctrl`或根目录`make ctrl`
- **THEN** 命令通过`linactl`内嵌的 GoFrame CLI module 执行`gen ctrl`
- **AND** 命令不得调用`gf`、`gf -v`、`gf install`或 GitHub release 下载地址
- **AND** 开发者不需要在`PATH`中提供`gf`可执行文件

#### Scenario: DAO 生成不依赖外部 `gf`

- **WHEN** 开发者运行`linactl dao`或根目录`make dao`
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

系统 SHALL 在执行内嵌 GoFrame CLI 生成命令时保持`apps/lina-core`作为工作目录，使`hack/config.yaml`、`api/`、`internal/`和`go.mod`解析结果与原外部`make ctrl`和`make dao`流程一致。

#### Scenario: controller 生成使用宿主工作目录

- **WHEN** `linactl ctrl`触发内嵌 GoFrame CLI
- **THEN** GoFrame CLI 在仓库根目录下的`apps/lina-core`目录中执行
- **AND** `api/`和`internal/controller`路径按宿主目录解析

#### Scenario: DAO 生成使用宿主工作目录

- **WHEN** `linactl dao`触发内嵌 GoFrame CLI
- **THEN** GoFrame CLI 在仓库根目录下的`apps/lina-core`目录中执行
- **AND** `gfcli.gen.dao`配置从宿主`hack/config.yaml`解析
- **AND** `internal/dao`、`internal/model/do`和`internal/model/entity`路径按宿主目录解析
