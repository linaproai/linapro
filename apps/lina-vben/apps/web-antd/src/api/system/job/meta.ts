import type { JobRecord } from './model';
import type { CSSProperties } from 'vue';

import { h } from 'vue';

interface CronLogRetentionLike {
  mode?: string;
  value?: number;
}

function renderJobHelpContent(content: string) {
  return () =>
    h(
      'div',
      {
        style: {
          lineHeight: '1.65',
          maxWidth: '320px',
          whiteSpace: 'pre-line',
        },
      },
      content,
    );
}

export const JOB_CRON_CODE_CONTAINER_STYLE: CSSProperties = {
  background: 'var(--ant-color-fill-tertiary, #f5f5f5)',
  border: '1px solid var(--ant-color-border-secondary, #f0f0f0)',
  borderRadius: '8px',
  display: 'inline-block',
  fontFamily:
    "ui-monospace, 'SFMono-Regular', SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace",
  lineHeight: '1.5',
  maxWidth: '100%',
  padding: '4px 8px',
  whiteSpace: 'nowrap',
};

export const JOB_PLUGIN_PAUSED_LABEL = '插件处理器不可用';

export const JOB_PLUGIN_PAUSED_TOOLTIP =
  '该任务依赖插件提供的处理器；当插件被禁用、卸载或处理器未注册时，系统会自动暂停任务，待插件恢复后可重新启用。';

export const JOB_STATUS_FILTER_OPTIONS = [
  { label: '启用', value: 'enabled' },
  { label: '停用', value: 'disabled' },
  { label: JOB_PLUGIN_PAUSED_LABEL, value: 'paused_by_plugin' },
];

export const JOB_SCOPE_OPTIONS = [
  { label: '仅主节点执行', value: 'master_only' },
  { label: '所有节点执行', value: 'all_node' },
];

export const JOB_CONCURRENCY_OPTIONS = [
  { label: '单例执行', value: 'singleton' },
  { label: '允许并行执行', value: 'parallel' },
];

export type JobSourceKind =
  | 'host_builtin'
  | 'plugin_builtin'
  | 'user_created';

export function parsePluginIdFromHandlerRef(handlerRef?: string) {
  const matched = (handlerRef || '').trim().match(/^plugin:([^/]+)\//);
  return matched?.[1] || '';
}

export function getJobSourceKind(record?: Partial<JobRecord> | null): JobSourceKind {
  if ((record?.isBuiltin || 0) !== 1) {
    return 'user_created';
  }
  return (record?.handlerRef || '').trim().startsWith('plugin:')
    ? 'plugin_builtin'
    : 'host_builtin';
}

export function getJobSourceLabel(source: JobSourceKind) {
  switch (source) {
    case 'host_builtin':
      return '宿主内置';
    case 'plugin_builtin':
      return '插件内置';
    case 'user_created':
    default:
      return '用户创建';
  }
}

export function getJobSourceColor(source: JobSourceKind) {
  switch (source) {
    case 'host_builtin':
      return 'geekblue';
    case 'plugin_builtin':
      return 'purple';
    case 'user_created':
    default:
      return 'gold';
  }
}

export const JOB_CRON_FIELD_HELP = renderJobHelpContent(
  '支持 5 段或 6 段 Cron。\n5 段按“分 时 日 月 周”解析，运行时会自动补 # 秒占位。\n6 段可显式配置秒位。',
);

export const JOB_TIMEOUT_FIELD_HELP = renderJobHelpContent(
  '任务实例单次运行允许的最长时长。\n超过该时长后，系统会尝试中断本次执行，并将执行日志标记为超时。\n建议按任务正常耗时预留一定余量，避免误判。',
);

export const JOB_MAX_EXECUTIONS_FIELD_HELP = renderJobHelpContent(
  '用于限制会计入执行计数的累计执行次数。\n达到上限后，任务会自动停用并记录停止原因。\n设置为 0 表示不限制执行次数；手动“立即执行”不会计入该次数。',
);

export const JOB_SCOPE_FIELD_HELP = renderJobHelpContent(
  '仅主节点执行：只有当前主节点会执行。\n所有节点执行：每个在线节点都会各自执行一次。',
);

export const JOB_CONCURRENCY_FIELD_HELP = renderJobHelpContent(
  '单例执行：本节点已有实例运行时，新触发会跳过。\n允许并行执行：本节点可同时运行多个实例，并受“最大并发”限制。',
);

export function renderJobCronExpression(
  expr?: string,
  attrs?: Record<string, any>,
) {
  const trimmedExpr = (expr || '').trim();
  if (!trimmedExpr) {
    return '-';
  }

  return h(
    'code',
    {
      ...attrs,
      style: JOB_CRON_CODE_CONTAINER_STYLE,
      title: trimmedExpr,
    },
    trimmedExpr,
  );
}

export function getJobScopeLabel(value: string) {
  const matched = JOB_SCOPE_OPTIONS.find((item) => item.value === value);
  return matched?.label || value || '-';
}

export function getJobConcurrencyLabel(value: string) {
  const matched = JOB_CONCURRENCY_OPTIONS.find((item) => item.value === value);
  return matched?.label || value || '-';
}

export function formatCronLogRetentionSummary(
  logRetention?: CronLogRetentionLike,
) {
  const mode = (logRetention?.mode || 'days').trim();
  const value = Number(logRetention?.value || 0);

  switch (mode) {
    case 'count': {
      return value > 0 ? `按条数保留最近 ${value} 条日志` : '按条数保留日志';
    }
    case 'none': {
      return '不自动清理日志';
    }
    case 'days':
    default: {
      return value > 0 ? `按天保留最近 ${value} 天日志` : '按天保留日志';
    }
  }
}

export function getJobRetentionFieldHelp(logRetention?: CronLogRetentionLike) {
  return renderJobHelpContent(
    `跟随系统：任务会按系统级日志保留策略执行清理。\n当前系统策略：${formatCronLogRetentionSummary(logRetention)}。\n如果任务单独设置了覆盖策略，则优先使用任务级配置。`,
  );
}
