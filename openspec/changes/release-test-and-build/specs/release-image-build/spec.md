## ADDED Requirements

### Requirement: Release workflow 必须复用共享测试模板并运行简要测试门禁

系统 SHALL 提供 `Release Test and Build` GitHub Actions workflow，用于替代只发布镜像的 release workflow。该 workflow 在 tag push 触发后 SHALL 像 `Nightly Test and Build` 一样复用共享测试验证套件，并采用与 `Main CI` 一致的简要测试范围：host-only 与 plugin-full 的 Windows 命令冒烟、Go 单元测试、前端单元测试、插件命令冒烟、常用 make 命令冒烟、Redis integration、SQLite smoke、host-only build smoke 和 Redis cluster smoke。Release workflow 不 SHALL 运行 host-only E2E 或 plugin-full E2E；完整 E2E 验证由 nightly workflow 覆盖。

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
