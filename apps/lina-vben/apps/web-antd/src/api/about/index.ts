import { requestClient } from '#/api/request';

export interface ComponentInfo {
  name: string;
  version: string;
  url: string;
  description: string;
}

export interface SystemInfoResult {
  goVersion: string;
  gfVersion: string;
  os: string;
  arch: string;
  dbVersion: string;
  startTime: string;
  runDuration: string;
  backendComponents: ComponentInfo[];
  frontendComponents: ComponentInfo[];
}

/** 获取系统运行信息 */
export function getSystemInfo() {
  return requestClient.get<SystemInfoResult>('/system/info');
}
