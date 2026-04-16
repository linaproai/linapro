import { requestClient } from '#/api/request';

export interface Role {
  id: number;
  name: string;
  key: string;
  sort: number;
  dataScope: number;
  status: number;
  remark: string;
  createdAt: string;
  updatedAt: string;
  menuIds?: number[];
}

export interface RoleOption {
  id: number;
  name: string;
  key: string;
}

export interface RoleListParams {
  name?: string;
  key?: string;
  status?: number;
  page?: number;
  size?: number;
}

export interface RoleUser {
  id: number;
  username: string;
  nickname: string;
  email: string;
  phone: string;
  status: number;
  createdAt: string;
}

export interface RoleUsersParams {
  id: number;
  username?: string;
  phone?: string;
  status?: number;
  page?: number;
  size?: number;
}

/** 角色列表 */
export async function roleList(params?: RoleListParams) {
  const res = await requestClient.get<{ list: Role[]; total: number }>('/role', { params });
  return { items: res.list, total: res.total };
}

/** 角色详情 */
export function roleInfo(id: number) {
  return requestClient.get<Role>(`/role/${id}`);
}

/** 新增角色 */
export function roleAdd(data: Partial<Role>) {
  return requestClient.post<{ id: number }>('/role', data);
}

/** 更新角色 */
export function roleUpdate(id: number, data: Partial<Role>) {
  return requestClient.put(`/role/${id}`, data);
}

/** 删除角色 */
export function roleRemove(id: number) {
  return requestClient.delete(`/role/${id}`);
}

/** 修改角色状态 */
export function roleStatusChange(id: number, status: number) {
  return requestClient.put(`/role/${id}/status`, { status });
}

/** 角色下拉选项 */
export async function roleOptions() {
  const res = await requestClient.get<{ list: RoleOption[] }>('/role/options');
  return res.list;
}

/** 角色用户列表 */
export async function roleUsers(params: RoleUsersParams) {
  const { id, ...rest } = params;
  const res = await requestClient.get<{ list: RoleUser[]; total: number }>(`/role/${id}/users`, { params: rest });
  return { items: res.list, total: res.total };
}

/** 分配用户到角色 */
export function roleAssignUsers(roleId: number, userIds: number[]) {
  return requestClient.post(`/role/${roleId}/users`, { userIds });
}

/** 取消用户授权 */
export function roleUnassignUser(roleId: number, userId: number) {
  return requestClient.delete(`/role/${roleId}/users/${userId}`);
}

/** 批量取消用户授权 */
export function roleUnassignUsers(roleId: number, userIds: number[]) {
  return requestClient.delete(`/role/${roleId}/users`, { data: { userIds } });
}