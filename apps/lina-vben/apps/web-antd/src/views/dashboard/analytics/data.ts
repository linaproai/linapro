export type AnalyticsRangeKey = 'today' | 'week' | 'month';

export type AnalyticsTone = 'cyan' | 'emerald' | 'amber';

export type AnalyticsOverviewMetricKey =
  | 'hostCalls'
  | 'pluginActivity'
  | 'workspaceVisits'
  | 'regressionRuns';

export interface AnalyticsOverviewMetric {
  key: AnalyticsOverviewMetricKey;
  title: string;
  totalTitle: string;
  totalValue: number;
  value: number;
}

export interface AnalyticsInsightItem {
  title: string;
  value: string;
  description: string;
  tone: AnalyticsTone;
}

export interface AnalyticsLineSeries {
  color: string;
  data: number[];
  name: string;
}

export interface AnalyticsPieItem {
  name: string;
  value: number;
}

export interface AnalyticsRadarItem {
  name: string;
  value: number[];
}

export interface AnalyticsRangeDataset {
  cadenceAxis: string[];
  cadenceLabel: string;
  cadenceSeries: number[];
  insights: AnalyticsInsightItem[];
  overview: AnalyticsOverviewMetric[];
  radarIndicators: string[];
  radarSeries: AnalyticsRadarItem[];
  salesItems: AnalyticsPieItem[];
  sourceItems: AnalyticsPieItem[];
  summary: string;
  touchpointLabel: string;
  trendAxis: string[];
  trendSeries: AnalyticsLineSeries[];
  updatedAt: string;
}

export const analyticsRangeOptions: Array<{
  label: string;
  value: AnalyticsRangeKey;
}> = [
  { label: '今日', value: 'today' },
  { label: '最近 7 天', value: 'week' },
  { label: '最近 30 天', value: 'month' },
];

export const analyticsRangeData: Record<AnalyticsRangeKey, AnalyticsRangeDataset> = {
  today: {
    cadenceAxis: ['09:00', '10:00', '11:00', '12:00', '13:00', '14:00', '15:00', '16:00'],
    cadenceLabel: '今日任务节奏',
    cadenceSeries: [4, 6, 8, 5, 9, 7, 6, 4],
    insights: [
      {
        description: '1 个动态插件待安装授权，1 个已安装 release 待复核',
        title: '安装审查',
        tone: 'amber',
        value: '2 项',
      },
      {
        description: '核心系统页与插件管理回归链路已完成全量自检',
        title: '自动化回归',
        tone: 'emerald',
        value: '36 / 36',
      },
      {
        description: '宿主 API、插件页与消息中心形成稳定访问峰值',
        title: '访问热点',
        tone: 'cyan',
        value: '15:00',
      },
    ],
    overview: [
      { key: 'hostCalls', title: '宿主调用量', totalTitle: '累计调用量', totalValue: 18_620, value: 824 },
      { key: 'pluginActivity', title: '插件活跃数', totalTitle: '已安装插件', totalValue: 18, value: 14 },
      { key: 'workspaceVisits', title: '工作台访问', totalTitle: '累计访问量', totalValue: 52_800, value: 1280 },
      { key: 'regressionRuns', title: '回归执行数', totalTitle: '累计执行数', totalValue: 420, value: 36 },
    ],
    radarIndicators: ['核心 API', '插件运行时', '权限治理', '文件资产', '接口文档', '消息中心'],
    radarSeries: [
      { name: '当前', value: [88, 72, 84, 76, 68, 64] },
      { name: '基线', value: [70, 60, 66, 58, 54, 48] },
    ],
    salesItems: [
      { name: '核心宿主治理', value: 38 },
      { name: '插件迭代交付', value: 27 },
      { name: '管理工作台优化', value: 21 },
      { name: '自动化回归', value: 14 },
    ],
    sourceItems: [
      { name: '系统管理', value: 460 },
      { name: '插件治理', value: 320 },
      { name: '监控诊断', value: 280 },
      { name: '系统信息', value: 220 },
    ],
    summary: '面向近 24 小时的宿主工作区运行概览，重点关注插件治理与回归节奏。',
    touchpointLabel: '今日触点覆盖',
    trendAxis: ['09:00', '10:00', '11:00', '12:00', '13:00', '14:00', '15:00', '16:00'],
    trendSeries: [
      { color: '#1677ff', data: [110, 126, 142, 132, 168, 180, 194, 176], name: '宿主调用' },
      { color: '#13c2c2', data: [62, 74, 82, 78, 96, 118, 126, 112], name: '工作台访问' },
    ],
    updatedAt: '更新于今日 16:00',
  },
  week: {
    cadenceAxis: ['周一', '周二', '周三', '周四', '周五', '周六', '周日'],
    cadenceLabel: '近 7 天交付节奏',
    cadenceSeries: [5, 8, 6, 9, 7, 4, 3],
    insights: [
      {
        description: '默认工作台与插件管理页完成 3 轮交互收敛',
        title: '页面优化',
        tone: 'cyan',
        value: '3 次',
      },
      {
        description: '登录、系统信息与插件生命周期链路持续通过回归',
        title: '关键路径',
        tone: 'emerald',
        value: '100%',
      },
      {
        description: '高峰时段集中在角色权限和插件安装审查场景',
        title: '治理热点',
        tone: 'amber',
        value: '权限 + 插件',
      },
    ],
    overview: [
      { key: 'hostCalls', title: '宿主调用量', totalTitle: '累计调用量', totalValue: 18_620, value: 4_860 },
      { key: 'pluginActivity', title: '插件活跃数', totalTitle: '已安装插件', totalValue: 18, value: 15 },
      { key: 'workspaceVisits', title: '工作台访问', totalTitle: '累计访问量', totalValue: 52_800, value: 8_540 },
      { key: 'regressionRuns', title: '回归执行数', totalTitle: '累计执行数', totalValue: 420, value: 92 },
    ],
    radarIndicators: ['核心 API', '插件运行时', '权限治理', '文件资产', '接口文档', '消息中心'],
    radarSeries: [
      { name: '当前', value: [92, 80, 88, 82, 76, 70] },
      { name: '基线', value: [74, 66, 70, 62, 60, 56] },
    ],
    salesItems: [
      { name: '核心宿主治理', value: 34 },
      { name: '插件迭代交付', value: 31 },
      { name: '管理工作台优化', value: 19 },
      { name: '自动化回归', value: 16 },
    ],
    sourceItems: [
      { name: '系统管理', value: 1780 },
      { name: '插件治理', value: 1260 },
      { name: '监控诊断', value: 980 },
      { name: '系统信息', value: 620 },
    ],
    summary: '最近 7 天重点围绕权限治理、插件发布与关键页面体验优化展开。',
    touchpointLabel: '近 7 天触点覆盖',
    trendAxis: ['周一', '周二', '周三', '周四', '周五', '周六', '周日'],
    trendSeries: [
      { color: '#1677ff', data: [520, 680, 610, 760, 720, 540, 480], name: '宿主调用' },
      { color: '#13c2c2', data: [280, 360, 330, 420, 390, 300, 260], name: '工作台访问' },
    ],
    updatedAt: '更新于最近 7 天汇总',
  },
  month: {
    cadenceAxis: ['第 1 周', '第 2 周', '第 3 周', '第 4 周'],
    cadenceLabel: '近 30 天交付节奏',
    cadenceSeries: [18, 22, 20, 24],
    insights: [
      {
        description: '默认管理工作台、插件治理链路与宿主文案完成统一收敛',
        title: '迭代主题',
        tone: 'cyan',
        value: '工作台治理',
      },
      {
        description: 'E2E 用例已覆盖认证、系统、监控、插件与系统信息核心场景',
        title: '测试覆盖',
        tone: 'emerald',
        value: '79 条',
      },
      {
        description: '插件安装授权和默认仪表盘是最近一个月最常回看的管理入口',
        title: '关注入口',
        tone: 'amber',
        value: '插件 + 仪表盘',
      },
    ],
    overview: [
      { key: 'hostCalls', title: '宿主调用量', totalTitle: '累计调用量', totalValue: 18_620, value: 16_240 },
      { key: 'pluginActivity', title: '插件活跃数', totalTitle: '已安装插件', totalValue: 18, value: 16 },
      { key: 'workspaceVisits', title: '工作台访问', totalTitle: '累计访问量', totalValue: 52_800, value: 31_400 },
      { key: 'regressionRuns', title: '回归执行数', totalTitle: '累计执行数', totalValue: 420, value: 268 },
    ],
    radarIndicators: ['核心 API', '插件运行时', '权限治理', '文件资产', '接口文档', '消息中心'],
    radarSeries: [
      { name: '当前', value: [94, 86, 90, 88, 84, 78] },
      { name: '基线', value: [76, 70, 74, 68, 62, 60] },
    ],
    salesItems: [
      { name: '核心宿主治理', value: 32 },
      { name: '插件迭代交付', value: 29 },
      { name: '管理工作台优化', value: 23 },
      { name: '自动化回归', value: 16 },
    ],
    sourceItems: [
      { name: '系统管理', value: 6_280 },
      { name: '插件治理', value: 4_560 },
      { name: '监控诊断', value: 3_440 },
      { name: '系统信息', value: 1_820 },
    ],
    summary: '最近 30 天的视角更适合观察宿主治理、工作台入口使用和插件扩展协同趋势。',
    touchpointLabel: '近 30 天触点覆盖',
    trendAxis: ['第 1 周', '第 2 周', '第 3 周', '第 4 周'],
    trendSeries: [
      { color: '#1677ff', data: [3_420, 3_980, 4_120, 4_720], name: '宿主调用' },
      { color: '#13c2c2', data: [1_860, 2_140, 2_320, 2_860], name: '工作台访问' },
    ],
    updatedAt: '更新于最近 30 天汇总',
  },
};
