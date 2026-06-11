# 修正 pkg/plugin 依赖方向

## Why

`apps/lina-core/pkg/plugin`的预期分层是`capability`（传输无关的领域能力契约层）在底层，`pluginhost`（源码插件接入）与`pluginbridge`（动态插件 guest SDK）在其上各自独立。当前存在两处违背该分层的反向依赖：`capability/recordstore`在 wasip1 构建中直接 import `pluginbridge/protocol`并内嵌 host-service 传输编码（契约层依赖传输层，分层倒置）；`pluginhost`的源码插件升级回调契约通过`pluginbridge/contract.ManifestSnapshotV1`暴露动态插件 ABI 类型（源码插件公开 API 被动态插件 ABI 污染）。这两处缺陷使 protocol wire 变更会穿透到契约层，且依赖方向仅靠约定维持、没有任何治理门禁防止回归。

## What Changes

- **BREAKING** 将`pkg/plugin/capability/recordstore`整体迁移到`pkg/plugin/pluginbridge/recordstore`（含`internal/plan`子包），所有调用方 import 路径同步更新，不保留旧路径。RecordStore 本就只出现在`pluginbridge.Services`、不在`capability.Services`，是动态插件专属 guest SDK，迁移后其对`pluginbridge/protocol`的依赖成为同层内聚。
- **BREAKING** 将`ManifestSnapshotV1`类型定义从`pluginbridge/contract`迁移到`capability/capmodel`；`pluginbridge/contract`保留`type ManifestSnapshotV1 = capmodel` 别名以维持 protocol facade 的别名转发惯例，JSON wire 格式不变。`pluginhost`及其`internal/manifestview`改为依赖`capmodel`，切断`pluginhost`对`pluginbridge/contract`的依赖。
- 新增 import 边界治理测试，固化`pkg/plugin`三大子包的依赖方向：`capability/**`（非测试代码）不得 import `pluginbridge`或`pluginhost`；`pluginhost/**`（非测试代码）不得 import `pluginbridge`任何子包。
- 在`pkg/plugin`双语 README 中记录设计决定：`Runtime`、`Network`、`RecordStore`是动态插件专属能力，不进入`capability.Services`目录；源码插件对前两者有宿主原生等价物（日志组件、HTTP 客户端）。
- 宿主侧消费方（`internal/service/plugin/internal/datahost`等使用`QueryPlan`契约的代码、`internal/service/plugin/internal/sourceupgrade`等使用 manifest snapshot 的代码）import 路径同步更新，无行为变更。

## Capabilities

### New Capabilities

无。本变更不引入新的能力语义，仅修正既有组件的包归属和依赖方向治理。

### Modified Capabilities

- `plugin-package-boundary-governance`：record store SDK 的规范位置从`pkg/plugin/capability/recordstore`改为`pkg/plugin/pluginbridge/recordstore`；`capability`下`*cap`命名豁免列表移除`recordstore`；新增`pkg/plugin`子包依赖方向要求（`capability`不依赖`pluginhost`/`pluginbridge`，`pluginhost`不依赖`pluginbridge`）及其治理验证要求。
- `pluginbridge-subcomponent-architecture`：`pluginbridge`新增公开`recordstore`子组件（guest record store SDK 及其`internal/plan`）；明确该子组件依赖`protocol`的方向合法性。

## Impact

- **代码**：`apps/lina-core/pkg/plugin/capability/recordstore/`（7 个文件 + `internal/plan` 5 个文件）整体迁移；`capability/capmodel`新增 manifest snapshot 类型；`pluginbridge/contract/contract_lifecycle.go`改为别名；`pluginbridge`根包（directory、stub）、`pluginhost`（manifest 文件、`internal/manifestview`）、宿主侧`internal/service/plugin/internal/{datahost,wasm,sourceupgrade,runtime}`中引用旧路径的 import 更新。
- **行为**：无运行时行为变更。`LifecycleRequest`的 JSON wire 格式、host-service 协议、recordstore 查询计划语义全部保持不变。
- **测试**：迁移包的既有测试随包移动；新增 import 边界治理测试（单元测试形态，扫描 import 声明）。
- **文档**：`pkg/plugin/README.md`与`README.zh-CN.md`同步更新（组件职责表、动态插件专属能力说明）。
- **无影响判断**：无 HTTP API、DTO、路由、数据库、SQL、前端、i18n 运行时文案变更；无缓存一致性影响；无数据权限路径变更（recordstore 的宿主侧授权执行路径仅改 import）；无新增运行期依赖（纯类型迁移，DI 装配不变）。
