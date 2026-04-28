# 默认管理工作台英文多语言巡检

## 巡检方法

- 页面巡检：在 `en-US` 语言下使用 `playwright-cli` 登录默认管理后台，并逐个访问当前可见的 22 个菜单路由
- 资源诊断：调用 `GET /api/v1/i18n/messages/missing?locale=en-US`，返回 `total=0`，说明当前问题不主要来自运行时消息包缺键
- 源码扫描：扫描宿主 `apps/lina-vben/apps/web-antd/src/views`、`apps/lina-vben/apps/web-antd/src/components`、`apps/lina-vben/apps/web-antd/src/router` 与内建插件 `apps/lina-plugins/*` 的前端页面、插件清单和安装 SQL

## 巡检结论

- 已巡检的 22 个菜单路由全部仍可见中文展示内容，问题覆盖宿主页面、宿主共享组件、内建插件页面和默认种子/演示数据
- 当前英文态残留中文的主要来源不是运行时消息包缺键，而是未接入 i18n 的页面硬编码文案、插件交付字面量、菜单与页签标题未刷新、以及默认 `seed/mock/demo` 数据未做本地化投影
- 后续跟进不能只修单个页面表头或按钮，必须同时覆盖：页面静态文案、动态元数据、共享组件、插件交付页面、默认演示内容、安装 SQL 初始化数据与页签/路由标题

## 分区域问题清单

### 1. 仪表盘与共享壳层

- `/dashboard/analytics` 仍显示 `用户量`、`总用户量`、`访问量`、`总访问量`、`下载量`、`使用量`、`流量趋势`、`月访问量`、`访问来源` 等统计卡片与图表标题
  - 来源范围：`apps/lina-vben/apps/web-antd/src/views/dashboard/analytics/index.vue`、`apps/lina-vben/apps/web-antd/src/views/dashboard/analytics/analytics-visits-data.vue`、`apps/lina-vben/apps/web-antd/src/views/dashboard/analytics/analytics-visits-source.vue`
- `/dashboard/workspace` 仍显示 `早安, 大人, 开始您一天的工作吧！`、`今日晴，20℃ - 32℃！`、`待办`、`项目`、`团队`、中文励志语句、中文动态流和待办事项描述
  - 来源范围：`apps/lina-vben/apps/web-antd/src/views/dashboard/workspace/index.vue`
- 共享壳层在多个页面持续暴露中文，例如当前用户展示名 `大人`、访问过调度模块后其他路由页签仍显示 `任务管理`、`分组管理`、`执行日志`
  - 来源范围：`apps/lina-core/manifest/sql/mock-data/003-mock-users.sql`、`apps/lina-core/manifest/sql/014-scheduled-job-management.sql`、`apps/lina-vben/apps/web-antd/src/router/routes/index.ts`、`apps/lina-vben/apps/web-antd/src/router/guard.ts`

### 2. 访问管理模块

- `/system/user` 仍显示 `用户账号`、`用户昵称`、`手机号码`、`用户状态`、`创建时间`、`用户列表`、`导出/导入/删除/新增`、`名称/头像/昵称/部门/角色/性别/邮箱/状态/操作` 等页面系统文案
- `/system/user` 同时暴露默认组织与演示用户内容，例如 `Lina科技`、`研发部门`、`市场部门`、`未分配部门`、`修改后的E2E用户`、`柳园`、`华倩`
  - 来源范围：`apps/lina-vben/apps/web-antd/src/views/system/user/index.vue`、`apps/lina-vben/apps/web-antd/src/views/system/user/data.ts`、`apps/lina-vben/apps/web-antd/src/views/system/user/user-drawer.vue`、`apps/lina-vben/apps/web-antd/src/views/system/user/user-import-modal.vue`、`apps/lina-vben/apps/web-antd/src/views/system/user/user-reset-pwd-modal.vue`、`apps/lina-vben/apps/web-antd/src/views/system/user/dept-tree.vue`、`apps/lina-core/manifest/sql/mock-data/003-mock-users.sql`
- `/system/role` 仍显示 `角色名称`、`权限字符`、`请选择状态`、`角色列表`、`数据权限`、`超级管理员`、`普通用户`、`全部数据权限`、`仅本人数据权限`、`分配`
  - 来源范围：`apps/lina-vben/apps/web-antd/src/views/system/role/index.vue`、`apps/lina-vben/apps/web-antd/src/views/system/role/data.ts`、`apps/lina-vben/apps/web-antd/src/views/system/role/role-drawer.vue`、`apps/lina-vben/apps/web-antd/src/views/system/role-auth/index.vue`、`apps/lina-core/manifest/sql/008-menu-role-management.sql`
- `/system/menu` 仍显示 `菜单名称`、`菜单状态`、`显示状态`、`菜单列表`、`级联删除`、`组件类型`、`权限标识`、`组件路径`、`目录/菜单/按钮`、以及按钮权限说明 `动态路由权限:*`
  - 来源范围：`apps/lina-vben/apps/web-antd/src/views/system/menu/index.vue`、`apps/lina-vben/apps/web-antd/src/views/system/menu/data.ts`、`apps/lina-vben/apps/web-antd/src/views/system/menu/menu-drawer.vue`、`apps/lina-core/manifest/sql/008-menu-role-management.sql`

### 3. 组织中心模块

- `/system/dept` 仍显示 `部门名称`、`部门状态`、`部门列表`、`折叠/展开/新增`、`部门编码`、`排序`、`操作`
- `/system/post` 仍显示 `岗位编码`、`岗位名称`、`岗位列表`、`排序`、`总经理`、`技术总监`、`开发工程师` 等默认组织与岗位内容
  - 来源范围：`apps/lina-plugins/org-center/frontend/pages/dept-management.vue`、`apps/lina-plugins/org-center/frontend/pages/dept-drawer.vue`、`apps/lina-plugins/org-center/frontend/pages/dept-data.ts`、`apps/lina-plugins/org-center/frontend/pages/post-management.vue`、`apps/lina-plugins/org-center/frontend/pages/post-drawer.vue`、`apps/lina-plugins/org-center/frontend/pages/post-data.ts`、`apps/lina-plugins/org-center/plugin.yaml`、`apps/lina-plugins/org-center/manifest/sql/001-org-center-schema.sql`

### 4. 系统设置模块

- `/system/dict` 仍显示 `字典名称`、`字典类型`、`字典类型列表`、`字典数据列表`、`字典键值`、`字典标签`、`备注` 以及 `导出/导入/删除/新增`
  - 来源范围：`apps/lina-vben/apps/web-antd/src/views/system/dict/type/index.vue`、`apps/lina-vben/apps/web-antd/src/views/system/dict/type/data.ts`、`apps/lina-vben/apps/web-antd/src/views/system/dict/type/dict-type-modal.vue`、`apps/lina-vben/apps/web-antd/src/views/system/dict/type/dict-type-import-modal.vue`、`apps/lina-vben/apps/web-antd/src/views/system/dict/data/index.vue`、`apps/lina-vben/apps/web-antd/src/views/system/dict/data/data.ts`、`apps/lina-vben/apps/web-antd/src/views/system/dict/data/dict-data-drawer.vue`、`apps/lina-vben/apps/web-antd/src/views/system/dict/data/dict-data-import-modal.vue`、`apps/lina-vben/apps/web-antd/src/components/dict/src/index.vue`
- `/system/config` 仍显示 `参数名称`、`参数键名`、`参数设置列表`、`参数键值`、`备注`、`修改时间`，并暴露中文默认配置内容，例如 `演示-首页公告文案`、`欢迎使用 LinaPro`、`仅用于演示自定义参数能力...`
  - 来源范围：`apps/lina-vben/apps/web-antd/src/views/system/config/index.vue`、`apps/lina-vben/apps/web-antd/src/views/system/config/data.ts`、`apps/lina-vben/apps/web-antd/src/views/system/config/config-modal.vue`、`apps/lina-vben/apps/web-antd/src/views/system/config/config-import-modal.vue`、`apps/lina-core/manifest/sql/007-config-management.sql`、`apps/lina-core/manifest/sql/mock-data/006-mock-configs.sql`
- `/system/file` 仍显示 `原始文件名`、`文件类型`、`使用场景`、`上传时间`、`文件列表`、`文件上传`、`图片上传`、`文件预览`、`详情/下载`、`用户头像`、`通知公告附件`
  - 来源范围：`apps/lina-vben/apps/web-antd/src/views/system/file/index.vue`、`apps/lina-vben/apps/web-antd/src/views/system/file/data.tsx`、`apps/lina-vben/apps/web-antd/src/views/system/file/file-detail-modal.vue`、`apps/lina-vben/apps/web-antd/src/views/system/file/file-upload-modal.vue`、`apps/lina-vben/apps/web-antd/src/views/system/file/image-upload-modal.vue`、`apps/lina-vben/apps/web-antd/src/components/upload/src/file-upload.vue`、`apps/lina-vben/apps/web-antd/src/components/upload/src/image-upload.vue`、`apps/lina-core/manifest/sql/005-file-storage.sql`

### 5. 内容管理模块

- `/system/notice` 仍显示 `公告标题`、`公告类型`、`创建人`、`通知公告列表`、`通知/公告`、`草稿/已发布`、`预览/编辑`
- 默认通知内容仍为中文，例如 `新功能上线预告`、`关于规范使用系统的公告`、`系统升级通知`
  - 来源范围：`apps/lina-plugins/content-notice/frontend/pages/notice-management.vue`、`apps/lina-plugins/content-notice/frontend/pages/notice-modal.vue`、`apps/lina-plugins/content-notice/frontend/pages/notice-preview-modal.vue`、`apps/lina-plugins/content-notice/frontend/pages/data.ts`、`apps/lina-plugins/content-notice/plugin.yaml`、`apps/lina-plugins/content-notice/manifest/sql/001-content-notice-schema.sql`

### 6. 系统监控模块

- `/monitor/online` 仍显示 `在线用户列表`、`登录账号`、`部门名称`、`浏览器`、`操作系统`、`登录时间`、`强制下线`
  - 来源范围：`apps/lina-plugins/monitor-online/frontend/pages/online-user.vue`、`apps/lina-plugins/monitor-online/frontend/pages/data.ts`、`apps/lina-plugins/monitor-online/plugin.yaml`
- `/monitor/server` 仍显示 `数据库信息`、`服务器信息`、`服务信息`、`Go 版本`、`服务启动时间`、`系统运行时长`、`网络流量`、`总发送/总接收`
  - 来源范围：`apps/lina-plugins/monitor-server/frontend/pages/server-monitor.vue`、`apps/lina-plugins/monitor-server/plugin.yaml`
- `/monitor/operlog` 仍显示 `模块名称`、`操作人员`、`操作类型`、`操作结果`、`操作日志列表`、`清空`、`日志编号`、`动态插件示例`、`同步源码插件`
  - 来源范围：`apps/lina-plugins/monitor-operlog/frontend/pages/operlog-management.vue`、`apps/lina-plugins/monitor-operlog/frontend/pages/data.ts`、`apps/lina-plugins/monitor-operlog/frontend/pages/operlog-detail-drawer.vue`、`apps/lina-plugins/monitor-operlog/plugin.yaml`、`apps/lina-plugins/monitor-operlog/manifest/sql/001-monitor-operlog-schema.sql`
- `/monitor/loginlog` 仍显示 `登录状态`、`登录日期`、`登录日志列表`、`提示信息`、`登录成功`
  - 来源范围：`apps/lina-plugins/monitor-loginlog/frontend/pages/loginlog-management.vue`、`apps/lina-plugins/monitor-loginlog/frontend/pages/data.ts`、`apps/lina-plugins/monitor-loginlog/frontend/pages/loginlog-detail-modal.vue`、`apps/lina-plugins/monitor-loginlog/plugin.yaml`、`apps/lina-plugins/monitor-loginlog/manifest/sql/001-monitor-loginlog-schema.sql`

### 7. 调度中心模块

- 侧边菜单和页签标题仍显示 `任务管理`、`分组管理`、`执行日志`，并在离开调度模块后继续污染其他页面的页签展示
- `/system/job` 仍显示 `任务分组`、`请选择分组`、`任务状态`、`关键字`、`定时任务列表`、`任务名称`、`所属分组`、`任务来源`、`定时表达式`、`调度范围`、`并发策略`、`停止原因`、`立即执行`
- `/system/job-group` 仍显示 `任务分组列表`、`分组编码`、`分组名称`、`任务数`、`系统默认任务分组，删除其他分组时任务会迁移至此。`
- `/system/job-log` 仍显示 `执行状态`、`执行节点`、`执行日志列表`、`触发方式`、`错误摘要`
- 默认调度数据仍为中文，例如 `在线会话清理`、`服务监控采集`、`服务监控清理`、`源码插件回显巡检`、`默认分组`、`宿主内置`、`插件内置`
  - 来源范围：`apps/lina-vben/apps/web-antd/src/views/system/job/index.vue`、`apps/lina-vben/apps/web-antd/src/views/system/job/form.vue`、`apps/lina-vben/apps/web-antd/src/views/system/job/form-handler.vue`、`apps/lina-vben/apps/web-antd/src/views/system/job/form-shell.vue`、`apps/lina-vben/apps/web-antd/src/views/system/job-group/index.vue`、`apps/lina-vben/apps/web-antd/src/views/system/job-group/modal.vue`、`apps/lina-vben/apps/web-antd/src/views/system/job-log/index.vue`、`apps/lina-vben/apps/web-antd/src/views/system/job-log/detail.vue`、`apps/lina-core/manifest/sql/014-scheduled-job-management.sql`

### 8. 扩展中心、开发中心与插件交付页面

- `/system/plugin` 仍显示 `插件标识`、`插件名称`、`插件类型`、`接入态`、`插件列表`、`上传插件`、`同步插件`、`详情/卸载/安装`、`源码插件/动态插件/启用/禁用`
  - 来源范围：`apps/lina-vben/apps/web-antd/src/views/system/plugin/index.vue`、`apps/lina-vben/apps/web-antd/src/views/system/plugin/plugin-detail-modal.vue`、`apps/lina-vben/apps/web-antd/src/views/system/plugin/plugin-uninstall-modal.vue`、`apps/lina-vben/apps/web-antd/src/views/system/plugin/plugin-dynamic-upload-modal.vue`、`apps/lina-vben/apps/web-antd/src/views/system/plugin/plugin-host-service-auth-modal.vue`、`apps/lina-vben/apps/web-antd/src/views/system/plugin/plugin-host-service-view.ts`、`apps/lina-vben/apps/web-antd/src/views/system/plugin/plugin-route-review-list.vue`
- `/about/api-docs` 与 `/about/system-info` 当前页面虽然主体内容大多可用英文，但仍会因为页签缓存继续显示中文调度标签，说明共享路由/页签标题链路没有完全国际化
  - 来源范围：`apps/lina-vben/apps/web-antd/src/router/routes/index.ts`、`apps/lina-vben/apps/web-antd/src/router/guard.ts`、`apps/lina-vben/apps/web-antd/src/views/about/system-info/index.vue`、`apps/lina-vben/apps/web-antd/src/views/about/config.ts`
- 动态插件示例页 `/link-6865-plugin-assets-plugin-demo-dynamic-v0-1-0-mount-js` 仍显示 `动态插件 SQL 示例记录` 和整段中文说明
  - 来源范围：`apps/lina-plugins/plugin-demo-dynamic/frontend/pages/mount.js`、`apps/lina-plugins/plugin-demo-dynamic/plugin.yaml`、`apps/lina-plugins/plugin-demo-dynamic/manifest/sql/001-plugin-demo-dynamic-records.sql`

### 9. 共享组件、弹窗与非菜单入口页面

- 宿主共享组件仍包含大量中文字面量，后续若打开导入、导出、裁剪、树选择、个人设置、安全设置、密码设置等流程，英文态仍会继续漏出中文
  - 来源范围：`apps/lina-vben/apps/web-antd/src/components/global/export-confirm-modal.vue`、`apps/lina-vben/apps/web-antd/src/components/tree/src/data.tsx`、`apps/lina-vben/apps/web-antd/src/components/tree/src/helper.tsx`、`apps/lina-vben/apps/web-antd/src/components/tree/src/tree-select-panel.vue`、`apps/lina-vben/apps/web-antd/src/components/upload/src/file-upload.vue`、`apps/lina-vben/apps/web-antd/src/components/upload/src/image-upload.vue`、`apps/lina-vben/apps/web-antd/src/components/cropper/src/cropper-modal.vue`
- 个人中心与认证相关页面也仍存在中文资源，需要纳入同一轮收口，避免用户从非菜单入口进入后继续遇到中文内容
  - 来源范围：`apps/lina-vben/apps/web-antd/src/views/_core/profile/index.vue`、`apps/lina-vben/apps/web-antd/src/views/_core/profile/base-setting.vue`、`apps/lina-vben/apps/web-antd/src/views/_core/profile/security-setting.vue`、`apps/lina-vben/apps/web-antd/src/views/_core/profile/password-setting.vue`、`apps/lina-vben/apps/web-antd/src/views/_core/profile/notification-setting.vue`、`apps/lina-vben/apps/web-antd/src/views/_core/authentication/code-login.vue`

### 10. 宿主与插件交付型种子数据

- 宿主 SQL 仍写入大量直接面向 UI 的中文默认数据，包括菜单、角色、调度任务、配置示例、组织与用户演示数据；这些内容在英文态下会直接投射到列表、页签、详情和下拉选项中
  - 来源范围：`apps/lina-core/manifest/sql/002-dict-dept-post.sql`、`apps/lina-core/manifest/sql/007-config-management.sql`、`apps/lina-core/manifest/sql/008-menu-role-management.sql`、`apps/lina-core/manifest/sql/014-scheduled-job-management.sql`、`apps/lina-core/manifest/sql/mock-data/003-mock-users.sql`、`apps/lina-core/manifest/sql/mock-data/006-mock-configs.sql`
- 多个内建插件仍通过 `plugin.yaml`、插件前端源码与插件安装 SQL 直接交付中文内容，而不是通过 `manifest/i18n/en-US/*.json` 或业务内容多语言模型进行英文投影
  - 来源范围：`apps/lina-plugins/org-center/plugin.yaml`、`apps/lina-plugins/content-notice/plugin.yaml`、`apps/lina-plugins/monitor-online/plugin.yaml`、`apps/lina-plugins/monitor-server/plugin.yaml`、`apps/lina-plugins/monitor-operlog/plugin.yaml`、`apps/lina-plugins/monitor-loginlog/plugin.yaml`、`apps/lina-plugins/plugin-demo-dynamic/plugin.yaml`

## 跟进建议

- 以页面分组而不是按单文件零散修补，保证同一模块的列表页、抽屉、弹窗、默认数据和插件资源一次收口
- 对框架自带的默认展示内容建立明确边界：凡是宿主和内建插件交付的内容，都必须在 `en-US` 下给出完整本地化结果
- 修复完成后补一轮英文态 E2E 巡检，至少覆盖菜单导航、页签标题、列表页、弹窗/抽屉和内建插件页面
