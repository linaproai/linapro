# Tasks

## Summary

- [x] 交付跨平台`linactl`主入口、工具整合、代码生成`dir=`、插件`hack/config.yaml`/`build.commands`、Agent 桥接、月度 OpenSpec、release/nightly、安装脚本、`lina-perf-audit`、`Go`静态检查与`lint dir=`、公开参数收敛。
- [x] 交付`linactl upgrade` / `make upgrade`：官方`linapro` remote、默认最新稳定 tag、`v=`版本/分支、脏工作区拒绝、`force`、合并不更新`apps/lina-plugins`；单测与临时仓库路径覆盖。
- [x] 扩展`i18n.check`校验宿主 seed/常量与`i18n.enabled`插件`SysConfigKey`的`config.<key>.name/remark`；补齐宿主/插件缺译；同步 i18n 规则与 linactl 文档。
- [x] 反馈：代码生成目标与插件 Makefile、smoke 旧`init`、lint/wasip1/builder 配置、兼容面删除、升级源固定官方且保留本地插件、config 展示键缺译。
- [x] 验证：`go test ./hack/tools/linactl/...`、`make i18n.check`/`plugins.check`、lint 定向/全量 smoke、OpenSpec 校验；跨平台入口为 Go/`linactl`。
- [x] 治理：框架`upgrade`与 lint`dir=`无运行时 HTTP/权限/缓存影响；`i18n.check`扩展有语言包与规则影响；插件运行时升级契约归 plugin-framework。
