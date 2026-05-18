import { requestClient } from '#/api/request';

export interface Menu {
  id: number;
  parentId: number;
  name: string;
  path: string;
  component: string;
  perms: string;
  icon: string;
  type: string; // D=Directory M=Menu B=Button
  sort: number;
  visible: number;
  status: number;
  isFrame: number;
  isCache: number;
  queryParam: string;
  remark: string;
  createdAt: string;
  updatedAt: string;
  children?: Menu[];
}

export interface MenuTreeNode {
  id: number;
  parentId: number;
  label: string;
  type: string; // D=Directory M=Menu B=Button
  icon?: string;
  children?: MenuTreeNode[];
}

export interface RoleMenuTreeResp {
  menus: MenuTreeNode[];
  checkedKeys: number[];
}

export interface MenuListParams {
  name?: string;
  status?: number;
  visible?: number;
}

/** 菜单列表 */
export async function menuList(params?: MenuListParams) {
  const res = await requestClient.get<{ list: Menu[] }>('/menu', { params });
  return res.list;
}

/** 菜单详情 */
export function menuInfo(id: number) {
  return requestClient.get<Menu>(`/menu/${id}`);
}

/** 新增菜单 */
export function menuAdd(data: Partial<Menu>) {
  return requestClient.post('/menu', data);
}

/** 更新菜单 */
export function menuUpdate(id: number, data: Partial<Menu>) {
  return requestClient.put(`/menu/${id}`, data);
}

/** 删除菜单 */
export function menuRemove(id: number, cascadeDelete?: boolean) {
  const params = cascadeDelete ? { cascadeDelete: true } : {};
  return requestClient.delete(`/menu/${id}`, { params });
}

/** 菜单下拉树 */
export async function menuTreeSelect() {
  const res = await requestClient.get<{ list: MenuTreeNode[] }>('/menu/treeselect');
  return res.list;
}

/** 角色菜单树 */
export async function roleMenuTreeSelect(roleId: number) {
  return requestClient.get<RoleMenuTreeResp>(`/menu/role/${roleId}`);
}
