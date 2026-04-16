import type { Dept, DeptTree, DeptUser } from './model';

import { requestClient } from '#/api/request';

/** 部门列表 */
export async function deptList(params?: Record<string, any>) {
  const res = await requestClient.get<{ list: Dept[] }>('/dept', { params });
  return res.list;
}

/** 新增部门 */
export function deptAdd(data: Partial<Dept>) {
  return requestClient.post('/dept', data);
}

/** 更新部门 */
export function deptUpdate(id: number, data: Partial<Dept>) {
  return requestClient.put(`/dept/${id}`, data);
}

/** 删除部门 */
export function deptDelete(id: number) {
  return requestClient.delete(`/dept/${id}`);
}

/** 获取部门详情 */
export function deptInfo(id: number) {
  return requestClient.get<Dept>(`/dept/${id}`);
}

/** 部门树 */
export async function deptTree() {
  const res = await requestClient.get<{ list: DeptTree[] }>('/dept/tree');
  return res.list;
}

/** 部门树（排除指定节点及其子节点） */
export async function deptExclude(id: number) {
  const res = await requestClient.get<{ list: DeptTree[] }>(
    `/dept/exclude/${id}`,
  );
  return res.list;
}

/** 获取部门下的用户列表（负责人选择） */
export async function deptUsers(
  id: number,
  params?: { keyword?: string; limit?: number },
) {
  const res = await requestClient.get<{ list: DeptUser[] }>(
    `/dept/${id}/users`,
    { params },
  );
  return res.list;
}
