## 技术方案

### 1. 数据库设计

新增 `sys_file` 文件管理表，使用场景（scene）字段直接存储在表中，不设独立关联表：

```sql
CREATE TABLE sys_file (
  id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '文件ID',
  name       VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '存储文件名',
  original   VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '原始文件名',
  suffix     VARCHAR(32)     NOT NULL DEFAULT '' COMMENT '文件后缀',
  size       BIGINT UNSIGNED NOT NULL DEFAULT 0  COMMENT '文件大小（字节）',
  hash       VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '文件SHA-256散列值，用于去重',
  url        VARCHAR(512)    NOT NULL DEFAULT '' COMMENT '文件访问URL',
  path       VARCHAR(512)    NOT NULL DEFAULT '' COMMENT '文件存储路径',
  engine     VARCHAR(32)     NOT NULL DEFAULT 'local' COMMENT '存储引擎：local=本地',
  scene      VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '使用场景：avatar/notice_image/notice_attachment等',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0  COMMENT '上传者用户ID',
  created_at DATETIME        NOT NULL COMMENT '上传时间',
  updated_at DATETIME        NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  INDEX idx_engine (engine),
  INDEX idx_created_by (created_by),
  INDEX idx_suffix (suffix),
  INDEX idx_hash (hash),
  INDEX idx_scene (scene)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件管理';
```

使用场景直接记录在 `scene` 字段中，系统预定义场景包括：`avatar`（用户头像）、`notice_image`（通知公告图片）、`notice_attachment`（通知公告附件）。相同散列值的文件如在不同场景使用，则各自创建独立的文件记录。

### 2. 后端架构

#### 2.1 文件存储抽象层

在 `internal/service/file/` 中设计存储接口，默认实现本地存储，后续可扩展 OSS：

```
service/file/
├── file.go          # 主服务：上传/下载/列表/删除/后缀列表等业务逻辑
├── file_code.go     # 错误码定义
├── storage.go       # Storage 接口定义
└── storage_local.go # 本地存储实现
```

**Storage 接口**：
- `Put(ctx, filename, data) (path, error)` -- 保存文件
- `Get(ctx, path) (data, error)` -- 读取文件
- `Delete(ctx, path) error` -- 删除文件
- `Url(ctx, path) string` -- 获取访问 URL

#### 2.2 API 端点设计

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/file/upload` | 上传文件（multipart/form-data），支持可选参数 scene |
| GET | `/file` | 文件列表（分页+筛选：文件名、后缀、使用场景、时间范围），支持排序（大小、上传时间） |
| GET | `/file/download/{id}` | 下载文件 |
| DELETE | `/file/{ids}` | 删除文件（支持批量） |
| GET | `/file/suffixes` | 获取数据库中已存在的文件后缀列表 |
| GET | `/file/scenes` | 获取系统预定义的使用场景列表 |

上传接口返回：
```json
{
  "id": 1,
  "name": "20260319_abc123.png",
  "original": "avatar.png",
  "url": "http://host:port/api/v1/uploads/2026/03/20260319_abc123.png",
  "suffix": "png",
  "size": 102400
}
```

文件列表接口返回的 url 字段为完整的 HTTP 地址，方便前端直接使用。

#### 2.3 文件去重机制

上传时计算文件 SHA-256 散列值，与已有记录比对。相同散列值的文件直接复用已有存储文件，仅新增一条文件记录。每条记录独立关联使用场景和业务上下文。

### 3. 前端架构

#### 3.1 通用上传组件

参考 ruoyi-plus-vben5 的实现，在 `src/components/upload/` 创建：

- `FileUpload` -- 文件上传组件（支持拖拽上传）
- `ImageUpload` -- 图片上传组件（picture-card 样式）
- 通用上传 hook（`useUpload`）处理进度、校验、错误等

组件通过 `v-model:value` 绑定文件 ID（单文件为 string，多文件为 string[]），回显时通过文件 ID 查询文件信息。上传时支持传入 scene 参数标识使用场景。

#### 3.2 文件管理页面

在系统管理菜单下新增"文件管理"子菜单，页面功能包含：

- 文件列表表格（文件类型、原始名、预览、大小、上传时间、上传者、操作）
- 搜索条件（文件名、文件类型下拉选择、使用场景、上传时间范围）
- 工具栏按钮（批量删除、文件上传、图片上传）
- 操作列（详情、下载、删除）
- 图片预览模式默认开启，非图片非 PDF 文件展示可点击的 URL 地址（过长省略，悬停显示完整链接）
- 详情弹窗展示文件完整信息及使用场景

文件类型筛选为 Select 下拉选择，选项从后端 `/file/suffixes` 接口动态获取已存在的后缀列表（不含点号）。

#### 3.3 现有功能改造

**通知公告编辑器**：
- 给 TiptapEditor 传入 `uploadHandler` prop，调用文件上传接口（scene=notice_image）
- 上传成功后返回完整 URL 插入编辑器
- 移除 Base64 内嵌模式依赖
- 新增 FileUpload 组件用于附件上传（scene=notice_attachment）

**用户头像上传**：
- 改为调用通用文件上传接口 `POST /file/upload`，传入 scene=avatar
- 上传后使用返回的 URL 更新用户头像
- 删除原有的 `POST /user/profile/avatar` 独立上传端点及静态文件服务路由

### 4. 文件存储路径

本地存储按年月组织目录结构：
```
temp/upload/
  2026/
    03/
      20260319_abc123.png
      20260319_def456.docx
```

### 5. 配置变更

`config.yaml` 中 `upload.path` 配置项保持不变（`temp/upload`），新增 `upload.maxSize` 配置项控制最大上传文件大小。
