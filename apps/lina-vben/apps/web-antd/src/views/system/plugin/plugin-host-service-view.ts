import type {
  HostServicePermissionCronItem,
  HostServicePermissionItem,
  HostServicePermissionResourceItem,
  HostServicePermissionTableItem,
} from '#/api/system/plugin/model';

import {
  getJobConcurrencyLabel,
  getJobScopeLabel,
} from '#/api/system/job/meta';
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
  scopes: HostServiceScopeView[];
  service: string;
  title: string;
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

const hostServiceOrder: Record<string, number> = {
  data: 0,
  storage: 1,
  network: 2,
  cron: 3,
  runtime: 4,
};

export function sortHostServices(items?: HostServicePermissionItem[]) {
  return [...(items ?? [])].sort((left, right) => {
    const leftOrder = hostServiceOrder[left.service] ?? Number.MAX_SAFE_INTEGER;
    const rightOrder = hostServiceOrder[right.service] ?? Number.MAX_SAFE_INTEGER;
    if (leftOrder !== rightOrder) {
      return leftOrder - rightOrder;
    }
    return left.service.localeCompare(right.service);
  });
}

export function formatServiceLabel(service: string) {
  switch (service) {
    case 'data': {
      return $t('pages.system.plugin.hostServices.service.data');
    }
    case 'network': {
      return $t('pages.system.plugin.hostServices.service.network');
    }
    case 'cron': {
      return $t('pages.system.plugin.hostServices.service.cron');
    }
    case 'runtime': {
      return $t('pages.system.plugin.hostServices.service.runtime');
    }
    case 'storage': {
      return $t('pages.system.plugin.hostServices.service.storage');
    }
    default: {
      return service;
    }
  }
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
            scopeKey: `${serviceKey}-effective`,
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
              scopeKey: `${serviceKey}-requested`,
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
              scopeKey: `${serviceKey}-authorized`,
              service: authorizedService,
            }),
          );
        }
      }

      return {
        scopes,
        service: serviceKey,
        title: formatServiceLabel(serviceKey),
      } satisfies HostServiceCardView;
    })
    .filter((item): item is HostServiceCardView => item !== null);
}

export function buildPluginAuthorizationHostServiceCards(
  requested?: HostServicePermissionItem[],
  options: HostServiceCardBuildOptions = {},
) {
  return sortHostServices(requested).map((service) => ({
    scopes: [
      buildScopeView({
        badgeColor: 'green',
        buildScopeContainerTestId: options.buildScopeContainerTestId,
        buildScopeItemTestIdPrefix: options.buildScopeItemTestIdPrefix,
        kind: options.authorizationRequired ? 'requested' : 'authorized',
        label: options.authorizationRequired
          ? $t('pages.system.plugin.hostServices.scope.requested')
          : $t('pages.system.plugin.hostServices.scope.effective'),
        scopeKey: `${service.service}-review`,
        service,
        targetSummaryBadgeColor: options.targetSummaryBadgeColor,
      }),
    ],
    service: service.service,
    title: formatServiceLabel(service.service),
  }));
}

function buildServiceMap(items?: HostServicePermissionItem[]) {
  const map = new Map<string, HostServicePermissionItem>();
  for (const item of sortHostServices(items)) {
    map.set(item.service, item);
  }
  return map;
}

function sortServiceKeys(
  requested: Map<string, HostServicePermissionItem>,
  authorized: Map<string, HostServicePermissionItem>,
) {
  return [...new Set([...requested.keys(), ...authorized.keys()])].sort(
    (left, right) => {
      const leftOrder = hostServiceOrder[left] ?? Number.MAX_SAFE_INTEGER;
      const rightOrder = hostServiceOrder[right] ?? Number.MAX_SAFE_INTEGER;
      if (leftOrder !== rightOrder) {
        return leftOrder - rightOrder;
      }
      return left.localeCompare(right);
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
    emptyText: resolveScopeEmptyText(input.service),
    hint: input.hint,
    itemTestIdPrefix: input.buildScopeItemTestIdPrefix?.(
      input.service.service,
      input.scopeKey,
    ),
    key: input.scopeKey,
    label: input.label,
    methods: normalizeStringList(input.service.methods),
    targetSummaryBadgeColor:
      input.targetSummaryBadgeColor ?? resolveTargetSummaryBadgeColor(input.kind),
    targetSummaryLabel: resolveTargetSummaryLabel(input.service),
    targets: resolveServiceTargets(input.service),
  } satisfies HostServiceScopeView;
}

function buildServiceCompareKey(service: HostServicePermissionItem) {
  return JSON.stringify({
    methods: normalizeStringList(service.methods),
    paths: normalizeStringList(service.paths),
    resources: normalizeStringList(
      (service.resources ?? []).map((item) => item.ref),
    ),
    tables: normalizeStringList(resolveDataTableNames(service)),
  });
}

function resolveServiceTargets(service: HostServicePermissionItem) {
  if (service.service === 'cron') {
    return resolveCronTargets(service);
  }
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
    (service.cronItems ?? []).length === 0 &&
    (service.tableItems ?? []).length === 0 &&
    (service.resources ?? []).length === 0
  ) {
    return undefined;
  }
  switch (service.service) {
    case 'cron': {
      return undefined;
    }
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
  return normalizeStringList(service.tables).map<HostServicePermissionTableItem>(
    (table) => ({
      name: table,
    }),
  );
}

function resolveDataTableNames(service: HostServicePermissionItem) {
  if ((service.tableItems ?? []).length > 0) {
    return (service.tableItems ?? []).map((table) => table.name);
  }
  return service.tables ?? [];
}

function resolveCronTargets(service: HostServicePermissionItem) {
  return (service.cronItems ?? []).map((item) => ({
    details: buildCronTargetDetails(item),
    label: formatCronTargetLabel(item),
    testIdValue: item.name,
    variant: 'panel' as const,
  }));
}

function formatCronTargetLabel(item: HostServicePermissionCronItem) {
  const displayName = (item.displayName || '').trim();
  const name = (item.name || '').trim();
  if (displayName && name && displayName !== name) {
    return `${displayName} (${name})`;
  }
  return displayName || name || '-';
}

function buildCronTargetDetails(item: HostServicePermissionCronItem) {
  const lines: HostServiceTargetDetailView[] = [
    {
      label: $t('pages.system.plugin.hostServices.cron.pattern'),
      value: (item.pattern || '').trim() || '-',
    },
    {
      label: $t('pages.system.job.fields.scope'),
      value: getJobScopeLabel(item.scope),
    },
    {
      label: $t('pages.system.job.fields.concurrency'),
      value: formatCronConcurrencySummary(item),
    },
  ];
  const timezone = (item.timezone || '').trim();
  if (timezone) {
    lines.push({
      label: $t('pages.system.job.fields.timezone'),
      value: timezone,
    });
  }
  const description = (item.description || '').trim();
  if (description) {
    lines.push({
      label: $t('pages.fields.description'),
      value: description,
    });
  }
  return lines;
}

function formatCronConcurrencySummary(item: HostServicePermissionCronItem) {
  const label = getJobConcurrencyLabel(item.concurrency);
  if (item.concurrency === 'parallel' && Number(item.maxConcurrency || 0) > 0) {
    return $t('pages.system.plugin.hostServices.cron.parallelLimit', {
      label,
      value: item.maxConcurrency,
    });
  }
  return label;
}

function resolveScopeEmptyText(service: HostServicePermissionItem) {
  if (service.service === 'cron') {
    return $t('pages.system.plugin.hostServices.messages.cronEmpty');
  }
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
  return [...new Set((items ?? []).map((item) => item.trim()).filter(Boolean))].sort(
    (left, right) => left.localeCompare(right),
  );
}
