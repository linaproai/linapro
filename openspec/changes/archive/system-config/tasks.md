# Tasks

## Summary

- [x] 运行时配置治理：注册并保护 JWT/会话/上传/黑名单与公共前端键；`sys.upload.maxSize` 统一 20 MB；本地快照 + 共享修订号；`GetRaw` 数据驱动有效快照与统一读取优先级；插件 `plugin.<id>` 最高优先。
- [x] 编辑元数据本地化：`GetById` 投影 name/remark、value 库原文；mutation 用 raw 加载；内置 Update 忽略 name/remark；前端内置元数据只读。
- [x] 参数值类型：`value_type`/`options` 入表与 seed；CRUD/导入导出；封闭类型校验；内置锁定类型与选项；管理面按类型渲染（含 Tiptap 富文本与密度策略）。
- [x] 管理面分流：`system_manageable`；List/Export 仅 `=1`；Get/Update/Delete/Import 锁定 `=0`；`SetValue`/`BatchSetValue` 四参 options；插件 settings 批量写入且默认不可系统维护；不按 `plugin.*` 命名空间硬过滤。
- [x] 登录 slogan 参数：`sys.auth.sloganImage` 规格/校验/公共前端投影/种子与 i18n 元数据；允许空串隐藏。
- [x] FB-1（plugin.* 管理面收敛已撤销）：曾实现隐藏/锁定 `plugin.*` 后确认需求有误并还原；根因：以命名空间硬过滤与 `system_manageable` 分流目标冲突；处理：撤销过滤/拒绝/掩码与相关错误码；验证：管理面恢复按 `system_manageable` 与既有可见规则展示。
- [x] 登录后首页 SQL：会话单次读取判定；插件 release 请求/列表级复用。
- [x] FB（编辑元数据）：英文环境内置编辑展示与保存不污染；mutation 与投影解耦。
- [x] FB-1～5（类型化）：下拉弹层宽度；options 简单行格式；选项列表布局与空标签；富文本接 Tiptap；类型驱动宽弹窗/全屏/视口高度。
- [x] 验证：config/sysconfig/hostconfig 等 Go 单测与覆盖率门禁；E2E（英文编辑元数据、TC007 类型化编辑）；会话与 release 复用测试；`make lint`；`openspec validate`；`git diff --check`。
- [x] 治理：i18n 有影响（字段/类型/错误/内置展示键与 option 键）；缓存 revision 语义未改、写路径仍 `MarkRuntimeParamsChanged`；数据权限无实质变更；无新增运行期 DI；跨平台工具无影响；交叉契约由对应 owner/`openspec/specs` 承载。
