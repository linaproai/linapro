# Dict Management

## Purpose

定义字典类型与字典数据的查询、维护、导入导出能力，确保系统枚举和可配置选项能够通过统一字典体系进行管理而不是散落在代码中。

## Requirements

### Requirement: 字典类型列表查询
系统 SHALL 提供字典类型的分页列表查询接口。

#### Scenario: 查询字典类型列表
- **WHEN** 调用 `GET /api/v1/dict/type` 并传入分页参数 `pageNum` 和 `pageSize`
- **THEN** 返回字典类型列表和总数，格式为 `{list: [...], total: number}`

#### Scenario: 字典类型列表支持条件筛选
- **WHEN** 查询时传入筛选参数 `name`（字典名称）或 `type`（字典类型）
- **THEN** `name` 和 `type` 使用模糊匹配（LIKE）
- **THEN** 返回符合条件的字典类型列表

#### Scenario: 字典类型列表排除已删除记录
- **WHEN** 查询字典类型列表
- **THEN** 结果中不包含已软删除的记录

### Requirement: 创建字典类型
系统 SHALL 提供创建字典类型的接口。

#### Scenario: 创建字典类型成功
- **WHEN** 调用 `POST /api/v1/dict/type` 并提交 name（字典名称）和 type（字典类型）
- **THEN** 系统创建字典类型并返回成功

#### Scenario: 字典类型重复
- **WHEN** 创建字典类型时提交已存在的 type 值
- **THEN** 系统返回错误信息，提示字典类型已存在

#### Scenario: 必填字段校验
- **WHEN** 创建字典类型时缺少 name 或 type
- **THEN** 系统返回参数校验错误

### Requirement: 更新字典类型
系统 SHALL 提供更新字典类型的接口。

#### Scenario: 更新字典类型成功
- **WHEN** 调用 `PUT /api/v1/dict/type/{id}` 并提交要更新的字段
- **THEN** 系统更新对应字典类型信息并返回成功

#### Scenario: 更新不存在的字典类型
- **WHEN** 更新一个不存在的字典类型 ID
- **THEN** 系统返回错误信息

### Requirement: 删除字典类型
系统 SHALL 提供删除字典类型的接口。

#### Scenario: 删除字典类型成功
- **WHEN** 调用 `DELETE /api/v1/dict/type/{id}`
- **THEN** 字典类型被软删除（设置 deleted_at）

#### Scenario: 删除已关联数据的字典类型
- **WHEN** 删除一个已有关联字典数据的字典类型
- **THEN** 系统返回错误信息，提示该类型下存在字典数据，须先删除字典数据

### Requirement: 导出字典类型
系统 SHALL 提供将字典类型列表导出为 Excel 文件的功能。

#### Scenario: 导出字典类型
- **WHEN** 调用 `GET /api/v1/dict/type/export` 并传入筛选参数
- **THEN** 返回 Excel 文件流
- **THEN** 导出字段包括：字典名称、字典类型、状态、备注、创建时间

### Requirement: 字典类型选项列表
系统 SHALL 提供获取所有字典类型选项的接口，供下拉选择使用。

#### Scenario: 获取字典类型选项
- **WHEN** 调用 `GET /api/v1/dict/type/options`
- **THEN** 返回所有正常状态的字典类型列表（不分页）

### Requirement: 字典数据列表查询
系统 SHALL 提供按字典类型查询字典数据的分页列表接口。

#### Scenario: 按字典类型查询数据列表
- **WHEN** 调用 `GET /api/v1/dict/data` 并传入 `dictType` 参数和分页参数
- **THEN** 返回该字典类型下的数据列表和总数

#### Scenario: 字典数据列表支持标签筛选
- **WHEN** 查询时传入 `label` 参数
- **THEN** 使用模糊匹配（LIKE）筛选字典标签

### Requirement: 创建字典数据
系统 SHALL 提供创建字典数据的接口。

#### Scenario: 创建字典数据成功
- **WHEN** 调用 `POST /api/v1/dict/data` 并提交 dictType、label、value、sort 等字段
- **THEN** 系统创建字典数据并返回成功

#### Scenario: 必填字段校验
- **WHEN** 创建字典数据时缺少 dictType、label 或 value
- **THEN** 系统返回参数校验错误

### Requirement: 更新字典数据
系统 SHALL 提供更新字典数据的接口。

#### Scenario: 更新字典数据成功
- **WHEN** 调用 `PUT /api/v1/dict/data/{id}` 并提交要更新的字段
- **THEN** 系统更新对应字典数据信息并返回成功

### Requirement: 删除字典数据
系统 SHALL 提供删除字典数据的接口。

#### Scenario: 删除字典数据成功
- **WHEN** 调用 `DELETE /api/v1/dict/data/{id}`
- **THEN** 字典数据被软删除

### Requirement: 导出字典数据
系统 SHALL 提供将字典数据导出为 Excel 文件的功能。

#### Scenario: 导出字典数据
- **WHEN** 调用 `GET /api/v1/dict/data/export` 并传入 dictType 和筛选参数
- **THEN** 返回 Excel 文件流
- **THEN** 导出字段包括：字典标签、字典值、排序、Tag 样式、CSS 类、状态、备注、创建时间

### Requirement: 按字典类型获取选项数据
系统 SHALL 提供按字典类型获取选项数据的接口，供全局缓存使用。

#### Scenario: 获取指定类型的字典选项
- **WHEN** 调用 `GET /api/v1/dict/data/type/{dictType}`
- **THEN** 返回该类型下所有正常状态的字典数据，按 sort 升序排列
- **THEN** 返回字段包括：label、value、tagStyle、cssClass

### Requirement: 字典数据表设计
系统 SHALL 提供 sys_dict_type 和 sys_dict_data 两张表。

#### Scenario: sys_dict_type 表结构
- **WHEN** 查看 sys_dict_type 表结构
- **THEN** 表包含：id、name、type（UNIQUE）、status、remark、created_at、updated_at、deleted_at

#### Scenario: sys_dict_data 表结构
- **WHEN** 查看 sys_dict_data 表结构
- **THEN** 表包含：id、dict_type（字符串关联 sys_dict_type.type）、label、value、sort、tag_style、css_class、status、remark、created_at、updated_at、deleted_at

### Requirement: Tag 样式系统
系统 SHALL 提供完整的 Tag 样式配置能力，包括预设色和自定义色。

#### Scenario: 预设色渲染
- **WHEN** 字典数据的 tag_style 值为预设色名称（cyan、green、orange、pink、purple、red、danger、success、warning、info、primary、default）
- **THEN** DictTag 组件使用 Ant Design Tag 的 color 属性渲染对应颜色标签

#### Scenario: 自定义色渲染
- **WHEN** 字典数据的 tag_style 值为 hex 颜色值（如 #1677ff）
- **THEN** DictTag 组件使用自定义样式渲染该颜色

#### Scenario: CSS 类支持
- **WHEN** 字典数据配置了 css_class 值
- **THEN** DictTag 组件将该值作为 CSS 类名应用到标签元素

#### Scenario: TagStylePicker 组件
- **WHEN** 编辑字典数据时配置 Tag 样式
- **THEN** 提供"默认颜色"和"自定义颜色"两种模式切换
- **THEN** 默认颜色模式显示 12 种预设色下拉选择（带颜色预览）
- **THEN** 自定义颜色模式显示颜色选择器（hex 格式，禁用透明度）

### Requirement: 字典管理前端双面板布局
系统 SHALL 在字典管理页面采用左右双面板布局。

#### Scenario: 双面板交互
- **WHEN** 打开字典管理页面
- **THEN** 左侧显示字典类型列表，右侧显示字典数据列表
- **THEN** 右侧初始为空，提示选择字典类型

#### Scenario: 类型与数据联动
- **WHEN** 点击左侧某个字典类型行
- **THEN** 右侧自动加载并显示该类型下的字典数据

#### Scenario: 响应式布局
- **WHEN** 在桌面端查看
- **THEN** 左右面板并排显示（flex-row）
- **WHEN** 在移动端查看
- **THEN** 上下堆叠显示（flex-col）

### Requirement: 全局 DictTag 组件
系统 SHALL 提供全局可复用的 DictTag 组件，用于在表格等场景中渲染字典值。

#### Scenario: DictTag 渲染
- **WHEN** 使用 DictTag 组件并传入 dicts（字典选项数组）和 value（当前值）
- **THEN** 匹配对应的字典标签并按配置的 Tag 样式渲染

#### Scenario: 加载中状态
- **WHEN** dicts 数据尚未加载完成（空数组）
- **THEN** 显示加载中指示器

#### Scenario: 无匹配值
- **WHEN** value 在 dicts 中无匹配项
- **THEN** 显示 fallback 文本（默认为 "unknown"）

### Requirement: Pinia 字典缓存 Store
系统 SHALL 提供 Pinia Store 缓存字典数据，避免重复请求。

#### Scenario: 字典数据缓存
- **WHEN** 首次请求某字典类型的数据
- **THEN** 从 API 获取并缓存到 Store 中
- **WHEN** 再次请求相同字典类型的数据
- **THEN** 直接从缓存返回，不重复请求 API

#### Scenario: 请求去重
- **WHEN** 多个组件同时请求同一字典类型的数据
- **THEN** 仅发出一次 API 请求，所有组件共享同一结果

#### Scenario: 缓存刷新
- **WHEN** 在字典管理页面点击"刷新缓存"按钮
- **THEN** 清除所有缓存的字典数据，下次使用时重新从 API 获取

### Requirement: 字典初始化数据
系统 SHALL 提供基础的字典初始化数据。

#### Scenario: 初始化通用字典
- **WHEN** 执行 v0.2.0 数据库迁移脚本
- **THEN** 创建以下字典类型和数据：
  - `sys_normal_disable`（系统开关）：正常/停用
  - `sys_user_sex`（用户性别）：男/女/未知

### Requirement: Import dictionary types from Excel
The system SHALL support importing dictionary type records from an Excel file. The system SHALL provide a template download endpoint and validate imported data before persisting.

#### Scenario: Download dictionary type import template
- **WHEN** user requests GET `/dict/type/import-template`
- **THEN** system returns an Excel template with example data showing required columns: 字典名称、字典类型、状态、备注

#### Scenario: Import dictionary types with valid data
- **WHEN** user uploads a valid Excel file to POST `/dict/type/import`
- **THEN** system validates all rows, creates records, and returns success count

#### Scenario: Import dictionary types with validation errors
- **WHEN** user uploads an Excel file with invalid data (missing required fields, duplicate type values)
- **THEN** system rejects the entire import and returns error details with row numbers and reasons

#### Scenario: Import dictionary types with overwrite mode
- **WHEN** user uploads an Excel file with `updateSupport=true` and the file contains type values that already exist
- **THEN** system updates existing records with the imported values

#### Scenario: Import dictionary types with ignore mode
- **WHEN** user uploads an Excel file with `updateSupport=false` (default) and the file contains type values that already exist
- **THEN** system skips existing records and only creates new records

#### Scenario: Dictionary type import modal UI
- **WHEN** user clicks "导入" button on the dictionary type management page
- **THEN** system displays a modal with template download link, drag-and-drop upload area, file type hint (xlsx/xls), and overwrite/ignore mode switch

### Requirement: Import dictionary data from Excel
The system SHALL support importing dictionary data records from an Excel file. The system SHALL provide a template download endpoint and validate imported data before persisting.

#### Scenario: Download dictionary data import template
- **WHEN** user requests GET `/dict/data/import-template`
- **THEN** system returns an Excel template with example data showing required columns: 字典类型、字典标签、字典键值、排序、Tag样式、CSS类名、状态、备注

#### Scenario: Import dictionary data with valid data
- **WHEN** user uploads a valid Excel file to POST `/dict/data/import`
- **THEN** system validates all rows, creates records, and returns success count

#### Scenario: Import dictionary data with validation errors
- **WHEN** user uploads an Excel file with invalid data (missing required fields, non-existent dict type)
- **THEN** system rejects the entire import and returns error details with row numbers and reasons

#### Scenario: Import dictionary data with overwrite mode
- **WHEN** user uploads an Excel file with `updateSupport=true` and the file contains records with matching dictType+value combination
- **THEN** system updates existing records with the imported values

#### Scenario: Import dictionary data with ignore mode
- **WHEN** user uploads an Excel file with `updateSupport=false` (default) and the file contains records with matching dictType+value combination
- **THEN** system skips existing records and only creates new records

#### Scenario: Dictionary data import modal UI
- **WHEN** user clicks "导入" button on the dictionary data management page
- **THEN** system displays a modal with template download link, drag-and-drop upload area, file type hint (xlsx/xls), and overwrite/ignore mode switch

### Requirement: Dictionary combined export and import
The system SHALL support combined export and import of dictionary types and dictionary data in a single Excel file with two sheets.

#### Scenario: Combined dictionary export
- **WHEN** user requests GET `/dict/export`
- **THEN** system returns an Excel file with two sheets: "字典类型" and "字典数据"
- **THEN** export file name is "字典导出.xlsx"

#### Scenario: Combined dictionary import
- **WHEN** user uploads an Excel file with two sheets to POST `/dict/import`
- **THEN** system validates both sheets, imports dictionary types first, then imports dictionary data
- **THEN** system returns combined success count for both types and data

#### Scenario: Combined dictionary import template download
- **WHEN** user requests GET `/dict/import-template`
- **THEN** system returns an Excel template with two sheets containing example data for dictionary types and dictionary data
