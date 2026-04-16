import { requestClient } from '#/api/request';

export interface SysUser {
  id: number;
  username: string;
  nickname: string;
  email: string;
  phone: string;
  sex: number;
  avatar: string;
  status: number;
  remark: string;
  loginDate: string;
  createdAt: string;
  updatedAt: string;
  deptId: number;
  deptName: string;
  postIds: number[];
  roleIds: number[];
  roleNames: string[];
}

export interface DeptTree {
  id: number;
  label: string;
  children?: DeptTree[];
}

export interface UserListParams {
  pageNum?: number;
  pageSize?: number;
  username?: string;
  nickname?: string;
  status?: number;
  phone?: string;
  beginTime?: string;
  endTime?: string;
  orderBy?: string;
  orderDirection?: string;
  deptId?: number;
}

export interface UserListResult {
  list: SysUser[];
  total: number;
}

export interface UserCreateParams {
  username: string;
  password: string;
  nickname?: string;
  email?: string;
  phone?: string;
  sex?: number;
  status?: number;
  remark?: string;
  deptId?: number;
  postIds?: number[];
  roleIds?: number[];
}

export interface UserUpdateParams {
  id: number;
  username?: string;
  password?: string;
  nickname?: string;
  email?: string;
  phone?: string;
  sex?: number;
  status?: number;
  remark?: string;
  deptId?: number;
  postIds?: number[];
  roleIds?: number[];
}

/** 用户列表 */
export async function userList(params?: UserListParams) {
  const res = await requestClient.get<UserListResult>('/user', { params });
  // VXE-Grid proxy expects { items, total } format
  return { items: res.list, total: res.total };
}

/** 创建用户 */
export function userAdd(data: UserCreateParams) {
  return requestClient.post('/user', data);
}

/** 更新用户 */
export function userUpdate(data: UserUpdateParams) {
  return requestClient.put(`/user/${data.id}`, data);
}

/** 删除用户 */
export function userDelete(id: number) {
  return requestClient.delete(`/user/${id}`);
}

/** 获取用户详情 */
export function userInfo(id: number) {
  return requestClient.get<SysUser>(`/user/${id}`);
}

/** 修改用户状态 */
export function userStatusChange(id: number, status: number) {
  return requestClient.put(`/user/${id}/status`, { status });
}

/** 获取当前用户信息 */
export function getProfile() {
  return requestClient.get<SysUser>('/user/profile');
}

/** 更新当前用户信息 */
export function updateProfile(data: {
  nickname?: string;
  email?: string;
  phone?: string;
  sex?: number;
  password?: string;
}) {
  return requestClient.put('/user/profile', data);
}

/** 导出用户列表为 Excel */
export function userExport(params?: { ids?: number[] }) {
  return requestClient.download<Blob>('/user/export', {
    params,
  });
}

/** 导入用户 */
export function userImport(file: File, updateSupport?: boolean) {
  const formData = new FormData();
  formData.append('file', file);
  if (updateSupport) {
    formData.append('updateSupport', '1');
  }
  return requestClient.post<{
    success: number;
    fail: number;
    failList: Array<{ row: number; reason: string }>;
  }>('/user/import', formData);
}

/** 下载导入模板 */
export function userImportTemplate() {
  return requestClient.download<Blob>('/user/import-template');
}

/** 重置用户密码 */
export function userResetPassword(id: number, password: string) {
  return requestClient.put(`/user/${id}/reset-password`, { password });
}

/** 上传头像（通过通用文件上传接口上传，再更新头像URL） */
export async function userUpdateAvatar(fileCallback: {
  file: Blob;
  filename: string;
}) {
  const { file, filename } = fileCallback;
  const uniqueName = filename || `${Date.now()}_${Math.random().toString(36).slice(2, 10)}.png`;
  const uploadFile = new File([file], uniqueName);
  // Step 1: Upload file via generic upload API with scene=avatar
  const formData = new FormData();
  formData.append('file', uploadFile);
  formData.append('scene', 'avatar');
  const uploadResult = await requestClient.post<{ url: string }>(
    '/file/upload',
    formData,
    { headers: { 'Content-Type': 'multipart/form-data' } },
  );
  // Step 2: Update avatar URL
  await requestClient.put('/user/profile/avatar', { avatar: uploadResult.url });
  return { url: uploadResult.url };
}

/** 获取部门树 */
export async function getDeptTree() {
  const res = await requestClient.get<{ list: DeptTree[] }>('/user/dept-tree');
  return res.list;
}
