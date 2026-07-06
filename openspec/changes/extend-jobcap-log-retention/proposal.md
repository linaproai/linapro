# 为 `jobcap` 增加日志清理策略支持

## 背景

当前宿主定时任务管理 API 和底层 `jobmeta.SaveJobInput` 已支持任务级日志清理策略，但插件可消费的 `jobcap.SaveInput` 未暴露该字段，导致源码插件和动态插件通过 `Jobs().Create` 或 `Jobs().Update` 创建、更新定时任务时只能使用全局 `sys.cron.log.retention` 策略。需要补齐 `jobcap` 运行期领域能力契约，使插件创建的任务也能声明与宿主管理 API 一致的任务级日志清理策略。

## 范围

- 扩展 `jobcap.SaveInput`，增加可选任务级日志清理策略。
- 复用现有 `jobs.*` 动态 host service 方法，不新增方法或授权资源。
- 将策略映射到宿主 `jobmeta.SaveJobInput.LogRetentionOverride`，沿用现有校验、持久化和清理执行逻辑。
- 补充自动化测试覆盖源码插件能力适配和动态 host service JSON 转发。

## 非范围

- 不修改 `sys_job` 数据表结构。
- 不修改普通 HTTP 定时任务管理 API。
- 不修改定时任务日志清理执行算法。
- 不新增前端 UI 或 E2E 用例。
