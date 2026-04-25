## Context

默认管理工作台基于 `apps/lina-vben` 实现，当前 Logo 默认值来自 `packages/@core/preferences/src/config.ts` 中的 Vben 远程静态资源，认证页和基础布局都会通过 `preferences.logo.source` 消费该配置。启动阶段还会读取宿主 `/config/public/frontend` 公开配置，初始化 SQL 中的 `sys.app.logo` 和 `sys.app.logoDark` 会覆盖前端默认值。favicon 位于 `apps/web-antd/public/favicon.ico`，由 Vite 静态资源机制直接服务。

本次变更只替换项目品牌资产和内置配置种子值，并在现有公开前端配置白名单中补充用户默认头像参数；不改变页面结构、主题系统或运行时国际化链路。

## Goals / Non-Goals

**Goals:**

- 将用户提供的 `linapro-mark.png` 拷贝到管理工作台可随构建交付的本地静态资源目录。
- 将用户提供的 `favicon.ico` 替换为管理工作台当前 favicon。
- 让默认管理工作台壳层和认证页统一使用本地图标 Logo，并保留应用名称文本展示，避免继续加载 Vben 远程默认资源。
- 新增用户默认头像系统参数，默认值为 `/avatar.webp`，并让前端头像兜底逻辑消费该配置。
- 保持改动范围局限在前端静态资源、默认偏好配置与宿主内置配置种子值。

**Non-Goals:**

- 不新增品牌配置管理页面或新的后端 API 端点。
- 不调整登录页、侧边栏、顶部栏等布局结构。
- 不新增数据库表。
- 不处理历史缓存中的用户个性化偏好覆盖问题。

## Decisions

- 决定将 Logo 放在 `apps/lina-vben/apps/web-antd/public/` 下，并通过 `/linapro-mark.png` 引用。
  - 原因：`public` 目录资源会按原文件名进入构建产物，适合 favicon、品牌图片这类无需打包指纹引用的静态资产。
  - 备选方案：放入 `src/assets` 并通过模块导入。该方案需要新增导入层，且默认偏好配置位于共享包中，不如公共路径直接稳定。

- 决定只修改 `packages/@core/preferences/src/config.ts` 中的默认 Logo 来源。
  - 原因：认证页和基础布局都已经消费统一的 `preferences.logo.source`，修改默认值即可保持入口一致。
  - 备选方案：在页面组件内分别覆盖 Logo。该方案会制造重复配置，并增加后续维护成本。

- 决定同步修改宿主配置初始化 SQL、服务兜底值与接口示例中 `sys.app.logo` 与 `sys.app.logoDark` 的默认值。
  - 原因：默认管理工作台会在启动时读取公开前端配置，若种子值仍指向 Vben 远程资源，真实运行态会覆盖前端默认 Logo。
  - 备选方案：忽略后端种子配置，仅依赖前端默认值。该方案只能覆盖后端配置接口不可用的兜底场景，不满足默认运行态要求。

- 决定将用户默认头像作为受保护系统参数 `sys.user.defaultAvatar` 暴露在现有公开前端配置中。
  - 原因：头像兜底需要由宿主统一治理，且默认管理工作台已经通过公开前端配置同步品牌与界面参数。
  - 备选方案：只修改前端 `defaultAvatar` 默认值。该方案无法通过系统参数调整运行态头像兜底地址。

- 决定直接替换 `apps/web-antd/public/favicon.ico`。
  - 原因：当前项目已使用该路径作为 favicon 交付位置，原位替换符合最小改动原则。
  - 备选方案：新增另一个 favicon 文件并修改 `index.html` 引用。该方案没有实际收益。

## Risks / Trade-offs

- 本地浏览器或持久化偏好可能缓存旧 Logo → 通过刷新页面、清理前端偏好缓存或无痕窗口验证默认态。
- Logo 尺寸与现有侧边栏空间可能不完全匹配 → 保持 `fit: 'contain'`，由现有图片适配策略控制展示。
- favicon 可能被浏览器强缓存 → 验证时以文件哈希和构建产物为准，必要时强制刷新浏览器缓存。
