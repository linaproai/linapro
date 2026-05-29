## 1. 规则与历史语义核对

- [ ] 1.1 实现前重新读取命中的规则文件：`openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`testing.md`、`database.md`、`i18n.md`、`cache-consistency.md`，并在任务记录中写明影响判断。
- [ ] 1.2 静态检索历史 `Metadata`、`metadata.get`、`service: metadata`、`metadata.yaml`、`声明型资源`、`HostServices.Manifest()` 相关引用，按设计中的语义分流表确认需要删除、重命名、保留或仅更新注释的项目。
- [ ] 1.3 确认宿主框架自身 `apps/lina-core/manifest/config/metadata.yaml`、动态路由元数据、审计元数据、错误元数据、数据库表元数据和 cron 展示 metadata 不属于插件 manifest 资源读取服务，避免误删历史功能。

## 2. 插件 Manifest 契约收敛

- [ ] 2.1 更新 `pkg/plugin/capability/contract`、`pkg/plugin/capability`、`pkg/plugin/capability/guest` 中 `Manifest()` 的注释和接口语义，明确其为插件自有 `manifest/` 原始资源只读视图。
- [ ] 2.2 删除或迁移任何插件可见的旧 `Metadata()`、`MetadataService`、`metadata.get`、`service: metadata` 或等价读取入口，不保留 deprecated alias。
- [ ] 2.3 更新 `Scan` 相关注释和测试命名，明确它是 YAML 便捷扫描能力，不代表 `Manifest()` 只能读取 YAML。

## 3. 源码插件读取实现

- [ ] 3.1 修改源码插件 manifest resource path 规范化逻辑，移除 `config/`、`sql/`、`i18n/` 排除规则，保留空根、绝对路径、URL、Windows drive path、路径穿越、重复 `manifest/` 前缀和跨插件路径拒绝。
- [ ] 3.2 保持源码插件读取来源顺序和作用域：优先读取当前插件嵌入文件系统，再读取当前仓库开发目录中的同一插件 `manifest/`，不得读取宿主或其他插件资源。
- [ ] 3.3 补充源码插件单元测试，覆盖读取 `metadata.yaml`、`config/config.example.yaml`、`sql/001-schema.sql`、`i18n/zh-CN/plugin.json`、缺失资源、路径穿越和跨插件拒绝。

## 4. 动态插件授权与 host service

- [ ] 4.1 修改 `pluginbridge` host service path 校验和 WASM host call 授权逻辑，允许 `manifest` 服务声明和访问 `config/`、`sql/`、`i18n/` 相对路径。
- [ ] 4.2 保持动态插件 `service: manifest` 的 `methods: [get]` 和 `resources.paths` 授权快照校验，未授权路径必须拒绝。
- [ ] 4.3 补充 `pluginbridge` 和 `wasm` 单元测试，覆盖动态插件授权读取 `config/config.example.yaml`、`sql/001-schema.sql`、`i18n/zh-CN/plugin.json`，以及未授权路径拒绝。

## 5. 动态 artifact 资源视图和打包语义

- [ ] 5.1 修改 dynamic active release artifact manifest resource projection，完整投影 artifact 中 `manifest/` 下实际携带的资源，不再只投影 `*.yaml` 或排除 `config/`、`sql/`、`i18n/`。
- [ ] 5.2 保持 `HostServices.Config()` 的默认配置读取语义，继续从 active release 中识别 `manifest/config/config.yaml`，但不把 `Manifest()` 读取结果作为运行期有效配置自动生效。
- [ ] 5.3 检查动态插件构建和 artifact 解析逻辑，确保 `go:embed` 和目录扫描资源来源都保留 `manifest/config/`、`manifest/sql/`、`manifest/i18n/` 及其他 manifest 资源原始路径。
- [ ] 5.4 更新资源计数、示例插件 `plugin.yaml` 或测试 fixture，使动态插件通过 `service: manifest` 显式声明需要读取的 manifest 资源路径。

## 6. 历史命名和文档清理

- [ ] 6.1 清理生产代码、测试、fixture 和注释中把 `Manifest()` 描述为 `Metadata` 或“声明型资源读取器”的历史表述。
- [ ] 6.2 更新相关 README 或开发说明时同步维护中英文镜像；若最终无需修改目录级说明文档，在任务记录中明确无文档镜像影响。
- [ ] 6.3 运行静态检索确认插件资源读取路径中不存在旧 `Metadata()`、`MetadataService`、`metadata.get`、`service: metadata` 或 deprecated alias。

## 7. 验证与审查

- [ ] 7.1 运行 `go test ./pkg/plugin/capability/manifest ./pkg/plugin/capability/guest ./pkg/plugin/pluginbridge/internal/hostservice -count=1`。
- [ ] 7.2 运行 `go test ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/wasm -count=1`。
- [ ] 7.3 若修改动态插件构建器或官方示例插件，运行对应构建器/示例插件测试，并记录跨平台影响；若未修改开发工具，记录开发工具无影响。
- [ ] 7.4 运行 `openspec validate replace-plugin-metadata-with-manifest --strict`。
- [ ] 7.5 运行 `git diff --check` 和旧 Metadata 静态检索，确认没有格式问题和旧读取入口残留。
- [ ] 7.6 完成任务后调用 `lina-review`，审查结论必须覆盖 `i18n`、缓存一致性、数据权限、SQL、开发工具跨平台、DI 来源和测试策略影响。
