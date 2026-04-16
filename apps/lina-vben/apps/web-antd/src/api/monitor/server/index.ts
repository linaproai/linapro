import type { ServerMonitorParams, ServerMonitorResult } from './model';

import { requestClient } from '#/api/request';

/** 服务监控数据 */
export function getServerMonitor(params?: ServerMonitorParams) {
  return requestClient.get<ServerMonitorResult>('/monitor/server', { params });
}
