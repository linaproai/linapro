# 内部工具包

`internal/utility`用于归口`Lina Core`内部工具包。这些工具包不属于业务模块、运行时服务、公开`pkg`契约、生成模型或打包资源。

## 目录规则

| 规则 | 要求 |
|------|------|
| 根包 | 不得在`internal/utility`根目录直接放置`Go`文件。 |
| 子包命名 | 使用窄职责名称，例如`testsupport`、`repoenv`或`sourceasset`。 |
| 泛化命名 | 不得创建名为`util`、`utils`、`common`、`helper`或`misc`的二级包。 |
| 复用门槛 | 只有被多个`Lina Core`包复用，或明确属于跨包治理能力时，才放入该目录。 |
| 依赖方向 | 工具包不得依赖`internal/service`、`internal/controller`或`internal/cmd`中的业务实现。 |
| 公开边界 | 工具包只是`Lina Core`内部辅助能力，不是面向插件或外部调用方的稳定`pkg`契约。 |

## 放置指南

测试支撑、仓库环境探测、静态治理检查、本地源码资源发现等内部辅助能力可以放入`internal/utility`。

业务服务、`API`契约、插件协议、持有缓存状态的运行时组件、权限边界、生成代码，或仅被单个包使用的一次性辅助逻辑，不应放入`internal/utility`。单包私有辅助逻辑应保留在其所属包内。

## 当前包

| 包 | 职责 |
|----|------|
| `testsupport` | 共享测试辅助能力，用于探测仓库和官方插件工作区状态。 |
