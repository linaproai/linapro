# 默认管理工作台英文布局巡检

## 巡检方法

- 巡检时间：`2026-04-25`
- 巡检环境：本地开发环境，前端 `http://127.0.0.1:5666`，后端 `http://127.0.0.1:8080`
- 巡检账号：`admin / admin123`
- 巡检语言：`en-US`
- 巡检视口：桌面 `1366 x 900`
- 巡检方式：使用 Playwright 自动登录、切换英文、逐页访问核心宿主页面与内建插件页面，并对关键列表页、个人中心页以及典型新增/上传表单截图取证
- 截图目录：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/`

## 巡检范围

### 路由页

- `/profile`
- `/system/user`
- `/system/role`
- `/system/menu`
- `/system/dept`
- `/system/post`
- `/system/dict`
- `/system/config`
- `/system/file`
- `/system/notice`
- `/monitor/online`
- `/monitor/server`
- `/monitor/operlog`
- `/monitor/loginlog`
- `/system/job`
- `/system/job-group`
- `/system/job-log`
- `/system/plugin`
- `/about/system-info`
- `/about/api-docs`

### 关键表单/弹窗

- 个人中心基础信息与修改密码
- 用户新增抽屉
- 部门新增抽屉
- 岗位新增抽屉
- 参数新增弹窗
- 动态插件上传弹窗

## 总体结论

- 当前英文态已经基本消除了大面积中文残留，但**布局适配**仍有一批稳定可复现的问题，主要集中在“窄标签列 + 长英文文案”的组合上。
- 影响最大的区域不是单个页面，而是几类共性模式：
  - 左侧导航和顶部页签的空间预算仍按中文长度设计；
  - 列表页搜索区的标签列偏窄，`Phone Number`、`Permission Key`、`Operation Result` 之类文案容易换行；
  - `vxe-table` 列宽在英文态下偏紧，导致表头换行、固定操作列压缩主内容区域，甚至把关键列默认挤出首屏；
  - 600px 级抽屉/520px 级弹窗中的表单标签列过窄，`Parent Department`、`Parameter Value`、`Phone Number` 等标签容易断成两行。
- 本轮巡检中，`/system/dict`、`/system/user`、`/system/menu`、`/monitor/operlog`、`/system/job-log` 和若干抽屉/弹窗问题最明显，建议优先处理。

## 优先级排序

### P0：共享壳层与全局布局策略

- 左侧导航在展开态下会截断较长英文菜单，例如 `Dynamic Plugin Demo`、`Organization Management`。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/profile-base.png`
  - 风险：用户无法在不 hover 的情况下快速完整辨识菜单含义。
- 顶部页签在同时打开多个英文路由后更早进入溢出状态，文件、监控、调度等页面组合时尤为明显。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/file-list.png`
  - 风险：长英文标题更容易把页签栏推入滚动，降低定位效率。
- 建议：
  - 将展开态侧栏宽度提升到更适合英文的区间，例如 `232px ~ 248px`；
  - 对被截断的菜单统一补 tooltip，而不是只依赖视觉省略号；
  - 为页签栏补充更明确的滚动阴影/渐隐提示，并评估英文态下的默认页签最小宽度。

### P1：个人中心与高频表单

- `/profile` 基础信息页：`Phone Number` 标签换行，右侧输入框对齐被拉高。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/profile-base.png`
  - 建议：在紧凑表单里可直接缩短为 `Phone`；若希望保留完整语义，需扩大表单标签列宽。
- `/profile` 修改密码页：`Current Password`、`New Password`、`Confirm Password` 全部换行为两行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/profile-password.png`
  - 建议：这里更适合**增加标签列宽或改为顶部标签布局**，不建议把语义过度缩短到难以理解。
- `/system/user` 新增抽屉：`Phone Number` 再次换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/user-add-drawer.png`
  - 建议：抽屉表单统一建立英文态标签宽度基线，例如 `120px+`，避免每个表单单独打补丁。
- `/system/dept` 新增抽屉：`Parent Department`、`Department Name`、`Phone Number` 都发生换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/dept-add-drawer.png`
  - 建议：部门类抽屉改成更宽的 label 列，或把 `Parent Department` 简化为 `Parent Dept.`。
- `/system/post` 新增抽屉：`Position Name`、`Position Code` 均换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/post-add-drawer.png`
  - 建议：`Position` 系列表单适合使用 `Name` / `Code` 的上下文缩写，或改成两列表单时放宽标签区域。
- `/system/config` 新增弹窗：`Parameter Name`、`Parameter Key`、`Parameter Value` 三个标签全部断行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/config-add-modal.png`
  - 建议：参数管理弹窗可直接把英文短文案改成 `Name` / `Key` / `Value`，完整语义放在标题或 tooltip 中。
- `/system/plugin` 动态插件上传弹窗：开关说明 `Allow upload to overwrite an existing plugin package with the same ID and version` 被迫换成两行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/plugin-upload-modal.png`
  - 建议：这类长说明不适合作为单行 label，建议改为开关下方 help text，或拆成短标签 + 说明文本。

### P1：列表页搜索区

- `/system/user`：`User Account`、`User Nickname`、`Phone Number` 在搜索区换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/user-list.png`
- `/system/role`：`Permission Key` 换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/role-list.png`
- `/system/menu`：`Menu Status`、`Visible Status` 换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/menu-list.png`
- `/system/dept`：`Department Name` 换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/dept-list.png`
- `/system/post`：`Position Code`、`Position Name` 换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/post-list.png`
- `/system/config`：`Parameter Name`、`Parameter Key` 换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/config-list.png`
- `/system/file`：`Original Name`、`Usage Scene`、`Uploaded At` 换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/file-list.png`
- `/monitor/online`：`User Account` 换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/online-list.png`
- `/monitor/operlog`：`Module Name`、`Operation Type`、`Operation Result`、`Operation Time` 全部换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/operlog-list.png`
- `/monitor/loginlog`：`User Account`、`Login Status` 换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/loginlog-list.png`
- `/system/job-group`：`Group Name` 换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/job-group-list.png`
- `/system/job-log`：`Execution Status`、`Execution Node` 在搜索区有明显长度压力。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/job-log-list.png`
- `/system/plugin`：`Plugin Name` 换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/plugin-list.png`
- 建议：
  - 为搜索区建立英文态专用 label 宽度，而不是沿用中文态默认值；
  - 对高频紧凑文案采用统一缩写策略，例如 `Phone`、`Permission`、`Uploaded`、`Parent Dept.`、`Exec Status`；
  - 当页面可用宽度不足时，优先切换为“标签在上、输入框在下”的响应式表单布局，而不是继续挤压单行标签。

### P1：列表表头与固定列压缩

- `/system/user`：英文态下首屏无法同时容纳所有关键列，`Phone Number` 列会被整体挤出默认可见区域。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/user-list.png`
  - 建议：用户页应重新分配列宽，或将 `Phone Number` 简化为 `Phone`，并避免右侧固定操作列过早压缩中间数据列。
- `/system/role`：`Sort Order` 表头换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/role-list.png`
  - 建议：在表格中可直接缩短为 `Order`，并设置英文态最小列宽。
- `/system/menu`：`Sort Order`、`Component Type`、`Permission Key`、`Component Path`、`Created At` 均出现换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/menu-list.png`
  - 建议：菜单页不适合继续使用中文态列宽，建议成组调整列宽，并为超长列标题补 tooltip。
- `/system/post`：`Sort Order` 换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/post-list.png`
- `/system/dict`：左右双栏都存在严重表头换行，`Dictionary Name`、`Dictionary Type`、`Dictionary Label`、`Dictionary Value`、`Sort Order`、`Created At` 都被压成窄列。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/dict-list.png`
  - 建议：这是当前最明显的英文布局问题之一。建议在英文态下扩大双栏最小宽度、允许左右面板改为上下堆叠，或直接在双栏里使用更短的英文表头并统一 tooltip。
- `/system/file`：`Original Name` 表头换行，文件名与使用场景也有明显压缩。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/file-list.png`
  - 建议：表头可缩短为 `Filename` / `Usage` / `Uploaded`，同时对单元格统一启用 ellipsis + tooltip。
- `/monitor/operlog`：`Operation Summary`、`Operation Type`、`Operation Result`、`Operation Date` 均发生换行。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/operlog-list.png`
- `/monitor/loginlog`：`Login Status` 表头换行，浏览器与操作系统值也出现截断。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/loginlog-list.png`
  - 建议：监控日志页除了缩短表头，还应把长值字段统一改为省略 + hover tooltip，而不是默认硬截断。
- `/system/job`、`/system/job-log`：固定操作列明显侵占主内容区，`Cron Expression`、`Trigger Type`、`Execution Status` 等英文表头和数据列被压缩。
  - 证据：`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/job-list.png`、`openspec/changes/framework-i18n-foundation/artifacts/english-layout-audit/job-log-list.png`
  - 建议：调度中心应优先重新设计操作列宽度与主数据列的最小宽度，必要时把次级操作收纳进下拉菜单。

## 页面级改进建议清单

### 1. 共享导航与壳层

- 左侧菜单：扩展态宽度不够，建议统一提升宽度，并对 ellipsis 菜单补 tooltip。
- 顶部页签：在英文态下更容易溢出，建议增加滚动提示、限制最小宽度抖动，并评估长标题的 locale-specific 截断策略。

### 2. 个人中心

- `Base Info`：`Phone Number` 可在紧凑表单改为 `Phone`。
- `Password`：不建议把 `Current Password`、`New Password`、`Confirm Password` 继续强行缩短，优先增加标签宽度或切换为顶部标签。

### 3. 用户、角色、菜单、组织中心

- 用户页：搜索区和新增抽屉都建议把 `Phone Number` 缩短为 `Phone`；表格中为 `Phone` 列设更高优先级的默认可见宽度。
- 角色页：`Permission Key` 在搜索区可考虑简化为 `Permission`，表头 `Sort Order` 简化为 `Order`。
- 菜单页：`Menu Status` / `Visible Status` 建议分别简化为 `Status` / `Visibility`，`Component Path` 建议在表头用 `Component`，完整文案放 tooltip。
- 部门/岗位页：`Parent Department`、`Department Name`、`Position Code`、`Position Name` 可在抽屉里增加 label width，在表格里采用 `Parent Dept.`、`Code`、`Name` 等上下文内缩写。

### 4. 系统设置

- 字典页：优先级最高，建议直接调整双栏布局规则，而不是只改单个文案。
- 参数页：弹窗内的 `Parameter Name/Key/Value` 建议缩短为 `Name/Key/Value`。
- 文件页：`Original Name` 改为 `Filename`，`Usage Scene` 改为 `Usage`，`Uploaded At` 改为 `Uploaded`，可显著降低换行概率。

### 5. 系统监控与调度中心

- 在线用户、操作日志、登录日志：统一采用“短表头 + tooltip + ellipsis”的英文策略。
- 调度中心：重点不是单词翻译，而是固定列、最小列宽和列优先级分配需要重新设计。

### 6. 扩展中心

- 插件上传弹窗中的长开关说明不应继续占用 label 角色，建议改为开关下方帮助文本。

## 本轮未发现明显英文长度问题的页面

- `/system/notice` 列表与创建弹窗整体可读性尚可，当前主要问题更偏向个别未翻译占位符，而不是英文长度挤压。
- `/monitor/server` 在 `1366 x 900` 视口下没有出现明显的英文表头换行问题。
- `/about/system-info`、`/about/api-docs` 在本轮视口下未发现由英文长度直接引发的显著布局异常。

## 后续落点建议

- 共享壳层：`apps/lina-vben/apps/web-antd/src/layouts/basic.vue`
- 个人中心：`apps/lina-vben/apps/web-antd/src/views/_core/profile/base-setting.vue`、`apps/lina-vben/apps/web-antd/src/views/_core/profile/password-setting.vue`
- 用户/角色/菜单：`apps/lina-vben/apps/web-antd/src/views/system/user/`、`apps/lina-vben/apps/web-antd/src/views/system/role/`、`apps/lina-vben/apps/web-antd/src/views/system/menu/`
- 系统设置：`apps/lina-vben/apps/web-antd/src/views/system/dict/`、`apps/lina-vben/apps/web-antd/src/views/system/config/`、`apps/lina-vben/apps/web-antd/src/views/system/file/`
- 组织中心插件：`apps/lina-plugins/org-center/frontend/pages/`
- 监控插件：`apps/lina-plugins/monitor-online/frontend/pages/`、`apps/lina-plugins/monitor-operlog/frontend/pages/`、`apps/lina-plugins/monitor-loginlog/frontend/pages/`
- 调度中心：`apps/lina-vben/apps/web-antd/src/views/system/job/`、`apps/lina-vben/apps/web-antd/src/views/system/job-group/`、`apps/lina-vben/apps/web-antd/src/views/system/job-log/`
- 插件管理：`apps/lina-vben/apps/web-antd/src/views/system/plugin/`

## 建议的执行顺序

1. 先做共享壳层和表单/表格的英文长度基线能力（侧栏宽度、页签宽度、通用 label 宽度、通用列最小宽度、ellipsis + tooltip 规范）。
2. 再集中处理 `用户 / 菜单 / 字典 / 监控日志 / 调度中心` 这五类问题最集中的页面。
3. 最后回头清理个别文案长度策略，例如 `Phone Number -> Phone`、`Sort Order -> Order`、`Original Name -> Filename` 等。
