import type {
  PluginAuthorizationPayload,
  PluginListParams,
  PluginDynamicState,
  PluginUploadDynamicResult,
  SystemPlugin,
} from './model';

import { requestClient } from '#/api/request';

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

/** 卸载插件 */
export function pluginUninstall(pluginId: string) {
  return requestClient.delete(`/plugins/${pluginId}`);
}
