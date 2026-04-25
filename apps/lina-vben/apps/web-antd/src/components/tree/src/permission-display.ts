import { preferences } from '@vben/preferences';

const dynamicRoutePermissionPrefix = '动态路由权限:';

const zhSegmentGlossary: Record<string, string> = {
  action: '动作',
  add: '新增',
  audit: '审计',
  auth: '授权',
  backend: '后端',
  config: '配置',
  create: '创建',
  delete: '删除',
  demo: '示例',
  dept: '部门',
  detail: '详情',
  dict: '字典',
  disable: '停用',
  edit: '编辑',
  enable: '启用',
  export: '导出',
  file: '文件',
  health: '健康',
  import: '导入',
  inspect: '检查',
  install: '安装',
  list: '列表',
  log: '日志',
  menu: '菜单',
  notice: '公告',
  plugin: '插件',
  post: '岗位',
  query: '查询',
  record: '记录',
  remove: '删除',
  resource: '资源',
  review: '审核',
  role: '角色',
  run: '执行',
  status: '状态',
  summary: '摘要',
  sync: '同步',
  uninstall: '卸载',
  update: '修改',
  user: '用户',
  view: '查看',
};

const enSegmentGlossary: Record<string, string> = {
  add: 'Add',
  audit: 'Audit',
  auth: 'Authorization',
  backend: 'Backend',
  config: 'Configuration',
  create: 'Create',
  delete: 'Delete',
  demo: 'Demo',
  dept: 'Department',
  detail: 'Details',
  dict: 'Dictionary',
  disable: 'Disable',
  edit: 'Edit',
  enable: 'Enable',
  export: 'Export',
  file: 'File',
  health: 'Health',
  import: 'Import',
  inspect: 'Inspect',
  install: 'Install',
  list: 'List',
  log: 'Log',
  menu: 'Menu',
  notice: 'Notice',
  plugin: 'Plugin',
  post: 'Post',
  query: 'Query',
  record: 'Record',
  remove: 'Remove',
  review: 'Review',
  role: 'Role',
  run: 'Run',
  status: 'Status',
  summary: 'Summary',
  sync: 'Sync',
  uninstall: 'Uninstall',
  update: 'Update',
  user: 'User',
  view: 'View',
};

function getActiveLocale() {
  if (typeof document !== 'undefined' && document.documentElement.lang) {
    return document.documentElement.lang;
  }
  return preferences.app.locale;
}

function isEnglishLocale() {
  return getActiveLocale().startsWith('en');
}

function toTitleCase(rawValue: string) {
  if (!rawValue) {
    return '';
  }

  return rawValue
    .split(/\s+/)
    .filter(Boolean)
    .map((token) => token.slice(0, 1).toUpperCase() + token.slice(1))
    .join(' ');
}

function humanizePermissionSegment(rawValue: string) {
  const normalized = String(rawValue || '').trim();
  if (!normalized) {
    return '';
  }

  const glossary = isEnglishLocale() ? enSegmentGlossary : zhSegmentGlossary;
  const tokens = normalized.split(/[-_/]+/).filter(Boolean);
  if (tokens.length === 0) {
    return normalized;
  }

  const transformed = tokens.map((token) => {
    const lowerToken = token.toLowerCase();
    const mapped = glossary[lowerToken];
    if (mapped) {
      return mapped;
    }
    return isEnglishLocale() ? toTitleCase(token) : token;
  });

  return isEnglishLocale() ? transformed.join(' ') : transformed.join('');
}

function extractDynamicRoutePermission(rawValue: string) {
  const normalized = String(rawValue || '').trim();
  if (!normalized) {
    return '';
  }

  if (normalized.startsWith(dynamicRoutePermissionPrefix)) {
    return normalized.slice(dynamicRoutePermissionPrefix.length).trim();
  }

  const parts = normalized.split(':');
  if (
    parts.length === 3 &&
    parts.every((part) => part.trim() !== '') &&
    /^plugin[-_]/.test(parts[0].trim())
  ) {
    return normalized;
  }

  return '';
}

export function formatMenuPermissionLabel(rawValue: string | null | undefined) {
  const normalized = String(rawValue || '').trim();
  if (!normalized) {
    return '';
  }

  const permission = extractDynamicRoutePermission(normalized);
  if (!permission) {
    return normalized;
  }

  const parts = permission.split(':');
  if (parts.length !== 3) {
    return isEnglishLocale() ? 'Dynamic Route Permission' : '动态路由权限';
  }

  const resourceLabel = humanizePermissionSegment(parts[1]);
  const actionLabel = humanizePermissionSegment(parts[2]);

  if (isEnglishLocale()) {
    return `Dynamic Route Permission (resource: ${resourceLabel}, action: ${actionLabel})`;
  }

  return `动态路由权限（资源：${resourceLabel}，动作：${actionLabel}）`;
}
