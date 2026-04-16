# Tiptap Editor

## Purpose

定义基于 Tiptap 的富文本编辑器组件能力，确保管理工作台表单和内容编辑场景拥有统一、可扩展且可复用的富文本输入基础设施。

## Requirements

### Requirement: Tiptap 编辑器组件
系统 SHALL 提供基于 Tiptap 的富文本编辑器 Vue 组件。

#### Scenario: 组件基本功能
- **WHEN** 在表单中使用 Tiptap 编辑器组件
- **THEN** 显示带有工具栏的富文本编辑区域
- **THEN** 支持 v-model 双向绑定，输出 HTML 格式内容

#### Scenario: 工具栏功能
- **WHEN** 查看编辑器工具栏
- **THEN** 包含以下功能按钮：加粗、斜体、下划线、删除线、标题（H1-H3）、有序列表、无序列表、引用块、代码块、分割线、撤销、重做、图片上传、链接

#### Scenario: 图片上传按钮
- **WHEN** 用户点击工具栏的图片上传按钮
- **THEN** 弹出文件选择对话框，支持选择图片文件
- **THEN** 选择后将图片以 Base64 格式内联插入编辑器

#### Scenario: 图片上传扩展点
- **WHEN** 查看图片上传处理逻辑
- **THEN** 上传逻辑通过 `uploadHandler` prop 传入，默认使用 Base64 内联
- **THEN** 后续接入 OSS 时仅需传入新的 `uploadHandler` 函数

#### Scenario: 链接插入
- **WHEN** 用户点击工具栏的链接按钮
- **THEN** 弹出输入框，输入 URL 后将选中文本设为超链接

#### Scenario: 组件禁用状态
- **WHEN** 传入 `disabled` prop 为 true
- **THEN** 编辑器变为只读模式，工具栏隐藏或禁用

#### Scenario: 内容高度自适应
- **WHEN** 使用编辑器组件
- **THEN** 支持通过 `height` prop 设置编辑区域高度，默认 300px

### Requirement: Tiptap 依赖包
系统 SHALL 引入以下 Tiptap 相关 npm 依赖包。

#### Scenario: 依赖安装
- **WHEN** 查看前端项目依赖
- **THEN** 包含 `@tiptap/vue-3`、`@tiptap/starter-kit`、`@tiptap/extension-image`、`@tiptap/extension-link`、`@tiptap/extension-placeholder`、`@tiptap/extension-underline`

### Requirement: Tiptap 组件文件结构
系统 SHALL 将 Tiptap 编辑器组件放置在前端通用组件目录下。

#### Scenario: 文件组织
- **WHEN** 查看 Tiptap 组件文件结构
- **THEN** 组件位于 `src/components/tiptap/` 目录下
- **THEN** 包含 `index.ts`（导出）、`src/editor.vue`（主组件）、`src/toolbar.vue`（工具栏组件）、`src/extensions.ts`（扩展配置）
