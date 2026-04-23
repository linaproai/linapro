## 1. 后端路由审查数据投影

- [ ] 1.1 在插件管理 API DTO 中新增动态路由审查字段，定义方法、真实公开路径、访问级别、权限标识与摘要等展示模型
- [ ] 1.2 在 `apps/lina-core/internal/service/plugin/` 与 `apps/lina-core/internal/controller/plugin/` 中把当前 release 的动态路由合同投影到插件列表行数据，确保仅动态插件返回该清单并复用宿主公开路径构建逻辑
- [ ] 1.3 为插件列表投影补充单元测试，覆盖含 `public`/`login` 路由、权限标识以及空路由清单场景

## 2. 前端授权弹窗展示增强

- [ ] 2.1 在 `apps/lina-vben/apps/web-antd/src/api/system/plugin/` 与视图模型中接入动态路由清单字段
- [ ] 2.2 更新动态插件安装/启用授权弹窗，新增与宿主服务卡片并列的路由信息区，并按只读方式展示方法、公开路径、访问级别、权限标识与摘要
- [ ] 2.3 处理空路由场景，确保未声明动态路由的插件不会渲染冗余路由块，且现有宿主服务授权视图保持不变

## 3. 回归验证与 E2E

- [ ] 3.1 扩展 `hack/tests/e2e/extension/plugin/TC0073-plugin-host-service-authorization-review.ts`，增加动态路由列表展示断言，覆盖安装审查窗口中的公开路径、访问级别与权限标识展示
- [ ] 3.2 运行相关后端单元测试与 `TC0073-plugin-host-service-authorization-review.ts`，确认动态插件授权审查流程与既有宿主服务展示回归通过
