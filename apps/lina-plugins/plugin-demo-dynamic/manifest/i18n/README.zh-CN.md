# 运行时国际化样例

该目录用于存放 `plugin-demo-dynamic` 示例插件的交付语言包。

宿主会把`manifest/i18n/<locale>/`下的直属`JSON`文件按语言合并并快照进动态插件产物中，以便运行时国际化接口在安装、启用、升级、停用、卸载之后正确聚合插件自有翻译消息。

插件接口文档翻译资源存放在`manifest/i18n/<locale>/apidoc/**/*.json`。它们会独立嵌入动态插件产物，并且只在渲染`/api.json`时由宿主合并。

当前示例覆盖的归一化 key 包括：

- 插件元数据，例如 `plugin.plugin-demo-dynamic.name`
- 菜单元数据，例如 `menu.plugin:plugin-demo-dynamic:main-entry.title`
- 内嵌和独立页面文案，例如 `plugin.plugin-demo-dynamic.page.*`
- `<locale>/apidoc/`下的`plugins.plugin_demo_dynamic.*`接口文档元数据

运行时 UI 消息文件可使用层级 JSON 或扁平 dotted key。宿主会把两种格式统一归一化为扁平 key，用于聚合和诊断，并在返回前端运行时语言包时再转换为嵌套对象。

接口文档消息文件也遵循相同的层级或扁平编写规则，并归一化为稳定的 `plugins.plugin_demo_dynamic.*` key。重复的标准响应元数据由宿主自有的 `core.common.*` fallback key 提供。

请采用`zh-CN`、`en-US`这类规范化语言目录名，并使用`plugin.json`、`menu.json`、`job.json`这类语义化运行时文件名。
