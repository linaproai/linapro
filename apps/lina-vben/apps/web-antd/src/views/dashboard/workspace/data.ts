import type {
  WorkbenchProjectItem,
  WorkbenchTodoItem,
  WorkbenchTrendItem,
} from '@vben/common-ui';

export interface WorkspaceFocusItem {
  description: string;
  key: string;
  title: string;
  value: string;
}

export interface WorkspaceQuickActionItem {
  badge: string;
  description: string;
  icon: string;
  key: string;
  title: string;
  url: string;
}

export const workspaceFocusItems: WorkspaceFocusItem[] = [
  {
    description: '核心宿主服务、治理组件与统一扩展边界持续稳定运行。',
    key: 'host',
    title: '核心宿主',
    value: '18 项',
  },
  {
    description: '默认管理工作台入口均已收敛到当前实际可达的管理页面。',
    key: 'workspace',
    title: '工作台入口',
    value: '12 个',
  },
  {
    description: '插件安装、启用、卸载与授权审查均纳入统一治理流程。',
    key: 'plugin',
    title: '插件治理',
    value: '9 项',
  },
  {
    description: 'OpenSpec、任务清单、反馈和回归验证在同一工作流内协同。',
    key: 'delivery',
    title: '协作流程',
    value: 'OpenSpec',
  },
];

export const workspaceQuickActionItems: WorkspaceQuickActionItem[] = [
  {
    badge: '推荐',
    description: '统一处理插件安装、启用、卸载与授权确认。',
    icon: 'lucide:puzzle',
    key: 'plugin-management',
    title: '插件治理',
    url: '/system/plugin',
  },
  {
    badge: '高频',
    description: '维护管理员、角色与授权边界。',
    icon: 'lucide:users',
    key: 'user-management',
    title: '用户权限',
    url: '/system/user',
  },
  {
    badge: '组织',
    description: '查看部门树、岗位分配与组织结构。',
    icon: 'lucide:network',
    key: 'dept-management',
    title: '组织架构',
    url: '/system/dept',
  },
  {
    badge: '配置',
    description: '管理系统参数、字典与工作台运行配置。',
    icon: 'lucide:sliders-horizontal',
    key: 'config-management',
    title: '参数配置',
    url: '/system/config',
  },
  {
    badge: '资产',
    description: '统一查看附件、图片与上传记录。',
    icon: 'lucide:folder-open',
    key: 'file-management',
    title: '文件资产',
    url: '/system/file',
  },
  {
    badge: '文档',
    description: '快速查看宿主接口文档与版本信息。',
    icon: 'lucide:file-code',
    key: 'api-docs',
    title: '系统接口',
    url: '/about/api-docs',
  },
];

export const workspaceProjectItems: WorkbenchProjectItem[] = [
  {
    color: '#1677ff',
    content: '查看宿主服务、前后端组件与当前项目定位信息。',
    date: '持续可用',
    group: '宿主治理',
    icon: 'lucide:server',
    title: '核心宿主服务',
    url: '/about/system-info',
  },
  {
    color: '#13c2c2',
    content: '围绕默认管理工作台入口、路由与布局交互进行巡检。',
    date: '本周重点',
    group: '工作台体验',
    icon: 'lucide:layout-dashboard',
    title: '默认管理工作台',
    url: '/dashboard/analytics',
  },
  {
    color: '#52c41a',
    content: '安装、启用和审查插件 release 的授权与资源边界。',
    date: '需持续关注',
    group: '插件扩展',
    icon: 'lucide:plug',
    title: '插件交付治理',
    url: '/system/plugin',
  },
  {
    color: '#faad14',
    content: '统一查看用户、角色、菜单与部门的配置状态。',
    date: '高频入口',
    group: '系统治理',
    icon: 'lucide:shield-check',
    title: '系统基础配置',
    url: '/system/role',
  },
  {
    color: '#722ed1',
    content: '通过 OpenSpec 追踪提案、任务和反馈闭环。',
    date: '研发协同',
    group: '流程协作',
    icon: 'lucide:clipboard-list',
    title: 'OpenSpec 工作流',
    url: 'https://github.com/gqcn/lina',
  },
  {
    color: '#eb2f96',
    content: '查看前后端能力边界与项目说明，对齐框架定位。',
    date: '文档入口',
    group: '项目说明',
    icon: 'lucide:book-open',
    title: '项目资料',
    url: '/about/system-info',
  },
];

export const workspaceTodoItems: WorkbenchTodoItem[] = [
  {
    completed: false,
    content: '核对动态插件当前 release 的授权快照与宿主服务声明是否一致。',
    date: '今天 18:00 前',
    title: '复核插件发布授权快照',
  },
  {
    completed: true,
    content: '完成默认管理工作台文案与入口定位收敛，保持 LinaPro 项目叙事统一。',
    date: '今天 12:30',
    title: '统一项目定位文案',
  },
  {
    completed: false,
    content: '执行系统管理、监控和插件模块的 E2E 回归，确认关键链路可用。',
    date: '今天 20:00 前',
    title: '执行关键页面回归',
  },
  {
    completed: false,
    content: '检查文件管理与参数导入模板，确保管理员导入路径可观测。',
    date: '明天 10:00 前',
    title: '复核导入与文件资产',
  },
  {
    completed: false,
    content: '回看用户、角色和菜单权限调整，确认默认工作台入口没有漏项。',
    date: '本周内',
    title: '审查权限治理边界',
  },
];

export const workspaceTrendItems: WorkbenchTrendItem[] = [
  {
    avatar: 'svg:avatar-1',
    content: '在 <a>插件治理</a> 中完成了一次动态插件安装审查与授权确认。',
    date: '刚刚',
    title: '管理员',
  },
  {
    avatar: 'svg:avatar-2',
    content: '更新了 <a>系统信息</a> 页的项目定位与组件说明。',
    date: '1 小时前',
    title: 'LinaPro Bot',
  },
  {
    avatar: 'svg:avatar-3',
    content: '执行了 <a>系统管理</a> 模块的关键链路回归并记录结果。',
    date: '今天 14:00',
    title: 'QA',
  },
  {
    avatar: 'svg:avatar-4',
    content: '整理了 <a>OpenSpec</a> 反馈任务，准备进入下一轮实现。',
    date: '今天 10:30',
    title: '研发协作',
  },
  {
    avatar: 'svg:avatar-1',
    content: '在 <a>默认管理工作台</a> 中完成快捷入口与仪表盘语义收敛。',
    date: '昨天',
    title: '前端',
  },
  {
    avatar: 'svg:avatar-2',
    content: '同步检查了 <a>系统接口</a> 页，确认接口文档入口可正常访问。',
    date: '本周',
    title: '平台',
  },
];

export const workspaceTrafficSourceItems = [
  { name: '系统管理', value: 42 },
  { name: '插件治理', value: 28 },
  { name: '系统监控', value: 18 },
  { name: '系统信息', value: 12 },
];
