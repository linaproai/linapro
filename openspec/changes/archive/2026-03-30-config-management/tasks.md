## Tasks

### Task 1: Create database table and seed data
**Status**: done
**Description**: Create `v0.8.0.sql` in `manifest/sql/` with:
1. `sys_config` table DDL (id, name, key, value, remark, created_at, updated_at, deleted_at) with unique index on `key`
2. Mock data in `manifest/sql/mock-data/` with sample config records (e.g., `sys.app.name`, `sys.user.initPassword`)

After creating the SQL file, run `make init` and `make mock` to apply, then run `make dao` to generate DAO/DO/Entity files.

**Files**:
- `apps/lina-core/manifest/sql/v0.8.0.sql` (create)
- `apps/lina-core/manifest/sql/mock-data/06_mock_configs.sql` (create)

---

### Task 2: Create backend API definitions
**Status**: done
**Description**: Create API request/response definitions for config module in `api/config/v1/`. Each endpoint in a separate file following dict module pattern. Include proper `g.Meta` tags with `dc`, `eg` tags on all fields.

Endpoints:
- `config_list.go` — GET `/config` (paginated list with name/key/time filters)
- `config_get.go` — GET `/config/{id}` (get by ID)
- `config_create.go` — POST `/config` (create)
- `config_update.go` — PUT `/config/{id}` (update)
- `config_delete.go` — DELETE `/config/{id}` (delete)
- `config_by_key.go` — GET `/config/key/{key}` (get by key)
- `config_export.go` — GET `/config/export` (export Excel)

After creating API files, run `make ctrl` to generate controller skeletons.

**Files**:
- `apps/lina-core/api/config/v1/*.go` (create, 7 files)

---

### Task 3: Implement backend service layer
**Status**: done
**Description**: Create config service in `internal/service/config/` (note: current `internal/service/config/` exists for app config reading — the new config management service should be in a new package like `internal/service/sysconfig/` to avoid conflict, OR integrate into the existing config service). Follow dict service pattern.

Methods to implement:
- `List()` — paginated query with filters (name, key, time range)
- `GetById()` — get single record
- `Create()` — create with key uniqueness check
- `Update()` — update with key uniqueness check (exclude self)
- `Delete()` — soft delete single record
- `GetByKey()` — get by key name
- `Export()` — generate Excel file

**Files**:
- `apps/lina-core/internal/service/sysconfig/sysconfig.go` (create)

---

### Task 4: Implement backend controller layer
**Status**: done
**Description**: Fill in the auto-generated controller skeletons (from `make ctrl`) with service calls. Register the controller in `cmd_http.go` route bindings.

**Files**:
- `apps/lina-core/internal/controller/config/*.go` (edit auto-generated files)
- `apps/lina-core/internal/cmd/cmd_http.go` (edit — add import and Bind)

---

### Task 5: Create frontend API layer
**Status**: done
**Description**: Create TypeScript API functions and type definitions for config module following dict API pattern.

**Files**:
- `apps/lina-vben/apps/web-antd/src/api/system/config/index.ts` (create — API functions)
- `apps/lina-vben/apps/web-antd/src/api/system/config/model.d.ts` (create — type definitions)

---

### Task 6: Create frontend config management page
**Status**: done
**Description**: Create config management page with VXE-Grid table, search form, toolbar, and create/edit modal. Follow dict type page pattern and ruoyi-plus-vben5 reference.

Components:
1. `index.vue` — Main page with search bar, toolbar (export, batch delete, add), VXE-Grid table
2. `config-modal.vue` — Create/edit modal with form fields (name, key, value as textarea, remark as textarea)
3. `data.ts` — Column definitions, query form schema, modal form schema

Table columns: checkbox, 参数名称, 参数键名, 参数键值, 备注, 修改时间, 操作(编辑/删除)
Search form: 参数名称(Input), 参数键名(Input), 创建时间(RangePicker)
Toolbar: 导出, 批量删除, 新增

**Files**:
- `apps/lina-vben/apps/web-antd/src/views/system/config/index.vue` (create)
- `apps/lina-vben/apps/web-antd/src/views/system/config/config-modal.vue` (create)
- `apps/lina-vben/apps/web-antd/src/views/system/config/data.ts` (create)

---

### Task 7: Add frontend route
**Status**: done
**Description**: Add config management route to the system routes module.

**Files**:
- `apps/lina-vben/apps/web-antd/src/router/routes/modules/system.ts` (edit — add ConfigManagement route)

---

### Task 8: Create E2E tests
**Status**: done
**Description**: Create Playwright E2E tests covering all config management features:
- Page navigation and display
- Search/filter functionality
- Create config via modal
- Edit config via modal
- Delete config (single and batch)
- Export functionality

**Files**:
- `hack/tests/e2e/system/TC0020-config.ts` (create — check existing TC numbers first)

---

## Feedback

- [x] **FB-1**：user/dept/file 服务的事务管理缺失，Create/Update 方法应使用事务确保数据一致性
- [x] **FB-2**：user 和 post 服务中存在重复的部门树遍历逻辑，应抽取到 dept 服务复用
- [x] **FB-3**：错误处理不一致，部分忽略删除关联错误，应统一使用 gerror 并记录日志
- [x] **FB-4**：Session TouchOrValidate 执行两次查询，优化为先验证存在性再更新
- [x] **FB-5**：文件上传添加文件名清洗防止路径遍历攻击（MIME类型校验已移除）
- [x] **FB-6**：字典类型更新时未校验 Type 字段唯一性，可能导致重复
- [x] **FB-7**：日志导出方法无条数限制，大数据量可能导致内存溢出
- [x] **FB-8**：操作类型映射和文件场景标签保持硬编码（用户确认不使用字典模块）
- [x] **FB-9**：用户列表查询存在 N+1 问题，应批量查询部门信息
- [x] **FB-10**：字典类型增加导入功能，支持下载模板、批量导入、数据校验、覆盖/忽略模式切换
- [x] **FB-11**：字典数据增加导入功能，支持下载模板、批量导入、数据校验、覆盖/忽略模式切换
- [x] **FB-12**：参数设置增加导入功能，支持下载模板、批量导入、数据校验、覆盖/忽略模式切换
- [x] **FB-13**：全局分页选项增加100条/页：修改 `adapter/vxe-table.ts` 中 `pageSizes` 为 `[10, 20, 50, 100]`
- [x] **FB-14**：抽象通用导出确认弹窗组件 `ExportConfirmModal`：支持选中N条提示和全部导出提示，复用至所有导出模块
- [x] **FB-15**：参数设置(config)导出支持选择导出：当选中记录时传递 `ids` 参数，未选中时导出全部；统一其他模块(dept/post/notice/file)的行为
- [x] **FB-16**：字典类型/字典数据/部门/岗位/公告/文件管理等模块的导出按钮增加二次确认弹窗
- [x] **FB-17**：为所有导出模块的 E2E 测试增加导出确认弹窗的测试用例
- [x] **FB-18**：补充参数设置导入功能的实际文件上传和导入成功测试
- [x] **FB-19**：补充覆盖模式开关效果验证测试（重复key处理）
- [x] **FB-20**：补充导入结果提示消息验证测试
- [x] **FB-21**：统一所有模块导出文件命名规范，使用描述性名称（字典类型导出.xlsx、字典数据导出.xlsx、参数设置导出.xlsx等）
- [x] **FB-22**：新增字典合并导出接口（GET /dict/export），同时导出字典类型和字典数据到双Sheet Excel文件
- [x] **FB-23**：新增字典合并导入接口（POST /dict/import），支持同时导入字典类型和字典数据
- [x] **FB-24**：新增字典导入模板下载接口，返回包含两个Sheet的模板文件
- [x] **FB-25**：前端字典类型面板更新导出导入功能，使用合并接口
- [x] **FB-26**：前端字典数据面板移除导出和导入按钮
- [x] **FB-27**：更新E2E测试用例以覆盖新的导出导入功能
- [x] **FB-28**：字典合并导入接口（`/dict/import`）未正确处理覆盖模式开关，Controller层未读取`updateSupport`参数，Service层`CombinedImport`方法不支持更新已存在记录
- [x] **FB-29**：字典类型删除逻辑改为级联删除，删除字典类型时同时删除关联的字典数据
- [x] **FB-30**：字典类型删除确认弹窗增加提示信息，告知用户将同时删除关联的字典数据
