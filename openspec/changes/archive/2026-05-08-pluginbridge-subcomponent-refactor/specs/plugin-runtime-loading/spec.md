## MODIFIED Requirements

### Requirement: WASM 自定义段解析能力必须由 pluginbridge 集中提供

宿主系统 SHALL 通过 `apps/lina-core/pkg/pluginbridge` 体系提供 `ReadCustomSection(content []byte, name string) ([]byte, bool, error)` 和 `ListCustomSections(content []byte) (map[string][]byte, error)` 公共能力，集中实现 `wasm` 文件头验证、段遍历和 ULEB128 解码。该能力可以由 `pluginbridge` 根包 facade 或 `pluginbridge/artifact` 等职责明确的子组件公开，但协议实现必须只有一个权威位置。`apps/lina-core/internal/service/i18n`、`apps/lina-core/internal/service/apidoc` 和插件运行时必须通过此公共能力从动态插件运行时产物中读取自定义段（如 `i18n_assets`、`apidoc_assets`），不得在业务包中维护重复的 WASM 解析实现。`pluginbridge.WasmSection*` 段名常量或其子组件等价常量必须由 `pluginbridge` 体系集中维护。

#### Scenario: i18n 通过 pluginbridge 读取动态插件 i18n 段

- **当** 系统需要从动态插件运行时产物中读取 `i18n_assets` 自定义段时
- **则** 调用方通过 `pluginbridge.ReadCustomSection(content, pluginbridge.WasmSectionI18NAssets)` 或 `pluginbridge/artifact` 的等价入口完成
- **且** `i18n` 包中不存在 `parseWasmCustomSectionsForI18N` / `readWasmULEB128ForI18N` 等专用解析函数

#### Scenario: 修复 WASM 解析缺陷只需修改 pluginbridge 体系

- **当** WASM 解析需要扩展以支持新段、修复解码错误或添加边界检查时
- **则** 修改 `pkg/pluginbridge` 对应 artifact/wasm section 子组件的权威实现即可
- **且** `i18n` 包和插件运行时不需要重复变更
