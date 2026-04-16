## 1. OpenSpec 规范治理

- [x] 1.1 审计并修复 `openspec/specs/` 中不符合当前 schema 的主规范结构
- [x] 1.2 清理归档残留文件并验证相关 capability 可正常执行 `openspec validate` 与归档
- [x] 1.3 为本次整改补齐 `spec-governance` 相关规范与说明

## 2. 后端 GoFrame 合规整改

- [x] 2.1 审计 `apps/lina-core/internal/controller/` 与 `apps/lina-core/internal/service/` 中的 GoFrame 规范违规项
- [x] 2.2 修复生产代码中的手写软删除过滤、非推荐 ORM 用法和依赖注入问题
- [x] 2.3 补齐导出方法、结构体和关键字段的规范注释

## 3. API 合同一致性整改

- [ ] 3.1 统一 `apps/lina-core/api/` 中的 REST 方法、路径参数风格和参数绑定方式
- [x] 3.2 补齐 API DTO 的 `dc` / `eg` 标签并修正文档不一致项
- [x] 3.3 若接口合同有变更，同步更新前端调用和相关 E2E 用例

## 4. 模块解耦与验证

- [ ] 4.1 为高耦合业务模块引入启用/禁用配置与后端降级处理
- [x] 4.2 执行 `go test ./...`、OpenSpec 校验和相关 E2E 回归测试
- [x] 4.3 完成 `openspec-review` 审查并准备归档

## Feedback

- [x] **FB-1**: 后端 API 输入 DTO 参数标签统一使用 `json`，禁止混用 `p` 与 `json`
- [x] **FB-2**: 移除本轮超出范围的 `dept/post` 模块开关实现，恢复纯规范整改范围
- [x] **FB-3**: 保持现有 API 路由地址不变，本轮仅整改参数标签、文档标签和注释一致性
