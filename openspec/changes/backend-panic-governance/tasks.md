## 1. Panic 使用边界梳理

- [x] 1.1 建立后端生产代码 `panic` allowlist，标注启动期、注册期、`Must*` 或未知异常重抛等保留原因
- [x] 1.2 新增静态检查脚本或测试，扫描生产 Go 文件并阻断 allowlist 外的 `panic` 调用

## 2. 运行期错误返回整改

- [x] 2.1 将 Excel 单元格坐标和文件关闭辅助逻辑中的非必要 `panic` 改为显式错误返回
- [x] 2.2 将动态插件 hostServices 规范化流程拆分为返回错误的运行期路径和必要的 Must 路径
- [x] 2.3 将运行时配置读取中的解析错误改为显式 `error` 返回，避免静默降级绕过规范

## 3. 测试与验证

- [x] 3.1 补充 Go 单元测试覆盖 Excel 辅助函数、hostServices 非法输入、运行时配置异常值返回和 panic allowlist 检查
- [x] 3.2 运行受影响后端包测试与 OpenSpec 校验，确认不需要 i18n 资源变更
- [x] 3.3 调用 lina-review 完成本次变更审查并处理发现项

## Feedback

- [x] **FB-1**: Cron 运行时配置读取不应以日志降级替代显式 error 返回
- [x] **FB-2**: closeutil 和 excelutil 的关闭错误日志应说明 nil error 指针误用并接收调用上下文
- [x] **FB-3**: 项目规范和 lina-review 应要求日志调用沿调用链传递 ctx 以保留链路追踪
- [x] **FB-4**: panic allowlist 检查应从 lina-core 根目录迁移到 internal/cmd 测试目录，并避免把 test helper 当作生产 panic 边界维护
- [x] **FB-5**: panic allowlist 测试应降低定制化字符串拼接和扫描逻辑耦合，提升可维护性
