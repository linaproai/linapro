import type { MenuPermissionOption, Permission } from './data';

import type { useVbenVxeGrid } from '#/adapter/vxe-table';
import type { MenuTreeNode } from '#/api/system/menu';

import { isEmpty, isUndefined } from '@vben/utils';

import { notification } from 'ant-design-vue';

import { $t } from '#/locales';
import { treeToList } from '#/utils/tree';

import { formatMenuPermissionLabel } from './permission-display';

/**
 * 数组差集 - 返回在第一个数组但不在第二个数组的元素
 */
function difference<T>(arr1: T[], arr2: T[]): T[] {
  const set2 = new Set(arr2);
  return arr1.filter((item) => !set2.has(item));
}

/**
 * 保留可持久化的真实菜单 ID，过滤权限树展示用的合成节点 ID。
 */
export function filterPersistedMenuIds(ids: (number | string)[]) {
  return ids.filter((id) => Number(id) > 0);
}

/**
 * 权限列设置是否全选
 */
export function setPermissionsChecked(
  record: MenuPermissionOption,
  checked: boolean,
) {
  if (record?.permissions?.length > 0) {
    record.permissions.forEach((permission) => {
      permission.checked = checked;
    });
  }
}

/**
 * 设置当前行 & 所有子节点选中状态
 */
export function rowAndChildrenChecked(
  record: MenuPermissionOption,
  checked: boolean,
) {
  setPermissionsChecked(record, checked);
  record?.children?.forEach?.((permission) => {
    rowAndChildrenChecked(permission as MenuPermissionOption, checked);
  });
}

/**
 * void方法 会直接修改原始数据
 * 将树结构转为 tree+permissions结构
 */
export function menusWithPermissions(menus: MenuTreeNode[]) {
  const processNode = (item: MenuPermissionOption) => {
    item.label = formatMenuPermissionLabel(item.label);
    validateMenuTree(item);
    if (item.children && item.children.length > 0) {
      const permissions = item.children.filter(
        (child: MenuTreeNode) => child.type === 'B' && item.type !== 'D',
      );
      const diffCollection = difference(item.children, permissions);
      item.children = diffCollection;

      const permissionsArr = permissions.map((permission: MenuTreeNode) => {
        return {
          id: permission.id,
          label: formatMenuPermissionLabel(permission.label),
          checked: false,
        };
      });
      item.permissions = permissionsArr;

      // 递归处理子节点
      diffCollection.forEach((child: MenuTreeNode) => {
        processNode(child as MenuPermissionOption);
      });
    }
  };

  menus.forEach((menu) => {
    processNode(menu as MenuPermissionOption);
  });
}

/**
 * 设置表格选中
 */
export function setTableChecked(
  checkedKeys: (number | string)[],
  menus: MenuPermissionOption[],
  tableApi: ReturnType<typeof useVbenVxeGrid>['1'],
  association: boolean,
) {
  const menuList: MenuPermissionOption[] = treeToList(menus);
  let checkedRows = menuList.filter((item) => checkedKeys.includes(item.id));

  if (!association) {
    checkedRows = checkedRows.filter(
      (item) => isUndefined(item.children) || isEmpty(item.children),
    );
  }

  checkedRows.forEach((item) => {
    tableApi.grid.setCheckboxRow(item, true);
    if (item?.permissions?.length > 0) {
      item.permissions.forEach((permission) => {
        if (checkedKeys.includes(permission.id)) {
          permission.checked = true;
        }
      });
    }
  });

  if (!association) {
    const emptyRows = checkedRows.filter((item) => {
      if (isUndefined(item.permissions) || isEmpty(item.permissions)) {
        return false;
      }
      return (item.permissions as Permission[]).every(
        (permission) => permission.checked === false,
      );
    });
    tableApi.grid.setCheckboxRow(emptyRows, false);
  }
}

/**
 * 校验是否符合规范 给出warning提示
 */
function validateMenuTree(menu: MenuTreeNode) {
  if (menu.type === 'M') {
    menu.children?.forEach?.((item) => {
      if (['M', 'D'].includes(item.type || '')) {
        const description = $t('pages.tree.validation.menuChildInvalid', {
          childLabel: item.label,
          menuLabel: menu.label,
        });
        console.warn(description);
        notification.warning({
          message: $t('pages.common.confirmTitle'),
          description,
          duration: 0,
        });
      }
    });
  }
  if (menu.type === 'B') {
    menu.children?.forEach?.((item) => {
      if (['B', 'D', 'M'].includes(item.type || '')) {
        const description = $t('pages.tree.validation.buttonChildInvalid', {
          childLabel: item.label,
          menuLabel: menu.label,
        });
        console.warn(description);
        notification.warning({
          message: $t('pages.common.confirmTitle'),
          description,
          duration: 0,
        });
      }
    });
  }
}
