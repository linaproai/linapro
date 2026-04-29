-- Mock data: file metadata records for file-management screens.
-- 模拟数据：文件管理页面使用的文件元数据记录。

INSERT IGNORE INTO sys_file (name, original, suffix, scene, size, hash, url, path, engine, created_by, created_at, updated_at)
SELECT
    'mock/avatar-admin.webp',
    'admin-avatar.webp',
    'webp',
    'avatar',
    18432,
    'a3f2d9c1b5e6478091234567890abcdef1234567890abcdef1234567890abcd',
    '/uploads/mock/avatar-admin.webp',
    'mock/avatar-admin.webp',
    'local',
    u.id,
    '2026-04-20 08:30:00',
    '2026-04-20 08:30:00'
FROM sys_user u
WHERE u.username = 'admin'
  AND NOT EXISTS (
      SELECT 1
      FROM sys_file
      WHERE hash = 'a3f2d9c1b5e6478091234567890abcdef1234567890abcdef1234567890abcd'
  );

INSERT IGNORE INTO sys_file (name, original, suffix, scene, size, hash, url, path, engine, created_by, created_at, updated_at)
SELECT
    'mock/notice-maintenance.pdf',
    '系统维护说明.pdf',
    'pdf',
    'notice_attachment',
    245760,
    'b4e3d0c2a6f758901234567890abcdef1234567890abcdef1234567890abcde',
    '/uploads/mock/notice-maintenance.pdf',
    'mock/notice-maintenance.pdf',
    'local',
    u.id,
    '2026-04-20 09:15:00',
    '2026-04-20 09:15:00'
FROM sys_user u
WHERE u.username = 'admin'
  AND NOT EXISTS (
      SELECT 1
      FROM sys_file
      WHERE hash = 'b4e3d0c2a6f758901234567890abcdef1234567890abcdef1234567890abcde'
  );

INSERT IGNORE INTO sys_file (name, original, suffix, scene, size, hash, url, path, engine, created_by, created_at, updated_at)
SELECT
    'mock/report-screenshot.png',
    '巡检截图.png',
    'png',
    'other',
    98304,
    'c5f4e1d3b7a86901234567890abcdef1234567890abcdef1234567890abcdef',
    '/uploads/mock/report-screenshot.png',
    'mock/report-screenshot.png',
    'local',
    u.id,
    '2026-04-21 15:45:00',
    '2026-04-21 15:45:00'
FROM sys_user u
WHERE u.username = 'user002'
  AND NOT EXISTS (
      SELECT 1
      FROM sys_file
      WHERE hash = 'c5f4e1d3b7a86901234567890abcdef1234567890abcdef1234567890abcdef'
  );
