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
