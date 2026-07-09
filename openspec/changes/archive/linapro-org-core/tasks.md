## Summary

- [x] FB-1: 部门管理页面在部门树为空时无法创建首个顶级部门
- [x] FB-2: 字典标签文字更新后已打开列表未同步显示最新标签
- [x] FB-3: 切换工作台 Tab 后前一个列表页分页状态丢失
- [x] FB-4: 动态插件样例 CI 单测因 config.yaml 缺失失败
- [x] FB-5: 动态插件样例不应将运行配置契约从 config.yaml 改为 config.example.yaml

### 验证记录

- 前端类型检查：通过
- 字典 Store 单元测试：通过，覆盖数组引用稳定、标签同步、并发去重
- E2E TC004 部门空树创建顶级部门：通过
- E2E TC010 字典标签同步与 Tab 分页保持：通过
- `make db.init confirm=init`：通过，无额外补丁 SQL
- 动态插件样例 Go 单测：通过
- `make test.go plugins=1 race=true`：通过
- `openspec validate --strict`：通过
- `git diff --check`：通过
