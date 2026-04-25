import type { JobRecord } from './model';
import type { CSSProperties } from 'vue';

import { h } from 'vue';
import { $t } from '@vben/locales';

interface CronLogRetentionLike {
  mode?: string;
  value?: number;
}

function renderJobHelpContent(resolveContent: () => string) {
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
      resolveContent(),
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

export function getJobPluginPausedLabel() {
  return $t('pages.system.job.status.pluginUnavailable');
}

export function getJobPluginPausedTooltip() {
  return $t('pages.system.job.messages.pluginPausedTooltip');
}

export const JOB_STATUS_FILTER_OPTIONS = [
  { get label() { return $t('pages.system.job.status.enabled'); }, value: 'enabled' },
  { get label() { return $t('pages.system.job.status.disabled'); }, value: 'disabled' },
  { get label() { return getJobPluginPausedLabel(); }, value: 'paused_by_plugin' },
];

export const JOB_SCOPE_OPTIONS = [
  { get label() { return $t('pages.system.job.scope.masterOnly'); }, value: 'master_only' },
  { get label() { return $t('pages.system.job.scope.allNodes'); }, value: 'all_node' },
];

export const JOB_CONCURRENCY_OPTIONS = [
  { get label() { return $t('pages.system.job.concurrency.singleton'); }, value: 'singleton' },
  { get label() { return $t('pages.system.job.concurrency.parallel'); }, value: 'parallel' },
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
      return $t('pages.system.job.source.hostBuiltin');
    case 'plugin_builtin':
      return $t('pages.system.job.source.pluginBuiltin');
    case 'user_created':
    default:
      return $t('pages.system.job.source.userCreated');
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

export const JOB_CRON_FIELD_HELP = renderJobHelpContent(() =>
  $t('pages.system.job.help.cron'),
);

export const JOB_TIMEOUT_FIELD_HELP = renderJobHelpContent(() =>
  $t('pages.system.job.help.timeout'),
);

export const JOB_MAX_EXECUTIONS_FIELD_HELP = renderJobHelpContent(() =>
  $t('pages.system.job.help.maxExecutions'),
);

export const JOB_SCOPE_FIELD_HELP = renderJobHelpContent(() =>
  $t('pages.system.job.help.scope'),
);

export const JOB_CONCURRENCY_FIELD_HELP = renderJobHelpContent(() =>
  $t('pages.system.job.help.concurrency'),
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
      return value > 0
        ? $t('pages.system.job.retention.summary.countWithValue', { value })
        : $t('pages.system.job.retention.summary.count');
    }
    case 'none': {
      return $t('pages.system.job.retention.summary.none');
    }
    case 'days':
    default: {
      return value > 0
        ? $t('pages.system.job.retention.summary.daysWithValue', { value })
        : $t('pages.system.job.retention.summary.days');
    }
  }
}

export function getJobRetentionFieldHelp(logRetention?: CronLogRetentionLike) {
  return renderJobHelpContent(() =>
    $t('pages.system.job.retention.followSystemHelp', {
      currentPolicy: formatCronLogRetentionSummary(logRetention),
    }),
  );
}
