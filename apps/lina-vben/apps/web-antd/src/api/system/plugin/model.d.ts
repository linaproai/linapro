export type PluginType = 'dynamic' | 'source' | string;
export type PluginDistribution = 'builtin' | 'managed' | string;

export interface PluginListParams {
  pageNum?: number;
  pageSize?: number;
  id?: string;
  includeBuiltin?: boolean;
  installed?: number;
  name?: string;
  status?: number;
  type?: PluginType;
}

export interface PluginListItem {
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
  installedAt: number | null;
  enabled: number;
  autoEnableManaged: number;
  autoEnableForNewTenants?: boolean;
  supportsMultiTenant?: boolean;
  statusKey: string;
  updatedAt: number | null;
  authorizationRequired: number;
  authorizationStatus: 'confirmed' | 'not_required' | 'pending' | string;
  dependencyCheck?: PluginDependencyCheckResult;
  distribution?: PluginDistribution;
  hasMockData: number;
  installMode?: 'global' | 'tenant_scoped' | string;
  scopeNature?: 'platform_only' | 'tenant_aware' | string;
}

export interface SystemPlugin extends PluginListItem {
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
  installed: boolean;
  discovered: boolean;
  status: 'missing' | 'satisfied' | 'version_unsatisfied' | string;
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
  owner?: string;
  service: string;
  version?: string;
  methods: string[];
  paths?: string[];
  tables?: string[];
  tableItems?: HostServicePermissionTableItem[];
  resources?: HostServicePermissionResourceItem[];
}

export interface HostServicePermissionTableItem {
  name: string;
  comment?: string;
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
      owner?: string;
      methods?: string[];
      paths?: string[];
      resourceRefs?: string[];
      tables?: string[];
      service: string;
      version?: string;
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
  distribution?: PluginDistribution;
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
  owner?: string;
  service: string;
  version?: string;
  fromMethods?: string[];
  toMethods?: string[];
  fromResourceCount: number;
  toResourceCount: number;
  fromResourceRefs?: string[];
  toResourceRefs?: string[];
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
