# Tasks

## Summary

- [x] 澄清 `Storage()` / `Files()` 领域边界；demo 源码插件附件改走 `storagecap.Service`；更新 `pkg/plugin` 与 demo 双语 README
- [x] 动态插件 storage 分片上传：`put.init`/`put.chunk`/`put.commit`/`put.abort`、会话绑定、guest 自动选择单次/分片；协议与 WASM/guest 单测
- [x] Provider 扩展落地：唯一可服务自动选中、冲突码、配置无效 fail-closed；移除主配置 active provider 选择语义
- [x] 官方云存储 source 插件：cos / oss / obs / qiniu / aws / azure / s3（settings、Provider 全方法、连通性探测、i18n、README）
- [x] 配置菜单挂 `setting`；不维护 `storage` 一级目录；品牌图标在插件侧；唯一启用与 fail-closed 文案
- [x] 运行时选择与冒烟：0→local、1→云、≥2→conflict；可选 MinIO/S3 集成路径

## Feedback

- [x] **FB-1**: COS/OSS 等菜单品牌图标（去掉误用 QQ 图标等）
- [x] **FB-2**: 存储相关目录排序（后收敛为挂系统设置，无独立一级目录）
- [x] **FB-3**: 新增 `linapro-storage-aws`；`linapro-storage-s3` 改为通用 S3 协议插件
- [x] **FB-4**: COS 腾讯云品牌图标；S3 命名去掉「兼容」字样
- [x] **FB-5 / P0**: 新增 `linapro-storage-obs`、`linapro-storage-azure`
- [x] **FB-6~7**: 连接测试失败展示（短 Toast + 详情；标题字号）；后收敛为 Modal.error
- [x] **FB-8**: COS 启用后上传「重置文件读取位置失败」
- [x] **FB-9**: 品牌 SVG 不得进宿主 `packages/icons`；插件 `frontend/icons` + 构建期注册
- [x] **FB-10**: 多云冲突时哈希复用跳过 Put 仍成功 → fail-closed
- [x] **FB-11**: 去掉宿主「存储管理」一级目录；云配置挂「系统设置」
- [x] **FB-12~13**: S3 菜单命名迭代为「存储管理-S3」
- [x] **FB-14**: 测试连接失败统一 Modal.error；更新 S3 TC003
- [x] **P0 扩展**: 七牛 `linapro-storage-qiniu`（sort=35）

## Verification

- [x] `openspec validate` 相关变更 strict 通过
- [x] 边界/分片/Provider/settings 相关 Go 单测通过
- [x] 云插件菜单挂载与 settings 掩码/权限 E2E（或记录环境阻断）
- [x] `lina-review` 审查通过

## Governance

- [x] **i18n**：云插件菜单/设置页/错误文案双语；宿主不维护 `storage` 菜单标题；边界澄清无新增运行时文案
- [x] **数据权限**：settings 平台配置控制面；Storage 插件/租户 key 作用域；动态分片 path 授权不变
- [x] **缓存**：无新增跨节点业务缓存权威数据；`sys_config` 既有读路径
- [x] **测试策略**：契约单测 + 至少一家 settings E2E + 菜单挂载 E2E；Provider mock 覆盖冲突/missing/batch/cursor
- [x] **DI**：settings/sys_config 与 Provider factory owner 在插件侧；宿主复用既有 `storagecap` 解析与 local 后端
- [x] **跨平台开发工具**：无影响
