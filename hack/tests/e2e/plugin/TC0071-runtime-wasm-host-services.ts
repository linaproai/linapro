import { execFileSync } from "node:child_process";
import { mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { createServer, type Server } from "node:http";
import type { AddressInfo } from "node:net";
import path from "node:path";

import type { APIRequestContext, APIResponse } from "@playwright/test";
import { request as playwrightRequest } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";
import { config } from "../../fixtures/config";

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? "http://127.0.0.1:8080/api/v1/";
const mysqlBin = process.env.E2E_MYSQL_BIN ?? "mysql";
const mysqlUser = process.env.E2E_DB_USER ?? "root";
const mysqlPassword = process.env.E2E_DB_PASSWORD ?? "12345678";
const mysqlDatabase = process.env.E2E_DB_NAME ?? "lina";

const successPluginID = "plugin-host-services-e2e";
const deniedPluginID = "plugin-host-services-denied-e2e";
const rawSQLPluginID = "plugin-host-services-raw-sql-e2e";

type PluginListItem = {
  id: string;
  enabled?: number;
  installed?: number;
};

function repoRoot() {
  return path.resolve(process.cwd(), "../..");
}

function tempRoot() {
  return path.join(repoRoot(), "temp", "e2e-host-services");
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

function ensurePluginStateTable() {
  execFileSync(
    mysqlBin,
    [
      `-u${mysqlUser}`,
      `-p${mysqlPassword}`,
      mysqlDatabase,
      "-e",
      [
        "CREATE TABLE IF NOT EXISTS sys_plugin_state (",
        "  id INT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',",
        "  plugin_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '插件唯一标识（kebab-case）',",
        "  state_key VARCHAR(255) NOT NULL DEFAULT '' COMMENT '状态键',",
        "  state_value LONGTEXT COMMENT '状态值（支持JSON）',",
        "  created_at DATETIME COMMENT '创建时间',",
        "  updated_at DATETIME COMMENT '更新时间',",
        "  UNIQUE KEY uk_plugin_state (plugin_id, state_key)",
        ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='插件键值状态存储表';",
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

function buildSuccessPluginSource(upstreamBaseURL: string) {
  const pluginDir = sourcePluginDir(successPluginID);
  const moduleName = "lina-plugin-host-services-e2e";
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
name: Host Services E2E
version: v0.1.0
type: dynamic
hostServices:
  - service: runtime
    methods:
      - log.write
      - state.get
      - state.set
      - state.delete
      - info.now
      - info.uuid
      - info.node
  - service: storage
    methods:
      - put
      - get
      - delete
      - list
      - stat
    resources:
      paths:
        - e2e/
  - service: network
    methods:
      - request
    resources:
      - url: ${upstreamBaseURL}
  - service: data
    methods:
      - list
      - get
      - create
      - update
      - delete
      - transaction
    resources:
      tables:
        - sys_plugin_node_state
`,
  );
  writeTestFile(
    path.join(pluginDir, "backend", "api", "dynamic", "v1", "host_services.go"),
    `package v1

import "github.com/gogf/gf/v2/frame/g"

type HostServicesReq struct {
	g.Meta \`path:"/host-services" method:"get" tags:"动态插件 E2E" summary:"核心 host service 演示" dc:"验证 runtime、storage、network、data 四类核心宿主服务在动态插件路由内的成功调用" access:"login" permission:"${successPluginID}:host-services:view" operLog:"other"\`
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
    path.join(pluginDir, "backend", "internal", "controller", "dynamic", "host_services.go"),
    `package dynamic

import (
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugindb"
	"lina-core/pkg/pluginbridge"
)

const (
	networkURL = "${upstreamBaseURL}"
	dataTable  = "sys_plugin_node_state"
)

func (c *Controller) HostServices(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	var (
		runtimeSvc = pluginbridge.Runtime()
		storageSvc = pluginbridge.Storage()
		httpSvc    = pluginbridge.HTTP()
		dataSvc    = plugindb.Open()
	)

	nowValue, err := runtimeSvc.Now()
	if err != nil {
		return nil, err
	}
	uuidValue, err := runtimeSvc.UUID()
	if err != nil {
		return nil, err
	}
	nodeValue, err := runtimeSvc.Node()
	if err != nil {
		return nil, err
	}

	stateKey := "e2e-state-" + uuidValue
	if err = runtimeSvc.StateSet(stateKey, request.PluginID); err != nil {
		return nil, err
	}
	stateValue, stateFound, err := runtimeSvc.StateGet(stateKey)
	if err != nil {
		return nil, err
	}
	if !stateFound || stateValue != request.PluginID {
		return nil, gerror.New("runtime state round trip failed")
	}
	if err = runtimeSvc.StateDelete(stateKey); err != nil {
		return nil, err
	}
	_, stateFoundAfterDelete, err := runtimeSvc.StateGet(stateKey)
	if err != nil {
		return nil, err
	}
	if stateFoundAfterDelete {
		return nil, gerror.New("runtime state delete failed")
	}
	if err = runtimeSvc.Log(int(pluginbridge.LogLevelInfo), "host services e2e success", map[string]string{
		"pluginId": request.PluginID,
		"requestId": request.RequestID,
		"demoKey": uuidValue,
	}); err != nil {
		return nil, err
	}

	objectPath := "e2e/" + uuidValue + ".json"
	storageBody, err := json.Marshal(map[string]string{
		"pluginId": request.PluginID,
		"uuid": uuidValue,
	})
	if err != nil {
		return nil, gerror.Wrap(err, "marshal storage body failed")
	}
	objectMeta, err := storageSvc.Put(objectPath, storageBody, "application/json", true)
	if err != nil {
		return nil, err
	}
	readBody, _, storageFound, err := storageSvc.Get(objectPath)
	if err != nil {
		return nil, err
	}
	statObject, statFound, err := storageSvc.Stat(objectPath)
	if err != nil {
		return nil, err
	}
	listObjects, err := storageSvc.List("e2e", 10)
	if err != nil {
		return nil, err
	}
	if err = storageSvc.Delete(objectPath); err != nil {
		return nil, err
	}
	_, deletedFound, err := storageSvc.Stat(objectPath)
	if err != nil {
		return nil, err
	}

	networkResponse, err := httpSvc.Request(networkURL+"/ping?plugin="+request.PluginID, &pluginbridge.HostServiceNetworkRequest{
		Method: "GET",
		Headers: map[string]string{
			"x-request-id": request.RequestID,
		},
	})
	if err != nil {
		return nil, err
	}

	err = dataSvc.Transaction(func(tx *plugindb.Tx) error {
		_, txErr := tx.Table(dataTable).Insert(map[string]any{
			"pluginId": request.PluginID,
			"releaseId": 0,
			"nodeKey": "e2e-" + uuidValue,
			"desiredState": "running",
			"currentState": "pending",
			"generation": 1,
			"errorMessage": "",
		})
		return txErr
	})
	if err != nil {
		return nil, err
	}
	listRecords, listTotal, err := dataSvc.Table(dataTable).
		Fields("id", "nodeKey", "currentState").
		WhereEq("pluginId", request.PluginID).
		WhereLike("nodeKey", uuidValue).
		WhereIn("currentState", []string{"pending", "running"}).
		OrderDesc("id").
		Page(1, 10).
		All()
	if err != nil {
		return nil, err
	}
	if listTotal != 1 || len(listRecords) != 1 {
		return nil, gerror.New("plugindb list did not return one record")
	}
	recordKey := listRecords[0]["id"]
	countTotal, err := dataSvc.Table(dataTable).
		WhereEq("pluginId", request.PluginID).
		WhereLike("nodeKey", uuidValue).
		Count()
	if err != nil {
		return nil, err
	}
	updateResult, err := dataSvc.Table(dataTable).WhereKey(recordKey).Update(map[string]any{
		"currentState": "running",
		"errorMessage": "",
	})
	if err != nil {
		return nil, err
	}
	record, dataFound, err := dataSvc.Table(dataTable).Fields("currentState").WhereKey(recordKey).One()
	if err != nil {
		return nil, err
	}
	deleteResult, err := dataSvc.Table(dataTable).WhereKey(recordKey).Delete()
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"pluginId": request.PluginID,
		"runtime": map[string]any{
			"now": nowValue,
			"uuid": uuidValue,
			"node": nodeValue,
			"stateRoundTrip": stateFound && stateValue == request.PluginID,
			"stateDeleted": !stateFoundAfterDelete,
		},
		"storage": map[string]any{
			"objectPath": objectPath,
			"stored": storageFound && string(readBody) == string(storageBody),
			"listedCount": len(listObjects),
			"statFound": statFound,
			"deleted": !deletedFound,
			"writtenPath": objectMeta.Path,
			"statPath": statObject.Path,
		},
		"network": map[string]any{
			"statusCode": networkResponse.StatusCode,
			"contentType": networkResponse.ContentType,
			"body": string(networkResponse.Body),
			"requestId": networkResponse.Headers["X-Upstream-Request-Id"],
		},
		"data": map[string]any{
			"recordKey": fmt.Sprint(recordKey),
			"transactionAffectedRows": 1,
			"listTotal": listTotal,
			"countTotal": countTotal,
			"updated": updateResult != nil && updateResult.AffectedRows >= 1,
			"deleted": deleteResult != nil && deleteResult.AffectedRows >= 1,
			"currentState": fmt.Sprint(record["currentState"]),
			"found": dataFound,
		},
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal host services payload failed")
	}
	return pluginbridge.NewJSONResponse(200, content), nil
}
`,
  );
  writeTestFile(
    path.join(pluginDir, "frontend", "pages", "placeholder.html"),
    "<!doctype html><html><body>host services e2e</body></html>\n",
  );
  writeTestFile(
    path.join(pluginDir, "manifest", "README.md"),
    "runtime host services e2e fixture\n",
  );
  return pluginDir;
}

function buildDeniedPluginSource() {
  const pluginDir = sourcePluginDir(deniedPluginID);
  const moduleName = "lina-plugin-host-services-denied-e2e";
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
name: Host Services Denied E2E
version: v0.1.0
type: dynamic
hostServices:
  - service: storage
    methods:
      - put
    resources:
      paths:
        - authorized-files/
`,
  );
  writeTestFile(
    path.join(pluginDir, "backend", "api", "dynamic", "v1", "denied_routes.go"),
    `package v1

import "github.com/gogf/gf/v2/frame/g"

type DeniedMethodReq struct {
	g.Meta \`path:"/denied-method" method:"get" tags:"动态插件 E2E" summary:"未授权 method 调用" dc:"验证插件调用未声明的 host service method 会被宿主拒绝" access:"login" permission:"${deniedPluginID}:denied-method:view" operLog:"other"\`
}

type DeniedResourceReq struct {
	g.Meta \`path:"/denied-resource" method:"get" tags:"动态插件 E2E" summary:"未授权资源标识调用" dc:"验证插件调用未授权资源标识时会被宿主拒绝" access:"login" permission:"${deniedPluginID}:denied-resource:view" operLog:"other"\`
}

type DeniedServiceReq struct {
	g.Meta \`path:"/denied-service" method:"get" tags:"动态插件 E2E" summary:"未授权 service 调用" dc:"验证插件调用未声明的 host service capability 会被宿主拒绝" access:"login" permission:"${deniedPluginID}:denied-service:view" operLog:"other"\`
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
	"lina-core/pkg/plugindb"
	"lina-core/pkg/pluginbridge"
)

func (c *Controller) DeniedMethod(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	_, _, _, err := pluginbridge.Storage().Get("authorized-files/blocked.txt")
	if err != nil {
		return nil, err
	}
	return pluginbridge.NewJSONResponse(200, []byte("{}")), nil
}

func (c *Controller) DeniedResource(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	_, err := pluginbridge.Storage().PutText("denied-files/blocked.txt", "blocked", "text/plain", true)
	if err != nil {
		return nil, err
	}
	return pluginbridge.NewJSONResponse(200, []byte("{}")), nil
}

func (c *Controller) DeniedService(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	_, _, err := plugindb.Open().Table("sys_plugin_node_state").Page(1, 1).All()
	if err != nil {
		return nil, err
	}
	return pluginbridge.NewJSONResponse(200, []byte("{}")), nil
}
`,
  );
  writeTestFile(
    path.join(pluginDir, "frontend", "pages", "placeholder.html"),
    "<!doctype html><html><body>denied host services e2e</body></html>\n",
  );
  writeTestFile(
    path.join(pluginDir, "manifest", "README.md"),
    "denied host services e2e fixture\n",
  );
  return pluginDir;
}

function buildDynamicPluginArtifact(pluginDir: string, pluginID: string) {
  mkdirSync(buildOutputDir(), { recursive: true });
  const goWorkPath = path.join(pluginDir, ".e2e-host-services.go.work");
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

function appendWasmULEB128(target: number[], value: number) {
  let current = value >>> 0;
  while (true) {
    let byte = current & 0x7f;
    current >>>= 7;
    if (current !== 0) {
      byte |= 0x80;
    }
    target.push(byte);
    if (current === 0) {
      return;
    }
  }
}

function appendCustomSection(target: number[], name: string, payload: Buffer) {
  const section: number[] = [];
  appendWasmULEB128(section, Buffer.byteLength(name));
  section.push(...Buffer.from(name));
  section.push(...payload);

  target.push(0x00);
  appendWasmULEB128(target, section.length);
  target.push(...section);
}

function buildRawSQLInvalidArtifact() {
  mkdirSync(buildOutputDir(), { recursive: true });
  const bytes: number[] = [0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00];

  appendCustomSection(
    bytes,
    "lina.plugin.manifest",
    Buffer.from(
      JSON.stringify({
        id: rawSQLPluginID,
        name: "Host Services Raw SQL E2E",
        version: "v0.1.0",
        type: "dynamic",
      }),
    ),
  );
  appendCustomSection(
    bytes,
    "lina.plugin.dynamic",
    Buffer.from(
      JSON.stringify({
        runtimeKind: "wasm",
        abiVersion: "v1",
      }),
    ),
  );
  appendCustomSection(
    bytes,
    "lina.plugin.backend.capabilities",
    Buffer.from(JSON.stringify(["host:runtime", "host:db:query"])),
  );
  appendCustomSection(
    bytes,
    "lina.plugin.backend.host-services",
    Buffer.from(
      JSON.stringify([
        {
          service: "runtime",
          methods: ["info.uuid"],
        },
      ]),
    ),
  );

  const artifactPath = builtArtifactPath(rawSQLPluginID);
  writeFileSync(artifactPath, Buffer.from(bytes));
  return artifactPath;
}

async function startUpstreamServer() {
  const server = createServer((request, response) => {
    if (request.url?.startsWith("/ping")) {
      response.statusCode = 200;
      response.setHeader("Content-Type", "application/json");
      response.setHeader(
        "X-Upstream-Request-Id",
        String(request.headers["x-request-id"] ?? ""),
      );
      response.end(
        JSON.stringify({
          ok: true,
          path: request.url,
          requestId: request.headers["x-request-id"] ?? "",
        }),
      );
      return;
    }
    response.statusCode = 404;
    response.end("not found");
  });

  await new Promise<void>((resolve) => {
    server.listen(0, "127.0.0.1", () => resolve());
  });

  const address = server.address() as AddressInfo;
  return {
    server,
    baseURL: `http://127.0.0.1:${address.port}`,
  };
}

test.describe("TC-71 Runtime Wasm Host Services", () => {
  let adminApi: APIRequestContext | null = null;
  let upstreamServer: Server | null = null;
  let successArtifact = "";
  let deniedArtifact = "";
  let rawSQLArtifact = "";

  test.beforeAll(async () => {
    rmSync(tempRoot(), { force: true, recursive: true });
    mkdirSync(sourceRoot(), { recursive: true });
    mkdirSync(buildOutputDir(), { recursive: true });
    ensurePluginStateTable();

    const upstream = await startUpstreamServer();
    upstreamServer = upstream.server;

    successArtifact = buildDynamicPluginArtifact(
      buildSuccessPluginSource(upstream.baseURL),
      successPluginID,
    );
    deniedArtifact = buildDynamicPluginArtifact(
      buildDeniedPluginSource(),
      deniedPluginID,
    );
    rawSQLArtifact = buildRawSQLInvalidArtifact();
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    if (adminApi) {
      await resetPlugin(adminApi, successPluginID);
      await resetPlugin(adminApi, deniedPluginID);
      await adminApi.dispose();
    }
    if (upstreamServer) {
      await new Promise<void>((resolve, reject) => {
        upstreamServer?.close((error) => {
          if (error) {
            reject(error);
            return;
          }
          resolve();
        });
      });
    }
    cleanupPluginRows([successPluginID, deniedPluginID, rawSQLPluginID]);
    cleanupArtifacts([successPluginID, deniedPluginID, rawSQLPluginID]);
    rmSync(tempRoot(), { force: true, recursive: true });
  });

  test.beforeEach(async () => {
    await resetPlugin(adminApi!, successPluginID);
    await resetPlugin(adminApi!, deniedPluginID);
    cleanupPluginRows([successPluginID, deniedPluginID, rawSQLPluginID]);
    cleanupArtifacts([successPluginID, deniedPluginID, rawSQLPluginID]);
  });

  test.afterEach(async () => {
    await resetPlugin(adminApi!, successPluginID);
    await resetPlugin(adminApi!, deniedPluginID);
    cleanupPluginRows([successPluginID, deniedPluginID, rawSQLPluginID]);
    cleanupArtifacts([successPluginID, deniedPluginID, rawSQLPluginID]);
  });

  test("TC-71a: 已授权的 runtime、storage、network 和 data 宿主服务调用成功", async () => {
    await uploadDynamicPlugin(adminApi!, successArtifact);
    await installPlugin(adminApi!, successPluginID);
    await setPluginEnabled(adminApi!, successPluginID, true);

    const response = await adminApi!.get(`extensions/${successPluginID}/host-services`);
    const responseText = await response.text();
    expect(
      response.status(),
      `调用核心 host service 演示路由失败: ${responseText}`,
    ).toBe(200);

    const payload = JSON.parse(responseText) as {
      pluginId: string;
      runtime: Record<string, any>;
      storage: Record<string, any>;
      network: Record<string, any>;
      data: Record<string, any>;
    };

    expect(payload.pluginId).toBe(successPluginID);
    expect(payload.runtime.stateRoundTrip).toBeTruthy();
    expect(payload.runtime.stateDeleted).toBeTruthy();
    expect(payload.runtime.uuid).toBeTruthy();
    expect(payload.runtime.node).toBeTruthy();

    expect(payload.storage.stored).toBeTruthy();
    expect(payload.storage.statFound).toBeTruthy();
    expect(payload.storage.deleted).toBeTruthy();
    expect(payload.storage.listedCount).toBeGreaterThanOrEqual(1);
    expect(payload.storage.writtenPath).toBe(payload.storage.objectPath);
    expect(payload.storage.statPath).toBe(payload.storage.objectPath);

    expect(payload.network.statusCode).toBe(200);
    expect(payload.network.contentType).toContain("application/json");
    expect(payload.network.body).toContain(successPluginID);
    expect(payload.network.requestId).toBeTruthy();

    expect(payload.data.transactionAffectedRows).toBeGreaterThanOrEqual(1);
    expect(payload.data.listTotal).toBe(1);
    expect(payload.data.countTotal).toBe(1);
    expect(payload.data.updated).toBeTruthy();
    expect(payload.data.deleted).toBeTruthy();
    expect(payload.data.currentState).toBe("running");
    expect(payload.data.found).toBeTruthy();
    expect(payload.data.recordKey).toBeTruthy();
  });

  test("TC-71b: 未声明的 service、method 或未授权资源标识调用会被宿主拒绝", async () => {
    await uploadDynamicPlugin(adminApi!, deniedArtifact);
    await installPlugin(adminApi!, deniedPluginID);
    await setPluginEnabled(adminApi!, deniedPluginID, true);

    const deniedMethodResponse = await adminApi!.get(
      `extensions/${deniedPluginID}/denied-method`,
    );
    expect(deniedMethodResponse.status()).toBe(500);
    await expectApiFailure(
      deniedMethodResponse,
      "未声明 method 的调用应被拒绝",
      "storage.get",
    );

    const deniedResourceResponse = await adminApi!.get(
      `extensions/${deniedPluginID}/denied-resource`,
    );
    expect(deniedResourceResponse.status()).toBe(500);
    await expectApiFailure(
      deniedResourceResponse,
      "未授权资源标识的调用应被拒绝",
      "resource=denied-files",
    );

    const deniedServiceResponse = await adminApi!.get(
      `extensions/${deniedPluginID}/denied-service`,
    );
    expect(deniedServiceResponse.status()).toBe(500);
    await expectApiFailure(
      deniedServiceResponse,
      "未声明 service capability 的调用应被拒绝",
      "host:data:read",
    );
  });

  test("TC-71c: 插件尝试申请 raw SQL 旧能力时会在上传链路被拒绝", async () => {
    const response = await adminApi!.post("plugins/dynamic/package", {
      multipart: {
        overwriteSupport: "0",
        file: {
          name: path.basename(rawSQLArtifact),
          mimeType: "application/wasm",
          buffer: readFileSync(rawSQLArtifact),
        },
      },
    });

    await expectApiFailure(
      response,
      "raw SQL 旧能力必须被宿主拒绝",
      "host:db:query",
    );
    expect(await findPlugin(adminApi!, rawSQLPluginID)).toBeNull();
  });
});
