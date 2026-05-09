# monitor-online

`monitor-online` 是 LinaPro 官方提供的在线用户治理源码插件。

## 能力范围

该插件负责：

- 在线会话投影查询
- 强制下线治理 API 与工作台入口

该插件复用宿主持有的 `sys_online_session` 会话真相源，不额外创建插件自有表。
