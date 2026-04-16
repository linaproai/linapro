import type { VxeTableGridOptions } from '@vben/plugins/vxe-table';

import { Fragment, defineComponent, h } from 'vue';

import {
  setupVbenVxeTable,
  useVbenVxeGrid as useBaseVbenVxeGrid,
} from '@vben/plugins/vxe-table';

import { Button, Image } from 'ant-design-vue';

import PluginSlotOutlet from '#/components/plugin/plugin-slot-outlet.vue';
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
