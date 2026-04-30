## 1. 回归反馈修复

- [x] 1.1 检查并修复监控组件数据上报定时任务默认间隔，将 `monitor.interval`、内置定时任务投影和 fallback 从 30 秒统一为 1 分钟
- [x] 1.2 检查并修复工作台页面英文环境残留中文内容，补齐运行时 i18n 资源和页面引用
- [x] 1.3 检查并修复用户管理页面英文环境角色名称仍显示 `超级管理员` 的问题，确保与角色管理页面一致展示 `Administrator`
- [x] 1.4 检查并修复角色管理页面英文环境默认种子角色 `普通用户` 仍显示中文的问题
- [x] 1.5 检查并修复菜单管理页面中 `Dynamic Route Permission:plugin-demo-dynamic:` 前缀按钮的挂载位置，确保按钮挂在所属动态插件菜单下
- [x] 1.6 检查并优化菜单管理树形列表交互，使可展开目录/菜单悬停显示可点击指针、点击标题区域可展开或折叠，并移除标题后的冗余图标提示
- [x] 1.7 检查并修复部门相关页面英文环境下系统生成的 `未分配部门` 虚拟节点本地化
- [x] 1.8 检查并优化岗位管理页面英文环境下新增/编辑岗位表单状态选择项换行问题
- [x] 1.9 检查并优化字典管理页面英文环境下新增/编辑字典类型表单 `Dictionary Type` 标签换行问题
- [x] 1.10 检查并优化字典管理页面英文环境下新增/编辑字典数据表单 `Tag Style` 标签换行问题，并修复标签样式下拉选项显示 i18n key 的问题
- [x] 1.11 检查并修复系统参数页面英文环境下 `用户登录-IP 黑名单列表`、`登录展示-登录副标题`、`登录展示-页面说明`、`登录展示-页面标题` 的名称或内容中文展示问题
- [x] 1.12 检查并优化服务监控页面英文环境下 `Disk Usage` 表格列宽，缩小 `Mount Path` 并增加 `File System`、`Total`、`Used`、`Available` 可读宽度
- [x] 1.13 为定时任务 `立即执行` 按钮增加二次确认弹窗，复用删除操作同类确认组件并使用执行语义样式

## 2. 自动化验证

- [x] 2.1 新增或更新 `TC0138-workbench-english-i18n.ts`，验证英文工作台不残留中文系统文案
- [x] 2.2 新增或更新 `TC0139-role-user-english-seed-display.ts`，验证用户管理与角色管理中内置角色英文展示一致
- [x] 2.3 新增或更新 `TC0140-menu-dynamic-permission-tree.ts`，验证动态插件按钮挂载和菜单树可展开节点点击交互
- [x] 2.4 新增或更新 `TC0141-org-dict-config-english-layout.ts`，验证未分配部门、岗位表单、字典表单和系统参数英文展示
- [x] 2.5 新增或更新 `TC0142-server-monitor-disk-table-english.ts`，使用 Playwright 截图检查 `File System`、`Total`、`Used`、`Available` 表头和列值未换行
- [x] 2.6 新增或更新 `TC0143-job-manual-trigger-confirm.ts`，验证点击立即执行先出现确认弹窗，取消时不触发接口，确认后再触发
- [x] 2.7 运行本变更相关 Playwright 用例并保存结果；若环境阻塞，记录具体命令和失败原因

## 3. 资源与审查

- [x] 3.1 检查宿主和源码插件 `manifest/i18n`、packed manifest、默认配置模板和 SQL 资源的一致性，明确记录本次 i18n 影响面
- [x] 3.2 运行相关 Go/前端静态检查或单元测试，覆盖本次修改的后端投影、插件挂载和前端表单/表格逻辑
- [x] 3.3 执行 `/lina-review` 范围审查，确认 OpenSpec、i18n、SQL、GoFrame、前端和 E2E 规范符合项目要求

## Feedback

- [x] **FB-1**: Add a subtle cyan edge glow to the management workbench logo in dark mode
- [x] **FB-2**: Fix scheduled-job action rendering so Shell rows show a single edit action and Run Now remains clickable through confirmation
- [x] **FB-3**: Set the default `cron.shell.enabled` runtime parameter to `true`
- [x] **FB-4**: Preserve explicit user theme preference across page refresh when public frontend theme default is different
- [x] **FB-5**: Fix dynamic plugin permission button names in the English menu management page
- [x] **FB-6**: Improve dictionary type add/edit form layout so `Dictionary Type` stays on one line in English
- [x] **FB-7**: Change the login password placeholder to use localized password input prompts
- [x] **FB-8**: Replace remaining frontend runtime Chinese hardcoded copy with i18n-backed or non-localized stable handling
- [x] **FB-9**: Preserve the actual user nickname in the dashboard workspace greeting instead of overriding built-in admin names
- [x] **FB-10**: Update dashboard workspace project cards to reflect LinaPro admin stack projects with 2026-05-01 dates
- [x] **FB-11**: Rename the system information menu to version information and complete the LinaPro description punctuation
- [x] **FB-12**: Use local WebP logo assets for LinaPro, GoFrame, and Vben workbench project cards and keep English descriptions single-line with ellipsis
- [x] **FB-13**: Update dashboard workspace quick navigation destinations and LinaPro demo activity/todo copy
- [x] **FB-14**: Remove markdown-style backticks from frontend locale JSON translations so raw backticks are not shown in UI text
- [x] **FB-15**: Stabilize dynamic plugin English menu regression by expanding the menu tree before asserting nested plugin rows
- [x] **FB-16**: Protect built-in dictionary types, dictionary data, and system parameters from deletion while keeping them editable (TC0154)
- [x] **FB-17**: Remove the demo-control plugin-governance write whitelist so enabled demo mode blocks plugin install, uninstall, enable, and disable operations
- [x] **FB-18**: Ensure installed-but-disabled demo-control does not block write requests before the plugin is enabled
