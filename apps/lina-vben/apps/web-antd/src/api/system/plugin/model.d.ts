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
  type: PluginType;
  description: string;
  installed: number;
  installedAt: string;
  enabled: number;
  statusKey: string;
  updatedAt: string;
  authorizationRequired: number;
  authorizationStatus: 'confirmed' | 'not_required' | 'pending' | string;
  requestedHostServices?: HostServicePermissionItem[];
  authorizedHostServices?: HostServicePermissionItem[];
}

export interface HostServicePermissionItem {
  service: string;
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
      methods?: string[];
      paths?: string[];
      resourceRefs?: string[];
      tables?: string[];
      service: string;
    }>;
  };
}

export interface PluginDynamicState {
  id: string;
  installed: number;
  enabled: number;
  version: string;
  generation: number;
  statusKey: string;
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
