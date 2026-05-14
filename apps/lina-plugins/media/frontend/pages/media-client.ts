import { requestClient } from "#/api/request";

export interface MediaStrategy {
  id: number;
  name: string;
  strategy: string;
  global: number;
  enable: number;
  creatorId: number;
  updaterId: number;
  createTime: string;
  updateTime: string;
}

export interface MediaStrategyListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  enable?: number;
  global?: number;
}

export interface MediaStrategyInput {
  name: string;
  strategy: string;
  enable: number;
  global: number;
}

export interface MediaDeviceBinding {
  rowKey: string;
  deviceId: string;
  strategyId: number;
  strategyName: string;
}

export interface MediaTenantBinding {
  rowKey: string;
  tenantId: string;
  strategyId: number;
  strategyName: string;
}

export interface MediaTenantDeviceBinding {
  rowKey: string;
  tenantId: string;
  deviceId: string;
  strategyId: number;
  strategyName: string;
}

export type MediaBindingKind = "device" | "tenant" | "tenantDevice";

export interface MediaBindingListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
}

export interface MediaResolveParams {
  tenantId?: string;
  deviceId?: string;
}

export interface MediaResolveResult {
  matched: boolean;
  source: string;
  sourceLabel: string;
  strategyId: number;
  strategyName: string;
  strategy: string;
}

export interface MediaAlias {
  id: number;
  alias: string;
  autoRemove: number;
  streamPath: string;
  createTime: string;
}

export interface MediaAliasListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
}

export interface MediaAliasInput {
  alias: string;
  autoRemove: number;
  streamPath: string;
}

export interface MediaTenantWhite {
  rowKey?: string;
  tenantId: string;
  ip: string;
  description: string;
  enable: number;
  creatorId: number;
  createTime: string;
  updaterId: number;
  updateTime: string;
}

export interface MediaTenantWhiteListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  enable?: number;
}

export interface MediaTenantWhiteInput {
  tenantId: string;
  ip: string;
  description: string;
  enable: number;
}

export interface MediaNode {
  id: number;
  nodeNum: number;
  name: string;
  qnUrl: string;
  basicUrl: string;
  dnUrl: string;
  creatorId: number;
  createTime: string;
  updaterId: number;
  updateTime: string;
}

export interface MediaNodeListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
}

export interface MediaNodeInput {
  nodeNum: number;
  name: string;
  qnUrl: string;
  basicUrl: string;
  dnUrl: string;
}

export interface MediaDeviceNode {
  deviceId: string;
  nodeNum: number;
  nodeName: string;
}

export interface MediaDeviceNodeListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
}

export interface MediaDeviceNodeInput {
  deviceId: string;
  nodeNum: number;
}

export interface MediaTenantStreamConfig {
  tenantId: string;
  maxConcurrent: number;
  nodeNum: number;
  nodeName: string;
  enable: number;
  creatorId: number;
  createTime: string;
  updaterId: number;
  updateTime: string;
}

export interface MediaTenantStreamConfigListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  enable?: number;
}

export interface MediaTenantStreamConfigInput {
  tenantId: string;
  maxConcurrent: number;
  nodeNum: number;
  enable: number;
}

export async function listMediaStrategies(params?: MediaStrategyListParams) {
  const res = await requestClient.get<{
    list: MediaStrategy[];
    total: number;
  }>("/media/strategies", { params });
  return { items: res.list, total: res.total };
}

export function getMediaStrategy(id: number) {
  return requestClient.get<MediaStrategy>(`/media/strategies/${id}`);
}

export function createMediaStrategy(data: MediaStrategyInput) {
  return requestClient.post<{ id: number }>("/media/strategies", data);
}

export function updateMediaStrategy(id: number, data: MediaStrategyInput) {
  return requestClient.put<{ id: number }>(`/media/strategies/${id}`, data);
}

export function deleteMediaStrategy(id: number) {
  return requestClient.delete(`/media/strategies/${id}`);
}

export function setGlobalMediaStrategy(id: number) {
  return requestClient.put<{ id: number }>(`/media/strategies/${id}/global`);
}

export function updateMediaStrategyEnable(id: number, enable: number) {
  return requestClient.put<{ id: number }>(`/media/strategies/${id}/enable`, {
    enable,
  });
}

function encodePathSegment(value: string) {
  return encodeURIComponent(value);
}

export async function listMediaDeviceBindings(params?: MediaBindingListParams) {
  const res = await requestClient.get<{
    list: MediaDeviceBinding[];
    total: number;
  }>("/media/device-bindings", { params });
  return { items: res.list, total: res.total };
}

export function saveMediaDeviceBinding(deviceId: string, strategyId: number) {
  return requestClient.put<MediaDeviceBinding>(
    `/media/device-bindings/${encodePathSegment(deviceId)}`,
    { deviceId, strategyId },
  );
}

export function deleteMediaDeviceBinding(deviceId: string) {
  return requestClient.delete(
    `/media/device-bindings/${encodePathSegment(deviceId)}`,
  );
}

export async function listMediaTenantBindings(params?: MediaBindingListParams) {
  const res = await requestClient.get<{
    list: MediaTenantBinding[];
    total: number;
  }>("/media/tenant-bindings", { params });
  return { items: res.list, total: res.total };
}

export function saveMediaTenantBinding(
  tenantId: string,
  strategyId: number,
) {
  return requestClient.put<MediaTenantBinding>(
    `/media/tenant-bindings/${encodePathSegment(tenantId)}`,
    { tenantId, strategyId },
  );
}

export function deleteMediaTenantBinding(tenantId: string) {
  return requestClient.delete(
    `/media/tenant-bindings/${encodePathSegment(tenantId)}`,
  );
}

export async function listMediaTenantDeviceBindings(
  params?: MediaBindingListParams,
) {
  const res = await requestClient.get<{
    list: MediaTenantDeviceBinding[];
    total: number;
  }>("/media/tenant-device-bindings", { params });
  return { items: res.list, total: res.total };
}

export function saveMediaTenantDeviceBinding(
  tenantId: string,
  deviceId: string,
  strategyId: number,
) {
  return requestClient.put<MediaTenantDeviceBinding>(
    `/media/tenant-device-bindings/${encodePathSegment(
      tenantId,
    )}/${encodePathSegment(deviceId)}`,
    { tenantId, deviceId, strategyId },
  );
}

export function deleteMediaTenantDeviceBinding(
  tenantId: string,
  deviceId: string,
) {
  return requestClient.delete(
    `/media/tenant-device-bindings/${encodePathSegment(
      tenantId,
    )}/${encodePathSegment(deviceId)}`,
  );
}

export function resolveMediaStrategy(params: MediaResolveParams) {
  return requestClient.get<MediaResolveResult>("/media/strategies/resolve", {
    params,
  });
}

export async function listMediaAliases(params?: MediaAliasListParams) {
  const res = await requestClient.get<{ list: MediaAlias[]; total: number }>(
    "/media/stream-aliases",
    { params },
  );
  return { items: res.list, total: res.total };
}

export function getMediaAlias(id: number) {
  return requestClient.get<MediaAlias>(`/media/stream-aliases/${id}`);
}

export function createMediaAlias(data: MediaAliasInput) {
  return requestClient.post<{ id: number }>("/media/stream-aliases", data);
}

export function updateMediaAlias(id: number, data: MediaAliasInput) {
  return requestClient.put<{ id: number }>(`/media/stream-aliases/${id}`, data);
}

export function deleteMediaAlias(id: number) {
  return requestClient.delete(`/media/stream-aliases/${id}`);
}

export async function listMediaTenantWhites(
  params?: MediaTenantWhiteListParams,
) {
  const res = await requestClient.get<{
    list: MediaTenantWhite[];
    total: number;
  }>("/media/tenant-whites", { params });
  return {
    items: res.list.map((item) => ({
      ...item,
      rowKey: `${item.tenantId}:${item.ip}`,
    })),
    total: res.total,
  };
}

export function getMediaTenantWhite(tenantId: string, ip: string) {
  return requestClient.get<MediaTenantWhite>(
    `/media/tenant-whites/${encodePathSegment(
      tenantId,
    )}/${encodePathSegment(ip)}`,
  );
}

export function createMediaTenantWhite(data: MediaTenantWhiteInput) {
  return requestClient.post<{ tenantId: string; ip: string }>(
    "/media/tenant-whites",
    data,
  );
}

export function updateMediaTenantWhite(
  oldTenantId: string,
  oldIp: string,
  data: MediaTenantWhiteInput,
) {
  return requestClient.put<{ tenantId: string; ip: string }>(
    `/media/tenant-whites/${encodePathSegment(
      oldTenantId,
    )}/${encodePathSegment(oldIp)}`,
    data,
  );
}

export function deleteMediaTenantWhite(tenantId: string, ip: string) {
  return requestClient.delete(
    `/media/tenant-whites/${encodePathSegment(
      tenantId,
    )}/${encodePathSegment(ip)}`,
  );
}

export async function listMediaNodes(params?: MediaNodeListParams) {
  const res = await requestClient.get<{ list: MediaNode[]; total: number }>(
    "/media/nodes",
    { params },
  );
  return { items: res.list, total: res.total };
}

export function getMediaNode(nodeNum: number) {
  return requestClient.get<MediaNode>(`/media/nodes/${nodeNum}`);
}

export function createMediaNode(data: MediaNodeInput) {
  return requestClient.post<{ nodeNum: number }>("/media/nodes", data);
}

export function updateMediaNode(oldNodeNum: number, data: MediaNodeInput) {
  return requestClient.put<{ nodeNum: number }>(
    `/media/nodes/${oldNodeNum}`,
    data,
  );
}

export function deleteMediaNode(nodeNum: number) {
  return requestClient.delete(`/media/nodes/${nodeNum}`);
}

export async function listMediaDeviceNodes(
  params?: MediaDeviceNodeListParams,
) {
  const res = await requestClient.get<{
    list: MediaDeviceNode[];
    total: number;
  }>("/media/device-nodes", { params });
  return { items: res.list, total: res.total };
}

export function getMediaDeviceNode(deviceId: string) {
  return requestClient.get<MediaDeviceNode>(
    `/media/device-nodes/${encodePathSegment(deviceId)}`,
  );
}

export function createMediaDeviceNode(data: MediaDeviceNodeInput) {
  return requestClient.post<{ deviceId: string }>(
    "/media/device-nodes",
    data,
  );
}

export function updateMediaDeviceNode(
  oldDeviceId: string,
  data: MediaDeviceNodeInput,
) {
  return requestClient.put<{ deviceId: string }>(
    `/media/device-nodes/${encodePathSegment(oldDeviceId)}`,
    data,
  );
}

export function deleteMediaDeviceNode(deviceId: string) {
  return requestClient.delete(
    `/media/device-nodes/${encodePathSegment(deviceId)}`,
  );
}

export async function listMediaTenantStreamConfigs(
  params?: MediaTenantStreamConfigListParams,
) {
  const res = await requestClient.get<{
    list: MediaTenantStreamConfig[];
    total: number;
  }>("/media/tenant-stream-configs", { params });
  return { items: res.list, total: res.total };
}

export function getMediaTenantStreamConfig(tenantId: string) {
  return requestClient.get<MediaTenantStreamConfig>(
    `/media/tenant-stream-configs/${encodePathSegment(tenantId)}`,
  );
}

export function createMediaTenantStreamConfig(
  data: MediaTenantStreamConfigInput,
) {
  return requestClient.post<{ tenantId: string }>(
    "/media/tenant-stream-configs",
    data,
  );
}

export function updateMediaTenantStreamConfig(
  oldTenantId: string,
  data: MediaTenantStreamConfigInput,
) {
  return requestClient.put<{ tenantId: string }>(
    `/media/tenant-stream-configs/${encodePathSegment(oldTenantId)}`,
    data,
  );
}

export function deleteMediaTenantStreamConfig(tenantId: string) {
  return requestClient.delete(
    `/media/tenant-stream-configs/${encodePathSegment(tenantId)}`,
  );
}
