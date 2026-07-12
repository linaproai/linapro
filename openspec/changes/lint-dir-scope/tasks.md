## 1. linactl 定向 lint 实现

- [x] 1.1 在`lint.go`注册与参数解析中支持可选`dir`
- [x] 1.2 实现`dir`→ module 解析（绝对路径、插件根→backend、向上找`go.mod`）
- [x] 1.3 在 workspace modules 中过滤目标 module；不匹配时明确失败
- [x] 1.4 plan/summary 日志输出`scope=dir|workspace`与目标路径

## 2. Make 入口与文档

- [x] 2.1 `hack/makefiles/lint.mk`透传`dir`
- [x] 2.2 更新`hack/tools/linactl/README.md`与`README.zh-CN.md`
- [x] 2.3 更新`.agents/rules/backend-go.md`中的`make lint`使用说明

## 3. 测试与验证

- [x] 3.1 补充 unit tests：module 解析、workspace 过滤、无效路径、插件根→backend
- [x] 3.2 运行`cd hack/tools/linactl && go test`相关用例
- [x] 3.3 手工 smoke：`make lint dir=hack/tools/linactl plugins=0`与无`dir`兼容路径
- [x] 3.4 `openspec validate lint-dir-scope --strict`
