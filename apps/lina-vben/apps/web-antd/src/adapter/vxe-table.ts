import type { VxeTableGridOptions } from '@vben/plugins/vxe-table';

import { Fragment, defineComponent, h } from 'vue';

import { Tag, Tooltip } from 'ant-design-vue';
import {
  setupVbenVxeTable,
  useVbenVxeGrid as useBaseVbenVxeGrid,
} from '@vben/plugins/vxe-table';

import { Button, Image } from 'ant-design-vue';

import PluginSlotOutlet from '#/components/plugin/plugin-slot-outlet.vue';
import {
  getJobConcurrencyLabel,
  renderJobCronExpression,
  getJobScopeLabel,
  JOB_PLUGIN_PAUSED_LABEL,
  JOB_PLUGIN_PAUSED_TOOLTIP,
} from '#/api/system/job/meta';
import { pluginSlotKeys } from '#/plugins/plugin-slots';

import { useVbenForm } from './form';

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

export function useVbenVxeGrid(...args: Parameters<typeof useBaseVbenVxeGrid>) {
  const [BaseGrid, gridApi] = useBaseVbenVxeGrid(...args);

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
    { field: 'code', title: '分组编码', minWidth: 160 },
    { field: 'name', title: '分组名称', minWidth: 180 },
    { field: 'sortOrder', title: '排序', width: 90 },
    { field: 'jobCount', title: '任务数', width: 100 },
    {
      field: 'isDefault',
      title: '默认分组',
      width: 110,
      slots: {
        default: ({ row }: any) =>
          row.isDefault === 1
            ? h(Tag, { color: 'gold' }, () => '默认分组')
            : '-',
      },
    },
    { field: 'remark', title: '备注', minWidth: 200 },
    { field: 'updatedAt', title: '更新时间', minWidth: 180 },
    {
      field: 'action',
      fixed: 'right',
      title: '操作',
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
    { field: 'name', title: '任务名称', minWidth: 180 },
    { field: 'groupName', title: '所属分组', minWidth: 140 },
    {
      field: 'taskType',
      title: '任务类型',
      width: 110,
      slots: {
        default: ({ row }: any) =>
          h(
            Tag,
            { color: row.taskType === 'shell' ? 'volcano' : 'blue' },
            () => (row.taskType === 'shell' ? 'Shell' : 'Handler'),
          ),
      },
    },
    {
      field: 'status',
      title: '状态',
      minWidth: 180,
      slots: {
        default: ({ row }: any) => {
          if (row.status === 'paused_by_plugin') {
            return h(
              Tooltip,
              { title: JOB_PLUGIN_PAUSED_TOOLTIP },
              {
                default: () =>
                  h(Tag, { color: 'error' }, () => JOB_PLUGIN_PAUSED_LABEL),
              },
            );
          }
          if (row.status === 'enabled') {
            return h(Tag, { color: 'success' }, () => '启用');
          }
          return h(Tag, {}, () => '停用');
        },
      },
    },
    {
      field: 'cronExpr',
      title: '定时表达式',
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
    { field: 'timezone', title: '时区', width: 140 },
    {
      field: 'scope',
      title: '调度范围',
      minWidth: 140,
      slots: {
        default: ({ row }: any) => getJobScopeLabel(row.scope),
      },
    },
    {
      field: 'concurrency',
      title: '并发策略',
      minWidth: 140,
      slots: {
        default: ({ row }: any) => getJobConcurrencyLabel(row.concurrency),
      },
    },
    { field: 'executedCount', title: '已执行次数', width: 110 },
    {
      field: 'stopReason',
      title: '停止原因',
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
    { field: 'updatedAt', title: '更新时间', minWidth: 180 },
    {
      field: 'action',
      fixed: 'right',
      title: '操作',
      width: 320,
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
    { field: 'jobName', title: '任务名称', minWidth: 180 },
    { field: 'trigger', title: '触发方式', width: 100 },
    { field: 'nodeId', title: '执行节点', minWidth: 140 },
    {
      field: 'status',
      title: '状态',
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
            () => row.status,
          );
        },
      },
    },
    { field: 'startAt', title: '开始时间', minWidth: 180 },
    { field: 'endAt', title: '结束时间', minWidth: 180 },
    { field: 'durationMs', title: '耗时(ms)', width: 100 },
    {
      field: 'errMsg',
      title: '错误摘要',
      minWidth: 240,
      slots: {
        default: ({ row }: any) => row.errMsg || '-',
      },
    },
    {
      field: 'action',
      fixed: 'right',
      title: '操作',
      width: 220,
      slots: { default: 'action' },
    },
  ];
}
