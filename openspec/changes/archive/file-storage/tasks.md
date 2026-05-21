# Tasks

## 数据模型与存储后端

- [x] 新增 `sys_file` 数据表及版本 SQL 文件，统一记录文件名、原始名、大小、后缀、散列值、访问 URL、存储路径、使用场景和上传者等元信息。
- [x] 设计并实现后端文件存储抽象层（Storage 接口 + 本地存储实现），预留 OSS 扩展能力。
- [x] 基于 SHA-256 散列值与租户维度实现文件去重，相同内容重复上传时复用已有物理文件。
- [x] 移除 `sys_file_usage` 关联表，将使用场景直接收敛到 `sys_file.scene` 字段以简化架构。
- [x] 调整普通文件本地存储路径生成规则，新上传物理文件使用 `<tenantId>/<yyyy>/<MM>/<generated-file-name>`，去掉额外的 `t` 目录层。
- [x] 保留历史 `sys_file.path` 记录，不迁移旧文件，并确保旧 `t/<tenantId>/...` 路径继续按数据库记录访问。

## 后端 API 与业务集成

- [x] 实现文件管理后端 API（`POST /file/upload`、`GET /file`、`GET /file/download/{id}`、`DELETE /file/{ids}`、`GET /file/suffixes`、`GET /file/scenes`），并完成相关 DAO/Controller 骨架生成与业务实现。
- [x] 为上传接口增加可选 `scene` 参数，提供系统预定义场景列表，并修复场景接口路径冲突，统一使用 `/file/scenes`。
- [x] 修复文件列表返回值，输出完整预览 URL、正确的上传者账号名称，并支持按文件大小和上传时间排序。
- [x] 同步更新上传路径注释、接口示例和相关说明文本，避免继续展示旧路径格式。
- [x] 改造 TiptapEditor 富文本编辑器图片上传，从 Base64 内嵌改为调用通用文件上传接口（`scene=notice_image`），并新增通知公告附件上传能力（`scene=notice_attachment`）。
- [x] 改造用户头像上传，使用通用文件上传接口（`scene=avatar`）替代原有独立实现，移除旧的头像上传端点和静态文件服务路由。

## 前端组件与页面交互

- [x] 创建前端通用文件上传 API，以及 FileUpload / ImageUpload 组件，统一处理上传、回显、预览和错误反馈。
- [x] 新增文件管理页面（系统管理 > 文件管理），提供文件列表、搜索、上传、下载、批量删除、图片预览和详情展示能力。
- [x] 新增文件详情弹窗，展示文件完整信息、使用场景和文件路径，并统一弹窗宽度与 `Descriptions` 列宽以避免折叠换行。
- [x] 优化文件列表展示：移除“文件名”列，将“文件后缀”调整为“文件类型”，详情弹窗同步使用一致标签名。
- [x] 将文件类型搜索框改为 Select 下拉选择，选项从 `/file/suffixes` 接口动态获取，并修复值提取与标签展示格式，确保不包含点号。
- [x] 优化文件下载与预览交互：下载统一使用 `requestClient.download`，不可预览文件展示可点击 URL，文件管理列表默认开启预览模式。

## 测试、验证与审查

- [x] 编写并维护文件管理模块及相关上传改造的 E2E 测试用例，覆盖上传、列表、预览、下载、删除等核心流程。
- [x] 补充文件服务单元测试，覆盖新上传路径格式与旧 `t/...` 路径兼容读取边界。
- [x] 运行 `openspec validate simplify-upload-storage-path --strict`，确认新增规范与文档记录一致。
- [x] 运行 `cd apps/lina-core && go test ./internal/service/file -count=1` 与 `cd apps/lina-core && go test ./internal/cmd -count=1`，完成变更包与启动绑定编译烟测。
- [x] 完成 i18n、缓存一致性、数据权限与开发工具跨平台影响评估，并执行 `lina-review` 审查。
