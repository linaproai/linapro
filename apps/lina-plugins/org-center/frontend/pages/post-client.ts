import { requestClient } from '#/api/request';

export interface Post {
  id: number;
  deptId: number;
  code: string;
  name: string;
  sort: number;
  status: number;
  remark: string;
  createdAt: string;
}

export interface PostListParams {
  pageNum?: number;
  pageSize?: number;
  deptId?: number;
  code?: string;
  name?: string;
  status?: number;
}

export interface PostOption {
  postId: number;
  postName: string;
}

export interface DeptTreeNode {
  id: number;
  label: string;
  children?: DeptTreeNode[];
}

export async function postList(params?: PostListParams) {
  const res = await requestClient.get<{ list: Post[]; total: number }>('/post', { params });
  return { items: res.list, total: res.total };
}

export function postAdd(data: Partial<Post>) {
  return requestClient.post('/post', data);
}

export function postUpdate(id: number, data: Partial<Post>) {
  return requestClient.put(`/post/${id}`, data);
}

export function postDelete(ids: string) {
  return requestClient.delete(`/post/${ids}`);
}

export function postInfo(id: number) {
  return requestClient.get<Post>(`/post/${id}`);
}

export function postExport(params?: PostListParams) {
  return requestClient.download<Blob>('/post/export', { params });
}

export async function postDeptTree() {
  const res = await requestClient.get<{ list: DeptTreeNode[] }>('/post/dept-tree');
  return res.list;
}

export async function postOptionSelect(deptId?: number) {
  const res = await requestClient.get<{ list: PostOption[] }>('/post/option-select', {
    params: deptId ? { deptId } : {},
  });
  return res.list;
}
