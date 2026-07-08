import { describe, expect, it } from 'vitest';

import {
  filterPersistedMenuIds,
  setTableChecked,
  shouldUseAssociatedMenuSelection,
} from './helper';

describe('filterPersistedMenuIds', () => {
  it('drops synthetic display nodes from submitted menu ids', () => {
    expect(filterPersistedMenuIds([-1000001, 1, '2', 0, '-3'])).toEqual([
      1,
      '2',
    ]);
  });
});

describe('shouldUseAssociatedMenuSelection', () => {
  const menus = [
    {
      id: 1,
      label: '系统管理',
      parentId: 0,
      type: 'D',
      children: [
        {
          id: 2,
          label: '角色管理',
          parentId: 1,
          type: 'M',
          children: [
            { id: 3, label: '查询', parentId: 2, type: 'B' },
            { id: 4, label: '编辑', parentId: 2, type: 'B' },
          ],
        },
      ],
    },
  ];

  it('keeps linked mode when selected keys contain the required ancestors', () => {
    expect(shouldUseAssociatedMenuSelection(menus, [1, 2, 3])).toBe(true);
  });

  it('uses independent mode when a selected button is missing ancestors', () => {
    expect(shouldUseAssociatedMenuSelection(menus, [3])).toBe(false);
  });
});

describe('setTableChecked', () => {
  it('restores a button-only selection without checking extra menu rows', () => {
    const selectedRows: number[] = [];
    const menus = [
      {
        id: 2,
        label: '角色管理',
        parentId: 1,
        permissions: [{ checked: false, id: 3, label: '查询' }],
        type: 'M',
      },
    ] as any[];
    const tableApi = {
      grid: {
        setCheckboxRow(rows: any, checked: boolean) {
          if (!checked) {
            return;
          }
          const nextRows = Array.isArray(rows) ? rows : [rows];
          selectedRows.push(...nextRows.map((row) => row.id));
        },
      },
    } as any;

    setTableChecked([3], menus, tableApi, false);

    expect(menus[0].permissions[0].checked).toBe(true);
    expect(selectedRows).toEqual([]);
  });
});
