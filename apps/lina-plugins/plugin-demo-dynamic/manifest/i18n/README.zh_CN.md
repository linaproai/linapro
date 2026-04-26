# 运行时国际化样例

该目录用于存放 `plugin-demo-dynamic` 示例插件的交付语言包。

宿主会把 `manifest/i18n/<locale>.json` 快照进动态插件产物中，以便运行时国际化接口在安装、启用、升级、停用、卸载之后正确聚合插件自有翻译消息。

插件接口文档翻译资源存放在 `manifest/i18n/apidoc/<locale>.json`。它们会独立嵌入动态插件产物，并且只在渲染 `/api.json` 时由宿主合并。

当前示例覆盖的键包括：

- 插件元数据，例如 `plugin.plugin-demo-dynamic.name`
- 菜单元数据，例如 `menu.plugin:plugin-demo-dynamic:main-entry.title`
- 内嵌页面文案，例如 `plugin.plugin-demo-dynamic.page.*`
- `apidoc/` 下的 `plugins.plugin_demo_dynamic.*` 接口文档元数据

请统一使用扁平 key，并采用 `zh-CN.json`、`en-US.json` 这类规范化语言文件名。
