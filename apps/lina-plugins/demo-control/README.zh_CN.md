# demo-control

`demo-control` 是 LinaPro 官方提供的演示环境只读保护源码插件。

当目标环境需要开启演示只读模式时，可在宿主`plugin.autoEnable`列表中加入`demo-control`来启用该插件。

## 能力范围

该插件负责：

- 基于`HTTP Method`的环境级演示请求治理
- 在宿主`/*`作用域下拦截整个系统请求链路
- 对宿主与插件写请求进行统一拦截
- 演示模式下登录、登出最小会话白名单放行
