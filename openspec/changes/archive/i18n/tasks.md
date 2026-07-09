# Tasks

## Summary

- [x] 交付 i18n 基础设施：三层模型、文件资源单一事实来源、语言解析、默认双语基线、固定`ltr`、运行时翻译包和语言列表接口。
- [x] 交付运行时性能治理：翻译缓存按语言和扇区分层，显式作用域失效，`ETag`协商，单值翻译热路径避免整包克隆，前端持久缓存和后台校验。
- [x] 交付服务边界：`LocaleResolver`、`Translator`、`BundleProvider`、`Maintainer`小接口，共享`ResourceLoader`，`WASM`section 读取收敛到`pluginbridge`。
- [x] 交付消息治理：`bizerr`结构化错误、消息分类、导入导出本地化、插件桥接错误契约、硬编码中文扫描和前端`messageKey`优先渲染。
- [x] 交付工作台与文档治理：首次语言识别、语言切换刷新、英文布局回归、项目定位统一、README 中英文镜像规则。
- [x] 修复接口文档页面加载体验：iframe Loading 占位按`lang`参数维护双语文案，Stoplight Elements 渲染完成后自动隐藏（FB-1）。
- [x] 修复接口文档中文标题本地化：消除静态接口共享`dc`描述时的歧义删除，改为每个静态路由记录 DTO 稳定 key（FB-2）。
- [x] 修复源码插件接口文档本地化 key 选择：`/x/<plugin-id>/...`路径优先使用 DTO 稳定 key，不回退到动态插件路径派生键（FB-3）。
- [x] 交叉影响已迁移：配置、菜单、字典、调度、系统 API 文档、系统信息、工作台、登录页、数据库初始化和 demo-control 的完整契约由对应 owner 分组或`openspec/specs`承载。
- [x] 验证：缺失翻译检查、单元测试、基准测试、E2E、硬编码文案扫描、OpenSpec 校验、相关 README 与治理规则更新和`lina-review`均已作为归档维护证据保留。

## Feedback

- [x] **FB-1**: 接口文档页内容较多时缺少加载 Loading 状态。根因：iframe 内 Stoplight Elements 完成渲染前无可见占位。修复：新增 Loading 占位、长耗时提示和脚本加载失败提示。验证：宿主 E2E `TC001-api-docs-page.ts` 10 项通过。
- [x] **FB-2**: 中文语言环境下部分接口标题仍显示英文。根因：多个静态接口共享相同`dc`描述时歧义删除索引，回退到路径派生键。修复：为每个静态路由记录一次 DTO 稳定 key，不再因`dc`歧义删除。验证：`TestBuildLocalizesOpenAPIForRequestLocale`、宿主 E2E 通过。
- [x] **FB-3**: 源码插件接口文档中文标题未本地化。根因：`/x/<plugin-id>/...`路径被识别为动态插件路径，DTO 稳定 key 未被优先使用。修复：调整 operation key 优先级判断。验证：`TestOperationBaseKeyPrefersStaticMarkerForPluginNamespace`、插件 E2E `TC008` 通过。
