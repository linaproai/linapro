# Tasks

## Summary

- [x] `sys_file` 模型与本地 Storage：去重、`scene` 收敛、路径简化与历史 `t/...` 兼容
- [x] 文件管理 API/前端：上传列表下载删除、场景与后缀、通用上传组件、公告/头像接入
- [x] 统一对象后端：`ResolveProvider`、`files/` 键、本地 NamespaceFiles 路由、云未命中回退、`engine` 写入与 DI
- [x] 客户端直连：`storagecap` DirectAccess 契约与稳定码；宿主 resolving 探测与 local 降级；文件中心 init/complete/abort 会话与 API；官方七云 Provider 直连；前端 uploadApi 双模式
- [x] 验证：宿主/插件 Go 单测、local 环境 multipart E2E、`openspec validate --strict`、`lina-review`
- [x] 治理：i18n（API apidoc 与插件 CORS 文案双语）；数据权限（init/complete 继承上传权限与租户）；缓存（会话进程内，无新业务缓存语义）；DI（`file.New` 自建 `directSessions`，`storage.NewResolvingService` 复用启动期 runtime）；跨平台开发工具无影响；第一期不扩展 `filecap`、不新增 preferDirect 配置键
