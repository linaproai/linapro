## Context

`official-plugins-submodule-decoupling` 已经把官方源码插件工作区可选化，并让宿主默认 host-only、插件完整模式按 `apps/lina-plugins` 中的插件清单自动启用。该方案解决了官方插件仓库独立维护问题，但用户项目继续保留 `apps/lina-plugins` submodule 会带来新的使用成本：

- 用户下载或修改插件代码后，变更属于 submodule 仓库而不是用户主仓库。
- submodule 的 `origin` 容易继续指向 `linaproai/official-plugins.git`，用户难以提交到自己的仓库。
- 用户需要理解 `.gitmodules`、父仓库 gitlink、`.git/config`、`.git/modules` 和 submodule 内部 `.git` 文件的关系。
- 从多个来源安装插件时，单个 submodule 仓库无法自然表达每个插件的来源和更新状态。

因此本变更把用户项目插件工作区定义为普通目录，并把来源配置放到 `hack/config.yaml`。

## Goals / Non-Goals

**Goals:**

- 固定使用 `apps/lina-plugins` 作为源码插件维护目录，不在配置中重复声明路径。
- 通过 `make plugins.init` / `linactl plugins.init` 将现有 submodule 工作区转换为普通目录，保留插件代码。
- 通过 `hack/config.yaml` 的 `plugins.sources` 声明插件来源仓库、插件集合根目录、引用和待安装插件。
- `items` 仅支持字符串数组，保持配置简洁稳定。
- 通过 `plugins.install` 安装配置插件，通过 `plugins.update` 更新配置插件，通过 `plugins.status` 展示工作区和插件状态。
- 所有命令通过 `hack/tools/linactl` Go 工具实现，`Makefile` 和 `make.cmd` 只作为跨平台包装入口。
- 默认保护本地改动，更新或覆盖前发现 dirty 目录必须阻断，除非显式 `force=1`。

**Non-Goals:**

- 不实现运行时插件 marketplace，不新增后端 HTTP API。
- 不改变插件 `plugin.yaml` 清单契约和依赖管理语义。
- 不在 `items` 中支持对象形式、单插件 ref 或单插件覆盖策略。
- 不支持把 `apps/lina-plugins` 配置为其他路径。
- 不自动提交父仓库或插件目录改动。
- 不处理插件依赖安装顺序；依赖治理由 `plugin-dependency-management` 负责。

## Decisions

### 决策 1：用户项目插件工作区使用普通目录

`apps/lina-plugins` 在官方框架仓库可以继续作为 submodule 维护官方插件，但用户项目中的 `plugins.init` 应将其转换为普通目录。转换后插件代码由用户主仓库直接管理，避免 submodule 指针和远端归属造成的提交边界问题。

转换过程必须保留 `apps/lina-plugins` 当前文件内容，删除父仓库中 `apps/lina-plugins` 的 gitlink 跟踪，移除 `.gitmodules` 中对应 section，清理 `.git/config` 的 submodule 配置和 `.git/modules/apps/lina-plugins` 元数据，并移除或转换 `apps/lina-plugins/.git` 中的 submodule 指向信息。

如果 `.gitmodules` 中还存在其他 submodule，命令只能删除 `apps/lina-plugins` section；只有文件已无任何 section 时才删除 `.gitmodules` 文件。

### 决策 2：配置只声明来源，不声明工作区路径

配置形状：

```yaml
plugins:
  sources:
    official:
      repo: "https://github.com/linaproai/official-plugins.git"
      root: "."
      ref: "main"
      items:
        - multi-tenant
        - org-center
    custom:
      repo: "https://github.com/linaproai/linapro.git"
      root: "apps/lina-plugins"
      ref: "main"
      items:
        - content-notice
```

字段语义：

- `sources.<name>`：插件来源名称，仅用于诊断、锁文件和命令筛选。
- `repo`：远端 Git 仓库地址。
- `root`：仓库内插件集合根目录，工具读取 `<root>/<item>/plugin.yaml`。
- `ref`：该来源下所有插件共享的分支、tag 或 commit。
- `items`：待安装或更新插件 ID 字符串数组。

不支持 `items` 对象形式。若同一仓库不同插件需要不同 ref，应拆成多个 source。

### 决策 3：全局插件 ID 唯一且路径安全

配置校验必须在执行任何写入前完成：

- 所有 `items` 在全部 source 中必须全局唯一，防止多个来源写入同一个 `apps/lina-plugins/<plugin-id>`。
- source 名称、插件 ID 和路径必须符合安全规则，禁止空值、绝对路径、`..`、路径分隔符注入和 Windows drive path。
- `root` 必须为仓库内相对路径，`.` 表示仓库根目录。
- 远端必须存在 `<root>/<plugin-id>/plugin.yaml`，否则该插件安装或更新失败。

### 决策 4：安装与更新使用临时 checkout，再复制插件目录

`plugins.install` 和 `plugins.update` 不应把来源仓库作为 submodule 或嵌套 Git 仓库放进 `apps/lina-plugins`。命令应将来源仓库检出到 `temp/` 下的临时目录，解析指定 `ref` 到 commit，复制 `<root>/<plugin-id>` 到 `apps/lina-plugins/<plugin-id>`。

安装命令默认只安装缺失插件；目标目录已存在时应阻断并提示使用 update 或 `force=1`。更新命令默认只更新已配置插件；目标目录存在本地未提交改动时必须阻断，避免覆盖用户修改。

### 决策 5：锁文件记录工具状态，但配置仍以 `hack/config.yaml` 为准

工具应维护 `apps/lina-plugins/.linapro-plugins.lock.yaml`，记录每个插件的 source、repo、root、ref、resolvedCommit、manifest version 和内容摘要。该文件用于 `status` 判断本地版本、远端更新和来源漂移，不作为用户手写配置入口。

当配置变更时，`status` 应能指出：

- lock 中存在但配置中已删除的插件。
- 配置中存在但本地未安装的插件。
- 本地插件内容与 lock 摘要不一致。
- 当前远端 ref 解析出的 commit 与 lock 不一致。

### 决策 6：状态检查是只读诊断

`plugins.status` 不应修改工作区。它可以临时读取远端 metadata 或浅克隆到 `temp/`，但不得写入 `apps/lina-plugins`，不得变更 `.gitmodules` 或父仓库 Git index。

状态输出至少覆盖：

- `apps/lina-plugins` 是否存在、是否仍是 submodule/gitlink、是否是普通目录。
- 每个配置插件的本地存在性、`plugin.yaml` 版本、来源、lock commit、本地 dirty 状态和远端更新状态。
- 未纳入配置但本地存在的插件目录。
- 配置错误和潜在冲突。

### 决策 7：跨平台实现和 Git 命令边界

长期逻辑必须写入 `hack/tools/linactl`，使用 Go 标准库完成 YAML 解析、路径校验、文件复制、目录遍历、内容摘要和锁文件写入。需要 Git 协议能力时可以调用 `git` CLI，但必须通过 `linactl` 统一封装，避免新增 Bash、PowerShell 或平台专属脚本。

所有 destructive 操作必须显式限定在 `apps/lina-plugins`、`.gitmodules` 中的对应 section 和 `.git/modules/apps/lina-plugins` 元数据范围内，不得删除用户未授权路径。

## Risks / Trade-offs

- [风险] `plugins.init` 需要修改 Git index 和 submodule 元数据，错误实现可能丢失插件代码。缓解：先复制/备份目录状态，写测试覆盖 submodule、普通目录、缺失目录和混合 `.gitmodules` 场景；命令只移除 gitlink，不删除插件文件。
- [风险] 多来源配置同名插件导致覆盖。缓解：全局唯一校验，发现冲突时在写入前失败。
- [风险] 远端仓库很大，临时 clone 成本高。缓解：优先浅克隆指定 ref；后续可评估 sparse checkout，但第一阶段以正确性为主。
- [风险] 更新覆盖本地改动。缓解：默认 dirty 阻断，`force=1` 必须显式传入，并在输出中列出将被覆盖的插件。
- [风险] `status` 判断远端更新需要访问网络，离线时可能失败。缓解：离线或远端不可达时仍输出本地状态，并把远端更新状态标为 unknown。
- [风险] 官方框架仓库仍需要 submodule，而用户项目需要普通目录。缓解：`plugins.init` 是显式命令，不改变默认 clone 行为；文档说明官方维护和用户项目的不同工作流。

## Cross-Cutting Assessments

- i18n：本变更只新增开发工具命令、配置和文档，不新增前端运行时页面文案、后端接口文档或插件 manifest i18n。命令输出不进入运行时语言包；若实现新增前端可见提示，必须另行补充 i18n。
- 缓存一致性：不新增运行时缓存、缓存键、缓存失效、订阅或跨实例一致性逻辑。插件安装到源码目录后，运行时仍通过既有插件清单发现、同步和构建流程处理。
- 数据权限：不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- RESTful API：不新增后端 REST API。
- 开发工具脚本：涉及开发工具变更，必须使用 Go 工具链实现并运行 `cd hack/tools/linactl && go test ./... -count=1`、`go run ./hack/tools/linactl test.scripts` 或等价验证。

## Migration Plan

1. 扩展 `hack/config.yaml` 结构和 `linactl` 配置解析，支持 `plugins.sources`。
2. 实现配置校验、插件 ID 去重、路径安全校验和 source 解析。
3. 实现 `plugins.init`，覆盖 submodule 转普通目录并保留文件内容。
4. 实现临时 Git checkout、插件目录复制、安装、更新和锁文件维护。
5. 实现 `plugins.status` 的只读诊断和远端更新检测。
6. 更新 README/README.zh-CN、`hack/tools/linactl` README/README.zh-CN 和命令帮助。
7. 补充单元测试、cross-platform smoke、OpenSpec 校验和文档格式验证。

Rollback 策略：如果插件远端拉取和锁文件逻辑出现阻塞，可先交付 `plugins.init` 和 `plugins.status` 的本地诊断能力，暂缓 install/update 写入逻辑；普通目录工作区仍可手动维护插件代码。
