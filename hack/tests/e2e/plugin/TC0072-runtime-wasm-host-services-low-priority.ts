import { execFileSync } from "node:child_process";
import { mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";

import type { APIRequestContext, APIResponse } from "@playwright/test";
import { request as playwrightRequest } from "@playwright/test";

import { expect, test } from "../../fixtures/auth";
import { config } from "../../fixtures/config";

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? "http://127.0.0.1:8080/api/v1/";
const mysqlBin = process.env.E2E_MYSQL_BIN ?? "mysql";
const mysqlUser = process.env.E2E_DB_USER ?? "root";
const mysqlPassword = process.env.E2E_DB_PASSWORD ?? "12345678";
const mysqlDatabase = process.env.E2E_DB_NAME ?? "lina";

const successPluginID = "lp-host-e2e";
const deniedPluginID = "lp-host-denied-e2e";

type PluginListItem = {
  id: string;
  enabled?: number;
  installed?: number;
};

function repoRoot() {
  return path.resolve(process.cwd(), "../..");
}

function tempRoot() {
  return path.join(repoRoot(), "temp", "e2e-low-priority-host-services");
}

function sourceRoot() {
  return path.join(tempRoot(), "plugins");
}

function buildOutputDir() {
  return path.join(tempRoot(), "artifacts");
}

function builderDir() {
  return path.join(repoRoot(), "hack", "build-wasm");
}

function runtimeStorageDir() {
  return path.join(repoRoot(), "temp", "output");
}

function sourcePluginDir(pluginID: string) {
  return path.join(sourceRoot(), pluginID);
}

function builtArtifactPath(pluginID: string) {
  return path.join(buildOutputDir(), `${pluginID}.wasm`);
}

function uploadedArtifactPath(pluginID: string) {
  return path.join(runtimeStorageDir(), `${pluginID}.wasm`);
}

function writeTestFile(filePath: string, content: string) {
  mkdirSync(path.dirname(filePath), { recursive: true });
  writeFileSync(filePath, content);
}

function assertOk(response: APIResponse, message: string) {
  expect(response.ok(), `${message}, status=${response.status()}`).toBeTruthy();
}

async function expectApiSuccess<T = any>(
  response: APIResponse,
  message: string,
): Promise<T> {
  assertOk(response, message);

  const payload = (await response.json()) as {
    code?: number;
    data?: T;
    message?: string;
  };
  expect(
    payload?.code,
    `${message}, business code=${payload?.code}, business message=${payload?.message ?? ""}`,
  ).toBe(0);
  return (payload?.data ?? null) as T;
}

async function expectApiFailure(
  response: APIResponse,
  message: string,
  expectedText: string,
) {
  const text = await response.text();
  if (response.ok()) {
    const payload = JSON.parse(text) as {
      code?: number;
      message?: string;
    };
    expect(payload?.code, `${message}, body=${text}`).not.toBe(0);
    expect(text, `${message}, body=${text}`).toContain(expectedText);
    return;
  }
  expect(text, `${message}, status=${response.status()}`).toContain(expectedText);
}

async function createAdminApiContext(): Promise<APIRequestContext> {
  const anonymousApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const loginResponse = await anonymousApi.post("auth/login", {
    data: {
      username: config.adminUser,
      password: config.adminPass,
    },
  });
  const loginPayload = await expectApiSuccess<{ accessToken?: string }>(
    loginResponse,
    "管理员登录失败",
  );
  await anonymousApi.dispose();

  expect(loginPayload?.accessToken, "管理员登录后应返回 accessToken").toBeTruthy();
  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${loginPayload.accessToken as string}`,
    },
  });
}

async function apiLogin(username: string, password: string): Promise<string> {
  const anonymousApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  try {
    const loginResponse = await anonymousApi.post("auth/login", {
      data: {
        username,
        password,
      },
    });
    const loginPayload = await expectApiSuccess<{ accessToken?: string }>(
      loginResponse,
      `登录失败: ${username}`,
    );
    expect(loginPayload?.accessToken).toBeTruthy();
    return loginPayload.accessToken as string;
  } finally {
    await anonymousApi.dispose();
  }
}

async function apiUnreadCount(token: string): Promise<number> {
  const api = await playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${token}`,
    },
  });
  try {
    const response = await api.get("user/message/count");
    const payload = await expectApiSuccess<{ count?: number }>(
      response,
      "查询未读消息数失败",
    );
    return payload?.count ?? 0;
  } finally {
    await api.dispose();
  }
}

async function apiClearMessages(token: string): Promise<void> {
  const api = await playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${token}`,
    },
  });
  try {
    const response = await api.delete("user/message/clear");
    await expectApiSuccess(response, "清空消息失败");
  } finally {
    await api.dispose();
  }
}

async function listPlugins(adminApi: APIRequestContext): Promise<PluginListItem[]> {
  const response = await adminApi.get("plugins");
  const payload = await expectApiSuccess<{ list?: PluginListItem[] }>(
    response,
    "查询插件列表失败",
  );
  return payload?.list ?? [];
}

async function findPlugin(adminApi: APIRequestContext, pluginID: string) {
  const list = await listPlugins(adminApi);
  return list.find((item) => item.id === pluginID) ?? null;
}

async function uploadDynamicPlugin(
  adminApi: APIRequestContext,
  artifactPath: string,
  overwrite = false,
) {
  const response = await adminApi.post("plugins/dynamic/package", {
    multipart: {
      overwriteSupport: overwrite ? "1" : "0",
      file: {
        name: path.basename(artifactPath),
        mimeType: "application/wasm",
        buffer: readFileSync(artifactPath),
      },
    },
  });
  await expectApiSuccess(response, `上传动态插件失败: ${artifactPath}`);
}

async function installPlugin(adminApi: APIRequestContext, pluginID: string) {
  const response = await adminApi.post(`plugins/${pluginID}/install`);
  await expectApiSuccess(response, `安装动态插件失败: ${pluginID}`);
}

async function uninstallPlugin(adminApi: APIRequestContext, pluginID: string) {
  const response = await adminApi.delete(`plugins/${pluginID}`);
  await expectApiSuccess(response, `卸载动态插件失败: ${pluginID}`);
}

async function setPluginEnabled(
  adminApi: APIRequestContext,
  pluginID: string,
  enabled: boolean,
) {
  const response = await adminApi.put(
    enabled ? `plugins/${pluginID}/enable` : `plugins/${pluginID}/disable`,
  );
  await expectApiSuccess(
    response,
    `切换动态插件状态失败: ${pluginID} enabled=${enabled}`,
  );
}

async function resetPlugin(adminApi: APIRequestContext, pluginID: string) {
  const plugin = await findPlugin(adminApi, pluginID);
  if (!plugin) {
    return;
  }
  if (plugin.enabled === 1) {
    await setPluginEnabled(adminApi, pluginID, false);
  }
  if (plugin.installed === 1) {
    await uninstallPlugin(adminApi, pluginID);
  }
}

function ensureLowPriorityHostServiceTables() {
  execFileSync(
    mysqlBin,
    [
      `-u${mysqlUser}`,
      `-p${mysqlPassword}`,
      mysqlDatabase,
      "-e",
      [
        "CREATE TABLE IF NOT EXISTS sys_kv_cache (",
        "  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',",
        "  owner_type VARCHAR(16) NOT NULL DEFAULT '' COMMENT '所属类型：plugin=动态插件 module=宿主模块',",
        "  owner_key VARCHAR(64) NOT NULL DEFAULT '' COMMENT '所属标识：插件ID或模块名',",
        "  namespace VARCHAR(64) NOT NULL DEFAULT '' COMMENT '缓存命名空间，对应 host-cache 资源标识',",
        "  cache_key VARCHAR(128) NOT NULL DEFAULT '' COMMENT '缓存键',",
        "  value_kind TINYINT NOT NULL DEFAULT 1 COMMENT '值类型：1=字符串 2=整数',",
        "  value_bytes VARBINARY(4096) NOT NULL COMMENT '缓存字节值，供 get/set 使用',",
        "  value_int BIGINT NOT NULL DEFAULT 0 COMMENT '缓存整数值，供 incr 使用',",
        "  expire_at DATETIME NULL DEFAULT NULL COMMENT '过期时间，NULL表示永不过期',",
        "  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',",
        "  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',",
        "  UNIQUE KEY uk_owner_namespace_key (owner_type, owner_key, namespace, cache_key),",
        "  KEY idx_expire_at (expire_at)",
        ") ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='宿主分布式KV缓存表';",
        "CREATE TABLE IF NOT EXISTS sys_locker (",
        "  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',",
        "  name VARCHAR(64) NOT NULL COMMENT '锁名称，唯一标识',",
        "  reason VARCHAR(255) DEFAULT '' COMMENT '获取锁的原因',",
        "  holder VARCHAR(64) DEFAULT '' COMMENT '锁持有者标识（节点名）',",
        "  expire_time DATETIME NOT NULL COMMENT '锁过期时间',",
        "  created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',",
        "  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',",
        "  UNIQUE KEY uk_name (name),",
        "  INDEX idx_expire_time (expire_time)",
        ") ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COMMENT='分布式锁表';",
        "CREATE TABLE IF NOT EXISTS sys_notify_channel (",
        "  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',",
        "  channel_key VARCHAR(64) NOT NULL DEFAULT '' COMMENT '通道标识',",
        "  name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '通道名称',",
        "  channel_type VARCHAR(32) NOT NULL DEFAULT '' COMMENT '通道类型：inbox=站内信 email=邮件 webhook=Webhook',",
        "  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1=启用 0=停用',",
        "  config_json LONGTEXT NOT NULL COMMENT '通道配置JSON',",
        "  remark VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注',",
        "  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',",
        "  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',",
        "  deleted_at DATETIME NULL DEFAULT NULL COMMENT '删除时间',",
        "  UNIQUE KEY uk_channel_key (channel_key)",
        ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通知通道表';",
        "CREATE TABLE IF NOT EXISTS sys_notify_message (",
        "  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',",
        "  plugin_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '来源插件ID，宿主内建流程为空',",
        "  source_type VARCHAR(32) NOT NULL DEFAULT '' COMMENT '来源类型：notice=公告 plugin=插件 system=系统',",
        "  source_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '来源业务ID',",
        "  category_code VARCHAR(32) NOT NULL DEFAULT '' COMMENT '消息分类：notice=通知 announcement=公告 other=其他',",
        "  title VARCHAR(255) NOT NULL DEFAULT '' COMMENT '消息标题',",
        "  content LONGTEXT NOT NULL COMMENT '消息正文',",
        "  payload_json LONGTEXT NOT NULL COMMENT '扩展载荷JSON',",
        "  sender_user_id BIGINT NOT NULL DEFAULT 0 COMMENT '发送者用户ID',",
        "  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',",
        "  KEY idx_source (source_type, source_id)",
        ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通知消息主表';",
        "CREATE TABLE IF NOT EXISTS sys_notify_delivery (",
        "  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',",
        "  message_id BIGINT NOT NULL DEFAULT 0 COMMENT '通知消息ID',",
        "  channel_key VARCHAR(64) NOT NULL DEFAULT '' COMMENT '投递通道标识',",
        "  channel_type VARCHAR(32) NOT NULL DEFAULT '' COMMENT '投递通道类型',",
        "  recipient_type VARCHAR(32) NOT NULL DEFAULT '' COMMENT '接收者类型：user=用户 email=邮箱 webhook=Webhook',",
        "  recipient_key VARCHAR(128) NOT NULL DEFAULT '' COMMENT '接收者标识，如用户ID邮箱地址或Webhook标识',",
        "  user_id BIGINT NOT NULL DEFAULT 0 COMMENT '站内信用户ID，非站内信时为0',",
        "  delivery_status TINYINT NOT NULL DEFAULT 0 COMMENT '投递状态：0=待发送 1=成功 2=失败',",
        "  is_read TINYINT NOT NULL DEFAULT 0 COMMENT '是否已读：0=未读 1=已读',",
        "  read_at DATETIME NULL DEFAULT NULL COMMENT '已读时间',",
        "  error_message VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '失败原因',",
        "  sent_at DATETIME NULL DEFAULT NULL COMMENT '发送完成时间',",
        "  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',",
        "  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',",
        "  deleted_at DATETIME NULL DEFAULT NULL COMMENT '删除时间',",
        "  KEY idx_message_id (message_id),",
        "  KEY idx_user_inbox (user_id, channel_type, delivery_status, is_read),",
        "  KEY idx_channel_status (channel_key, delivery_status)",
        ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通知投递记录表';",
        "CREATE TABLE IF NOT EXISTS sys_plugin_state (",
        "  id INT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',",
        "  plugin_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '插件唯一标识（kebab-case）',",
        "  state_key VARCHAR(255) NOT NULL DEFAULT '' COMMENT '状态键',",
        "  state_value LONGTEXT COMMENT '状态值（支持JSON）',",
        "  created_at DATETIME COMMENT '创建时间',",
        "  updated_at DATETIME COMMENT '更新时间',",
        "  UNIQUE KEY uk_plugin_state (plugin_id, state_key)",
        ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='插件键值状态存储表';",
        "INSERT INTO sys_notify_channel (",
        "  channel_key, name, channel_type, status, config_json, remark, created_at, updated_at, deleted_at",
        ") VALUES (",
        "  'inbox', '站内信', 'inbox', 1, '{}', '系统内置站内信通道', NOW(), NOW(), NULL",
        ") ON DUPLICATE KEY UPDATE",
        "  name = VALUES(name),",
        "  channel_type = VALUES(channel_type),",
        "  status = VALUES(status),",
        "  config_json = VALUES(config_json),",
        "  remark = VALUES(remark),",
        "  deleted_at = NULL;",
      ].join(" "),
    ],
    { stdio: "ignore" },
  );
}

function cleanupPluginRows(pluginIDs: string[]) {
  const statements: string[] = [];
  for (const pluginID of pluginIDs) {
    const escapedID = pluginID.replaceAll("'", "''");
    statements.push(
      `DELETE FROM sys_notify_delivery WHERE message_id IN (SELECT id FROM sys_notify_message WHERE plugin_id = '${escapedID}');`,
      `DELETE FROM sys_notify_message WHERE plugin_id = '${escapedID}';`,
      `DELETE FROM sys_kv_cache WHERE owner_type = 'plugin' AND owner_key = '${escapedID}';`,
      `DELETE FROM sys_locker WHERE name LIKE 'plugin:${escapedID}:%';`,
      `DELETE FROM sys_role_menu WHERE menu_id IN (SELECT menu_ids.id FROM (SELECT id FROM sys_menu WHERE menu_key LIKE 'plugin:${escapedID}:%') AS menu_ids);`,
      `DELETE FROM sys_menu WHERE menu_key LIKE 'plugin:${escapedID}:%';`,
      `DELETE FROM sys_plugin_state WHERE plugin_id = '${escapedID}';`,
      `DELETE FROM sys_plugin_node_state WHERE plugin_id = '${escapedID}';`,
      `DELETE FROM sys_plugin_resource_ref WHERE plugin_id = '${escapedID}';`,
      `DELETE FROM sys_plugin_migration WHERE plugin_id = '${escapedID}';`,
      `DELETE FROM sys_plugin_release WHERE plugin_id = '${escapedID}';`,
      `DELETE FROM sys_plugin WHERE plugin_id = '${escapedID}';`,
    );
  }
  execFileSync(
    mysqlBin,
    [
      `-u${mysqlUser}`,
      `-p${mysqlPassword}`,
      mysqlDatabase,
      "-e",
      statements.join(" "),
    ],
    { stdio: "ignore" },
  );
}

function cleanupArtifacts(pluginIDs: string[]) {
  for (const pluginID of pluginIDs) {
    rmSync(uploadedArtifactPath(pluginID), { force: true });
  }
}

function buildPluginRuntimeMain(moduleName: string) {
  return `package main

import (
	"lina-core/pkg/pluginbridge"
	dynamicbackend "${moduleName}/backend"
)

var guestRuntime = pluginbridge.NewGuestRuntime(dynamicbackend.HandleRequest)

//go:wasmexport lina_dynamic_route_alloc
func linaDynamicRouteAlloc(size uint32) uint32 {
	return guestRuntime.Alloc(size)
}

//go:wasmexport lina_dynamic_route_execute
func linaDynamicRouteExecute(size uint32) uint64 {
	responsePointer, responseLength, err := guestRuntime.Execute(size)
	if err != nil {
		fallback, _ := pluginbridge.EncodeResponseEnvelope(pluginbridge.NewInternalErrorResponse(err.Error()))
		responsePointer, responseLength, _ = guestRuntime.ExposeResponseBuffer(fallback)
	}
	return uint64(responsePointer)<<32 | uint64(responseLength)
}

//go:wasmexport lina_host_call_alloc
func linaHostCallAlloc(size uint32) uint32 {
	return guestRuntime.HostCallAlloc(size)
}

func main() {}
`;
}

function buildPluginEmbedFile() {
  return `package main

import "embed"

//go:embed plugin.yaml frontend manifest
var EmbeddedFiles embed.FS
`;
}

function buildBackendPluginFile(moduleName: string) {
  return `package backend

import (
	"lina-core/pkg/pluginbridge"
	"${moduleName}/backend/internal/controller/dynamic"
)

var guestRouteDispatcher = pluginbridge.MustNewGuestControllerRouteDispatcher(dynamic.New())

func HandleRequest(
	request *pluginbridge.BridgeRequestEnvelopeV1,
) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return guestRouteDispatcher.HandleRequest(request)
}
`;
}

function buildSuccessPluginSource() {
  const pluginDir = sourcePluginDir(successPluginID);
  const moduleName = "lina-plugin-lp-host-e2e";
  rmSync(pluginDir, { force: true, recursive: true });

  writeTestFile(
    path.join(pluginDir, "go.mod"),
    `module ${moduleName}

go 1.25.0
`,
  );
  writeTestFile(path.join(pluginDir, "main.go"), buildPluginRuntimeMain(moduleName));
  writeTestFile(path.join(pluginDir, "plugin_embed.go"), buildPluginEmbedFile());
  writeTestFile(path.join(pluginDir, "backend", "plugin.go"), buildBackendPluginFile(moduleName));
  writeTestFile(
    path.join(pluginDir, "plugin.yaml"),
    `id: ${successPluginID}
name: Low Priority Host Services E2E
version: v0.1.0
type: dynamic
hostServices:
  - service: cache
    methods:
      - get
      - set
      - delete
      - incr
      - expire
    resources:
      - ref: e2e-cache
  - service: lock
    methods:
      - acquire
      - renew
      - release
    resources:
      - ref: e2e-lock
  - service: notify
    methods:
      - send
    resources:
      - ref: inbox
`,
  );
  writeTestFile(
    path.join(pluginDir, "backend", "api", "dynamic", "v1", "low_priority_host_services.go"),
    `package v1

import "github.com/gogf/gf/v2/frame/g"

type LowPriorityHostServicesReq struct {
	g.Meta \`path:"/low-priority-host-services" method:"get" tags:"动态插件 E2E" summary:"低优先级 host service 演示" dc:"验证 cache、lock、notify 三类低优先级宿主服务在动态插件路由内的成功调用" access:"login" permission:"${successPluginID}:host:view" operLog:"other"\`
}
`,
  );
  writeTestFile(
    path.join(pluginDir, "backend", "internal", "controller", "dynamic", "dynamic.go"),
    `package dynamic

type Controller struct{}

func New() *Controller {
	return &Controller{}
}
`,
  );
  writeTestFile(
    path.join(pluginDir, "backend", "internal", "controller", "dynamic", "low_priority_host_services.go"),
    `package dynamic

import (
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
)

const (
	cacheNamespace = "e2e-cache"
	lockName = "e2e-lock"
)

func (c *Controller) LowPriorityHostServices(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	var (
		cacheSvc = pluginbridge.Cache()
		lockSvc = pluginbridge.Lock()
		notifySvc = pluginbridge.Notify()
	)

	cacheSetValue, err := cacheSvc.Set(cacheNamespace, "profile", request.PluginID, 60)
	if err != nil {
		return nil, err
	}
	cacheGetValue, cacheFound, err := cacheSvc.Get(cacheNamespace, "profile")
	if err != nil {
		return nil, err
	}
	counterValue, err := cacheSvc.Incr(cacheNamespace, "counter", 2, 60)
	if err != nil {
		return nil, err
	}
	expireFound, expireAt, err := cacheSvc.Expire(cacheNamespace, "profile", 120)
	if err != nil {
		return nil, err
	}
	if err = cacheSvc.Delete(cacheNamespace, "profile"); err != nil {
		return nil, err
	}
	_, cacheFoundAfterDelete, err := cacheSvc.Get(cacheNamespace, "profile")
	if err != nil {
		return nil, err
	}

	lockAcquire, err := lockSvc.Acquire(lockName, 5000)
	if err != nil {
		return nil, err
	}
	if lockAcquire == nil || !lockAcquire.Acquired || lockAcquire.Ticket == "" {
		return nil, gerror.New("lock acquire failed")
	}
	lockRenew, err := lockSvc.Renew(lockName, lockAcquire.Ticket)
	if err != nil {
		return nil, err
	}
	if err = lockSvc.Release(lockName, lockAcquire.Ticket); err != nil {
		return nil, err
	}

	payloadJSON, err := json.Marshal(map[string]string{
		"requestId": request.RequestID,
	})
	if err != nil {
		return nil, gerror.Wrap(err, "marshal notify payload failed")
	}
	notifyResult, err := notifySvc.Send("inbox", &pluginbridge.HostServiceNotifySendRequest{
		Title: "低优先级宿主服务测试通知",
		Content: "cache/lock/notify success",
		SourceType: "plugin",
		SourceID: request.RequestID,
		CategoryCode: "other",
		RecipientUserIDs: []int64{1},
		PayloadJSON: payloadJSON,
	})
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"pluginId": request.PluginID,
		"cache": map[string]any{
			"setValue": cacheSetValue.Value,
			"found": cacheFound,
			"getValue": cacheGetValue.Value,
			"counterValue": counterValue.IntValue,
			"expireFound": expireFound,
			"expireAt": expireAt,
			"deleted": !cacheFoundAfterDelete,
		},
		"lock": map[string]any{
			"acquired": lockAcquire.Acquired,
			"ticket": lockAcquire.Ticket,
			"expireAt": lockAcquire.ExpireAt,
			"renewExpireAt": lockRenew.ExpireAt,
		},
		"notify": map[string]any{
			"messageId": notifyResult.MessageID,
			"deliveryCount": notifyResult.DeliveryCount,
		},
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal low priority host services payload failed")
	}
	return pluginbridge.NewJSONResponse(200, content), nil
}
`,
  );
  writeTestFile(
    path.join(pluginDir, "frontend", "pages", "placeholder.html"),
    "<!doctype html><html><body>low priority host services e2e</body></html>\n",
  );
  writeTestFile(
    path.join(pluginDir, "manifest", "README.md"),
    "low priority host services e2e fixture\n",
  );
  return pluginDir;
}

function buildDeniedPluginSource() {
  const pluginDir = sourcePluginDir(deniedPluginID);
  const moduleName = "lina-plugin-lp-host-denied-e2e";
  rmSync(pluginDir, { force: true, recursive: true });

  writeTestFile(
    path.join(pluginDir, "go.mod"),
    `module ${moduleName}

go 1.25.0
`,
  );
  writeTestFile(path.join(pluginDir, "main.go"), buildPluginRuntimeMain(moduleName));
  writeTestFile(path.join(pluginDir, "plugin_embed.go"), buildPluginEmbedFile());
  writeTestFile(path.join(pluginDir, "backend", "plugin.go"), buildBackendPluginFile(moduleName));
  writeTestFile(
    path.join(pluginDir, "plugin.yaml"),
    `id: ${deniedPluginID}
name: Low Priority Host Services Denied E2E
version: v0.1.0
type: dynamic
hostServices:
  - service: cache
    methods:
      - set
    resources:
      - ref: limited-cache
  - service: lock
    methods:
      - acquire
    resources:
      - ref: authorized-lock
  - service: notify
    methods:
      - send
    resources:
      - ref: inbox
`,
  );
  writeTestFile(
    path.join(pluginDir, "backend", "api", "dynamic", "v1", "denied_routes.go"),
    `package v1

import "github.com/gogf/gf/v2/frame/g"

type CacheLimitReq struct {
	g.Meta \`path:"/cache-limit" method:"get" tags:"动态插件 E2E" summary:"缓存长度超限" dc:"验证 cache host service 在超过字段字节上限时会被宿主拒绝" access:"login" permission:"${deniedPluginID}:host:view" operLog:"other"\`
}

type LockDeniedReq struct {
	g.Meta \`path:"/lock-denied" method:"get" tags:"动态插件 E2E" summary:"未授权锁资源" dc:"验证 lock host service 调用未授权逻辑锁名时会被宿主拒绝" access:"login" permission:"${deniedPluginID}:host:view" operLog:"other"\`
}

type NotifyDeniedReq struct {
	g.Meta \`path:"/notify-denied" method:"get" tags:"动态插件 E2E" summary:"未授权通知通道" dc:"验证 notify host service 调用未授权通知通道时会被宿主拒绝" access:"login" permission:"${deniedPluginID}:host:view" operLog:"other"\`
}
`,
  );
  writeTestFile(
    path.join(pluginDir, "backend", "internal", "controller", "dynamic", "dynamic.go"),
    `package dynamic

type Controller struct{}

func New() *Controller {
	return &Controller{}
}
`,
  );
  writeTestFile(
    path.join(pluginDir, "backend", "internal", "controller", "dynamic", "denied_routes.go"),
    `package dynamic

import (
	"strings"

	"lina-core/pkg/pluginbridge"
)

func (c *Controller) CacheLimit(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	_, err := pluginbridge.Cache().Set("limited-cache", "oversized", strings.Repeat("a", 4097), 0)
	if err != nil {
		return nil, err
	}
	return pluginbridge.NewJSONResponse(200, []byte("{}")), nil
}

func (c *Controller) LockDenied(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	_, err := pluginbridge.Lock().Acquire("blocked-lock", 1000)
	if err != nil {
		return nil, err
	}
	return pluginbridge.NewJSONResponse(200, []byte("{}")), nil
}

func (c *Controller) NotifyDenied(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	_, err := pluginbridge.Notify().Send("ops-webhook", &pluginbridge.HostServiceNotifySendRequest{
		Title: "denied notify",
		Content: "blocked",
		RecipientUserIDs: []int64{1},
	})
	if err != nil {
		return nil, err
	}
	return pluginbridge.NewJSONResponse(200, []byte("{}")), nil
}
`,
  );
  writeTestFile(
    path.join(pluginDir, "frontend", "pages", "placeholder.html"),
    "<!doctype html><html><body>low priority denied host services e2e</body></html>\n",
  );
  writeTestFile(
    path.join(pluginDir, "manifest", "README.md"),
    "low priority denied host services e2e fixture\n",
  );
  return pluginDir;
}

function buildDynamicPluginArtifact(pluginDir: string, pluginID: string) {
  mkdirSync(buildOutputDir(), { recursive: true });
  const goWorkPath = path.join(pluginDir, ".e2e-low-priority-host-services.go.work");
  writeFileSync(
    goWorkPath,
    [
      "go 1.25.0",
      "",
      "use (",
      `\t${path.join(repoRoot(), "apps", "lina-core")}`,
      `\t${builderDir()}`,
      `\t${pluginDir}`,
      ")",
      "",
    ].join("\n"),
  );
  try {
    execFileSync(
      "go",
      [
        "run",
        ".",
        "--plugin-dir",
        pluginDir,
        "--output-dir",
        buildOutputDir(),
      ],
      {
        cwd: builderDir(),
        env: {
          ...process.env,
          GOWORK: goWorkPath,
        },
        stdio: "pipe",
      },
    );
  } finally {
    rmSync(goWorkPath, { force: true });
  }
  return builtArtifactPath(pluginID);
}

test.describe("TC-72 Runtime Wasm Low Priority Host Services", () => {
  let adminApi: APIRequestContext | null = null;
  let adminToken = "";
  let successArtifact = "";
  let deniedArtifact = "";

  test.beforeAll(async () => {
    rmSync(tempRoot(), { force: true, recursive: true });
    mkdirSync(sourceRoot(), { recursive: true });
    mkdirSync(buildOutputDir(), { recursive: true });
    ensureLowPriorityHostServiceTables();

    successArtifact = buildDynamicPluginArtifact(
      buildSuccessPluginSource(),
      successPluginID,
    );
    deniedArtifact = buildDynamicPluginArtifact(
      buildDeniedPluginSource(),
      deniedPluginID,
    );
    adminApi = await createAdminApiContext();
    adminToken = await apiLogin(config.adminUser, config.adminPass);
  });

  test.afterAll(async () => {
    if (adminApi) {
      await resetPlugin(adminApi, successPluginID);
      await resetPlugin(adminApi, deniedPluginID);
      await adminApi.dispose();
    }
    if (adminToken) {
      await apiClearMessages(adminToken);
    }
    cleanupPluginRows([successPluginID, deniedPluginID]);
    cleanupArtifacts([successPluginID, deniedPluginID]);
    rmSync(tempRoot(), { force: true, recursive: true });
  });

  test.beforeEach(async () => {
    await resetPlugin(adminApi!, successPluginID);
    await resetPlugin(adminApi!, deniedPluginID);
    if (adminToken) {
      await apiClearMessages(adminToken);
    }
    cleanupPluginRows([successPluginID, deniedPluginID]);
    cleanupArtifacts([successPluginID, deniedPluginID]);
  });

  test.afterEach(async () => {
    await resetPlugin(adminApi!, successPluginID);
    await resetPlugin(adminApi!, deniedPluginID);
    if (adminToken) {
      await apiClearMessages(adminToken);
    }
    cleanupPluginRows([successPluginID, deniedPluginID]);
    cleanupArtifacts([successPluginID, deniedPluginID]);
  });

  test("TC-72a: 已授权的 cache、lock 和 notify 宿主服务调用成功", async () => {
    await uploadDynamicPlugin(adminApi!, successArtifact);
    await installPlugin(adminApi!, successPluginID);
    await setPluginEnabled(adminApi!, successPluginID, true);

    const unreadBefore = await apiUnreadCount(adminToken);
    expect(unreadBefore).toBe(0);

    const response = await adminApi!.get(
      `extensions/${successPluginID}/low-priority-host-services`,
    );
    const responseText = await response.text();
    expect(
      response.status(),
      `调用低优先级 host service 演示路由失败: ${responseText}`,
    ).toBe(200);

    const payload = JSON.parse(responseText) as {
      pluginId: string;
      cache: Record<string, any>;
      lock: Record<string, any>;
      notify: Record<string, any>;
    };

    expect(payload.pluginId).toBe(successPluginID);
    expect(payload.cache.setValue).toBe(successPluginID);
    expect(payload.cache.found).toBeTruthy();
    expect(payload.cache.getValue).toBe(successPluginID);
    expect(payload.cache.counterValue).toBe(2);
    expect(payload.cache.expireFound).toBeTruthy();
    expect(payload.cache.expireAt).toBeTruthy();
    expect(payload.cache.deleted).toBeTruthy();

    expect(payload.lock.acquired).toBeTruthy();
    expect(payload.lock.ticket).toBeTruthy();
    expect(payload.lock.expireAt).toBeTruthy();
    expect(payload.lock.renewExpireAt).toBeTruthy();

    expect(payload.notify.messageId).toBeGreaterThan(0);
    expect(payload.notify.deliveryCount).toBe(1);

    const unreadAfter = await apiUnreadCount(adminToken);
    expect(unreadAfter).toBe(1);
  });

  test("TC-72b: 低优先级宿主服务在未授权资源或超限场景下被宿主拒绝", async () => {
    await uploadDynamicPlugin(adminApi!, deniedArtifact);
    await installPlugin(adminApi!, deniedPluginID);
    await setPluginEnabled(adminApi!, deniedPluginID, true);

    const cacheLimitResponse = await adminApi!.get(
      `extensions/${deniedPluginID}/cache-limit`,
    );
    expect(cacheLimitResponse.status()).toBe(500);
    await expectApiFailure(
      cacheLimitResponse,
      "超限缓存值必须被宿主拒绝",
      "缓存值长度超出限制",
    );

    const lockDeniedResponse = await adminApi!.get(
      `extensions/${deniedPluginID}/lock-denied`,
    );
    expect(lockDeniedResponse.status()).toBe(500);
    await expectApiFailure(
      lockDeniedResponse,
      "未授权逻辑锁名必须被宿主拒绝",
      "resource=blocked-lock",
    );

    const notifyDeniedResponse = await adminApi!.get(
      `extensions/${deniedPluginID}/notify-denied`,
    );
    expect(notifyDeniedResponse.status()).toBe(500);
    await expectApiFailure(
      notifyDeniedResponse,
      "未授权通知通道必须被宿主拒绝",
      "resource=ops-webhook",
    );
  });
});
