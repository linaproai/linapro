import type {
  PluginAuthorizationPayload,
  PluginListParams,
  PluginDynamicState,
  PluginUploadDynamicResult,
  SystemPlugin,
} from './model';

import { requestClient } from '#/api/request';

type RuntimeEnvelope<T> = {
  code: number;
  data: T;
  errorCode?: string;
  message?: string;
  messageKey?: string;
  messageParams?: Record<string, unknown>;
};

/** 插件列表 */
export async function pluginList(params?: PluginListParams) {
  const res = await requestClient.get<{ list: SystemPlugin[]; total: number }>(
    '/plugins',
    { params },
  );
  return { items: res.list, total: res.total };
}

/** 公共插件运行时状态 */
export async function pluginDynamicList() {
  const res = await requestClient.get<{ list: PluginDynamicState[] }>(
    '/plugins/dynamic',
  );
  return res.list;
}

/** 同步源码插件 */
export function pluginSync() {
  return requestClient.post<{ total: number }>('/plugins/sync');
}

/** 安装插件 */
export function pluginInstall(
  pluginId: string,
  payload?: PluginAuthorizationPayload,
) {
  return requestClient.post(`/plugins/${pluginId}/install`, payload);
}

/** 上传动态插件 */
export function pluginDynamicUpload(file: File, overwriteSupport?: boolean) {
  const formData = new FormData();
  // Keep the original filename in multipart payload. The backend validates the
  // `.wasm` suffix from the uploaded filename before parsing the artifact.
  formData.append('file', file, file.name);
  if (overwriteSupport) {
    formData.append('overwriteSupport', '1');
  }
  return requestClient.post<PluginUploadDynamicResult>(
    '/plugins/dynamic/package',
    formData,
    {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    },
  );
}

/** 启用插件 */
export function pluginEnable(
  pluginId: string,
  payload?: PluginAuthorizationPayload,
) {
  return requestClient.put(`/plugins/${pluginId}/enable`, payload);
}

/** 禁用插件 */
export function pluginDisable(pluginId: string) {
  return requestClient.put(`/plugins/${pluginId}/disable`);
}

/** 更新插件新租户自动启用策略 */
export function pluginUpdateTenantProvisioningPolicy(
  pluginId: string,
  autoEnableForNewTenants: boolean,
) {
  return requestClient.put(`/plugins/${pluginId}/tenant-provisioning-policy`, {
    autoEnableForNewTenants,
  });
}

/** 卸载插件 */
export async function pluginUninstall(
  pluginId: string,
  options?: { force?: boolean; purgeStorageData?: boolean },
) {
  const params: Record<string, boolean | number> = {};
  if (typeof options?.purgeStorageData === 'boolean') {
    params.purgeStorageData = options.purgeStorageData ? 1 : 0;
  }
  if (options?.force) {
    params.force = true;
  }
  const res = await requestClient.delete<RuntimeEnvelope<unknown>>(
    `/plugins/${pluginId}`,
    {
      params: Object.keys(params).length > 0 ? params : undefined,
      responseReturn: 'body',
    },
  );
  if (!res || res.code !== 0) {
    throw res;
  }
  return res.data;
}
