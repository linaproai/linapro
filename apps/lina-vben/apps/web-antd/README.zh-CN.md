# LinaPro Web Workspace

`apps/lina-vben/apps/web-antd` 是 `LinaPro` 的默认管理工作台。它是项目中的参考 `Vue 3 + Vben 5` 应用，负责消费宿主 `API`、渲染内建系统模块，并承载插件感知页面。

## 职责边界

- 渲染用于系统治理与插件管理的默认宿主工作台。
- 通过共享请求客户端与运行时配置层消费 `/api/v1` 下的宿主接口。
- 提供与宿主模块一一对应的稳定路由、页面、表单和表格实现。
- 为定时任务管理提供默认 `UI` 入口，包括任务分组、持久化任务与执行日志。

## 定时任务页面入口

当前迭代在 `/system` 下新增了 3 个管理入口。

| 路由 | 页面 | 用途 |
| --- | --- | --- |
| `/system/job` | 任务管理 | 创建、编辑、启停、立即执行与重置持久化任务 |
| `/system/job-group` | 分组管理 | 管理任务分组，并保护默认分组不可删 |
| `/system/job-log` | 执行日志 | 查看执行历史、日志详情、批量清理与终止运行实例 |

任务表单同时支持 `Handler` 与 `Shell` 两类任务。当前端公开运行时配置判定 `Shell` 执行被禁用 / 不受支持，或当前用户缺少 `system:job:shell` 权限时，`Shell` 相关选项会自动隐藏。

## 关键目录

```text
src/api/            宿主 API 客户端
src/adapter/        表单与表格适配器
src/router/         工作台路由
src/runtime/        公开运行时配置加载
src/views/          页面实现
```

## 常用命令

```bash
pnpm install
pnpm -F @lina/web-antd dev --host 127.0.0.1
pnpm -F @lina/web-antd build
```

## 相关入口

- `src/router/routes/modules/system.ts`：系统管理路由注册入口。
- `src/views/system/job/`：定时任务页面与任务表单。
- `src/views/system/job-group/`：任务分组列表与弹窗。
- `src/views/system/job-log/`：执行日志列表与详情弹窗。
- `src/runtime/public-frontend.ts`：`Shell` 能力与品牌展示所需的公开运行时配置加载入口。
