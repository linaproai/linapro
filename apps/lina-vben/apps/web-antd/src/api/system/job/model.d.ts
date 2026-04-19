export type JobTaskType = 'handler' | 'shell' | string;

export type JobStatus = 'disabled' | 'enabled' | 'paused_by_plugin' | string;

export type JobScope = 'all_node' | 'master_only' | string;

export type JobConcurrency = 'parallel' | 'singleton' | string;

export type JobTrigger = 'cron' | 'manual' | string;

export type JobLogStatus =
  | 'cancelled'
  | 'failed'
  | 'running'
  | 'skipped_max_concurrency'
  | 'skipped_not_primary'
  | 'skipped_singleton'
  | 'success'
  | 'timeout'
  | string;

export interface JobRetentionOption {
  mode: 'count' | 'days' | 'none' | string;
  value: number;
}

export interface JobRecord {
  id: number;
  groupId: number;
  name: string;
  description: string;
  taskType: JobTaskType;
  handlerRef: string;
  params: string;
  timeoutSeconds: number;
  shellCmd: string;
  workDir: string;
  env: string;
  cronExpr: string;
  timezone: string;
  scope: JobScope;
  concurrency: JobConcurrency;
  maxConcurrency: number;
  maxExecutions: number;
  executedCount: number;
  stopReason: string;
  logRetentionOverride: string;
  status: JobStatus;
  isBuiltin: number;
  seedVersion: number;
  createdBy: number;
  updatedBy: number;
  createdAt: string;
  updatedAt: string;
  groupCode: string;
  groupName: string;
}

export interface JobListParams {
  pageNum?: number;
  pageSize?: number;
  groupId?: number;
  status?: string;
  taskType?: string;
  keyword?: string;
  scope?: string;
  concurrency?: string;
  orderBy?: string;
  orderDirection?: string;
}

export interface JobListResult {
  items: JobRecord[];
  total: number;
}

export interface JobPayload {
  groupId: number;
  name: string;
  description?: string;
  taskType: JobTaskType;
  handlerRef?: string;
  params?: Record<string, any>;
  timeoutSeconds: number;
  shellCmd?: string;
  workDir?: string;
  env?: Record<string, string>;
  cronExpr: string;
  timezone: string;
  scope: JobScope;
  concurrency: JobConcurrency;
  maxConcurrency: number;
  maxExecutions: number;
  status: Exclude<JobStatus, 'paused_by_plugin'>;
  logRetentionOverride?: JobRetentionOption | null;
}

export interface JobTriggerResult {
  logId: number;
}

export interface JobCronPreviewResult {
  times: string[];
}

export interface JobHandlerOption {
  ref: string;
  displayName: string;
  description: string;
  source: 'host' | 'plugin' | string;
  pluginId: string;
}

export interface JobHandlerDetail extends JobHandlerOption {
  paramsSchema: string;
}

export interface JobLogRecord {
  id: number;
  jobId: number;
  jobSnapshot: string;
  nodeId: string;
  trigger: JobTrigger;
  paramsSnapshot: string;
  startAt: string;
  endAt: string;
  durationMs: number;
  status: JobLogStatus;
  errMsg: string;
  resultJson: string;
  createdAt: string;
  jobName: string;
}

export interface JobLogListParams {
  pageNum?: number;
  pageSize?: number;
  jobId?: number;
  status?: string;
  trigger?: string;
  nodeId?: string;
  beginTime?: string;
  endTime?: string;
  orderBy?: string;
  orderDirection?: string;
}

export interface JobLogListResult {
  items: JobLogRecord[];
  total: number;
}
