/**
 * Tree utility functions
 */

/**
 * 遍历树形结构，对每个节点执行回调
 */
export function eachTree<T extends Record<string, any>>(
  tree: T[],
  callback: (node: T, index: number, parent?: T) => void,
  options?: { childProp?: string },
  parent?: T,
): void {
  const childProp = options?.childProp || 'children';
  for (let i = 0; i < tree.length; i++) {
    const node = tree[i];
    if (!node) continue;
    callback(node, i, parent);
    const children = node[childProp];
    if (children && Array.isArray(children) && children.length > 0) {
      eachTree(children as T[], callback, options, node);
    }
  }
}

/**
 * 将列表转换为树形结构
 * 如果节点已有 children 数组且非空，则保留
 */
export function listToTree<T extends Record<string, any>>(
  list: T[],
  options?: { id?: string; pid?: string; childProp?: string },
): T[] {
  const idProp = options?.id || 'id';
  const pidProp = options?.pid || 'parentId';
  const childProp = options?.childProp || 'children';

  const map = new Map<any, T & { [key: string]: any }>();
  const roots: (T & { [key: string]: any })[] = [];

  // 创建映射，保留已存在的 children 数组
  for (const item of list) {
    const existingChildren = item[childProp];
    map.set(item[idProp], {
      ...item,
      [childProp]: existingChildren && Array.isArray(existingChildren) && existingChildren.length > 0
        ? existingChildren
        : [],
    });
  }

  // 构建树
  for (const item of list) {
    const node = map.get(item[idProp]);
    if (!node) continue;
    const parentId = item[pidProp];
    if (parentId === 0 || parentId === null || parentId === undefined) {
      roots.push(node);
    } else {
      const parent = map.get(parentId);
      if (parent) {
        parent[childProp].push(node);
      } else {
        roots.push(node);
      }
    }
  }

  return roots as T[];
}

/**
 * 将树形结构转换为列表
 */
export function treeToList<T extends Record<string, any>>(
  tree: T[],
  options?: { childProp?: string },
): T[] {
  const childProp = options?.childProp || 'children';
  const result: T[] = [];

  const flatten = (nodes: T[]) => {
    for (const node of nodes) {
      if (!node) continue;
      const children = node[childProp] as T[] | undefined;
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { [childProp]: _, ...rest } = node;
      result.push(rest as T);
      if (children && Array.isArray(children) && children.length > 0) {
        flatten(children);
      }
    }
  };

  flatten(tree);
  return result;
}

/**
 * 为树节点添加全路径名称
 */
export function addFullName<T extends Record<string, any>>(
  tree: T[],
  labelField: string = 'label',
  separator: string = ' / ',
  parentName: string = '',
): T[] {
  for (const node of tree) {
    const label = node[labelField];
    (node as any).fullName = parentName ? `${parentName}${separator}${label}` : label;
    const children = node.children;
    if (children && Array.isArray(children) && children.length > 0) {
      addFullName(children, labelField, separator, (node as any).fullName);
    }
  }
  return tree;
}

/**
 * 查找所有父节点ID
 */
export function findGroupParentIds<T extends Record<string, any>>(
  tree: T[],
  targetIds: (number | string)[],
  options?: { idProp?: string; childProp?: string },
): (number | string)[] {
  const idProp = options?.idProp || 'id';
  const childProp = options?.childProp || 'children';
  const result: Set<number | string> = new Set();

  const find = (nodes: T[], targets: Set<number | string>): boolean => {
    let found = false;
    for (const node of nodes) {
      if (!node) continue;
      const id = node[idProp];
      const children = node[childProp] as T[] | undefined;
      const hasTargetChildren = children && Array.isArray(children) && children.length > 0
        ? find(children, targets)
        : false;

      if (targets.has(id) || hasTargetChildren) {
        result.add(id);
        found = true;
      }
    }
    return found;
  };

  find(tree, new Set(targetIds));
  return Array.from(result);
}

/**
 * 获取指定节点及其所有子孙节点的ID列表
 * @param tree 树形数据
 * @param nodeId 目标节点ID
 * @param options 配置选项
 * @returns 包含目标节点及其所有子孙节点的ID数组
 */
export function getDescendantIds<T extends Record<string, any>>(
  tree: T[],
  nodeId: number | string,
  options?: { idProp?: string; childProp?: string },
): (number | string)[] {
  const idProp = options?.idProp || 'id';
  const childProp = options?.childProp || 'children';
  const result: (number | string)[] = [];
  const targetId = String(nodeId); // 统一转为字符串比较

  const findAndCollect = (nodes: T[]): boolean => {
    for (const node of nodes) {
      if (!node) continue;
      const id = node[idProp];
      const children = node[childProp] as T[] | undefined;

      if (String(id) === targetId) {
        // 找到目标节点，收集它及其所有子孙节点
        result.push(id);
        if (children && Array.isArray(children)) {
          collectAllDescendants(children);
        }
        return true;
      }

      // 继续在子节点中查找
      if (children && Array.isArray(children) && children.length > 0) {
        if (findAndCollect(children)) {
          return true;
        }
      }
    }
    return false;
  };

  const collectAllDescendants = (nodes: T[]) => {
    for (const node of nodes) {
      if (!node) continue;
      const id = node[idProp];
      const children = node[childProp] as T[] | undefined;
      result.push(id);
      if (children && Array.isArray(children) && children.length > 0) {
        collectAllDescendants(children);
      }
    }
  };

  findAndCollect(tree);
  return result;
}