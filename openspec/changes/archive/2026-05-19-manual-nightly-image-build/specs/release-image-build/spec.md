## ADDED Requirements

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
- **THEN** Compose 启动`linapro`演示服务并暴露`8080`端口
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
