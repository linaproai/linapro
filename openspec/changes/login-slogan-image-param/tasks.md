## 1. 后端与数据

- [x] 1.1 新增 `sys.auth.sloganImage` 公共前端参数规格、可选文本校验与投影
- [x] 1.2 扩展公共前端 API DTO 字段 `auth.sloganImage`
- [x] 1.3 在 `005-config-management.sql` 种子中写入内置参数
- [x] 1.4 补充 config 元数据 i18n 与 apidoc 字段翻译
- [x] 1.5 更新公共前端控制器映射与后端单元测试

## 2. 前端

- [x] 2.1 扩展 `public-frontend` 运行时类型、状态与归一化
- [x] 2.2 登录布局绑定 `slogan-image` 并解析工作区资源路径
- [x] 2.3 更新前端相关单元测试

## 3. 测试与验证

- [x] 3.1 更新/补充 E2E：参数可见与 slogan 图片消费
- [x] 3.2 运行单元测试、lint、openspec validate 与相关 E2E

## Feedback

- [x] **FB-1**: `sys.auth.sloganImage` 默认改为 Vben 内置插画地址 `/slogan.svg`，空值表示不展示插画
