# 设计：登录 Slogan 插画系统参数

## 决策

1. **参数键**：`sys.auth.sloganImage`
2. **值类型**：`text`
3. **默认值**：`/slogan.svg`（Vben 内置插画静态资源）
   - 默认/非空：登录页渲染对应图片地址（`http(s)` 或站内绝对路径，语义对齐 `sys.app.logo`）
   - 空字符串：不渲染侧栏插画
4. **公开字段**：`auth.sloganImage`（不做文案类 i18n 投影，按 URL 原样下发；允许返回空串）
5. **读取语义**：该键使用“允许空串”的读取路径——库内已存在的空值不得回退到默认值，以支持“清空=隐藏插画”
6. **前端装配**：`auth.vue` 将非空配置解析为工作区静态资源 URL 后传入 `AuthPageLayout` 的 `slogan-image`；空值传空
7. **校验**：允许空；非空时最长 500 字符
8. **静态资源**：将 Vben 内置 slogan SVG 导出为 `public/slogan.svg`，构建时打入宿主公共静态资源
9. **SQL 落点**：参数种子合并写入 `005-config-management.sql`，不单独新增迭代 SQL 文件

## 数据流

```
sys_config(sys.auth.sloganImage)
  → GetPublicFrontend().Auth.SloganImage
  → GET /api/v1/config/public/frontend auth.sloganImage
  → publicFrontendSettings.auth.sloganImage
  → AuthPageLayout sloganImage prop
  → 有值渲染 <img>，无值渲染 SloganIcon
```

## 验证

- 后端：参数规格、校验、公共前端投影单测
- 前端：公共配置归一化单测
- E2E：参数列表可见；覆盖 slogan 后登录侧栏展示自定义图片（需左/右布局）
