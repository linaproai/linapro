# 发布镜像构建规范

## Purpose

定义宿主二进制跨平台构建、多架构 Docker 镜像发布和 nightly 发布工作流能力。
## Requirements
### Requirement:发布构建命令必须支持跨平台宿主二进制构建

系统 SHALL 允许运维人员通过 `make build` 构建指定目标平台的宿主二进制。目标平台 SHALL 由 `hack/config.yaml` 中的 `build.platforms` 数组统一配置，每个数组项 SHALL 支持 `<goos>/<goarch>` 形式，也 SHALL 支持 `auto`；`auto` SHALL 按当前执行环境的 `runtime.GOOS/runtime.GOARCH` 自动解析为一个目标平台。命令行 SHALL 支持通过 `platforms=<goos>/<goarch>[,<goos>/<goarch>...]` 覆盖配置文件。多平台构建时，前端生产资源、宿主 `manifest` 嵌入资源与动态插件 `Wasm` 产物 SHALL 只准备一次；宿主后端二进制 SHALL 按目标平台分别交叉编译并输出到可区分的平台目录。

#### Scenario:构建单一非本机架构宿主二进制
- **当** 运维人员运行 `make build platforms=linux/arm64`
- **则** 构建流程使用 `GOOS=linux` 与 `GOARCH=arm64` 编译宿主后端二进制
- **且** 构建流程产出标准宿主二进制文件

#### Scenario:使用 platforms 构建多平台宿主二进制
- **当** 运维人员运行 `make build platforms=linux/amd64,linux/arm64`
- **则** 构建流程分别使用 `GOOS=linux GOARCH=amd64` 与 `GOOS=linux GOARCH=arm64` 编译宿主后端二进制
- **且** 每个平台的二进制输出路径互不覆盖

#### Scenario:使用 auto 构建当前执行环境平台宿主二进制
- **当** 运维人员运行 `make build platforms=auto`
- **则** 构建流程将 `auto` 解析为当前执行环境的 `runtime.GOOS/runtime.GOARCH`
- **且** 构建流程使用解析后的 `GOOS` 与 `GOARCH` 编译宿主后端二进制

### Requirement:镜像构建命令必须支持多架构 Docker 镜像构建与推送

系统 SHALL 允许运维人员通过`make image platforms=linux/amd64,linux/arm64 registry=<registry> push=1`构建并推送多架构`Docker`镜像。多架构镜像构建 SHALL 使用每个目标平台对应的宿主二进制，而不是将单个平台二进制复用于所有镜像平台。多平台镜像推送 SHALL 通过`Docker buildx`生成并推送远端多架构 manifest。未启用推送时，多平台镜像构建 SHALL 快速失败并提示必须设置`push=1`，避免只写入本地构建缓存而没有可用镜像产物。镜像构建实现 SHALL 由`linactl`内部镜像构建组件执行，而不是依赖独立`hack/tools/image-builder`工具模块。

#### Scenario:构建并推送 amd64 与 arm64 多架构镜像

- **当** 运维人员运行`make image platforms=linux/amd64,linux/arm64 registry=ghcr.io/linaproai tag=v1.0.0 push=1`
- **则** 构建流程先分别准备`linux/amd64`与`linux/arm64`宿主二进制
- **且**`Dockerfile`在每个平台构建上下文中选择对应平台的宿主二进制
- **且** 镜像构建流程通过`Docker buildx`推送`ghcr.io/linaproai/linapro:v1.0.0`的多架构 manifest

#### Scenario:未启用 push 的多平台镜像构建快速失败

- **当** 运维人员运行`make image platforms=linux/amd64,linux/arm64 push=0`
- **则** 镜像构建流程在调用`Docker buildx`前失败
- **且** 错误消息说明多平台镜像构建需要`push=1`

#### Scenario:镜像命令不再调用独立 image-builder 工具

- **当** 运维人员运行`linactl image`或`linactl image.build`
- **则** 命令通过`linactl`内部镜像构建组件完成镜像构建或 staging
- **且** 命令不得再执行`go run ./hack/tools/image-builder`

### Requirement:发布构建命令必须支持自定义构建配置

系统 SHALL 允许运维人员通过 `make build config=<path>` 与 `make image config=<path>` 指定仓库相对路径的 image-builder 配置文件，用于替代默认 `hack/config.yaml`。该配置文件 SHALL 使用与 `hack/config.yaml` 相同的 `build` 与 `image` 结构，允许通过 `image.dockerfile` 指定自定义 Dockerfile。未指定 `config` 时，系统 SHALL 继续读取 `hack/config.yaml`。

#### Scenario:使用自定义构建配置构建宿主二进制
- **当** 运维人员运行 `make build config=.github/workflows/nightly-build/config.yaml`
- **则** 构建流程使用该配置文件中的 `build.platforms`、`build.outputDir` 与 `build.binaryName`

#### Scenario:使用自定义运行时配置构建镜像
- **当** 运维人员运行 `make image config=.github/workflows/nightly-build/config.yaml platforms=linux/amd64,linux/arm64 registry=ghcr.io/linaproai tag=nightly-20260507 push=1`
- **则** 镜像构建流程使用该配置文件中的 `image.dockerfile`
- **且** 自定义 Dockerfile 可决定容器内 `/app/config.yaml` 的来源

### Requirement:Nightly workflow 必须发布 GHCR 多架构镜像

系统 SHALL 提供 GitHub Actions nightly build workflow，每天凌晨自动运行一次，使用仓库标准 `make image` 入口构建 `linux/amd64` 与 `linux/arm64` 宿主二进制并发布多架构 `Docker` 镜像到 `ghcr.io`。workflow SHALL 使用 `Docker buildx` 发布远端多架构 manifest，镜像标签 SHALL 至少包含按日期生成的不可变 nightly 标签和浮动 `nightly` 标签。workflow SHALL 允许手动触发，以便发布链路可在需要时重新执行。

#### Scenario:nightly 定时发布多架构镜像
- **当** GitHub Actions nightly build workflow 按计划触发
- **则** workflow 安装前端、Go 与 Docker buildx 构建环境
- **且** workflow 运行 `make image config=.github/workflows/nightly-build/config.yaml platforms=linux/amd64,linux/arm64 registry=ghcr.io/<owner> image=linapro tag=nightly-<yyyymmdd> push=1`
- **且** workflow 将同一多架构 manifest 同步标记为 `ghcr.io/<owner>/linapro:nightly`

#### Scenario:手动触发 nightly 发布链路
- **当** 运维人员通过 GitHub Actions 手动触发 nightly build workflow
- **则** workflow 使用与定时任务相同的多架构构建与发布流程
- **且** 发布完成后可通过 `docker buildx imagetools inspect` 验证远端 manifest 包含 `linux/amd64` 与 `linux/arm64`

### Requirement: Release workflow 必须复用共享测试模板并运行简要测试门禁

系统 SHALL 提供 `Release Test and Build` GitHub Actions workflow，用于替代只发布镜像的 release workflow。该 workflow 在 tag push 触发后 SHALL 像 `Nightly Test and Build` 一样复用共享测试验证套件，并采用与 `Main CI` 一致的简要测试范围：host-only 与 plugin-full 的 Windows 命令冒烟、Go 单元测试、前端单元测试、插件命令冒烟、常用 make 命令冒烟、Redis integration、host-only build smoke 和 Redis cluster smoke。Release workflow 不 SHALL 运行 host-only E2E 或 plugin-full E2E；完整 E2E 验证由 nightly workflow 覆盖。

#### Scenario: Release tag 触发简要测试后发布镜像

- **WHEN** GitHub Actions 收到 tag push 事件
- **THEN** release workflow 先完成 release tag 与 framework version 校验
- **AND** release workflow 调用 `.github/workflows/reusable-test-verification-suite.yml`
- **AND** 共享测试套件使用与 `Main CI` 一致的不含 E2E 的简要测试开关
- **AND** release 镜像发布 job 通过 `needs` 依赖 tag 校验和共享测试套件
- **AND** 所有测试 job 成功后才执行 GHCR 登录、`make image push=1`、`latest` 标签发布和远端 manifest inspect

#### Scenario: Release 复用共享测试模板

- **WHEN** nightly workflow 和 main CI workflow 通过共享测试套件编排验证 job
- **THEN** release workflow SHALL 通过同一个 `reusable-test-verification-suite.yml` 编排验证 job
- **AND** release workflow 不得重复内联展开共享测试套件已经封装的验证 job

#### Scenario: Release 不运行完整 E2E

- **WHEN** release workflow 调用共享测试套件
- **THEN** `include-host-only-e2e-tests` SHALL 为 `false`
- **AND** `include-plugin-full-e2e-tests` SHALL 为 `false`
- **AND** release workflow 的镜像发布依赖不 SHALL 等待单独的 E2E job

#### Scenario: 任一简要测试失败阻止 release 镜像推送

- **WHEN** release workflow 中任一测试 job 失败、取消或超时
- **THEN** release 镜像发布 job 不得执行
- **AND** workflow 不得推送 release tag 对应镜像
- **AND** workflow 不得更新 `latest` 浮动镜像标签

#### Scenario: Release workflow 名称表达测试和构建职责

- **WHEN** 仓库维护 release 发布 workflow
- **THEN** workflow 文件名使用 `.github/workflows/release-test-and-build.yml`
- **AND** workflow 展示名称为 `Release Test and Build`
- **AND** 不再保留职责重叠的旧 `release-build.yml`

### Requirement: Release 插件镜像必须保留官方插件简要验证

系统 SHALL 将 release 发布链路视为 plugin-full 镜像发布路径。发布完整插件镜像前，workflow SHALL 通过共享测试套件运行 plugin-full Windows 命令冒烟、Go 单元测试和前端单元测试，并在镜像构建阶段使用官方插件工作区构建插件完整镜像。

#### Scenario: 官方插件工作区存在时继续发布简要验证
- **WHEN** release workflow checkout 完成
- **AND** `apps/lina-plugins` 包含官方插件目录与 `plugin.yaml`
- **THEN** workflow 继续执行 plugin-full Windows 命令冒烟、Go 单元测试、前端单元测试和镜像构建前置验证
- **AND** 官方插件源码属于 release 简要验证范围

#### Scenario: 官方插件工作区缺失时快速失败
- **WHEN** release workflow 需要发布完整插件镜像
- **AND** `apps/lina-plugins` 不存在、为空或缺少官方插件清单
- **THEN** plugin-full 简要验证或镜像构建在发布前失败
- **AND** 错误消息说明当前缺少官方插件工作区
- **AND** 错误消息包含初始化 submodule 的操作提示

#### Scenario: Release 镜像构建包含官方插件
- **WHEN** release workflow 发布多架构镜像
- **THEN** `release-image` job SHALL 继续使用 plugin-full 构建配置
- **AND** 发布出的 release tag 与 `latest` 浮动标签 SHALL 对应包含官方插件的镜像

### Requirement: Release workflow 必须在发布门禁完成后创建 GitHub Release

系统 SHALL 在 release tag 校验、共享测试验证套件和 GHCR 镜像发布全部成功后，为触发 workflow 的 Git tag 创建 GitHub Release。Release 标题 SHALL 使用 `LinaPro Release <tag>` 格式，其中 `<tag>` 是当前触发 workflow 的标签名称。

#### Scenario: 发布成功后创建 GitHub Release
- **WHEN** GitHub Actions 收到 tag push 事件
- **AND** release tag 校验成功
- **AND** 共享测试验证套件成功
- **AND** release 镜像发布成功
- **THEN** release workflow 创建当前标签对应的 GitHub Release
- **AND** Release 标题为 `LinaPro Release <tag>`

#### Scenario: 发布门禁失败时不创建 GitHub Release
- **WHEN** release tag 校验、共享测试验证套件或 release 镜像发布任一阶段失败、取消或超时
- **THEN** release workflow 不得创建 GitHub Release

### Requirement: Release tag 必须与框架元数据版本一致

系统 SHALL 将 `apps/lina-core/manifest/config/metadata.yaml` 中的 `framework.version` 作为 release tag 的唯一版本基线。任何 release tag 发布链路在执行测试、构建、镜像推送或更新浮动标签前，MUST 校验 Git tag 名称与 `framework.version` 完全一致。

#### Scenario: Tag 名称与框架版本一致时继续发布
- **WHEN** GitHub Actions 收到 release tag push 事件
- **AND** tag 名称等于 `apps/lina-core/manifest/config/metadata.yaml` 中的 `framework.version`
- **THEN** release workflow 继续执行后续测试和镜像发布门禁

#### Scenario: Tag 名称与框架版本不一致时阻止发布
- **WHEN** GitHub Actions 收到 release tag push 事件
- **AND** tag 名称不等于 `apps/lina-core/manifest/config/metadata.yaml` 中的 `framework.version`
- **THEN** release workflow 在测试和镜像发布前失败
- **AND** workflow 不得执行 GHCR 登录、`make image push=1`、`latest` 标签发布或远端 manifest inspect

#### Scenario: 框架版本格式不合法时阻止发布
- **WHEN** release workflow 校验 `framework.version`
- **AND** `framework.version` 不是 Docker tag 兼容的 release 版本格式
- **THEN** release workflow 失败
- **AND** 错误消息说明 `framework.version`、Git tag 名称和允许的版本格式

### Requirement: Release tag 校验必须通过跨平台工具复用

系统 SHALL 通过仓库跨平台工具入口执行 release tag 与框架元数据版本一致性校验。GitHub Actions、本地发布检查和后续发布自动化 SHALL 复用同一个校验命令，避免在不同 workflow 或脚本中维护重复的 YAML 解析和版本格式规则。

#### Scenario: 本地校验命令读取框架版本
- **WHEN** 维护者运行 release tag 校验命令并传入 tag 名称
- **THEN** 命令读取 `apps/lina-core/manifest/config/metadata.yaml`
- **AND** 命令比较传入 tag 与 `framework.version`
- **AND** 命令输出可读的成功或失败信息

#### Scenario: Workflow 复用同一个校验命令
- **WHEN** release workflow 需要校验 tag 名称
- **THEN** workflow 调用仓库跨平台校验命令
- **AND** workflow 不得使用独立的 ad hoc YAML 解析逻辑替代该命令

### Requirement: 受控发布入口必须在创建 tag 前校验框架版本

系统 SHALL 提供受控的 GitHub Actions 手动发布入口，用于读取 `framework.version` 并创建同名 release tag。该入口 MUST 在创建 tag 前运行 release tag 校验；校验失败时不得创建或推送 tag。

#### Scenario: 受控入口成功创建匹配 tag
- **WHEN** 维护者手动触发受控 release tag workflow
- **AND** `framework.version` 合法且对应远端 tag 尚不存在
- **THEN** workflow 创建名称等于 `framework.version` 的 Git tag
- **AND** workflow 将该 tag 推送到 GitHub 仓库

#### Scenario: 受控入口发现 tag 已存在
- **WHEN** 维护者手动触发受控 release tag workflow
- **AND** 远端已存在名称等于 `framework.version` 的 Git tag
- **THEN** workflow 失败
- **AND** workflow 不得移动、覆盖或删除既有 tag

#### Scenario: 仓库规则阻止人工直接创建 release tag
- **WHEN** 仓库配置了匹配 release tag 的 GitHub tag ruleset
- **THEN** 普通用户不得直接创建、更新或删除受保护 release tag
- **AND** 只有受控发布 actor 可以在校验通过后创建 release tag
- **AND** 受控发布 workflow 必须使用仓库变量 `RELEASE_APP_CLIENT_ID` 和仓库密钥 `RELEASE_APP_PRIVATE_KEY` 生成 GitHub App installation token
- **AND** ruleset bypass 必须配置到该 GitHub App actor，而不是配置到 token 字符串本身

### Requirement: Nightly image publishing must support a manual no-test entrypoint

系统 SHALL 提供一个独立的`GitHub Actions`手动 workflow，用于构建并发布`nightly`镜像。该 workflow MUST 仅通过`workflow_dispatch`触发，MUST 直接调用统一镜像发布 workflow，MUST 不依赖测试验证套件、单元测试、`E2E`测试、smoke 测试或其他前置测试 job。现有定时 nightly workflow MUST 继续保留测试门禁。

#### Scenario: 手动触发直接发布 nightly 镜像

- **WHEN** 维护者通过`GitHub Actions`手动触发 no-test nightly 镜像发布 workflow
- **THEN** workflow 直接调用统一镜像发布 workflow 构建并推送`linux/amd64`与`linux/arm64`多架构镜像
- **AND** workflow 发布日期型`nightly-<yyyymmdd>`不可变标签和`nightly`浮动标签
- **AND** workflow 不等待任何测试验证 job 完成

#### Scenario: 定时 nightly 继续受测试门禁保护

- **WHEN** 现有 nightly workflow 通过 schedule 触发
- **THEN** workflow 继续先运行共享测试验证套件
- **AND** 只有测试验证套件成功后才发布`nightly`镜像

### Requirement: Nightly demo image must provide a memory-only Compose launcher

系统 SHALL 在`hack/deploy/docker-compose.yaml`提供一个用于演示`nightly`镜像的`Docker Compose`启动入口。该启动入口 MUST 使用已发布的`linapro`镜像启动演示服务，MUST 使用`PostgreSQL`服务作为演示数据库，MUST 不挂载宿主数据目录或声明持久化卷，MUST 将应用运行期数据目录和`PostgreSQL`数据目录放在内存态`tmpfs`中，MUST 将运行时配置单独维护在`hack/deploy/config.yaml`并通过只读配置方式注入容器，MUST 在`PostgreSQL`健康后完成数据库初始化与`mock`演示数据加载再启动`HTTP`服务，MUST 使用必要注释说明镜像/端口覆盖、内存态数据、只读配置注入、数据库依赖、启动初始化顺序和演示保护插件用途。

#### Scenario: 启动内存态演示环境

- **WHEN** 体验者运行`docker compose -f hack/deploy/docker-compose.yaml up`
- **THEN** Compose 启动`linapro`演示服务并暴露`9120`端口
- **AND** 应用从`hack/deploy/config.yaml`读取只读运行时配置
- **AND** 应用连接 Compose 内的`PostgreSQL`服务作为数据库
- **AND** 应用运行期数据和`PostgreSQL`数据写入容器内`tmpfs`
- **AND** 停止并删除容器后演示数据不会通过宿主磁盘卷保留

### Requirement: Deployment test Compose must provide a manual development container

系统 SHALL 在`hack/deploy/tests/docker-compose.yaml`提供一个用于手动验证开发指令的`Docker Compose`测试入口。该入口 MUST 启动`PostgreSQL`服务，MUST 启动基于`loads/ubuntu:24.04-npm`的长期驻留开发容器，MUST 将当前仓库挂载到开发容器工作目录，MUST 通过`GF_GCFG_PATH`指向`hack/deploy/tests`配置目录并读取其中的`config.yaml`，MUST 等待`PostgreSQL`健康后再保持开发容器可进入状态，MUST NOT 自动执行`lina init`、`lina mock`或启动 LinaPro HTTP 服务。

#### Scenario: 进入开发容器手动执行验证命令

- **WHEN** 开发者运行`docker compose -f hack/deploy/tests/docker-compose.yaml up -d`
- **THEN** Compose 启动一次性`PostgreSQL`服务
- **AND** Compose 启动`loads/ubuntu:24.04-npm`开发容器并挂载当前仓库
- **AND** 开发者可以通过`docker compose -f hack/deploy/tests/docker-compose.yaml exec dev bash`进入容器
- **AND** 开发者在容器内手动执行初始化、构建、测试或其他开发验证命令

