# Tasks

## Summary

- [x] 补充 `plugin-storage-service` 和 `plugin-host-domain-capabilities` 增量规范，明确 `Storage()` 与 `Files()` 领域职责（clarify-plugin-storage-files-boundary）
- [x] 迁移 `linapro-demo-source` 示例，附件保存/下载/替换/删除/卸载清理改用 `storagecap.Service`
- [x] 移除源码插件示例中直接读取 `upload.path`、拼接宿主本地路径的实现
- [x] 更新 `apps/lina-core/pkg/plugin` 和 demo 中英文 README
- [x] 扩展 storage host service 方法目录和协议编解码，覆盖 `put.init`/`put.chunk`/`put.commit`/`put.abort`（stream-dynamic-plugin-storage-upload）
- [x] 在 WASM storage host service 中实现上传会话、chunk 顺序校验、commit 流式写入和 abort 清理
- [x] 更新动态插件 guest storage adapter，`Storage().Put` 按输入大小自动选择单次或分片上传
- [x] 补充协议、WASM host service 和 guest SDK 单元测试

## Verification

- [x] `openspec validate clarify-plugin-storage-files-boundary --strict` 通过
- [x] `openspec validate stream-dynamic-plugin-storage-upload --strict` 通过
- [x] 相关 Go 测试通过
- [x] `lina-review` 审查通过

## Governance

- [x] i18n：无运行时用户可见文案或语言包新增
- [x] 数据权限：分片上传授权继续使用 `storage.resources.paths` 匹配
- [x] 不修改 HTTP API、数据库结构或宿主文件中心 `Files()` 领域
