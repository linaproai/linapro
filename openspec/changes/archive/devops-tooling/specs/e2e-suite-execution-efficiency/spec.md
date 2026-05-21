## ADDED Requirements

### Requirement: E2E 测试执行必须避免共享状态级联超时

系统 SHALL 将共享状态密集的 E2E 用例（如 IAM 角色、IAM 用户、文件管理、用户设置、默认工作台跳转和部分 runtime i18n 回归用例）纳入串行隔离清单。共享测试验证套件和 release workflow SHALL 显式使用 `parallel-workers=1`。

### Requirement: E2E 全局默认超时必须适应长链路用例

Playwright 全局默认 `test.timeout` SHALL 设置为 180 秒，只扩大单个 test 的最大执行窗口。`expect.timeout` SHALL 保持 10 秒，避免定位或断言失败被无意义拉长。

### Requirement: 过长 E2E 流程必须拆分为独立子用例

超过 60 秒的 E2E 流程 SHALL 拆分为多个独立子用例，每个子用例独立准备 mock 状态和清理逻辑。非断言目标的重复 UI 导航和清理 SHALL 通过 API 调用替代。
