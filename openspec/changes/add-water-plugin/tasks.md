## 1. 变更文档

- [x] 1.1 完成 `add-water-plugin` 提案、设计、规格和任务清单
- [x] 1.2 运行 `openspec validate add-water-plugin --strict`

## 2. 插件骨架

- [x] 2.1 新增 `apps/lina-plugins/water` 源码插件目录、清单、嵌入文件和 README
- [x] 2.2 更新 `apps/lina-plugins/go.mod` 与 `lina-plugins.go` 注册 water 插件
- [x] 2.3 新增插件 SQL/mock/uninstall 资源占位，确认不创建 `hg_*` 表

## 3. 后端实现

- [x] 3.1 新增 water API DTO、控制器和路由注册
- [x] 3.2 生成或维护 water 插件只读 media DAO，读取 `media_*` 表
- [x] 3.3 实现 media 策略解析、watermark YAML 解析和错误码封装
- [x] 3.4 实现 base64 图片处理、纯 Go 水印绘制、异步任务队列、任务状态仓库和回调通知
- [x] 3.5 增加后端单元测试覆盖策略解析、配置解析、图片处理和参数错误

## 4. 前端实现

- [x] 4.1 新增 water 插件中文管理/测试页面和 API client
- [x] 4.2 实现任务提交、同步预览、任务查询、策略来源展示和图片展示
- [x] 4.3 确认本模块不新增运行时 i18n 资源，并记录该判断

## 5. 自动化验证

- [x] 5.1 新增 water 插件 E2E 测试用例
- [x] 5.2 运行插件 Go 测试、前端 typecheck、i18n 检查、water E2E 和 OpenSpec 校验
- [x] 5.3 按 `lina-review` 对实现、规范、REST、数据权限、缓存和 i18n 影响做审查

## Feedback

- [x] **FB-1**: Mock 数据缺少可直接测试的完整水印策略配置
- [x] **FB-2**: 水印页面头部不应展示无交互意义的 `读取 media_* 表` 标签
