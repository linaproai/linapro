-- 003: Operation Log, Login Log, Add dept code field

-- ============================================================
-- 操作日志表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_oper_log (
    id              INT PRIMARY KEY AUTO_INCREMENT COMMENT '日志ID',
    title           VARCHAR(50)   NOT NULL DEFAULT '' COMMENT '模块标题',
    oper_summary    VARCHAR(200)  NOT NULL DEFAULT '' COMMENT '操作摘要',
    oper_type       TINYINT       NOT NULL DEFAULT 0  COMMENT '操作类型（1新增 2修改 3删除 4导出 5导入 6其他）',
    method          VARCHAR(200)  NOT NULL DEFAULT '' COMMENT '方法名称',
    request_method  VARCHAR(10)   NOT NULL DEFAULT '' COMMENT '请求方式（GET/POST/PUT/DELETE）',
    oper_name       VARCHAR(50)   NOT NULL DEFAULT '' COMMENT '操作人员',
    oper_url        VARCHAR(500)  NOT NULL DEFAULT '' COMMENT '请求URL',
    oper_ip         VARCHAR(50)   NOT NULL DEFAULT '' COMMENT '操作IP地址',
    oper_param      TEXT          NOT NULL             COMMENT '请求参数',
    json_result     TEXT          NOT NULL             COMMENT '返回参数',
    status          TINYINT       NOT NULL DEFAULT 0   COMMENT '操作状态（0成功 1失败）',
    error_msg       TEXT          NOT NULL             COMMENT '错误消息',
    cost_time       INT           NOT NULL DEFAULT 0   COMMENT '耗时（毫秒）',
    oper_time       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='操作日志记录表';

-- ============================================================
-- 登录日志表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_login_log (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT '日志ID',
    user_name   VARCHAR(50)  NOT NULL DEFAULT '' COMMENT '登录账号',
    status      TINYINT      NOT NULL DEFAULT 0  COMMENT '登录状态（0成功 1失败）',
    ip          VARCHAR(50)  NOT NULL DEFAULT '' COMMENT '登录IP地址',
    browser     VARCHAR(200) NOT NULL DEFAULT '' COMMENT '浏览器类型',
    os          VARCHAR(200) NOT NULL DEFAULT '' COMMENT '操作系统',
    msg         VARCHAR(500) NOT NULL DEFAULT '' COMMENT '提示消息',
    login_time  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='系统登录日志表';

-- ============================================================
-- 字典数据表: 添加唯一约束防止重复数据
-- ============================================================
-- 先清理已有重复数据（保留每组 dict_type+value 中 id 最小的记录）
DELETE t1 FROM sys_dict_data t1
INNER JOIN sys_dict_data t2
WHERE t1.id > t2.id AND t1.dict_type = t2.dict_type AND t1.value = t2.value;

-- MySQL 中 CREATE INDEX IF NOT EXISTS 需要通过存储过程或直接创建（忽略已存在错误）
CREATE UNIQUE INDEX idx_sys_dict_data_type_value ON sys_dict_data(dict_type, value);

-- ============================================================
-- 字典初始化数据：操作类型
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('操作类型', 'sys_oper_type', 1, '操作日志操作类型列表', NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '新增', '1', 1, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '修改', '2', 2, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '删除', '3', 3, 'danger', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '导出', '4', 4, 'warning', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '导入', '5', 5, 'processing', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '其他', '6', 6, 'default', 1, NOW(), NOW());

-- ============================================================
-- 字典初始化数据：操作状态
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('操作状态', 'sys_oper_status', 1, '操作日志操作状态列表', NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_status', '成功', '0', 1, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_status', '失败', '1', 2, 'danger', 1, NOW(), NOW());
