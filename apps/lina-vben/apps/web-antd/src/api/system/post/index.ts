import type { DeptTreeNode, Post, PostListParams } from './model';

import { requestClient } from '#/api/request';

/** 岗位列表 */
export async function postList(params?: PostListParams) {
  const res = await requestClient.get<{ list: Post[]; total: number }>(
    '/post',
    { params },
  );
  return { items: res.list, total: res.total };
}

/** 新增岗位 */
export function postAdd(data: Partial<Post>) {
  return requestClient.post('/post', data);
}

/** 更新岗位 */
export function postUpdate(id: number, data: Partial<Post>) {
  return requestClient.put(`/post/${id}`, data);
}

/** 删除岗位 */
export function postDelete(ids: string) {
  return requestClient.delete(`/post/${ids}`);
}

/** 获取岗位详情 */
export function postInfo(id: number) {
  return requestClient.get<Post>(`/post/${id}`);
}

/** 导出岗位列表 */
export function postExport(params?: PostListParams) {
  return requestClient.download<Blob>('/post/export', { params });
}

/** 获取岗位部门树 */
export async function postDeptTree() {
  const res = await requestClient.get<{ list: DeptTreeNode[] }>(
    '/post/dept-tree',
  );
  return res.list;
}

/** 获取岗位选项列表 */
export async function postOptionSelect(deptId?: number) {
  const res = await requestClient.get<{
    list: Array<{ postId: number; postName: string }>;
  }>('/post/option-select', {
    params: deptId ? { deptId } : {},
  });
  return res.list;
}
