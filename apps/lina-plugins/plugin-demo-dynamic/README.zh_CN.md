# plugin-demo-dynamic

`plugin-demo-dynamic` 是 Lina 的动态 WASM 插件样例，用来演示一个受治理运行时插件的最小闭环。

## 样例覆盖内容

- 一个在默认管理工作台中渲染的菜单入口
- 一个不依赖宿主 UI 框架的独立静态页面
- 通过动态插件桥执行的后端演示路由
- 对 `runtime`、`storage`、`network`、`data` 宿主服务的受治理访问

## 目录结构

```text
plugin-demo-dynamic/
  main.go
  plugin_embed.go
  plugin.yaml
  backend/
  frontend/
  manifest/
```

## 构建方式

构建全部动态插件产物：

```bash
make wasm
```

只构建当前样例：

```bash
make wasm p=plugin-demo-dynamic
```

运行时产物会输出到 `temp/output/plugin-demo-dynamic.wasm`。

## 后端契约

该样例通过动态插件公共前缀暴露受治理路由，并将业务逻辑保留在 `backend/internal/service/` 中。

## 宿主服务

该样例在 `plugin.yaml` 中申请了以下宿主服务：

- `runtime`
- `storage`
- `network`
- `data`

这些声明会在插件生命周期流程中由宿主进行审查和授权。

## 审查要点

- `plugin.yaml` 中的元数据和宿主服务声明清晰可读。
- 前端资源与声明的访问模式一致。
- 构建得到的 WASM 产物可以由源码树稳定复现。
- 后端复杂逻辑保留在 service 组件中，而不是堆在 controller 中。
