## ADDED Requirements

### Requirement: WASM 自定义节解析能力必须由 pluginbridge 集中提供
宿主系统 SHALL 在 `apps/lina-core/pkg/pluginbridge/pluginbridge_wasm_section.go` 提供 `ReadCustomSection(content []byte, name string) ([]byte, bool, error)` 与 `ListCustomSections(content []byte) (map[string][]byte, error)` 公共能力,集中实现 `wasm` 文件头校验、节遍历与 ULEB128 解码。`apps/lina-core/internal/service/i18n`、`apps/lina-core/internal/service/apidoc` 与插件运行时 MUST 通过该公共能力读取动态插件运行时产物中的自定义节(如 `i18n_assets`、`apidoc_assets`),不得在业务包内维护重复 WASM 解析实现。`pluginbridge.WasmSection*` 节名常量 MUST 由 `pluginbridge` 包统一维护。

#### Scenario: i18n 通过 pluginbridge 读取动态插件 i18n 节
- **WHEN** 系统需要从某个动态插件运行时产物中读取 `i18n_assets` 自定义节
- **THEN** 调用方通过 `pluginbridge.ReadCustomSection(content, pluginbridge.WasmSectionI18NAssets)` 完成
- **AND** `i18n` 包内不存在 `parseWasmCustomSectionsForI18N` / `readWasmULEB128ForI18N` 这类专用解析函数

#### Scenario: 修复 WASM 解析缺陷只需修改 pluginbridge
- **WHEN** WASM 解析需要扩展支持新节、修复解码 bug 或增加边界校验
- **THEN** 修改 `pkg/pluginbridge/pluginbridge_wasm_section.go` 一处即可
- **AND** `i18n` 包与插件运行时无需重复改动
