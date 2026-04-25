import type { VxeTableGridOptions } from '@vben/plugins/vxe-table';

import { Fragment, defineComponent, h } from 'vue';

import { Tag, Tooltip } from 'ant-design-vue';
import { $t } from '@vben/locales';
import { preferences } from '@vben/preferences';
import {
  setupVbenVxeTable,
  useVbenVxeGrid as useBaseVbenVxeGrid,
} from '@vben/plugins/vxe-table';

import { Button, Image } from 'ant-design-vue';

import PluginSlotOutlet from '#/components/plugin/plugin-slot-outlet.vue';
import {
  getJobConcurrencyLabel,
  getJobPluginPausedLabel,
  getJobPluginPausedTooltip,
  getJobScopeLabel,
  getJobSourceColor,
  getJobSourceKind,
  getJobSourceLabel,
  renderJobCronExpression,
} from '#/api/system/job/meta';
import { pluginSlotKeys } from '#/plugins/plugin-slots';

import { useVbenForm } from './form';

const ENGLISH_SEARCH_LABEL_WIDTH = 120;

function normalizeClassName(value = '') {
  return value.trim().replace(/\s+/g, ' ');
}

function normalizeEnglishSearchWrapperClass(wrapperClass = '') {
  const normalized = normalizeClassName(wrapperClass);
  if (!normalized) {
    return 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-3 2xl:grid-cols-4';
  }

  const widened = normalized.replace(
    /\bxl:grid-cols-4\b/g,
    'xl:grid-cols-3 2xl:grid-cols-4',
  );

  if (/\b2xl:grid-cols-\d+\b/.test(widened)) {
    return normalizeClassName(widened);
  }
  if (/\bxl:grid-cols-3\b/.test(widened)) {
    return normalizeClassName(`${widened} 2xl:grid-cols-4`);
  }
  if (/\blg:grid-cols-3\b/.test(widened)) {
    return normalizeClassName(`${widened} xl:grid-cols-3 2xl:grid-cols-4`);
  }
  return normalizeClassName(widened);
}

function estimateEnglishColumnWidth(title: string) {
  const normalized = title.trim().replace(/\s+/g, ' ');
  if (!normalized) {
    return 0;
  }
  return Math.min(220, Math.max(96, Math.ceil(normalized.length * 7.5 + 32)));
}

function normalizeEnglishColumns(columns: any[] = []) {
  return columns.map((column) => {
    if (!column || typeof column !== 'object') {
      return column;
    }

    const nextColumn = { ...column };
    if (Array.isArray(nextColumn.children)) {
      nextColumn.children = normalizeEnglishColumns(nextColumn.children);
    }

    if (nextColumn.showHeaderOverflow === undefined) {
      nextColumn.showHeaderOverflow = 'tooltip';
    }

    const title = typeof nextColumn.title === 'string' ? nextColumn.title : '';
    const estimatedWidth = estimateEnglishColumnWidth(title);
    if (!estimatedWidth) {
      return nextColumn;
    }

    if (typeof nextColumn.width === 'number') {
      nextColumn.width = Math.max(nextColumn.width, estimatedWidth);
      return nextColumn;
    }

    if (typeof nextColumn.minWidth === 'number') {
      nextColumn.minWidth = Math.max(nextColumn.minWidth, estimatedWidth);
      return nextColumn;
    }

    if (!nextColumn.type) {
      nextColumn.minWidth = estimatedWidth;
    }
    return nextColumn;
  });
}

function adaptEnglishGridOptions<T extends Parameters<typeof useBaseVbenVxeGrid>[0]>(
  options: T,
) {
  if (preferences.app.locale !== 'en-US') {
    return options;
  }

  const nextOptions = { ...options };

  if (nextOptions.formOptions) {
    nextOptions.formOptions = { ...nextOptions.formOptions };
    nextOptions.formOptions.commonConfig = {
      ...(nextOptions.formOptions.commonConfig ?? {}),
      labelWidth: Math.max(
        nextOptions.formOptions.commonConfig?.labelWidth ?? 0,
        ENGLISH_SEARCH_LABEL_WIDTH,
      ),
    };
    nextOptions.formOptions.wrapperClass = normalizeEnglishSearchWrapperClass(
      nextOptions.formOptions.wrapperClass,
    );
  }

  if (nextOptions.gridOptions) {
    nextOptions.gridOptions = {
      ...nextOptions.gridOptions,
      showHeaderOverflow:
        nextOptions.gridOptions.showHeaderOverflow ?? 'tooltip',
      showOverflow: nextOptions.gridOptions.showOverflow ?? 'tooltip',
    };

    if (Array.isArray(nextOptions.gridOptions.columns)) {
      nextOptions.gridOptions.columns = normalizeEnglishColumns(
        nextOptions.gridOptions.columns as any[],
      ) as any;
    }
  }

  return nextOptions;
}

setupVbenVxeTable({
  configVxeTable: (vxeUI) => {
    vxeUI.setConfig({
      grid: {
        align: 'center',
        border: 'inner',
        columnConfig: {
          resizable: true,
        },
        minHeight: 180,
        formConfig: {
          // 全局禁用vxe-table的表单配置，使用formOptions
          enabled: false,
        },
        proxyConfig: {
          autoLoad: true,
          response: {
            result: 'items',
            total: 'total',
            list: 'items',
          },
          showActiveMsg: true,
          showResponseMsg: false,
        },
        round: true,
        showOverflow: true,
        size: 'medium',
        pagerConfig: {
          pageSize: 10,
          pageSizes: [10, 20, 50, 100],
        },
        rowConfig: {
          isHover: true,
          isCurrent: false,
        },
        toolbarConfig: {
          custom: true,
          zoom: true,
          refresh: true,
          refreshOptions: {
            code: 'query',
          },
        },
        customConfig: {
          storage: false,
        },
      } as VxeTableGridOptions,
    });

    // 表格配置项可以用 cellRender: { name: 'CellImage' },
    vxeUI.renderer.add('CellImage', {
      renderTableDefault(renderOpts, params) {
        const { props } = renderOpts;
        const { column, row } = params;
        return h(Image, { src: row[column.field], ...props });
      },
    });

    // 表格配置项可以用 cellRender: { name: 'CellLink' },
    vxeUI.renderer.add('CellLink', {
      renderTableDefault(renderOpts) {
        const { props } = renderOpts;
        return h(
          Button,
          { size: 'small', type: 'link' },
          { default: () => props?.text },
        );
      },
    });

    // 这里可以自行扩展 vxe-table 的全局配置，比如自定义格式化
    // vxeUI.formats.add
  },
  useVbenForm,
});

function resolveJobHandlerRef(jobSnapshot?: string) {
  if (!jobSnapshot) {
    return '';
  }
  try {
    const parsed = JSON.parse(jobSnapshot) as { handlerRef?: string };
    return parsed?.handlerRef || '';
  } catch {
    return '';
  }
}

export function useVbenVxeGrid(...args: Parameters<typeof useBaseVbenVxeGrid>) {
  const normalizedArgs = args.map((arg) =>
    adaptEnglishGridOptions(arg),
  ) as Parameters<typeof useBaseVbenVxeGrid>;
  const [BaseGrid, gridApi] = useBaseVbenVxeGrid(
    ...normalizedArgs,
  );

  const Grid = defineComponent(
    (props, { attrs, slots }) => {
      return () =>
        h(Fragment, null, [
          h(
            BaseGrid as any,
            { ...props, ...attrs },
            {
              ...slots,
              'toolbar-tools': (slotProps: Record<string, any>) => [
                slots['toolbar-tools']?.(slotProps),
                h(PluginSlotOutlet, {
                  class: 'ml-2',
                  slotKey: pluginSlotKeys.crudToolbarAfter,
                }),
              ],
            },
          ),
          h(PluginSlotOutlet, {
            class: 'mt-3',
            slotKey: pluginSlotKeys.crudTableAfter,
          }),
        ]);
    },
    {
      inheritAttrs: false,
      name: 'LinaPluginAwareGrid',
    },
  );

  return [Grid, gridApi] as const;
}

/**
 * 判断vxe-table的复选框是否选中
 */
export function vxeCheckboxChecked(
  tableApi: ReturnType<typeof useVbenVxeGrid>[1],
) {
  return tableApi?.grid?.getCheckboxRecords?.()?.length > 0;
}

export type * from '@vben/plugins/vxe-table';

/**
 * 构建任务分组列表列定义。
 */
export function buildJobGroupColumns(): VxeTableGridOptions['columns'] {
  return [
    { type: 'checkbox', width: 56 },
    { field: 'code', title: $t('pages.system.jobGroup.fields.code'), minWidth: 160 },
    {
      field: 'name',
      title: $t('pages.system.jobGroup.fields.name'),
      minWidth: 180,
    },
    { field: 'sortOrder', title: $t('pages.fields.sort'), width: 90 },
    { field: 'jobCount', title: $t('pages.system.jobGroup.fields.jobCount'), width: 100 },
    {
      field: 'isDefault',
      title: $t('pages.system.jobGroup.fields.defaultGroup'),
      width: 110,
      slots: {
        default: ({ row }: any) =>
          row.isDefault === 1
            ? h(Tag, { color: 'gold' }, () => $t('pages.system.jobGroup.fields.defaultGroup'))
            : '-',
      },
    },
    { field: 'remark', title: $t('pages.common.remark'), minWidth: 200 },
    { field: 'updatedAt', title: $t('pages.common.updatedAt'), minWidth: 180 },
    {
      field: 'action',
      fixed: 'right',
      title: $t('pages.common.actions'),
      width: 220,
      slots: { default: 'action' },
    },
  ];
}

/**
 * 构建任务列表列定义。
 */
export function buildJobColumns(): VxeTableGridOptions['columns'] {
  return [
    { type: 'checkbox', width: 56 },
    {
      field: 'name',
      title: $t('pages.system.job.fields.name'),
      minWidth: 180,
    },
    {
      field: 'groupName',
      title: $t('pages.system.job.fields.group'),
      minWidth: 140,
    },
    {
      field: 'source',
      title: $t('pages.system.job.fields.source'),
      minWidth: 120,
      slots: {
        default: ({ row }: any) => {
          const source = getJobSourceKind(row);
          return h(
            Tag,
            { color: getJobSourceColor(source) },
            () => getJobSourceLabel(source),
          );
        },
      },
    },
    {
      field: 'status',
      title: $t('pages.common.status'),
      minWidth: 180,
      slots: {
        default: ({ row }: any) => {
          if (row.status === 'paused_by_plugin') {
            return h(
              Tooltip,
              { title: getJobPluginPausedTooltip() },
              {
                default: () =>
                  h(Tag, { color: 'error' }, () =>
                    getJobPluginPausedLabel(),
                  ),
              },
            );
          }
          if (row.status === 'enabled') {
            return h(Tag, { color: 'success' }, () =>
              $t('pages.system.job.status.enabled'),
            );
          }
          return h(Tag, {}, () => $t('pages.system.job.status.disabled'));
        },
      },
    },
    {
      field: 'cronExpr',
      title: $t('pages.system.job.fields.cronExpr'),
      minWidth: 220,
      slots: {
        default: ({ row }: any) =>
          h(
            Tooltip,
            { title: row.cronExpr || '-' },
            {
              default: () =>
                renderJobCronExpression(row.cronExpr, {
                  'data-testid': `job-cron-expr-${row.id}`,
                }),
            },
          ),
      },
    },
    { field: 'timezone', title: $t('pages.system.job.fields.timezone'), width: 140 },
    {
      field: 'scope',
      title: $t('pages.system.job.fields.scope'),
      minWidth: 140,
      slots: {
        default: ({ row }: any) => getJobScopeLabel(row.scope),
      },
    },
    {
      field: 'concurrency',
      title: $t('pages.system.job.fields.concurrency'),
      minWidth: 140,
      slots: {
        default: ({ row }: any) => getJobConcurrencyLabel(row.concurrency),
      },
    },
    { field: 'executedCount', title: $t('pages.system.job.fields.executedCount'), width: 110 },
    {
      field: 'stopReason',
      title: $t('pages.system.job.fields.stopReason'),
      minWidth: 150,
      slots: {
        default: ({ row }: any) =>
          row.stopReason
            ? h(
                Tag,
                {
                  color:
                    row.stopReason === 'max_executions_reached'
                      ? 'warning'
                      : 'default',
                },
                () => row.stopReason,
              )
            : '-',
      },
    },
    { field: 'updatedAt', title: $t('pages.common.updatedAt'), minWidth: 180 },
    {
      field: 'action',
      fixed: 'right',
      title: $t('pages.common.actions'),
      width: 260,
      slots: { default: 'action' },
    },
  ];
}

/**
 * 构建任务日志列表列定义。
 */
export function buildJobLogColumns(): VxeTableGridOptions['columns'] {
  return [
    { type: 'checkbox', width: 56 },
    {
      field: 'jobName',
      title: $t('pages.system.jobLog.fields.jobName'),
      minWidth: 180,
      formatter: ({ row }: any) =>
        String(row?.jobName || resolveJobHandlerRef(String(row?.jobSnapshot || '')) || ''),
    },
    { field: 'trigger', title: $t('pages.system.jobLog.fields.trigger'), width: 100 },
    { field: 'nodeId', title: $t('pages.system.jobLog.fields.nodeId'), minWidth: 140 },
    {
      field: 'status',
      title: $t('pages.system.jobLog.fields.status'),
      width: 160,
      slots: {
        default: ({ row }: any) => {
          const colorMap: Record<string, string> = {
            cancelled: 'default',
            failed: 'error',
            running: 'processing',
            success: 'success',
            timeout: 'warning',
          };
          return h(
            Tag,
            { color: colorMap[row.status] || 'default' },
            () =>
              $t(`pages.system.jobLog.status.${row.status}`, {
                defaultValue: row.status,
              }),
          );
        },
      },
    },
    { field: 'startAt', title: $t('pages.system.jobLog.fields.startAt'), minWidth: 180 },
    { field: 'endAt', title: $t('pages.system.jobLog.fields.endAt'), minWidth: 180 },
    { field: 'durationMs', title: $t('pages.system.jobLog.fields.durationMs'), width: 100 },
    {
      field: 'errMsg',
      title: $t('pages.system.jobLog.fields.errorSummary'),
      minWidth: 240,
      slots: {
        default: ({ row }: any) => row.errMsg || '-',
      },
    },
    {
      field: 'action',
      fixed: 'right',
      title: $t('pages.common.actions'),
      width: 220,
      slots: { default: 'action' },
    },
  ];
}
