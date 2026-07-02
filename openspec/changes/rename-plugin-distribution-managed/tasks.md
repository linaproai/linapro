## 1. 契约和规范

- [x] 1.1 更新`plugin-manifest-lifecycle`增量规范，将普通分发治理枚举从`marketplace`改为`managed`。
- [x] 1.2 更新宿主`API`枚举、DTO 文档标签、manifest 校验、release snapshot 校验和分发归一化逻辑。
- [x] 1.3 更新数据库 SQL 基线、插件清单、前端类型、宿主/插件说明文档和`apidoc`翻译资源。

## 2. 测试和验证

- [x] 2.1 更新 Go 单元测试、契约测试和 E2E 静态期望中的分发枚举断言。
- [x] 2.2 运行覆盖变更包的 Go 测试和必要的前端/静态验证。
- [x] 2.3 运行`openspec validate rename-plugin-distribution-managed --strict`。
- [x] 2.4 运行静态扫描，确认除旧值拒绝测试和增量说明外，当前有效契约中不再残留`marketplace`作为普通插件分发枚举。

## 3. 影响记录和审查

- [x] 3.1 记录影响判断：有宿主`apidoc`和插件清单说明的`i18n`影响；无缓存一致性、数据权限、运行期服务依赖、开发工具跨平台和新增`E2E`影响。
- [x] 3.2 完成实现和验证后调用`lina-review`进行代码和规范审查，审查通过后再标记任务完成。
