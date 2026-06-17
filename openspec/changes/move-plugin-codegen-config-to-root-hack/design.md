## Context

当前官方插件的 GoFrame 代码生成配置统一放在`apps/lina-plugins/<plugin-id>/backend/hack/config.yaml`。该配置只服务开发期`make dao`和`linactl dao`，用于生成插件本地`backend/internal/dao`、`backend/internal/model/do`和`backend/internal/model/entity`，不属于插件运行时配置。

现有`linactl`实现把插件目标解析到`backend/`，并要求目标目录存在`hack/config.yaml`。如果简单把工作目录改成插件根，GoFrame 会把`internal/`、`api/`和`go.mod`按插件根解析，导致生成物落错目录。因此本变更必须解耦 GoFrame 工作目录和配置目录。

## Goals / Non-Goals

**Goals:**

- 插件开发期代码生成配置统一位于`apps/lina-plugins/<plugin-id>/hack/config.yaml`。
- GoFrame 生成命令继续以插件`backend/`为工作目录，保证生成物仍落到`backend/internal/**`。
- `linactl ctrl`和`linactl dao`只接受`dir=`目标参数，旧目标选择器不再兼容。
- `plugins.check`扫描新路径并阻断旧路径，避免配置事实源分叉。
- 更新 OpenSpec、规则和 README，使文档与工具行为一致。

**Non-Goals:**

- 不调整`apps/lina-core/hack/config.yaml`。
- 不调整插件运行时配置`manifest/config/config.yaml`。
- 不修改 GoFrame 上游源码。
- 不新增或修改 SQL 资产。
- 不改变插件后端业务分层、运行时接口、前端 UI、数据权限或缓存逻辑。

## Decisions

### 解耦`workDir`和`configDir`

`linactl`内部使用`Target{WorkDir, ConfigDir}`表示代码生成目标。宿主默认目标为`WorkDir=apps/lina-core`、`ConfigDir=apps/lina-core/hack`。标准插件后端目标为`WorkDir=apps/lina-plugins/<plugin-id>/backend`、`ConfigDir=apps/lina-plugins/<plugin-id>/hack`。

备选方案是让 GoFrame 在插件根执行，并把配置内路径改成`backend/internal`。该方案会牵动`api/`、`go.mod`、controller 生成路径和导入路径，风险高且不必要。

### 标准插件识别只作用于`dir=<plugin>/backend`

当`dir=`目标目录名为`backend`，父目录存在`plugin.yaml`，父目录位于仓库`apps/lina-plugins/`下，且`plugin.yaml`中的`id`与目录名一致时，目标被识别为标准插件后端并读取插件根`hack/`。其他目录继续使用`dir/hack`，保持非插件 GoFrame 项目语义。

### 只保留`dir=`目标参数

`p=`、`plugin=`和`target=`与`dir=`重复，会让插件 ID、目录和目标目录三套语义并存。项目无历史包袱，本次直接删除旧参数；收到未知参数时必须失败并说明只支持`dir=`。

### 不保留旧配置路径兼容

迁移后旧路径`backend/hack/config.yaml`不再作为有效配置。`plugins.check`发现旧路径必须失败，并提示迁移到插件根`hack/config.yaml`。双路径兼容会造成生成配置事实源不唯一，不符合本次治理目标。

## Risks / Trade-offs

- 生成物落错目录 → 保持`WorkDir=<plugin>/backend`，只迁移`ConfigDir`，并用`linactl`测试覆盖隐藏子命令传参。
- 治理扫描漏检 → `plugins.check`统一读取插件根`hack/config.yaml`，并新增旧路径阻断。
- `ctrl`缺配置行为回归 → `ctrl`继续允许配置缺失时使用临时空配置目录；`dao`继续强制要求配置文件存在。
- 非插件目标回归 → 标准插件识别失败时继续读取`dir/hack/config.yaml`。
- 文档与实现不一致 → 同步更新 OpenSpec 基线、规则文件和中英文 README，并用静态检索验证旧活动路径消除。

## Migration Plan

1. 创建本 OpenSpec 变更并更新三份相关能力规范。
2. 改造`linactl` GoFrame 目标解析、隐藏子命令参数和命令参数校验。
3. 改造`plugins.check`配置扫描和旧路径阻断。
4. 迁移官方插件`backend/hack/config.yaml`到插件根`hack/config.yaml`。
5. 更新规则、README 和测试夹具。
6. 运行`openspec validate`、`hack/tools/linactl`测试、`make plugins.check`、至少一个插件`make dao`和`make ctrl`烟测，并执行静态检索确认旧路径无活动配置残留。
