# 水印源码插件

水印源码插件提供基于媒体策略表的截图水印处理能力。

插件读取 `media_strategy`、`media_strategy_tenant`、`media_strategy_device`、`media_strategy_device_tenant`，不创建旧 `hg_*` 表，也不做宿主租户隔离存储。
