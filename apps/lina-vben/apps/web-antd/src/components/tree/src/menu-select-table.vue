<script setup lang="ts">
import type { RadioChangeEvent } from 'ant-design-vue';

import type { VxeGridProps } from '#/adapter/vxe-table';
import type { MenuTreeNode } from '#/api/system/menu';

import type { MenuPermissionOption } from './data';

import { nextTick, onMounted, ref, shallowRef, watch } from 'vue';

import { cloneDeep } from '@vben/utils';

import { Alert, Checkbox, RadioGroup, Space } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { findGroupParentIds } from '#/utils/tree';

/**
 * 数组去重
 */
function uniq<T>(arr: T[]): T[] {
  return [...new Set(arr)];
}

import { columns, nodeOptions } from './data';
import {
  menusWithPermissions,
  rowAndChildrenChecked,
  setPermissionsChecked,
  setTableChecked,
} from './helper';
import { useFullScreenGuide } from './hook';

defineOptions({
  name: 'MenuSelectTable',
  inheritAttrs: false,
});

const props = withDefaults(
  defineProps<{
    checkedKeys: (number | string)[];
    defaultExpandAll?: boolean;
    menus: MenuTreeNode[];
  }>(),
  {
    defaultExpandAll: true,
    checkedKeys: () => [],
  },
);

const association = defineModel<boolean>('association', {
  default: true,
});

const gridOptions: VxeGridProps = {
  checkboxConfig: {
    labelField: 'label',
    checkStrictly: !association.value,
  },
  size: 'small',
  columns,
  height: 'auto',
  keepSource: true,
  pagerConfig: {
    enabled: false,
  },
  proxyConfig: {
    enabled: false,
  },
  toolbarConfig: {
    refresh: false,
    custom: false,
  },
  rowConfig: {
    isHover: false,
    isCurrent: false,
    keyField: 'id',
  },
  scrollY: {
    enabled: true,
    gt: 0,
  },
  treeConfig: {
    parentField: 'parentId',
    rowField: 'id',
    transform: false,
  },
  showOverflow: false,
};

const checkedNum = ref(0);

function getTableRecords() {
  return tableApi.grid.getData() as MenuPermissionOption[];
}

function getCheckedRecords() {
  return (tableApi?.grid?.getCheckboxRecords?.(true) ?? []) as MenuPermissionOption[];
}

function updateCheckedNumber() {
  checkedNum.value = getCheckedKeys().length;
}

const [BasicTable, tableApi] = useVbenVxeGrid({
  gridOptions,
  gridEvents: {
    checkboxChange: (params: any) => {
      const checked = params.checked;
      const record = params.row;
      if (association.value) {
        rowAndChildrenChecked(record, checked);
      } else {
        setPermissionsChecked(record, checked);
      }
      updateCheckedNumber();
    },
    checkboxAll: (params: any) => {
      const records = params.$grid.getData();
      records.forEach((item: any) => {
        rowAndChildrenChecked(item, params.checked);
      });
      updateCheckedNumber();
    },
  },
});

function setCheckedByKeys(
  menus: MenuPermissionOption[],
  keys: (number | string)[],
  triggerOnchange: boolean,
) {
  menus.forEach((item) => {
    if (keys.includes(item.id)) {
      tableApi.grid.setCheckboxRow(item, true);
    }
    if (item.permissions && item.permissions.length > 0) {
      item.permissions.forEach((permission) => {
        if (keys.includes(permission.id)) {
          permission.checked = true;
          triggerOnchange && handlePermissionChange(item);
        }
      });
    }
    if (item.children && item.children.length > 0) {
      setCheckedByKeys(item.children as any, keys, triggerOnchange);
    }
  });
}

const { FullScreenGuide, openGuide } = useFullScreenGuide();
onMounted(() => {
  watch(
    () => props.menus,
    async (menus) => {
      const clonedMenus = cloneDeep(menus);
      menusWithPermissions(clonedMenus);
      await tableApi.grid.loadData(clonedMenus);
      if (props.defaultExpandAll) {
        await nextTick();
        setExpandOrCollapse(true);
      }
    },
  );

  watch(association, (value) => {
    tableApi.setGridOptions({
      checkboxConfig: {
        checkStrictly: !value,
      },
    });
  });

  watch(
    () => props.checkedKeys,
    (value) => {
      const allCheckedKeys = uniq([...value]);
      const records = getTableRecords();
      setCheckedByKeys(records, allCheckedKeys, association.value);
      updateCheckedNumber();
      setTimeout(openGuide, 1000);
    },
  );
});

const lastCheckedKeys = shallowRef<(number | string)[]>([]);
async function handleAssociationChange(e: RadioChangeEvent) {
  lastCheckedKeys.value = getCheckedKeys();
  const records = getTableRecords();
  records.forEach((item: any) => {
    rowAndChildrenChecked(item, false);
  });
  await tableApi.grid.clearCheckboxRow();
  await tableApi.grid.scrollTo(0, 0);
  setTableChecked(lastCheckedKeys.value, records, tableApi, !e.target.value);
  updateCheckedNumber();
}

function setExpandOrCollapse(expand: boolean) {
  tableApi.grid?.setAllTreeExpand(expand);
}

function handlePermissionChange(row: any) {
  if (association.value) {
    const checkedPermissions = row.permissions.filter(
      (item: any) => item.checked === true,
    );
    if (checkedPermissions.length > 0) {
      tableApi.grid.setCheckboxRow(row, true);
    }
    if (checkedPermissions.length === 0) {
      tableApi.grid.setCheckboxRow(row, false);
    }
  }
  updateCheckedNumber();
}

function getKeys(records: MenuPermissionOption[], addCurrent: boolean) {
  const allKeys: (number | string)[] = [];
  records.forEach((item) => {
    if (item.children && item.children.length > 0) {
      const keys = getKeys(item.children as MenuPermissionOption[], addCurrent);
      allKeys.push(...keys);
    } else {
      addCurrent && allKeys.push(item.id);
      if (item.permissions && item.permissions.length > 0) {
        const ids = item.permissions
          .filter((m) => m.checked === true)
          .map((m) => m.id);
        allKeys.push(...ids);
      }
    }
  });
  return uniq(allKeys);
}

function getCheckedKeys() {
  if (association.value) {
    const records = getCheckedRecords();
    const nodeKeys = getKeys(records, true);
    const parentIds = findGroupParentIds(props.menus, nodeKeys as number[]);
    const realKeys = uniq([...parentIds, ...nodeKeys]);
    return realKeys;
  }

  const records = getCheckedRecords();
  const allRecords = getTableRecords();
  const checkedIds = records.map((item: any) => item.id);
  const permissionIds = getKeys(allRecords, false);
  const allIds = uniq([...checkedIds, ...permissionIds]);
  return allIds;
}

defineExpose({
  getCheckedKeys,
});
</script>

<template>
  <div class="flex h-full flex-col" id="menu-select-table">
    <BasicTable>
      <template #toolbar-actions>
        <RadioGroup
          v-model:value="association"
          :options="nodeOptions"
          button-style="solid"
          option-type="button"
          @change="handleAssociationChange"
        />
        <Alert class="mx-2" type="info">
          <template #message>
            <div>
              已选中
              <span class="text-primary mx-1 font-semibold">
                {{ checkedNum }}
              </span>
              个节点
            </div>
          </template>
        </Alert>
      </template>
      <template #toolbar-tools>
        <Space>
          <a-button @click="setExpandOrCollapse(false)">
            折叠
          </a-button>
          <a-button @click="setExpandOrCollapse(true)">
            展开
          </a-button>
        </Space>
      </template>
      <template #permissions="{ row }">
        <div class="flex flex-wrap gap-x-3 gap-y-1">
          <Checkbox
            v-for="permission in row.permissions"
            :key="permission.id"
            v-model:checked="permission.checked"
            @change="() => handlePermissionChange(row)"
          >
            {{ permission.label }}
          </Checkbox>
        </div>
      </template>
    </BasicTable>
    <FullScreenGuide />
  </div>
</template>

<style scoped>
:deep(.ant-alert) {
  padding: 4px 8px;
}
</style>
