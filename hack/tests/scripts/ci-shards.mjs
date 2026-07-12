/**
 * CI E2E shard planning: host capability shards + plugin load-balanced packing.
 * Keep this module free of hard-coded official plugin business aliases.
 */
import {
  isPluginWorkspaceReady,
  listPluginTcFiles,
  listTcFiles,
  loadManifest,
  resolveEntries,
  resolveHostOnlyEntries,
} from './execution-governance.mjs';

const SHARD_NAME_PATTERN = /^[a-z0-9]+(?:-[a-z0-9]+)*$/u;

export function ciShardsConfig(manifest = loadManifest()) {
  return manifest.ciShards ?? null;
}

export function defaultParallelWorkers(manifest = loadManifest()) {
  const value = ciShardsConfig(manifest)?.defaults?.parallelWorkers;
  const parsed = Number.parseInt(`${value ?? 2}`, 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 2;
}

export function listSourcePluginIdsWithTests() {
  const files = listPluginTcFiles();
  const ids = new Set();
  for (const file of files) {
    const match = /^apps\/lina-plugins\/([^/]+)\//u.exec(file);
    if (match?.[1]) {
      ids.add(match[1]);
    }
  }
  return [...ids].sort();
}

/**
 * Greedy multiprocessor bin packing: heaviest-first into the lightest bin.
 * Stable when weights tie by sorting plugin id ascending within equal weights
 * after the primary weight descending sort.
 */
export function binPackByWeight(items, targetBinCount) {
  if (!Array.isArray(items) || items.length === 0) {
    return [];
  }
  const binCount = Math.max(1, Math.min(targetBinCount, items.length));
  const ordered = [...items].sort((left, right) => {
    if (right.weight !== left.weight) {
      return right.weight - left.weight;
    }
    return String(left.id).localeCompare(String(right.id));
  });

  const bins = Array.from({ length: binCount }, (_, index) => ({
    index,
    weight: 0,
    items: [],
  }));

  for (const item of ordered) {
    let best = bins[0];
    for (const bin of bins) {
      if (bin.weight < best.weight || (bin.weight === best.weight && bin.index < best.index)) {
        best = bin;
      }
    }
    best.items.push(item);
    best.weight += item.weight;
  }

  return bins
    .filter((bin) => bin.items.length > 0)
    .sort((left, right) => left.index - right.index)
    .map((bin, shardIndex, all) => ({
      name: `plugins-${shardIndex + 1}-of-${all.length}`,
      weight: bin.weight,
      pluginIds: bin.items.map((item) => item.id).sort(),
      entries: bin.items
        .map((item) => item.entries ?? [`plugins/${item.id}`])
        .flat()
        .sort(),
      files: uniqueFiles(bin.items.flatMap((item) => item.files ?? [])),
    }));
}

function uniqueFiles(files) {
  return [...new Set(files)].sort();
}

export function planHostShards(manifest = loadManifest()) {
  const config = ciShardsConfig(manifest);
  const host = config?.host;
  if (!Array.isArray(host)) {
    return { shards: [], errors: ['ciShards.host must be an array.'] };
  }

  const errors = [];
  const shards = [];
  const covered = new Map();

  for (const [index, item] of host.entries()) {
    const owner = `ciShards.host[${index}]`;
    if (!item || typeof item !== 'object') {
      errors.push(`${owner} must be an object.`);
      continue;
    }
    const name = typeof item.name === 'string' ? item.name.trim() : '';
    if (!name || !SHARD_NAME_PATTERN.test(name)) {
      errors.push(`${owner}.name must match ${SHARD_NAME_PATTERN}: ${name || '(empty)'}`);
    }
    const entries = item.entries;
    if (!Array.isArray(entries) || entries.length === 0) {
      errors.push(`${owner}.entries must be a non-empty array.`);
      continue;
    }
    for (const entry of entries) {
      if (typeof entry !== 'string' || entry.trim() === '') {
        errors.push(`${owner}.entries contains an invalid entry.`);
      }
    }
    const files = resolveHostOnlyEntries(entries);
    if (files.length === 0) {
      errors.push(`${owner} resolves to zero host-only TC files.`);
    }
    for (const file of files) {
      if (covered.has(file)) {
        errors.push(
          `Host TC file ${file} is covered by both host shards "${covered.get(file)}" and "${name}".`,
        );
      } else {
        covered.set(file, name);
      }
    }
    shards.push({
      name,
      kind: 'host',
      entries: [...entries],
      files,
      command: `node ./scripts/run-suite.mjs ci-shard host ${name}`,
    });
  }

  const expected = new Set(resolveHostOnlyEntries(['e2e']));
  for (const file of expected) {
    if (!covered.has(file)) {
      errors.push(`Host TC file is not covered by any ciShards.host entry: ${file}`);
    }
  }
  for (const file of covered.keys()) {
    if (!expected.has(file)) {
      errors.push(
        `Host shard covers non-host-only or unexpected file (check hostOnlyExcludedEntries): ${file}`,
      );
    }
  }

  const names = shards.map((shard) => shard.name);
  if (new Set(names).size !== names.length) {
    errors.push('ciShards.host names must be unique.');
  }

  return { shards, errors };
}

export function planPluginShards(manifest = loadManifest(), options = {}) {
  const config = ciShardsConfig(manifest)?.plugin ?? {};
  const packing = config.packing ?? 'file-count-binpack';
  const errors = [];

  if (packing !== 'file-count-binpack') {
    errors.push(`Unsupported ciShards.plugin.packing: ${packing}`);
  }

  if (!isPluginWorkspaceReady() && !options.allowEmptyWorkspace) {
    return {
      shards: [],
      errors: options.requireWorkspace
        ? ['Official plugin workspace is required to plan plugin CI shards.']
        : [],
      skipped: true,
    };
  }

  if (!isPluginWorkspaceReady()) {
    return { shards: [], errors: [], skipped: true };
  }

  const weightOverrides =
    config.weightOverrides && typeof config.weightOverrides === 'object'
      ? config.weightOverrides
      : {};
  const targetShardCount = Number.parseInt(`${config.targetShardCount ?? 8}`, 10);
  if (!Number.isFinite(targetShardCount) || targetShardCount < 1) {
    errors.push('ciShards.plugin.targetShardCount must be a positive integer.');
  }

  const items = [];
  for (const pluginId of listSourcePluginIdsWithTests()) {
    const files = listTcFiles(`plugins/${pluginId}`);
    if (files.length === 0) {
      continue;
    }
    const override = weightOverrides[pluginId];
    const weight =
      override === undefined || override === null
        ? files.length
        : Number.parseInt(`${override}`, 10);
    if (!Number.isFinite(weight) || weight < 1) {
      errors.push(`ciShards.plugin.weightOverrides["${pluginId}"] must be a positive integer.`);
      continue;
    }
    items.push({
      id: pluginId,
      files,
      weight,
      entries: [`plugins/${pluginId}`],
    });
  }

  if (items.length === 0) {
    errors.push('Plugin workspace is ready but no source-plugin TC files were discovered.');
    return { shards: [], errors };
  }

  const packed = binPackByWeight(items, targetShardCount);
  const covered = new Set();
  const shards = packed.map((bin) => {
    for (const file of bin.files) {
      if (covered.has(file)) {
        errors.push(`Plugin TC file packed into multiple shards: ${file}`);
      }
      covered.add(file);
    }
    return {
      name: bin.name,
      kind: 'plugin',
      entries: bin.entries,
      pluginIds: bin.pluginIds,
      files: bin.files,
      weight: bin.weight,
      command: `node ./scripts/run-suite.mjs ci-shard plugin ${bin.name}`,
    };
  });

  const expected = new Set(listPluginTcFiles());
  for (const file of expected) {
    if (!covered.has(file)) {
      errors.push(`Plugin TC file is not covered by plugin packing: ${file}`);
    }
  }

  return { shards, errors, skipped: false };
}

export function planPluginFullExtraShards(manifest = loadManifest()) {
  const config = ciShardsConfig(manifest);
  const extras = config?.pluginFullExtra;
  const errors = [];
  if (!Array.isArray(extras)) {
    return { shards: [], errors: ['ciShards.pluginFullExtra must be an array.'] };
  }

  const shards = [];
  for (const [index, item] of extras.entries()) {
    const owner = `ciShards.pluginFullExtra[${index}]`;
    if (!item || typeof item !== 'object') {
      errors.push(`${owner} must be an object.`);
      continue;
    }
    const name = typeof item.name === 'string' ? item.name.trim() : '';
    if (!name || !SHARD_NAME_PATTERN.test(name)) {
      errors.push(`${owner}.name must match ${SHARD_NAME_PATTERN}: ${name || '(empty)'}`);
    }
    const scope = typeof item.scope === 'string' ? item.scope.trim() : '';
    if (!scope) {
      errors.push(`${owner}.scope is required.`);
      continue;
    }
    const configuredEntries = manifest.moduleScopes?.[scope];
    if (!configuredEntries) {
      errors.push(`${owner}.scope is not a known moduleScopes key: ${scope}`);
      continue;
    }
    const files = resolveEntries(configuredEntries);
    if (files.length === 0) {
      errors.push(`${owner} scope resolves to zero TC files: ${scope}`);
    }
    shards.push({
      name,
      kind: 'plugin-full-extra',
      scope,
      entries: [...configuredEntries],
      files,
      command: `node ./scripts/run-suite.mjs ci-shard plugin-full-extra ${name}`,
    });
  }

  const names = shards.map((shard) => shard.name);
  if (new Set(names).size !== names.length) {
    errors.push('ciShards.pluginFullExtra names must be unique.');
  }

  return { shards, errors };
}

export function planAllCiShards(manifest = loadManifest(), options = {}) {
  const host = planHostShards(manifest);
  const plugin = planPluginShards(manifest, options);
  const extra = planPluginFullExtraShards(manifest);
  return {
    defaults: {
      parallelWorkers: defaultParallelWorkers(manifest),
    },
    host: host.shards,
    plugin: plugin.shards,
    pluginFullExtra: extra.shards,
    pluginFull: [...extra.shards, ...plugin.shards],
    errors: [...host.errors, ...plugin.errors, ...extra.errors],
    pluginSkipped: Boolean(plugin.skipped),
  };
}

export function findHostShard(name, manifest = loadManifest()) {
  const { shards, errors } = planHostShards(manifest);
  if (errors.length > 0) {
    throw new Error(errors.join('\n'));
  }
  const shard = shards.find((item) => item.name === name);
  if (!shard) {
    throw new Error(
      `Unknown host CI shard "${name}". Known: ${shards.map((item) => item.name).join(', ')}`,
    );
  }
  return shard;
}

export function findPluginShard(name, manifest = loadManifest()) {
  const { shards, errors } = planPluginShards(manifest, { requireWorkspace: true });
  if (errors.length > 0) {
    throw new Error(errors.join('\n'));
  }
  const shard = shards.find((item) => item.name === name);
  if (!shard) {
    throw new Error(
      `Unknown plugin CI shard "${name}". Known: ${shards.map((item) => item.name).join(', ')}`,
    );
  }
  return shard;
}

export function findPluginFullExtraShard(name, manifest = loadManifest()) {
  const { shards, errors } = planPluginFullExtraShards(manifest);
  if (errors.length > 0) {
    throw new Error(errors.join('\n'));
  }
  const shard = shards.find((item) => item.name === name);
  if (!shard) {
    throw new Error(
      `Unknown plugin-full-extra CI shard "${name}". Known: ${shards.map((item) => item.name).join(', ')}`,
    );
  }
  return shard;
}

export function validateCiShards(manifest = loadManifest(), options = {}) {
  const plan = planAllCiShards(manifest, {
    requireWorkspace: options.requireWorkspace ?? isPluginWorkspaceReady(),
    allowEmptyWorkspace: options.allowEmptyWorkspace ?? false,
  });
  return plan.errors;
}

export function toGitHubMatrix(shards) {
  return shards.map((shard) => ({
    name: shard.name,
    command: shard.command,
  }));
}
