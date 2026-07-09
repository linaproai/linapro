## Why

当前仓库缺少统一的`Go`静态检查入口，代码风格、错误处理和基础质量问题主要依赖人工审查发现，难以支撑面向可持续交付的`AI`原生全栈框架长期演进。参考`gf`项目引入`golangci-lint`后，应结合`LinaPro`已有`linactl`、多`Go module`和官方插件`submodule`工作区特点，建立可复用、可审计、可在`CI`中阻断的`Go`静态检查门禁。

## What Changes

- 新增仓库级`golangci-lint`配置和版本锁定文件，明确默认启用的`linters`、`formatters`、生成代码排除和`nolint`治理策略；导入分组 formatter 在后续单独规范化导入顺序后再启用。
- 新增`linactl lint.go`跨平台命令，并通过根`Makefile`暴露`make lint.go`与必要的聚合入口，避免把长期逻辑写入`Shell`或`Makefile`。
- 支持`plugins=0`宿主模式和`plugins=1`官方插件完整模式，插件完整模式复用现有临时`go.work`生成机制，使宿主、`linactl`和官方插件`Go module`都能进入静态检查范围。
- 新增可复用`GitHub Actions`工作流并接入现有验证套件，使主`CI`和发布验证能够运行`Go`静态检查。
- 更新开发工具与后端治理规范、`linactl`中英文说明文档，记录本地运行、`CI`门禁、跨平台边界和静态检查失败处理方式。

## Capabilities

### New Capabilities

- `go-static-lint-governance`：定义`Go`静态检查配置、命令入口、插件工作区覆盖、`CI`门禁和治理验证要求。

### Modified Capabilities

- 无。

## Impact

- 影响仓库根目录配置文件、根`Makefile`、`hack/makefiles/`、`hack/tools/linactl/`、`.github/workflows/`和相关工具文档。
- 新增开发期工具依赖`golangci-lint`，版本通过仓库文件锁定；默认入口必须在`Windows`、`Linux`、`macOS`上可执行。
- 不新增或修改运行时`HTTP API`、数据库结构、运行期服务依赖、权限模型、缓存路径或前端用户可见页面。
- `i18n`影响判断：本变更只新增开发工具命令、`CI`门禁和技术文档，不新增运行时用户可见文案、菜单、路由、接口文档源文本或翻译资源，确认无运行时`i18n`资源影响。
