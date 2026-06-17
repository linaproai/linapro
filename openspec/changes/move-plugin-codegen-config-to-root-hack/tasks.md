## Implementation

- [x] **T-1**：更新 OpenSpec delta specs，覆盖`linactl-build-tool-consolidation`、`module-decoupling`和`plugin-capability-boundary-governance`的新路径、目标模型与治理行为。
- [x] **T-2**：改造`hack/tools/linactl/internal/goframecli`，将目标解析升级为`workDir`和`configDir`，并保持非插件目标读取`dir/hack`。
- [x] **T-3**：收敛`linactl ctrl`和`linactl dao`参数，只允许`dir=`，删除`p=`、`plugin=`和`target=`，并补充失败用例。
- [x] **T-4**：改造`plugins.check`扫描插件根`hack/config.yaml`，阻断`backend/hack/config.yaml`，并覆盖已有`backend/internal/dao`但缺少根配置的失败用例。
- [x] **T-5**：迁移所有官方插件的`backend/hack/config.yaml`到插件根`hack/config.yaml`，不为无 DAO 配置的插件创建空配置。
- [x] **T-6**：同步更新`.agents/rules/plugin.md`、`apps/lina-plugins/README.md`、`apps/lina-plugins/README.zh-CN.md`、`hack/tools/linactl/README.md`和`hack/tools/linactl/README.zh-CN.md`。
- [x] **T-7**：审查`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`是否需要同步更新，并记录结论。结论：两份文档描述插件公开契约和 host service 边界，不描述开发期 GoFrame codegen 配置路径，无需修改。
- [x] **T-8**：运行验证：`openspec validate move-plugin-codegen-config-to-root-hack --strict`、`cd hack/tools/linactl && go test ./... -count=1`、`make plugins.check`、代表性插件`make dao`和`make ctrl`烟测、旧路径静态检索。

## Impact Records

- [x] **IR-1**：记录无运行时 HTTP API、DTO、前端 UI、SQL schema、运行时配置、缓存、数据权限和业务服务语义变化。
- [x] **IR-2**：记录开发工具跨平台影响：路径解析、文件检查和测试夹具使用 Go 标准库路径 API，不引入平台专属脚本语义。
- [x] **IR-3**：记录`i18n`影响：无运行时用户可见文案和语言包变化；仅工具错误提示和文档文本变化。
- [x] **IR-4**：记录数据库规则影响：不修改 SQL；仅迁移 DAO 生成配置，并通过生成路径、治理检查和 Go 编译门禁验证。

## Feedback

- [x] **FB-1**：移除`ctrl`和`dao`参数校验中对`p=`、`plugin=`和`target=`的专门分支，统一只允许`dir=`。
- [x] **FB-2**：移除根`Makefile`的`ctrl`和`dao`对`p=`的包装兼容，只允许通过`dir=`转发显式目标。
- [x] **FB-3**：移除`apps/lina-plugins`工作区级`ctrl`和`dao`对`p=`的包装兼容，只允许通过`dir=`转发显式目标。
- [x] **FB-4**：移除`hack/makefiles/database.mk`中对旧参数的显式识别和报错分支，使该文件只表达`dir=`转发。
- [x] **FB-5**：移除`apps/lina-plugins/Makefile`中`ctrl`和`dao`对旧`p=`参数的显式识别和报错分支，确保该入口在删除前不保留兼容逻辑。
- [x] **FB-6**：删除`apps/lina-plugins/Makefile`工作区级重复入口，统一由仓库根目录维护`make wasm`，由插件目录本地`Makefile`维护`make ctrl`和`make dao`。
