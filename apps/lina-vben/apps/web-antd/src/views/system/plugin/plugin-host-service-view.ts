import type {
  HostServicePermissionItem,
  HostServicePermissionResourceItem,
  HostServicePermissionTableItem,
} from '#/api/system/plugin/model';

import { $t } from '#/locales';

export interface HostServiceScopeView {
  badgeColor: string;
  containerTestId?: string;
  emptyText: string;
  hint?: string;
  itemTestIdPrefix?: string;
  key: string;
  label: string;
  methods: string[];
  targetSummaryBadgeColor?: string;
  targetSummaryLabel?: string;
  targets: HostServiceTargetView[];
}

export interface HostServiceCardView {
  identityKey: string;
  identityTestIdValue: string;
  owner: string | undefined;
  scopes: HostServiceScopeView[];
  service: string;
  title: string;
  version: string | undefined;
}

export interface HostServiceTargetView {
  details?: HostServiceTargetDetailView[];
  label: string;
  testIdValue: string;
  variant?: 'panel' | 'tag';
}

export interface HostServiceTargetDetailView {
  label: string;
  value: string;
}

type HostServiceScopeKind = 'authorized' | 'requested';

interface HostServiceCardBuildOptions {
  authorizationRequired?: boolean;
  buildScopeContainerTestId?: (
    service: string,
    scopeKey: string,
  ) => string | undefined;
  buildScopeItemTestIdPrefix?: (
    service: string,
    scopeKey: string,
  ) => string | undefined;
  targetSummaryBadgeColor?: string;
}

/**
 * Stable display order for known host service families.
 * Values align with common authorization-review scanning order: data plane
 * first, then infra, identity, and remaining catalog services.
 */
const hostServiceOrder: Record<string, number> = {
  data: 0,
  storage: 1,
  network: 2,
  runtime: 3,
  jobs: 4,
  plugins: 5,
  notifications: 6,
  cache: 7,
  lock: 8,
  files: 9,
  users: 10,
  sessions: 11,
  auth: 12,
  tenant: 13,
  org: 14,
  dict: 15,
  route: 16,
  apidoc: 17,
  bizctx: 18,
  hostconfig: 19,
  manifest: 20,
  ai: 21,
};

/**
 * Known host service wire names that have dedicated display labels.
 * Keep in sync with pages.system.plugin.hostServices.service.* i18n keys
 * (excluding the shared "unknown" fallback key).
 */
export const knownHostServiceLabels = [
  'ai',
  'apidoc',
  'auth',
  'bizctx',
  'cache',
  'data',
  'dict',
  'files',
  'hostconfig',
  'jobs',
  'lock',
  'manifest',
  'network',
  'notifications',
  'org',
  'plugins',
  'route',
  'runtime',
  'sessions',
  'storage',
  'tenant',
  'users',
] as const;

const hostServiceLabelKeyPrefix = 'pages.system.plugin.hostServices.service';

export function sortHostServices(items?: HostServicePermissionItem[]) {
  return [...(items ?? [])].sort(compareHostServiceIdentity);
}

/**
 * Resolve a localized display title for a host service wire name.
 * Known services use dedicated i18n keys; unknown services use a consistent
 * fallback that still surfaces the raw wire identifier for operators.
 */
export function formatServiceLabel(service: string) {
  const normalized = (service ?? '').trim();
  if (!normalized) {
    return '';
  }
  if ((knownHostServiceLabels as readonly string[]).includes(normalized)) {
    return $t(`${hostServiceLabelKeyPrefix}.${normalized}`);
  }
  return $t(`${hostServiceLabelKeyPrefix}.unknown`, { service: normalized });
}

export function buildPluginDetailHostServiceCards(
  requested?: HostServicePermissionItem[],
  authorized?: HostServicePermissionItem[],
) {
  const requestedMap = buildServiceMap(requested);
  const authorizedMap = buildServiceMap(authorized);
  const serviceKeys = sortServiceKeys(requestedMap, authorizedMap);

  return serviceKeys
    .map((serviceKey) => {
      const requestedService = requestedMap.get(serviceKey);
      const authorizedService = authorizedMap.get(serviceKey);
      const displayService = authorizedService ?? requestedService;
      if (!displayService) {
        return null;
      }
      const identity = buildHostServiceIdentity(displayService);

      const sameAsAuthorized =
        requestedService &&
        authorizedService &&
        buildServiceCompareKey(requestedService) ===
          buildServiceCompareKey(authorizedService);

      const scopes: HostServiceScopeView[] = [];
      if (sameAsAuthorized && authorizedService) {
        scopes.push(
          buildScopeView({
            badgeColor: 'green',
            label: $t('pages.system.plugin.hostServices.scope.effective'),
            kind: 'authorized',
            scopeKey: `${identity.testIdValue}-effective`,
            service: authorizedService,
          }),
        );
      } else {
        if (requestedService) {
          scopes.push(
            buildScopeView({
              badgeColor: 'blue',
              kind: 'requested',
              label: $t('pages.system.plugin.hostServices.scope.requested'),
              scopeKey: `${identity.testIdValue}-requested`,
              service: requestedService,
            }),
          );
        }
        if (authorizedService) {
          scopes.push(
            buildScopeView({
              badgeColor: 'green',
              kind: 'authorized',
              label: $t('pages.system.plugin.hostServices.scope.authorized'),
              scopeKey: `${identity.testIdValue}-authorized`,
              service: authorizedService,
            }),
          );
        }
      }

      return {
        identityKey: identity.key,
        identityTestIdValue: identity.testIdValue,
        owner: identity.owner || undefined,
        scopes,
        service: identity.service,
        title: formatServiceLabel(identity.service),
        version: identity.version || undefined,
      } satisfies HostServiceCardView;
    })
    .filter((item): item is HostServiceCardView => item !== null);
}

export function buildPluginAuthorizationHostServiceCards(
  requested?: HostServicePermissionItem[],
  options: HostServiceCardBuildOptions = {},
) {
  return sortHostServices(requested).map((service) => {
    const identity = buildHostServiceIdentity(service);
    return {
      identityKey: identity.key,
      identityTestIdValue: identity.testIdValue,
      owner: identity.owner || undefined,
      scopes: [
        buildScopeView({
          badgeColor: 'green',
          buildScopeContainerTestId: options.buildScopeContainerTestId,
          buildScopeItemTestIdPrefix: options.buildScopeItemTestIdPrefix,
          kind: options.authorizationRequired ? 'requested' : 'authorized',
          label: options.authorizationRequired
            ? $t('pages.system.plugin.hostServices.scope.requested')
            : $t('pages.system.plugin.hostServices.scope.effective'),
          scopeKey: `${identity.testIdValue}-review`,
          service,
          targetSummaryBadgeColor: options.targetSummaryBadgeColor,
        }),
      ],
      service: identity.service,
      title: formatServiceLabel(identity.service),
      version: identity.version || undefined,
    } satisfies HostServiceCardView;
  });
}

function buildServiceMap(items?: HostServicePermissionItem[]) {
  const map = new Map<string, HostServicePermissionItem>();
  for (const item of sortHostServices(items)) {
    map.set(buildHostServiceIdentity(item).key, item);
  }
  return map;
}

function sortServiceKeys(
  requested: Map<string, HostServicePermissionItem>,
  authorized: Map<string, HostServicePermissionItem>,
) {
  return [...new Set([...requested.keys(), ...authorized.keys()])].sort(
    (left, right) => {
      const leftService = requested.get(left) ?? authorized.get(left);
      const rightService = requested.get(right) ?? authorized.get(right);
      if (!leftService || !rightService) {
        return left.localeCompare(right);
      }
      return compareHostServiceIdentity(leftService, rightService);
    },
  );
}

function buildScopeView(input: {
  badgeColor: string;
  buildScopeContainerTestId?: (
    service: string,
    scopeKey: string,
  ) => string | undefined;
  buildScopeItemTestIdPrefix?: (
    service: string,
    scopeKey: string,
  ) => string | undefined;
  hint?: string;
  kind: HostServiceScopeKind;
  label: string;
  scopeKey: string;
  service: HostServicePermissionItem;
  targetSummaryBadgeColor?: string;
}) {
  return {
    badgeColor: input.badgeColor,
    containerTestId: input.buildScopeContainerTestId?.(
      input.service.service,
      input.scopeKey,
    ),
    emptyText: resolveScopeEmptyText(),
    hint: input.hint,
    itemTestIdPrefix: input.buildScopeItemTestIdPrefix?.(
      input.service.service,
      input.scopeKey,
    ),
    key: input.scopeKey,
    label: input.label,
    methods: normalizeStringList(input.service.methods),
    targetSummaryBadgeColor:
      input.targetSummaryBadgeColor ??
      resolveTargetSummaryBadgeColor(input.kind),
    targetSummaryLabel: resolveTargetSummaryLabel(input.service),
    targets: resolveServiceTargets(input.service),
  } satisfies HostServiceScopeView;
}

function buildServiceCompareKey(service: HostServicePermissionItem) {
  return JSON.stringify({
    owner: normalizeIdentitySegment(service.owner),
    service: normalizeIdentitySegment(service.service),
    version: normalizeIdentitySegment(service.version),
    methods: normalizeStringList(service.methods),
    paths: normalizeStringList(service.paths),
    resources: normalizeStringList(
      (service.resources ?? []).map((item) => item.ref),
    ),
    tables: normalizeStringList(resolveDataTableNames(service)),
  });
}

function buildHostServiceIdentity(service: HostServicePermissionItem) {
  const owner = normalizeIdentitySegment(service.owner);
  const serviceName = normalizeIdentitySegment(service.service);
  const version = normalizeIdentitySegment(service.version);
  const key = owner ? `${owner}/${serviceName}/${version}` : serviceName;
  return {
    key,
    owner,
    service: serviceName,
    testIdValue: key.replaceAll(/[^A-Za-z0-9_-]+/gu, '-'),
    version,
  };
}

function compareHostServiceIdentity(
  left: HostServicePermissionItem,
  right: HostServicePermissionItem,
) {
  const leftIdentity = buildHostServiceIdentity(left);
  const rightIdentity = buildHostServiceIdentity(right);
  const leftOrder =
    hostServiceOrder[leftIdentity.service] ?? Number.MAX_SAFE_INTEGER;
  const rightOrder =
    hostServiceOrder[rightIdentity.service] ?? Number.MAX_SAFE_INTEGER;
  if (leftOrder !== rightOrder) {
    return leftOrder - rightOrder;
  }
  const serviceCompare = leftIdentity.service.localeCompare(
    rightIdentity.service,
  );
  if (serviceCompare !== 0) {
    return serviceCompare;
  }
  const ownerCompare = leftIdentity.owner.localeCompare(rightIdentity.owner);
  if (ownerCompare !== 0) {
    return ownerCompare;
  }
  return leftIdentity.version.localeCompare(rightIdentity.version);
}

function normalizeIdentitySegment(value?: string) {
  return (value ?? '').trim();
}

function resolveServiceTargets(service: HostServicePermissionItem) {
  if (service.service === 'storage') {
    return normalizeStringList(service.paths).map((path) => ({
      label: path,
      testIdValue: path,
      variant: 'tag' as const,
    }));
  }
  if (service.service === 'data') {
    return resolveDataTableItems(service).map((table) => ({
      label: formatDataTableLabel(table),
      testIdValue: table.name,
      variant: 'tag' as const,
    }));
  }
  return (service.resources ?? []).map((item) => ({
    label: formatResourceLabel(item),
    testIdValue: item.ref,
    variant: 'tag' as const,
  }));
}

function resolveTargetSummaryLabel(service: HostServicePermissionItem) {
  if (
    (service.paths ?? []).length === 0 &&
    (service.tables ?? []).length === 0 &&
    (service.tableItems ?? []).length === 0 &&
    (service.resources ?? []).length === 0
  ) {
    return undefined;
  }
  switch (service.service) {
    case 'data': {
      return $t('pages.system.plugin.hostServices.summary.table');
    }
    case 'network': {
      return $t('pages.system.plugin.hostServices.summary.path');
    }
    case 'storage': {
      return $t('pages.system.plugin.hostServices.summary.storage');
    }
    default: {
      return $t('pages.system.plugin.hostServices.summary.resource');
    }
  }
}

function resolveTargetSummaryBadgeColor(kind: HostServiceScopeKind) {
  return kind === 'requested' ? 'cyan' : 'gold';
}

function resolveDataTableItems(service: HostServicePermissionItem) {
  if ((service.tableItems ?? []).length > 0) {
    return [...(service.tableItems ?? [])];
  }
  return normalizeStringList(
    service.tables,
  ).map<HostServicePermissionTableItem>((table) => ({
    name: table,
  }));
}

function resolveDataTableNames(service: HostServicePermissionItem) {
  if ((service.tableItems ?? []).length > 0) {
    return (service.tableItems ?? []).map((table) => table.name);
  }
  return service.tables ?? [];
}

function resolveScopeEmptyText() {
  return $t('pages.system.plugin.hostServices.messages.defaultEmpty');
}

function formatDataTableLabel(table: HostServicePermissionTableItem) {
  return table.comment ? `${table.name} (${table.comment})` : table.name;
}

function formatResourceLabel(resource: HostServicePermissionResourceItem) {
  const methods = normalizeStringList(resource.allowMethods);
  if (methods.length === 0) {
    return resource.ref;
  }
  return `${resource.ref} [${methods.join(', ')}]`;
}

function normalizeStringList(items?: string[]) {
  return [
    ...new Set((items ?? []).map((item) => item.trim()).filter(Boolean)),
  ].sort((left, right) => left.localeCompare(right));
}
