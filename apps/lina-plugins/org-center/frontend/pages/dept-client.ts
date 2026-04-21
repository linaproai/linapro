import { requestClient } from '#/api/request';

export interface Dept {
  id: number;
  parentId: number;
  ancestors: string;
  name: string;
  code: string;
  orderNum: number;
  leader: number;
  phone: string;
  email: string;
  status: number;
  remark: string;
  createdAt: string;
}

export interface DeptTree {
  id: number;
  label: string;
  userCount?: number;
  children?: DeptTree[];
}

export interface DeptUser {
  id: number;
  username: string;
  nickname: string;
}

export async function deptList(params?: Record<string, any>) {
  const res = await requestClient.get<{ list: Dept[] }>('/dept', { params });
  return res.list;
}

export function deptAdd(data: Partial<Dept>) {
  return requestClient.post('/dept', data);
}

export function deptUpdate(id: number, data: Partial<Dept>) {
  return requestClient.put(`/dept/${id}`, data);
}

export function deptDelete(id: number) {
  return requestClient.delete(`/dept/${id}`);
}

export function deptInfo(id: number) {
  return requestClient.get<Dept>(`/dept/${id}`);
}

export async function deptTree() {
  const res = await requestClient.get<{ list: DeptTree[] }>('/dept/tree');
  return res.list;
}

export async function deptExclude(id: number) {
  const res = await requestClient.get<{ list: DeptTree[] }>(`/dept/exclude/${id}`);
  return res.list;
}

export async function deptUsers(id: number, params?: { keyword?: string; limit?: number }) {
  const res = await requestClient.get<{ list: DeptUser[] }>(`/dept/${id}/users`, { params });
  return res.list;
}
