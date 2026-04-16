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

export interface PostListResult {
  items: Post[];
  total: number;
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
