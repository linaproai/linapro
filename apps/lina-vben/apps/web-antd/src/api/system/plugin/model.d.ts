export type PluginType = 'dynamic' | 'source' | string;

export interface PluginListParams {
  pageNum?: number;
  pageSize?: number;
  id?: string;
  installed?: number;
  name?: string;
  status?: number;
  type?: PluginType;
}

export interface SystemPlugin {
  id: string;
  name: string;
  version: string;
  runtimeState: PluginRuntimeState;
  effectiveVersion: string;
  discoveredVersion: string;
  upgradeAvailable: boolean;
  abnormalReason?: PluginRuntimeAbnormalReason;
  lastUpgradeFailure?: PluginUpgradeFailure;
  type: PluginType;
  description: string;
  installed: number;
  installedAt: string;
  enabled: number;
  autoEnableManaged: number;
  autoEnableForNewTenants?: boolean;
  supportsMultiTenant?: boolean;
  statusKey: string;
  updatedAt: string;
  authorizationRequired: number;
  authorizationStatus: 'confirmed' | 'not_required' | 'pending' | string;
  dependencyCheck?: PluginDependencyCheckResult;
  hasMockData: number;
  installMode?: 'global' | 'tenant_scoped' | string;
  scopeNature?: 'platform_only' | 'tenant_aware' | string;
  requestedHostServices?: HostServicePermissionItem[];
  authorizedHostServices?: HostServicePermissionItem[];
  declaredRoutes?: PluginRouteReviewItem[];
}

export type PluginRuntimeState =
  | 'abnormal'
  | 'normal'
  | 'pending_upgrade'
  | 'upgrade_failed'
  | 'upgrade_running'
  | string;

export type PluginRuntimeAbnormalReason =
  | 'discovered_version_lower_than_effective'
  | 'version_compare_failed'
  | string;

export interface PluginUpgradeFailure {
  phase: string;
  errorCode: string;
  messageKey: string;
  releaseId: number;
  releaseVersion: string;
  detail?: string;
}

export interface PluginDependencyCheckResult {
  targetId: string;
  framework?: PluginDependencyFrameworkCheck;
  dependencies?: PluginDependencyItem[];
  autoInstallPlan?: PluginDependencyAutoInstallItem[];
  autoInstalled?: PluginDependencyAutoInstallItem[];
  manualInstallRequired?: PluginDependencyItem[];
  softUnsatisfied?: PluginDependencyItem[];
  blockers?: PluginDependencyBlocker[];
  cycle?: string[];
  reverseDependents?: PluginDependencyReverseDependent[];
  reverseBlockers?: PluginDependencyBlocker[];
}

export interface PluginDependencyFrameworkCheck {
  requiredVersion: string;
  currentVersion: string;
  status: 'not_declared' | 'satisfied' | 'unsatisfied' | string;
}

export interface PluginDependencyItem {
  ownerId: string;
  dependencyId: string;
  dependencyName?: string;
  requiredVersion?: string;
  currentVersion?: string;
  required: boolean;
  installMode: 'auto' | 'manual' | string;
  installed: boolean;
  discovered: boolean;
  status:
    | 'auto_install_planned'
    | 'manual_install_required'
    | 'missing'
    | 'satisfied'
    | 'soft_unsatisfied'
    | 'version_unsatisfied'
    | string;
  chain?: string[];
}

export interface PluginDependencyAutoInstallItem {
  pluginId: string;
  name?: string;
  version?: string;
  requiredBy?: string;
  chain?: string[];
}

export interface PluginDependencyBlocker {
  code: string;
  pluginId?: string;
  dependencyId?: string;
  requiredVersion?: string;
  currentVersion?: string;
  chain?: string[];
  detail?: string;
}

export interface PluginDependencyReverseDependent {
  pluginId: string;
  name?: string;
  version?: string;
  requiredVersion?: string;
}

export interface PluginInstallResult {
  id: string;
  installed: number;
  enabled: number;
  dependencyCheck?: PluginDependencyCheckResult;
}

export interface PluginRouteReviewItem {
  method: string;
  publicPath: string;
  access: string;
  permission?: string;
  summary?: string;
  description?: string;
}

export interface HostServicePermissionItem {
  service: string;
  methods: string[];
  paths?: string[];
  tables?: string[];
  tableItems?: HostServicePermissionTableItem[];
  cronItems?: HostServicePermissionCronItem[];
  resources?: HostServicePermissionResourceItem[];
}

export interface HostServicePermissionTableItem {
  name: string;
  comment?: string;
}

export interface HostServicePermissionCronItem {
  name: string;
  displayName?: string;
  description?: string;
  pattern: string;
  timezone?: string;
  scope: string;
  concurrency: string;
  maxConcurrency?: number;
}

export interface HostServicePermissionResourceItem {
  ref: string;
  allowMethods?: string[];
  headerAllowList?: string[];
  timeoutMs?: number;
  maxBodyBytes?: number;
  attributes?: Record<string, string>;
}

export interface PluginAuthorizationPayload {
  authorization?: {
    services: Array<{
      methods?: string[];
      paths?: string[];
      resourceRefs?: string[];
      tables?: string[];
      service: string;
    }>;
  };
  installMockData?: boolean;
  installMode?: 'global' | 'tenant_scoped' | string;
  force?: boolean;
}

export interface PluginUpgradePreview {
  pluginId: string;
  runtimeState: PluginRuntimeState;
  effectiveVersion: string;
  discoveredVersion: string;
  fromManifest?: PluginManifestSnapshot;
  toManifest?: PluginManifestSnapshot;
  dependencyCheck?: PluginDependencyCheckResult;
  sqlSummary: PluginUpgradeSQLSummary;
  hostServicesDiff: PluginUpgradeHostServicesDiff;
  riskHints: string[];
}

export interface PluginManifestSnapshot {
  id: string;
  name: string;
  version: string;
  type: PluginType;
  scopeNature?: 'platform_only' | 'tenant_aware' | string;
  supportsMultiTenant: boolean;
  defaultInstallMode?: 'global' | 'tenant_scoped' | string;
  description?: string;
  runtimeKind?: string;
  runtimeAbiVersion?: string;
  manifestDeclared: boolean;
  installSqlCount: number;
  uninstallSqlCount: number;
  mockSqlCount: number;
  frontendPageCount: number;
  frontendSlotCount: number;
  menuCount: number;
  backendHookCount: number;
  resourceSpecCount: number;
  routeCount: number;
  routeExecutionEnabled: boolean;
  routeRequestCodec?: string;
  routeResponseCodec?: string;
  runtimeFrontendAssetCount: number;
  runtimeSqlAssetCount: number;
  hostServiceAuthRequired: boolean;
  hostServiceAuthConfirmed: boolean;
  requestedHostServices?: HostServicePermissionItem[];
  authorizedHostServices?: HostServicePermissionItem[];
}

export interface PluginUpgradeSQLSummary {
  installSqlCount: number;
  uninstallSqlCount: number;
  mockSqlCount: number;
  runtimeSqlAssetCount: number;
}

export interface PluginUpgradeHostServicesDiff {
  added: PluginUpgradeHostServiceChange[];
  removed: PluginUpgradeHostServiceChange[];
  changed: PluginUpgradeHostServiceChange[];
  authorizationRequired: boolean;
  authorizationChanged: boolean;
}

export interface PluginUpgradeHostServiceChange {
  service: string;
  fromMethods?: string[];
  toMethods?: string[];
  fromResourceCount: number;
  toResourceCount: number;
  fromTables?: string[];
  toTables?: string[];
  fromPaths?: string[];
  toPaths?: string[];
}

export interface PluginUpgradeResult {
  pluginId: string;
  runtimeState: PluginRuntimeState;
  effectiveVersion: string;
  discoveredVersion: string;
  fromVersion: string;
  toVersion: string;
  executed: boolean;
}

export interface PluginDynamicState {
  id: string;
  installed: number;
  enabled: number;
  version: string;
  generation: number;
  statusKey: string;
  runtimeState?: PluginRuntimeState;
}

export interface PluginUploadDynamicResult {
  id: string;
  name: string;
  version: string;
  type: PluginType;
  runtimeKind: string;
  runtimeAbi: string;
  installed: number;
  enabled: number;
}
