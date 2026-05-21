## Context

插件工作区管理命令需要读取远端 source 的 `<root>/<plugin-id>` 目录来安装、更新或计算远端状态。现有实现用 `os.MkdirTemp` 创建 `temp/plugin-source-<source>-*`，执行 `git clone --no-checkout` 后 checkout 指定 `ref`，命令结束再删除目录。这保证了命令之间完全隔离，但代价是每次执行都要重新下载仓库。

`plugins.status` 是只读诊断命令，用户可能频繁执行。对官方插件仓库而言，重复全量 clone 的网络和时间成本高，也会制造大量短生命周期临时目录。

## Goals / Non-Goals

**Goals:**

- 同一 source 在同一仓库工作区中复用本地 Git 缓存。
- 首次没有缓存时仍能自动 clone。
- 已有缓存时通过 `git fetch --prune origin` 获取增量更新，再把工作树重置到配置的 `ref`。
- 缓存损坏或配置 repo 变化时能自动重建。
- 保持安装、更新、状态查询对 `apps/lina-plugins` 的既有行为和保护语义。

**Non-Goals:**

- 不新增用户可配置的缓存目录。
- 不实现全局跨项目缓存。
- 不改变插件来源配置结构。
- 不改变锁文件字段含义或运行时插件生命周期。

## Decisions

1. 缓存目录固定为 `temp/plugin-sources/<source>`。

   source 名称已有安全校验，可作为目录名使用。缓存位于仓库 `temp/` 下，与现有开发工具生成物一致，不需要用户提交。

2. 缓存存在时优先 `fetch`，而不是重新 clone。

   更新流程为校验缓存目录是 Git 仓库、校验 `origin` URL 与配置 repo 一致、执行 `git fetch --prune origin`、checkout 配置的 `ref`、reset/clean 工作树，最后读取 `HEAD` commit。这样重复执行主要传输增量对象。

3. 缓存不可用时删除并重建。

   如果缓存目录不是 Git 仓库，或 `origin` URL 与当前配置不一致，继续复用可能产生错误来源。工具应限定删除范围为 `temp/plugin-sources/<source>`，然后重新 clone。

4. `plugins.status` 仍允许更新 source 缓存。

   该命令的只读边界针对用户插件工作区、父仓库 Git 配置/index 和锁文件。更新 `temp/` 下的开发工具缓存不改变用户插件目录，可接受并能显著减少重复下载。

5. `plugins.install`、`plugins.update` 与 `plugins.status` 在执行前统一确保插件工作区已初始化。

   命令入口复用与 `plugins.init` 相同的预检和转换逻辑。缺失目录时创建 `apps/lina-plugins`，历史 submodule/gitlink 状态时先移除 submodule metadata 并保留已有插件文件，再继续执行原命令。嵌套 `.git` metadata 或非目录路径仍然失败，因为这些状态可能覆盖用户维护的独立仓库或异常文件。

## Risks / Trade-offs

- [Risk] 缓存工作树残留本地变更导致远端状态计算不稳定。  
  Mitigation：每次 fetch 后执行 forced checkout、hard reset 和 clean，缓存目录只由工具维护。

- [Risk] 用户修改 `hack/config.yaml` 的 repo 后旧缓存指向错误仓库。  
  Mitigation：读取缓存 `origin` URL 并与配置 repo 比较，不一致时删除并重建。

- [Risk] `ref` 是 tag、branch 或 commit，fetch 后解析方式不同。  
  Mitigation：继续使用 `git checkout --quiet <ref>` 语义，与现有实现保持一致；fetch 只负责刷新 origin 对象。
