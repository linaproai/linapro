# 设计：登录框默认居中

## 决策

1. **默认值统一为 `panel-center`**
   - 后端运行时参数默认：`sys.auth.loginPanelLayout` → `panel-center`
   - 数据库迭代 Seed/对齐 SQL：将内置配置值更新为 `panel-center`
   - 前端公共配置状态与非法值回退：`panel-center`
   - 应用 preferences 覆盖：`authPageLayout: 'panel-center'`

2. **不引入新配置键**
   - 继续使用既有 `sys.auth.loginPanelLayout` 与公共前端字段 `auth.panelLayout`

3. **已有环境对齐策略**
   - 项目无兼容性负担；在 `005-config-management.sql` 种子中将内置参数 `sys.auth.loginPanelLayout` 默认写为 `panel-center`
   - 管理员之后仍可在参数设置中改回左/右

## 验证

- 后端：公共前端默认值/回退相关单元测试
- 前端：`public-frontend` 回退与应用 preferences 默认
- E2E：登录页默认布局断言改为居中；覆盖配置切换仍保留
