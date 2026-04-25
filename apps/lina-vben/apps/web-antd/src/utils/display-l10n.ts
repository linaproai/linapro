import { preferences } from '@vben/preferences';

function getActiveLocale() {
  if (typeof document !== 'undefined' && document.documentElement.lang) {
    return document.documentElement.lang;
  }
  return preferences.app.locale;
}

export function isEnglishLocale() {
  return getActiveLocale().startsWith('en');
}

function localizeByMap(
  rawValue: string | null | undefined,
  mapping: Record<string, string>,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }
  return mapping[rawValue] || rawValue;
}

function replaceByPatterns(
  rawValue: string | null | undefined,
  patterns: Array<[RegExp, (...args: string[]) => string]>,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }

  for (const [pattern, formatter] of patterns) {
    const matched = rawValue.match(pattern);
    if (matched) {
      return formatter(...matched.slice(1));
    }
  }
  return rawValue;
}

export function localizeSeedDeptName(
  code: string | null | undefined,
  rawValue: string | null | undefined,
) {
  const mappedByCode = localizeByMap(rawValue, {
    'Lina科技': 'LinaTech',
    '研发部门': 'R&D Department',
    '市场部门': 'Marketing Department',
    '测试部门': 'QA Department',
    '财务部门': 'Finance Department',
    '运维部门': 'Operations Department',
    '未分配部门': 'Unassigned Department',
  });

  if (!mappedByCode || !isEnglishLocale()) {
    return mappedByCode || '';
  }

  const byStableCode: Record<string, string> = {
    lina: 'LinaTech',
    dev: 'R&D Department',
    market: 'Marketing Department',
    qa: 'QA Department',
    finance: 'Finance Department',
    ops: 'Operations Department',
  };
  if (code && byStableCode[code]) {
    return byStableCode[code];
  }

  return replaceByPatterns(mappedByCode, [
    [/^编码测试部_(.+)$/, (suffix) => `Code Test Dept_${suffix}`],
    [/^编码测试二_(.+)$/, (suffix) => `Code Test Dept II_${suffix}`],
  ]);
}

export function localizeSeedDeptLabel(rawValue: string | null | undefined) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }

  const counted = rawValue.match(/^(.*?)(\(\d+\))$/);
  if (!counted?.[1] || !counted[2]) {
    return localizeSeedDeptName('', rawValue);
  }
  return `${localizeSeedDeptName('', counted[1])}${counted[2]}`;
}

export function localizeSeedPostName(
  code: string | null | undefined,
  rawValue: string | null | undefined,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }

  const byStableCode: Record<string, string> = {
    CEO: 'Chief Executive Officer',
    CTO: 'Chief Technology Officer',
    PM: 'Project Manager',
    DEV: 'Software Engineer',
    QA: 'QA Engineer',
  };
  if (code && byStableCode[code]) {
    return byStableCode[code];
  }

  return localizeByMap(rawValue, {
    '总经理': 'Chief Executive Officer',
    '技术总监': 'Chief Technology Officer',
    '项目经理': 'Project Manager',
    '开发工程师': 'Software Engineer',
    '测试工程师': 'QA Engineer',
  });
}

export function localizeSeedRoleName(
  roleKey: string | null | undefined,
  rawValue: string | null | undefined,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }

  const byStableKey: Record<string, string> = {
    admin: 'Administrator',
    user: 'Standard User',
  };
  if (roleKey && byStableKey[roleKey]) {
    return byStableKey[roleKey];
  }

  return localizeByMap(rawValue, {
    '超级管理员': 'Administrator',
    '普通用户': 'Standard User',
  });
}

export function localizeSeedRoleRemark(
  roleKey: string | null | undefined,
  rawValue: string | null | undefined,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }

  const byStableKey: Record<string, string> = {
    admin: 'Administrator with full permissions.',
    user: 'Standard user with access to personal data only.',
  };
  if (roleKey && byStableKey[roleKey]) {
    return byStableKey[roleKey];
  }

  return localizeByMap(rawValue, {
    '超级管理员，拥有所有权限': 'Administrator with full permissions.',
    '普通用户，仅查看本人数据': 'Standard user with access to personal data only.',
    'E2E测试角色': 'E2E Test Role',
  });
}

export function localizeSeedRoleNames(rawValue: string | null | undefined) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }
  return rawValue
    .split(/[,，]/)
    .map((item) => localizeSeedRoleName('', item.trim()))
    .join(', ');
}

export function localizeSeedConfigName(
  configKey: string | null | undefined,
  rawValue: string | null | undefined,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }

  const byKey: Record<string, string> = {
    'demo.support.email': 'Demo - Support Email',
    'demo.notice.banner': 'Demo - Homepage Banner Copy',
    'cron.log.retention': 'Scheduled Jobs - Execution Log Retention Policy',
    'cron.shell.enabled': 'Scheduled Jobs - Global Shell Mode Switch',
  };
  if (configKey && byKey[configKey]) {
    return byKey[configKey];
  }

  return replaceByPatterns(rawValue, [
    [/^测试参数_(.+)$/, (suffix) => `Test Parameter_${suffix}`],
    [/^重复测试配置$/, () => 'Duplicate Test Config'],
    [/^覆盖测试配置$/, () => 'Override Test Config'],
    [/^已存在配置$/, () => 'Existing Config'],
    [/^待覆盖配置$/, () => 'Pending Override Config'],
    [/^已更新配置$/, () => 'Updated Config'],
  ]);
}

export function localizeSeedConfigRemark(
  configKey: string | null | undefined,
  rawValue: string | null | undefined,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }

  const byKey: Record<string, string> = {
    'demo.support.email':
      'Used to demonstrate custom parameter management. It is not consumed directly by the host runtime.',
    'demo.notice.banner':
      'Used to demonstrate custom parameter management. It is not consumed directly by the host runtime.',
    'cron.log.retention':
      'Controls the default retention policy for scheduled-job execution logs. Use JSON: {"mode":"days|count|none","value":N}.',
    'cron.shell.enabled':
      'Controls whether Shell-type jobs can be created, updated, triggered, and stopped. Supported values: true, false.',
  };
  if (configKey && byKey[configKey]) {
    return byKey[configKey];
  }

  return replaceByPatterns(rawValue, [
    [/^测试备注(.*)$/, (suffix) => `Test Remark${suffix}`],
    [/^自动化测试创建$/, () => 'Created by automated test'],
    [/^覆盖更新$/, () => 'Updated by override test'],
  ]);
}

export function localizeSeedConfigValue(
  configKey: string | null | undefined,
  rawValue: string | null | undefined,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }

  const byKey: Record<string, string> = {
    'demo.notice.banner': 'Welcome to LinaPro',
  };
  if (configKey && byKey[configKey]) {
    return byKey[configKey];
  }

  return rawValue;
}

export function localizeSeedDictTypeName(
  dictType: string | null | undefined,
  rawValue: string | null | undefined,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }

  const byType: Record<string, string> = {
    cron_job_status: 'Scheduled Job Status',
    cron_job_task_type: 'Scheduled Job Type',
    cron_job_scope: 'Scheduled Job Scope',
    cron_job_concurrency: 'Scheduled Job Concurrency',
    cron_job_trigger: 'Scheduled Job Trigger',
    cron_job_log_status: 'Scheduled Job Log Status',
    cron_log_retention_mode: 'Scheduled Job Log Retention Mode',
  };
  if (dictType && byType[dictType]) {
    return byType[dictType];
  }

  return rawValue;
}

export function localizeSeedDictTypeRemark(
  dictType: string | null | undefined,
  rawValue: string | null | undefined,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }

  const byType: Record<string, string> = {
    cron_job_status: 'Scheduled job status options.',
    cron_job_task_type: 'Scheduled job type options.',
    cron_job_scope: 'Scheduled job scope options.',
    cron_job_concurrency: 'Scheduled job concurrency options.',
    cron_job_trigger: 'Scheduled job trigger options.',
    cron_job_log_status: 'Scheduled job execution log status options.',
    cron_log_retention_mode: 'Scheduled job log retention mode options.',
  };
  if (dictType && byType[dictType]) {
    return byType[dictType];
  }

  return rawValue;
}

export function localizeSeedJobGroupName(
  groupCode: string | null | undefined,
  rawValue: string | null | undefined,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }
  if (groupCode === 'default') {
    return 'Default Group';
  }
  return rawValue;
}

export function localizeSeedJobGroupRemark(
  groupCode: string | null | undefined,
  rawValue: string | null | undefined,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }
  if (groupCode === 'default') {
    return 'The system default job group. Jobs are moved here when other groups are deleted.';
  }
  return rawValue;
}

const seedJobByHandlerRef: Record<string, { description: string; name: string }> = {
  'host:cleanup-job-logs': {
    description:
      'Cleans up scheduled-job execution logs according to global and job-level retention policies.',
    name: 'Job Log Cleanup',
  },
  'host:session-cleanup': {
    description:
      'Cleans up inactive online sessions in the host according to the session-timeout policy.',
    name: 'Online Session Cleanup',
  },
  'plugin:monitor-server/cron:服务监控采集': {
    description: 'Built-in scheduled job registered by the monitor-server plugin.',
    name: 'Server Monitor Collection',
  },
  'plugin:monitor-server/cron:服务监控清理': {
    description: 'Built-in scheduled job registered by the monitor-server plugin.',
    name: 'Server Monitor Cleanup',
  },
  'plugin:plugin-demo-source/cron:源码插件回显巡检': {
    description: 'Built-in scheduled job registered by the plugin-demo-source plugin.',
    name: 'Source Plugin Echo Inspection',
  },
  'plugin:plugin-demo-dynamic/cron:heartbeat': {
    description:
      'Runs the dynamic plugin built-in job through the Wasm bridge and accumulates heartbeat executions.',
    name: 'Dynamic Plugin Heartbeat',
  },
};

export function localizeSeedJobName(
  handlerRef: string | null | undefined,
  rawValue: string | null | undefined,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }
  return seedJobByHandlerRef[handlerRef || '']?.name || rawValue;
}

export function localizeSeedJobDescription(
  handlerRef: string | null | undefined,
  rawValue: string | null | undefined,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }
  return seedJobByHandlerRef[handlerRef || '']?.description || rawValue;
}

export function localizeSeedNoticeTitle(
  noticeID: number | string | null | undefined,
  rawValue: string | null | undefined,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }

  const byID: Record<string, string> = {
    '1': 'System Upgrade Notice',
    '2': 'Notice on Standard System Usage',
    '3': 'New Feature Launch Preview',
  };
  return byID[String(noticeID ?? '')] || rawValue;
}

export function localizeSeedNoticeRemark(rawValue: string | null | undefined) {
  return localizeByMap(rawValue, {
    草稿状态: 'Draft status',
  });
}

export function localizeSeedNoticeContent(
  noticeID: number | string | null | undefined,
  rawValue: string | null | undefined,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }

  const byID: Record<string, string> = {
    '1':
      '<p>The system will undergo maintenance this Saturday from 2:00 AM to 4:00 AM and will be unavailable during that window.</p><p><strong>Upgrade scope:</strong></p><ul><li>Performance improvements</li><li>Security patch updates</li><li>New feature rollout</li></ul>',
    '2':
      '<p>To keep the system secure and stable, please follow these guidelines:</p><ol><li>Update your password regularly and keep it at least 8 characters long</li><li>Do not share your account credentials with others</li><li>Lock your screen when you leave your desk</li></ol><p>Thank you for your cooperation.</p>',
    '3':
      '<p>The following new capabilities are coming soon:</p><ul><li>Notice management</li><li>Message center</li><li>Rich-text editor</li></ul><p>Stay tuned.</p>',
  };
  return byID[String(noticeID ?? '')] || rawValue;
}

export function localizeSeedOperLogTitle(rawValue: string | null | undefined) {
  return localizeByMap(rawValue, {
    动态插件示例: 'Dynamic Plugin Demo',
    插件管理: 'Plugin Management',
    用户管理: 'User Management',
    角色管理: 'Role Management',
  });
}

export function localizeSeedOperLogSummary(rawValue: string | null | undefined) {
  return localizeByMap(rawValue, {
    分页查询动态插件示例记录: 'Paged query for dynamic plugin demo records',
    同步源码插件: 'Synchronize source plugins',
    安装插件: 'Install Plugin',
    启用插件: 'Enable Plugin',
    禁用插件: 'Disable Plugin',
    卸载插件: 'Uninstall Plugin',
    创建用户: 'Create User',
    创建角色: 'Create Role',
  });
}

export function localizeSeedLoginLogMessage(rawValue: string | null | undefined) {
  return localizeByMap(rawValue, {
    登录成功: 'Login successful',
  });
}

export function localizeDynamicPluginSeedRecordTitle(
  recordID: string | null | undefined,
  rawValue: string | null | undefined,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }
  if (recordID === 'plugin-demo-dynamic-seed-record') {
    return 'Dynamic Plugin SQL Demo Record';
  }
  return rawValue;
}

export function localizeDynamicPluginSeedRecordContent(
  recordID: string | null | undefined,
  rawValue: string | null | undefined,
) {
  if (!rawValue || !isEnglishLocale()) {
    return rawValue || '';
  }
  if (recordID === 'plugin-demo-dynamic-seed-record') {
    return 'This record is seeded by the plugin-demo-dynamic install SQL and demonstrates CRUD operations against the data table created during plugin installation.';
  }
  return rawValue;
}

export function localizeSeedPluginType(rawValue: string | null | undefined) {
  return localizeByMap(rawValue, {
    源码插件: 'Source Plugin',
    动态插件: 'Dynamic Plugin',
  });
}

export function localizeSeedPluginStatus(rawValue: string | null | undefined) {
  return localizeByMap(rawValue, {
    启用: 'Enabled',
    禁用: 'Disabled',
  });
}

export function localizeSeedMenuName(rawValue: string | null | undefined) {
  const mapped = localizeByMap(rawValue, {
    分析页: 'Analytics',
    分组修改: 'Update',
    分组删除: 'Delete',
    分组新增: 'Create',
    分组管理: 'Groups',
    参数修改: 'Update',
    参数删除: 'Delete',
    参数导入: 'Import',
    参数导出: 'Export',
    参数新增: 'Create',
    参数设置: 'Parameters',
    字典修改: 'Update',
    字典删除: 'Delete',
    字典导入: 'Import',
    字典导出: 'Export',
    字典新增: 'Create',
    字典管理: 'Dictionaries',
    字典查询: 'Query',
    岗位修改: 'Update',
    岗位删除: 'Delete',
    岗位导出: 'Export',
    岗位新增: 'Create',
    岗位查询: 'Query',
    岗位管理: 'Positions',
    工作台: 'Workspace',
    开发中心: 'Dev Tools',
    强制退出: 'Force Logout',
    任务修改: 'Update',
    任务删除: 'Delete',
    任务启停: 'Toggle Status',
    任务新增: 'Create',
    任务查询: 'Query',
    任务管理: 'Jobs',
    任务调度: 'Scheduler',
    任务重置: 'Reset Counter',
    任务日志清理: 'Job Log Cleanup',
    仅本人数据权限: 'Self Only',
    角色修改: 'Update',
    角色删除: 'Delete',
    角色新增: 'Create',
    角色授权用户: 'Role Members',
    角色管理: 'Roles',
    用户删除: 'Delete',
    用户导入: 'Import',
    用户导出: 'Export',
    用户新增: 'Create',
    用户查询: 'Query',
    用户管理: 'Users',
    用户修改: 'Update',
    文件上传: 'Upload',
    文件删除: 'Delete',
    文件下载: 'Download',
    文件查询: 'Query',
    文件管理: 'Files',
    插件上传: 'Upload',
    插件同步: 'Synchronize',
    插件管理: 'Plugins',
    扩展中心: 'Extensions',
    执行日志: 'Run Logs',
    日志删除: 'Delete',
    日志导出: 'Export',
    日志查询: 'Query',
    日志清空: 'Clear',
    日志终止: 'Terminate',
    服务监控: 'Server Metrics',
    权限管理: 'Access',
    登录日志: 'Login History',
    立即执行: 'Run Now',
    接口文档: 'API Docs',
    操作日志: 'Audit Logs',
    插件详情: 'Plugin Details',
    插件安装: 'Install',
    插件禁用: 'Disable',
    插件启用: 'Enable',
    插件查询: 'Query',
    插件卸载: 'Uninstall',
    数据权限: 'Data Scope',
    新增通知公告: 'Create',
    普通用户: 'Standard User',
    权限字符: 'Permission Key',
    消息列表: 'Messages',
    源码插件: 'Source Plugin',
    状态修改: 'Update Status',
    系统信息: 'System Info',
    系统监控: 'Monitoring',
    系统设置: 'Settings',
    组织管理: 'Organization',
    编辑通知公告: 'Update',
    联机用户: 'Sessions',
    菜单删除: 'Delete',
    菜单新增: 'Create',
    菜单管理: 'Menus',
    菜单修改: 'Update',
    菜单查询: 'Query',
    角色授权: 'Authorize',
    通知公告: 'Notices',
    部门删除: 'Delete',
    部门新增: 'Create',
    部门查询: 'Query',
    部门管理: 'Departments',
    部门修改: 'Update',
    配置导入: 'Import',
    配置导出: 'Export',
    重置密码: 'Reset Password',
    在线用户: 'Sessions',
  });

  if (!mapped || !isEnglishLocale()) {
    return mapped || '';
  }

  return replaceByPatterns(mapped, [
    [/^动态路由权限:(.+)$/, (permission) => `Dynamic Route Permission:${permission}`],
  ]);
}
