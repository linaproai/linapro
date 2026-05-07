import {
  expect,
  request as playwrightRequest,
  type APIRequestContext,
  type APIResponse,
} from "@playwright/test";

import { config } from "../../fixtures/config";

const apiBaseURL =
  process.env.E2E_API_BASE_URL ??
  new URL("/api/v1/", config.baseURL).toString();

export type ApiEnvelope<T> = {
  code: number;
  data: T;
  message?: string;
};

export type GroupListItem = {
  id: number;
  code: string;
  name: string;
  remark?: string;
  sortOrder: number;
  isDefault: number;
  jobCount: number;
};

export type JobDetail = {
  id: number;
  groupId: number;
  name: string;
  description: string;
  taskType: string;
  handlerRef: string;
  params: string;
  timeoutSeconds: number;
  shellCmd?: string;
  workDir?: string;
  env?: string;
  cronExpr: string;
  timezone: string;
  scope: string;
  concurrency: string;
  maxConcurrency: number;
  maxExecutions: number;
  executedCount: number;
  stopReason: string;
  status: string;
  logRetentionOverride?: string;
  isBuiltin?: number;
  seedVersion?: number;
  groupCode: string;
  groupName: string;
};

export type LogDetail = {
  id: number;
  jobId: number;
  trigger: string;
  status: string;
  errMsg?: string;
  resultJson?: string;
  jobName?: string;
};

export type HandlerListItem = {
  ref: string;
  displayName: string;
  description: string;
  source: string;
  pluginId: string;
};

export type ConfigItem = {
  id: number;
  key: string;
  value: string;
};

export type PluginItem = {
  id: string;
  name: string;
  description: string;
  version: string;
  type: string;
  installed: number;
  enabled: number;
};

export type MenuNode = {
  id: number;
  name: string;
  perms: string;
  children?: MenuNode[];
};

export type AccessibleMenuNode = {
  name?: string;
  path?: string;
  meta?: {
    title?: string;
  };
  children?: AccessibleMenuNode[];
};

function flattenMenus(list: MenuNode[]): MenuNode[] {
  return list.flatMap((item) => [item, ...flattenMenus(item.children ?? [])]);
}

function menuTreeHasPermission(node: MenuNode, permission: string): boolean {
  return (
    node.perms === permission ||
    Boolean(node.children?.some((child) => menuTreeHasPermission(child, permission)))
  );
}

export async function createApiContext(
  username: string,
  password: string,
): Promise<APIRequestContext> {
  const loginApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const loginResponse = await loginApi.post("auth/login", {
    data: {
      username,
      password,
    },
  });
  const loginPayload = await expectSuccess<{ accessToken: string }>(
    loginResponse,
  );
  await loginApi.dispose();

  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${loginPayload.accessToken}`,
    },
  });
}

export async function createAdminApiContext(): Promise<APIRequestContext> {
  return createApiContext(config.adminUser, config.adminPass);
}

export async function expectSuccess<T>(response: APIResponse): Promise<T> {
  expect(response.ok()).toBeTruthy();
  const payload = (await response.json()) as ApiEnvelope<T>;
  expect(payload.code).toBe(0);
  return payload.data;
}

export async function expectBusinessError(
  response: APIResponse,
  messageIncludes?: string,
) {
  const payload = (await response.json()) as ApiEnvelope<unknown>;
  expect(payload.code).not.toBe(0);
  if (messageIncludes) {
    expect(payload.message ?? "").toContain(messageIncludes);
  }
  return payload;
}

export async function listGroups(api: APIRequestContext, keyword = "") {
  return expectSuccess<{ list: GroupListItem[]; total: number }>(
    await api.get(
      `job-group?pageNum=1&pageSize=100&code=${encodeURIComponent(keyword)}`,
    ),
  );
}

export async function getDefaultGroup(
  api: APIRequestContext,
): Promise<GroupListItem> {
  const result = await listGroups(api);
  const group = result.list.find((item) => item.isDefault === 1);
  expect(group).toBeTruthy();
  return group!;
}

export async function createGroup(
  api: APIRequestContext,
  payload: { code: string; name: string; remark?: string; sortOrder?: number },
) {
  return expectSuccess<{ id: number }>(
    await api.post("job-group", { data: payload }),
  );
}

export async function updateGroup(
  api: APIRequestContext,
  id: number,
  payload: { code: string; name: string; remark?: string; sortOrder?: number },
) {
  await expectSuccess(await api.put(`job-group/${id}`, { data: payload }));
}

export async function deleteGroup(api: APIRequestContext, id: number) {
  await expectSuccess(await api.delete(`job-group/${id}`));
}

export async function listHandlers(api: APIRequestContext) {
  return expectSuccess<{ list: HandlerListItem[] }>(
    await api.get("job/handler"),
  );
}

export async function getMenuIdsByPerms(
  api: APIRequestContext,
  perms: string[],
) {
  const result = await expectSuccess<{ list: MenuNode[] }>(
    await api.get("menu"),
  );
  const flatMenus = flattenMenus(result.list);

  return perms.map((permission) => {
    const menu = flatMenus.find((item) => item.perms === permission);
    expect(menu, `missing menu permission: ${permission}`).toBeTruthy();
    return menu!.id;
  });
}

export async function getMenuIdsByPermsWithAncestors(
  api: APIRequestContext,
  perms: string[],
) {
  const result = await expectSuccess<{ list: MenuNode[] }>(
    await api.get("menu"),
  );
  const requiredPerms = new Set(perms);
  const selectedIds = new Set<number>();

  function visit(node: MenuNode, ancestors: number[]) {
    const nextAncestors = [...ancestors, node.id];
    if (requiredPerms.has(node.perms)) {
      nextAncestors.forEach((id) => selectedIds.add(id));
    }
    node.children?.forEach((child) => visit(child, nextAncestors));
  }

  result.list.forEach((node) => visit(node, []));
  for (const permission of requiredPerms) {
    expect(
      result.list.some((node) => menuTreeHasPermission(node, permission)),
      `missing menu permission: ${permission}`,
    ).toBeTruthy();
  }

  return [...selectedIds];
}

export async function getAccessibleMenus(api: APIRequestContext) {
  return expectSuccess<{ list: AccessibleMenuNode[] }>(
    await api.get("menus/all"),
  );
}

export async function getHandlerDetail(api: APIRequestContext, ref: string) {
  return expectSuccess<{
    ref: string;
    displayName: string;
    description: string;
    source: string;
    pluginId: string;
    paramsSchema: string;
  }>(await api.get(`job/handler/${encodeURIComponent(ref)}`));
}

export async function syncPlugins(api: APIRequestContext) {
  return expectSuccess<{ total: number }>(await api.post("plugins/sync"));
}

export async function listPlugins(
  api: APIRequestContext,
  id = "",
  lang?: string,
) {
  const params = new URLSearchParams();
  params.set("id", id);
  if (lang) {
    params.set("lang", lang);
  }
  return expectSuccess<{ list: PluginItem[]; total: number }>(
    await api.get(`plugins?${params.toString()}`),
  );
}

export async function getPlugin(api: APIRequestContext, id: string) {
  const result = await listPlugins(api, id);
  const item = result.list.find((plugin) => plugin.id === id);
  expect(item, `missing plugin: ${id}`).toBeTruthy();
  return item!;
}

export async function installPlugin(api: APIRequestContext, id: string) {
  return expectSuccess<{ id: string; installed: number; enabled: number }>(
    await api.post(`plugins/${id}/install`),
  );
}

export async function enablePlugin(api: APIRequestContext, id: string) {
  return expectSuccess<{ id: string; enabled: number }>(
    await api.put(`plugins/${id}/enable`),
  );
}

export async function disablePlugin(api: APIRequestContext, id: string) {
  return expectSuccess<{ id: string; enabled: number }>(
    await api.put(`plugins/${id}/disable`),
  );
}

export async function uninstallPlugin(api: APIRequestContext, id: string) {
  await expectSuccess(await api.delete(`plugins/${id}`));
}

export function buildHandlerJobPayload(
  overrides: Partial<Record<string, unknown>> = {},
) {
  return {
    groupId: 1,
    name: `e2e-job-${Date.now()}`,
    description: "E2E scheduled job",
    taskType: "handler",
    handlerRef: "host:cleanup-job-logs",
    params: {},
    timeoutSeconds: 300,
    cronExpr: "0 0 1 1 *",
    timezone: "Asia/Shanghai",
    scope: "master_only",
    concurrency: "singleton",
    maxConcurrency: 1,
    maxExecutions: 0,
    status: "disabled",
    ...overrides,
  };
}

export function buildShellJobPayload(
  overrides: Partial<Record<string, unknown>> = {},
) {
  return {
    groupId: 1,
    name: `e2e-shell-job-${Date.now()}`,
    description: "E2E shell scheduled job",
    taskType: "shell",
    handlerRef: "",
    params: {},
    timeoutSeconds: 300,
    shellCmd: "printf 'hello from shell'",
    workDir: "",
    env: {},
    cronExpr: "0 0 1 1 *",
    timezone: "Asia/Shanghai",
    scope: "master_only",
    concurrency: "singleton",
    maxConcurrency: 1,
    maxExecutions: 0,
    status: "disabled",
    ...overrides,
  };
}

export async function createJob(
  api: APIRequestContext,
  payload: Record<string, unknown>,
) {
  return expectSuccess<{ id: number }>(
    await api.post("job", { data: payload }),
  );
}

export async function updateJob(
  api: APIRequestContext,
  id: number,
  payload: Record<string, unknown>,
) {
  await expectSuccess(await api.put(`job/${id}`, { data: payload }));
}

export async function getJob(api: APIRequestContext, id: number) {
  return expectSuccess<JobDetail>(await api.get(`job/${id}`));
}

export async function listJobs(api: APIRequestContext, keyword = "") {
  return expectSuccess<{ list: JobDetail[]; total: number }>(
    await api.get(
      `job?pageNum=1&pageSize=100&keyword=${encodeURIComponent(keyword)}`,
    ),
  );
}

export async function updateJobStatus(
  api: APIRequestContext,
  id: number,
  status: "enabled" | "disabled",
) {
  await expectSuccess(await api.put(`job/${id}/status`, { data: { status } }));
}

export async function deleteJob(api: APIRequestContext, id: number) {
  await expectSuccess(await api.delete(`job/${id}`));
}

export async function triggerJob(api: APIRequestContext, id: number) {
  return expectSuccess<{ logId: number }>(await api.post(`job/${id}/trigger`));
}

export async function listLogs(api: APIRequestContext, jobId?: number) {
  const query = jobId
    ? `?pageNum=1&pageSize=100&jobId=${jobId}`
    : "?pageNum=1&pageSize=100";
  return expectSuccess<{ list: LogDetail[]; total: number }>(
    await api.get(`job/log${query}`),
  );
}

export async function getLog(api: APIRequestContext, id: number) {
  return expectSuccess<LogDetail>(await api.get(`job/log/${id}`));
}

export async function cancelLog(api: APIRequestContext, id: number) {
  await expectSuccess(await api.post(`job/log/${id}/cancel`));
}

export async function clearLogs(api: APIRequestContext, jobId?: number) {
  const query = jobId ? `?jobId=${jobId}` : "";
  await expectSuccess(await api.delete(`job/log${query}`));
}

export async function createRole(
  api: APIRequestContext,
  payload: {
    name: string;
    key: string;
    menuIds: number[];
    sort?: number;
    dataScope?: number;
    status?: number;
    remark?: string;
  },
) {
  return expectSuccess<{ id: number }>(
    await api.post("role", {
      data: {
        sort: 900,
        dataScope: 1,
        status: 1,
        remark: "",
        ...payload,
      },
    }),
  );
}

export async function deleteRole(api: APIRequestContext, id: number) {
  await expectSuccess(await api.delete(`role/${id}`));
}

export async function createUser(
  api: APIRequestContext,
  payload: {
    username: string;
    password: string;
    nickname: string;
    deptId?: number;
    roleIds?: number[];
  },
) {
  return expectSuccess<{ id: number }>(
    await api.post("user", {
      data: payload,
    }),
  );
}

export async function deleteUser(api: APIRequestContext, id: number) {
  await expectSuccess(await api.delete(`user/${id}`));
}

export async function previewCron(
  api: APIRequestContext,
  expr: string,
  timezone: string,
) {
  return expectSuccess<{ times: string[] }>(
    await api.get(
      `job/cron-preview?expr=${encodeURIComponent(expr)}&timezone=${encodeURIComponent(timezone)}`,
    ),
  );
}

export async function getConfigByKey(api: APIRequestContext, key: string) {
  return expectSuccess<ConfigItem>(
    await api.get(`config/key/${encodeURIComponent(key)}`),
  );
}

export async function updateConfigValue(
  api: APIRequestContext,
  id: number,
  value: string,
) {
  await expectSuccess(await api.put(`config/${id}`, { data: { value } }));
}

export async function setCronShellEnabled(
  api: APIRequestContext,
  enabled: boolean,
) {
  const item = await getConfigByKey(api, "cron.shell.enabled");
  const targetValue = enabled ? "true" : "false";
  if (item.value !== targetValue) {
    await updateConfigValue(api, item.id, targetValue);
  }
  return item;
}

export function buildPayloadFromJob(
  detail: JobDetail,
): Record<string, unknown> {
  return {
    groupId: detail.groupId,
    name: detail.name,
    description: detail.description,
    taskType: detail.taskType,
    handlerRef: detail.handlerRef,
    params: detail.params ? JSON.parse(detail.params) : {},
    timeoutSeconds: detail.timeoutSeconds,
    shellCmd: detail.shellCmd ?? "",
    workDir: detail.workDir ?? "",
    env: detail.env ? JSON.parse(detail.env) : {},
    cronExpr: detail.cronExpr,
    timezone: detail.timezone,
    scope: detail.scope,
    concurrency: detail.concurrency,
    maxConcurrency: detail.maxConcurrency,
    maxExecutions: detail.maxExecutions,
    status: detail.status,
    logRetentionOverride: detail.logRetentionOverride
      ? JSON.parse(detail.logRetentionOverride)
      : undefined,
  };
}
