## Why

插件市场需要在不读取源码和不依赖外部站点的前提下展示插件介绍、配置说明和更新日志。当前插件仅有面向开发者的`README.md`或清单资源，缺少面向终端用户的市场文档资源，也缺少统一的中英文内容。

## What Changes

- 为所有包含`plugin.yaml`的官方插件新增`manifest/docs`文档资源。
- 每个插件至少提供`zh-CN`和`en-US`两套文档，覆盖功能介绍、配置说明和更新日志。
- 文档不包含安装介绍、权限介绍或图片资产，避免插件市场首版展示范围过重。
- 文档内容以插件自己的`plugin.yaml`、现有`README.md`和`manifest`资源为事实来源，避免引入运行时实现变更。

## Impact

- 影响范围限定在`apps/lina-plugins/<plugin-id>/manifest/docs/**`和本变更的`OpenSpec`文档。
- `linapro-storage-core`目录缺少`plugin.yaml`，不是可安装插件，本次不为其生成插件市场文档。
- 不修改`apps/lina-core`、`apps/lina-vben`、HTTP API、数据库、前端运行时页面或插件生命周期逻辑。
- `i18n`影响：新增市场文档自身提供`zh-CN`和`en-US`双语内容，不修改运行时`manifest/i18n`语言包。
- 数据权限、缓存一致性、开发工具跨平台和运行期依赖无影响；验证通过静态文件存在性检查和`openspec validate`完成。
