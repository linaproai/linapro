## 1. 插件来源缓存

- [x] 1.1 将 `linactl` 插件来源 checkout 改为 `temp/plugin-sources/<source>` 持久缓存。
- [x] 1.2 缓存存在时使用 `git fetch --prune origin` 更新，并 reset/clean 到配置的 `ref`。
- [x] 1.3 缓存损坏或 origin repo 与配置不一致时安全重建缓存。
- [x] 1.4 保持插件安装、更新、状态查询的工作区写入边界和锁文件语义不变。

## 2. 测试与验证

- [x] 2.1 补充 `linactl` 测试，覆盖首次 clone 后再次执行通过 fetch 复用缓存，不再创建 `plugin-source-*` 临时 checkout。
- [x] 2.2 运行 `cd hack/tools/linactl && go test ./... -count=1`。
- [x] 2.3 运行 `openspec validate cache-plugin-source-checkouts --strict`。
- [x] 2.4 运行 `git diff --check -- hack/tools/linactl openspec/changes/cache-plugin-source-checkouts`。
- [x] 2.5 记录 i18n、缓存一致性、数据权限、REST API 与开发工具影响结论。

## Feedback

- [x] **FB-1**: `plugins.status/install/update` 每次都全量 clone 插件 source 到临时目录，执行结束删除，导致下次重复下载。

## Verification Notes

- FB-1 修复：`linactl` 插件来源同步改为复用 `temp/plugin-sources/<source>`；首次执行 clone，后续执行校验缓存 origin 后使用 `git fetch --prune origin` 更新，并 checkout/reset/clean 到配置 `ref` 对应 commit。缓存不是 Git 仓库或 origin 与配置 repo 不一致时，仅删除并重建该 source 的缓存目录。
- FB-1 测试：新增 `TestPluginsSourceCacheReusesCheckoutWithFetch`，覆盖首次安装创建缓存、远端新增 commit 后 update 复用缓存 fetch 到新版本，并断言不再创建旧 `temp/plugin-source-*` 一次性目录。
- 验证通过：`cd hack/tools/linactl && go test ./... -run 'TestPluginsInstallUpdateAndStatusUseConfiguredSources|TestPluginsSourceCacheReusesCheckoutWithFetch' -count=1`。
- 验证通过：`cd hack/tools/linactl && go test ./... -count=1`。
- 验证通过：`openspec validate cache-plugin-source-checkouts --strict`。
- 验证通过：`git diff --check -- hack/tools/linactl openspec/changes/cache-plugin-source-checkouts`。
- i18n 影响：本次仅调整开发工具终端输出、linactl README 和 OpenSpec 文档，不新增或修改前端运行时文案、接口文档、插件 manifest i18n 或 apidoc i18n。
- 缓存一致性影响：本次新增的是本地开发工具 Git checkout 缓存，位于仓库 `temp/`，不属于运行时业务缓存，不涉及集群、权限、配置、插件状态、租户隔离或 i18n 运行时缓存失效；缓存权威数据源为配置的 Git repo/ref，每次命令通过 fetch/checkout/reset 与远端 ref 对齐，损坏或 repo 变更时重建。
- 数据权限影响：本次不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- REST API 影响：本次不新增或修改 REST API。
- 开发工具影响：修改 `hack/tools/linactl` Go 工具，未新增 shell/PowerShell 脚本；`make`/`make.cmd` 继续作为薄包装入口。
- Review：已按 `lina-review` 口径完成审查。范围来源包括 `git status --short`、`git diff -- hack/tools/linactl openspec/changes/cache-plugin-source-checkouts`、`openspec status --change cache-plugin-source-checkouts --json`、静态扫描旧 `plugin-source-*`/`Downloading plugin source` 引用、OpenSpec 严格校验和 `linactl` Go 包测试。确认本次只修改开发工具、工具 README 和 OpenSpec 文档；未新增运行时 API、数据库访问、业务缓存、数据权限边界、前端运行时文案或平台专属脚本。严重问题 0；警告 0。
