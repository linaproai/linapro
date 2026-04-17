## Overview

本次反馈为插件管理页增加一个只读详情弹窗，管理员点击操作列中的“详情”按钮后，可直接查看该插件的完整治理信息。
实现保持前端内聚，不引入新的后端接口，直接复用当前列表接口返回的插件元数据。

## Design Decisions

### 复用现有列表行数据

- 插件列表接口已经返回详情弹窗需要的核心字段，包括名称、标识、类型、版本、描述、安装/启用状态、授权状态，以及宿主服务申请与授权快照
- 详情弹窗直接读取当前行数据，避免额外一次接口请求，减少实现复杂度

### 详情弹窗信息分层展示

- 第一层使用 `Descriptions` 展示基础治理信息，覆盖插件名称、标识、类型、版本、描述、接入状态、当前状态、授权要求、授权状态、安装时间与更新时间
- 第二层展示宿主服务申请清单和授权快照，按服务类型分组，并显示方法与资源边界
- 当插件未声明宿主服务时，使用明确提示文案替代空白区域

### 操作列保持原有动作不变

- “详情”按钮作为只读动作始终展示
- 原有安装、卸载、启停逻辑与权限控制保持不变

## Verification

- 新增 `hack/tests/e2e/plugin/TC0078-plugin-detail-dialog.ts`
- 回归执行 `hack/tests/e2e/plugin/TC0074-plugin-management-action-permissions.ts`

## Feedback Follow-Ups

### OpenSpec 文档语言治理

当前 `openspec/config.yaml` 将 OpenSpec 产物整体固定为简体中文，这会让英文上下文下的新迭代文档语言失真；与此同时，归档流程没有统一英文要求，难以把归档后的 change 文档与主规范作为对外协作的稳定英文资产使用。

### 决策一：新建迭代文档跟随用户上下文语言

- 新建 active change 时生成的 `proposal.md`、`design.md`、`tasks.md` 和增量规范，统一跟随用户当前输入的上下文语言。
- 语言判断优先级为：用户显式指定语言 > 当前需求描述的主语言。
- 同一个 active change 内默认保持单一文档语言，避免在一次迭代中混入中英文内容，除非用户明确要求对整套文档切换语言。

### 决策二：归档资产统一使用英文

- 执行归档时，归档目录中的 `proposal.md`、`design.md`、`tasks.md` 和增量规范统一使用英文。
- 若归档流程会将 delta spec 同步到 `openspec/specs/` 主规范，则同步后的主规范内容也统一使用英文，不再跟随当前会话语言。
- 这样可以把 archive 与主规范都沉淀为面向国际化支持和社区贡献的稳定英文资产。

### 决策三：仅通过配置与项目规范实现约束

- 不修改现有 skill 或 `/opsx:*` 斜杠指令内容。
- 通过 `openspec/config.yaml` 提供生成与归档阶段的语言约束，再由项目规范补充同一治理规则，确保后续执行时有统一依据。

### 本次治理验证

- 执行 `openspec validate openspec-language-governance --strict`
- 执行 `openspec instructions proposal --change openspec-language-governance --json`，确认输出上下文已包含“新建迭代跟随上下文语言、归档统一英文”的规则
