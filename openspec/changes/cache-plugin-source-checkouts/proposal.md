## Why

`plugins.install`、`plugins.update` 和 `plugins.status` 当前每次都会把配置的插件来源仓库 clone 到 `temp/plugin-source-<source>-*`，命令结束后删除。对于官方插件仓库，这会导致每次状态查询或安装更新都重复下载完整仓库，Windows 用户执行 `make.cmd plugins.status` 时能明显看到反复 clone 和删除临时目录。

## What Changes

- 将插件来源仓库获取从一次性临时 clone 改为 `temp/plugin-sources/<source>` 下的持久缓存。
- 首次访问某个 source 时 clone 到缓存目录；后续访问通过 `git fetch --prune origin` 更新缓存，再 checkout/reset 到 `hack/config.yaml` 中声明的 `ref`。
- 如果缓存目录损坏、不是 Git 仓库或 origin 与当前配置 repo 不一致，工具安全删除该缓存并重新 clone。
- 保持 `apps/lina-plugins` 写入语义不变：安装/更新仍只复制插件目录，不把 `.git` 元数据写入插件工作区；`plugins.status` 仍不得修改插件工作区、`.gitmodules`、父仓库 Git index 或锁文件。

## Capabilities

### Modified Capabilities

- `plugin-workspace-management`: 插件来源仓库应优先复用本地缓存，并通过 fetch 获取增量更新，避免每次命令重复全量 clone。

## Impact

- 影响 `hack/tools/linactl` 的插件来源 checkout 实现和相关测试。
- 影响 `temp/` 下开发工具缓存布局，新增可复用目录 `temp/plugin-sources/<source>`。
- 不新增运行时 REST API、数据库表、后端服务缓存、前端运行时 i18n 或业务数据权限边界。
