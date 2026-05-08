# 插件运行时加载规范

## 目的

定义动态插件运行时加载行为、集中式 Wasm 自定义段解析、跨节点派生缓存失效、Wasm 编译缓存键和产物刷新一致性。

## 需求

### 需求：WASM 自定义段解析能力必须由 pluginbridge 集中提供
宿主系统 SHALL 通过 `apps/lina-core/pkg/pluginbridge` 体系提供 `ReadCustomSection(content []byte, name string) ([]byte, bool, error)` 和 `ListCustomSections(content []byte) (map[string][]byte, error)` 公共能力，集中实现 `wasm` 文件头验证、段遍历和 ULEB128 解码。该能力可以由 `pluginbridge` 根包 facade 或 `pluginbridge/artifact` 等职责明确的子组件公开，但协议实现必须只有一个权威位置。`apps/lina-core/internal/service/i18n`、`apps/lina-core/internal/service/apidoc` 和插件运行时必须通过此公共能力从动态插件运行时产物中读取自定义段（如 `i18n_assets`、`apidoc_assets`），不得在业务包中维护重复的 WASM 解析实现。`pluginbridge.WasmSection*` 段名常量或其子组件等价常量必须由 `pluginbridge` 体系集中维护。

#### 场景：i18n 通过 pluginbridge 读取动态插件 i18n 段
- **当** 系统需要从动态插件运行时产物中读取 `i18n_assets` 自定义段时
- **则** 调用方通过 `pluginbridge.ReadCustomSection(content, pluginbridge.WasmSectionI18NAssets)` 或 `pluginbridge/artifact` 的等价入口完成
- **且** `i18n` 包中不存在 `parseWasmCustomSectionsForI18N` / `readWasmULEB128ForI18N` 等专用解析函数

#### 场景：修复 WASM 解析缺陷只需修改 pluginbridge 体系
- **当** WASM 解析需要扩展以支持新段、修复解码错误或添加边界检查时
- **则** 修改 `pkg/pluginbridge` 对应 artifact/wasm section 子组件的权威实现即可
- **且** `i18n` 包和插件运行时不需要重复变更

### 需求：动态插件运行时派生缓存必须跨节点失效

动态插件安装、启用、禁用、卸载、升级或同版本刷新后，系统 SHALL 使用统一缓存协调机制使所有节点上的插件运行时派生缓存失效或刷新。

#### 场景：非主节点观察到插件运行时修订号变更

- **当** 集群模式下主节点完成动态插件运行时状态转换并发布插件运行时缓存修订号时
- **则** 非主节点在下一个请求路径或监听路径上观察到新修订号
- **且** 非主节点刷新插件启用快照
- **且** 非主节点使插件前端包、运行时 i18n 包和 Wasm 编译缓存失效

#### 场景：插件禁用后非主节点不再暴露能力

- **当** 主节点上动态插件被禁用或卸载时
- **则** 非主节点不得在插件运行时缓存域允许的陈旧窗口之外继续从过期本地缓存暴露该插件的菜单、前端资产或动态路由能力

### 需求：Wasm 编译缓存必须绑定到产物校验和或 generation

系统 SHALL 将动态插件 Wasm 编译缓存绑定到当前活跃发布的产物校验和或 generation。不得仅通过可变产物路径决定缓存复用。

#### 场景：同版本动态插件刷新重新编译

- **当** 动态插件以相同版本但产物校验和变更进行刷新时
- **则** 节点观察到插件运行时修订号变更后，不得继续命中旧校验和的 Wasm 编译缓存
- **且** 下一次动态路由或动态任务执行必须从新产物编译或加载

#### 场景：相同产物路径但不同校验和

- **当** 活跃发布产物路径与旧缓存路径相同但校验和不同时
- **则** 系统将其视为不同的编译缓存条目
- **且** 旧条目必须失效或自然清理

### 需求：动态插件产物归档必须支持同版本刷新一致性

系统 SHALL 确保同版本刷新后的活跃发布指向可验证的新产物内容，并且其他节点可使用共享发布状态判断本地缓存是否过期。

#### 场景：同版本刷新写入新产物

- **当** 插件同版本刷新提交新产物内容时
- **则** 系统更新活跃发布的校验和或 generation
- **且** 发布插件运行时缓存修订号
- **且** 其他节点可使用活跃发布的校验和或 generation 判断本地缓存是否需要重建

#### 场景：旧产物清理不影响当前活跃发布

- **当** 系统清理旧动态插件产物时
- **则** 当前活跃发布引用的产物不得被删除
- **且** 仍被本地缓存引用但不再活跃的产物可根据保留策略稍后清理
