import { requestClient } from "#/api/request";

export interface WaterSubmitInput {
  deviceType: string;
  deviceId: string;
  tenant: string;
  image: string;
  deviceCode?: string;
  channelCode?: string;
  deviceIdx?: string;
  callbackUrl?: string;
  url?: string;
}

export interface WaterSubmitResult {
  success: boolean;
  taskId: string;
  status: string;
}

export interface WaterPreviewInput {
  tenant: string;
  deviceId?: string;
  deviceCode?: string;
  channelCode?: string;
  image: string;
}

export interface WaterPreviewResult {
  success: boolean;
  status: string;
  message: string;
  image: string;
  strategyId: number;
  strategyName: string;
  source: string;
  sourceLabel: string;
  durationMs: number;
}

export interface WaterTask {
  taskId: string;
  status: string;
  success: boolean;
  message: string;
  error: string;
  tenant: string;
  deviceId: string;
  strategyId: number;
  strategyName: string;
  source: string;
  sourceLabel: string;
  image: string;
  createdAt: string;
  updatedAt: string;
  durationMs: number;
}

export function submitWaterSnap(data: WaterSubmitInput) {
  const { deviceType, deviceId, ...body } = data;
  return requestClient.post<WaterSubmitResult>(
    `/water/snaps/${encodeURIComponent(deviceType)}/${encodeURIComponent(deviceId)}`,
    body,
  );
}

export function previewWatermark(data: WaterPreviewInput) {
  return requestClient.post<WaterPreviewResult>("/water/preview", data);
}

export function getWaterTask(taskId: string) {
  return requestClient.get<WaterTask>(
    `/water/tasks/${encodeURIComponent(taskId)}`,
  );
}
