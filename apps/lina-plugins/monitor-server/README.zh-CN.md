# monitor-server

`monitor-server` 是 LinaPro 官方提供的服务监控源码插件。

## 能力范围

该插件负责：

- 服务监控数据采集
- 过期快照清理
- 服务监控查询 API 与工作台入口

宿主保留 Cron 与插件生命周期底座，监控能力本身由该插件提供。
