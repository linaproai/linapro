## 1. 规范与默认值

- [x] 1.1 更新登录页展示规范：默认布局为 `panel-center`
- [x] 1.2 在 `005-config-management.sql` 将 `sys.auth.loginPanelLayout` 内置默认对齐为 `panel-center`
- [x] 1.3 更新后端公共前端运行时参数默认值为 `panel-center`
- [x] 1.4 更新前端 preferences、公共配置初始值与非法值回退为 `panel-center`

## 2. 测试

- [x] 2.1 更新后端公共前端相关单元测试默认期望
- [x] 2.2 更新前端 `public-frontend` 单元测试默认/回退期望
- [x] 2.3 更新登录页 E2E 默认布局断言为居中

## 3. 验证

- [x] 3.1 运行相关单元测试与 `openspec validate`
- [x] 3.2 运行登录页相关 E2E（若环境可用）
