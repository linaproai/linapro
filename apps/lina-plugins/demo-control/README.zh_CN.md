# demo-control

`demo-control` 是 LinaPro 官方提供的演示环境只读保护源码插件。

当目标环境需要开启演示只读模式时，可在宿主`plugin.autoEnable`列表中加入`demo-control`来启用该插件。

## 能力范围

该插件负责：

- 基于`HTTP Method`的环境级演示请求治理
- 对`/api/v1`下宿主与插件 API 的写操作拦截
- 演示模式下登录、登出最小会话白名单放行
