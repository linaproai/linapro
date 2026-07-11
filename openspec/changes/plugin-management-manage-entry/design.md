# 插件管理列表「管理」入口设计

## Context

- 源码插件通过 `frontend/pages/*.vue` 的 `pluginPageMeta.routePath` 与 `plugin.yaml` 菜单声明绑定管理页。
- 宿主前端在构建期将源码插件页面收集到 `page-registry`，运行时由动态页壳 `system/plugin/dynamic-page` 挂载。
- 插件管理列表首屏仅加载摘要列表，不得为每一行额外请求详情。

## Goals / Non-Goals

**Goals**

- 操作列提供稳定的「管理」按钮。
- 有管理页时可跳转到宿主已解析的完整路由。
- 无管理页时按钮置灰不可点击。
- 判断与跳转全部在前端完成，不增加列表接口字段。

**Non-Goals**

- 不为每个插件强制新增管理页。
- 不改变插件菜单同步、权限过滤或动态插件资产托管语义。
- 不在列表接口返回 `managementPath` 等新字段。

## Decisions

1. **管理页判定来源：前端页面注册表**
   - 以 `getPluginPages()` 中归属该 `pluginId` 的可导航页面为准。
   - 排除 `frontend/pages/components/**` 以及文件名含 `modal` / `drawer` 的辅助组件，避免把抽屉/弹窗当成管理页。
   - 理由：构建期已有稳定注册表，无需后端扩展，符合列表首屏性能约束。

2. **多管理页时的目标选择**
   - 以当前会话 `accessMenus` 的深度优先遍历顺序为准，选择该插件第一个匹配的菜单路径。
   - 若 access 菜单中尚无匹配（例如路由已注册但菜单尚未刷新），回退为 `router.getRoutes()` 注册顺序中的第一个匹配路径。
   - **禁止**按 `routePath` 字母序排序选页，否则会出现 `/ai/invocations` 排在 `/ai/providers` 前、误进最后一个菜单的问题。
   - 理由：侧边栏菜单顺序即用户感知的“第一个菜单”，与 `plugin.yaml` 的 `sort` 一致。

3. **跳转路径解析**
   - 路径匹配支持完整相等或后缀匹配，以兼容相对菜单路径挂到父目录后的完整 URL。
   - 若当前会话找不到任何匹配路由，保持在列表页并给出用户可见提示。
   - 理由：按钮启用只表达“插件声明了管理页且已安装”；真正可访问性仍受启用状态与权限约束。

4. **按钮状态**
   - 已安装 **且** 存在可导航管理页 → 可点击。
   - 未安装 → `disabled`，Tooltip 提示“请先安装该插件”。
   - 已安装但不存在管理页 → `disabled`，Tooltip 提示“该插件没有管理页面”。

## Risks / Trade-offs

- [动态插件仅 iframe/资产页、未进入 page-registry] → 可能被判定为无管理页。接受：当前托管工作台主要源码插件管理页走 page-registry；动态嵌入页若后续需要可扩展匹配 `x-assets/<plugin-id>`。
- [多管理页只进一个] → 可能不是用户想要的页。接受：优先 settings/management，后续若需要可改为下拉选择。
- [未启用时点击可能失败] → 用 message 提示而非把按钮一概禁用，避免与“无管理页置灰”语义混淆。

## Migration Plan

无数据迁移。前端发版后立即生效。

## Open Questions

无。
