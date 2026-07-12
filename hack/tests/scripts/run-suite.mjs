import os from 'node:os';
import { spawnSync } from 'node:child_process';

import {
  findHostShard,
  findPluginFullExtraShard,
  findPluginShard,
  defaultParallelWorkers,
} from './ci-shards.mjs';
import {
  listTcFiles,
  loadManifest,
  moduleHasPluginWorkspaceEntries,
  moduleRequiresPluginWorkspace,
  playwrightFileArg,
  pluginTestEntry,
  requirePluginWorkspace,
  resolveHostOnlyEntries,
  resolveEntries,
  splitHostOnlyBySerial,
  splitBySerial,
  summarizeIsolationCategories,
  testsDir,
} from './execution-governance.mjs';
import { resolveHostOnlyPluginsEnv } from './host-only-env.mjs';

const manifest = loadManifest();

const mode = process.argv[2] ?? 'full';
const extraArgs = normalizePassthroughArgs(process.argv.slice(3));
const configuredDefaultWorkers = defaultParallelWorkers(manifest);
const parallelWorkers = Number.parseInt(
  process.env.E2E_PARALLEL_WORKERS ??
    `${process.env.CI ? configuredDefaultWorkers : Math.min(Math.max(os.cpus().length - 1, 2), 4)}`,
  10,
);

function runPlaywright(files, workers, label) {
  if (files.length === 0) {
    return 0;
  }

  const runnableFiles = files.map(playwrightFileArg);
  const args = ['exec', 'playwright', 'test', ...runnableFiles, `--workers=${workers}`, ...extraArgs];
  console.log(`\n[${label}] playwright test ${files.length} file(s), workers=${workers}`);
  const env = {
    ...process.env,
    E2E_HOST_ONLY_PLUGINS: resolveHostOnlyPluginsEnv(
      label,
      process.env.E2E_HOST_ONLY_PLUGINS,
    ),
  };
  const result = spawnSync('pnpm', args, {
    cwd: testsDir,
    stdio: 'inherit',
    env,
  });
  return result.status ?? 1;
}

function normalizePassthroughArgs(args) {
  return args[0] === '--' ? args.slice(1) : args;
}

function formatCategorySummary(entries) {
  if (entries.length === 0) {
    return 'none';
  }
  return entries.map(([category, count]) => `${category}=${count}`).join(', ');
}

function runMode(entries, label) {
  const { files, parallelFiles, serialFiles } = splitBySerial(entries, manifest);
  const categorySummary = summarizeIsolationCategories(serialFiles, manifest);
  console.log(
    [
      `[${label}] selected=${files.length}`,
      `parallel=${parallelFiles.length}`,
      `serial=${serialFiles.length}`,
      `parallelWorkers=${Math.max(parallelWorkers, 1)}`,
      `serialIsolation=${formatCategorySummary(categorySummary)}`,
    ].join(' '),
  );

  const parallelStatus = runPlaywright(parallelFiles, Math.max(parallelWorkers, 1), `${label}:parallel`);
  if (parallelStatus !== 0) {
    return parallelStatus;
  }
  return runPlaywright(serialFiles, 1, `${label}:serial`);
}

function runHostMode(entries, label) {
  const { files, parallelFiles, serialFiles } = splitHostOnlyBySerial(entries, manifest);
  const categorySummary = summarizeIsolationCategories(serialFiles, manifest);
  console.log(
    [
      `[${label}] selected=${files.length}`,
      `parallel=${parallelFiles.length}`,
      `serial=${serialFiles.length}`,
      `parallelWorkers=${Math.max(parallelWorkers, 1)}`,
      `serialIsolation=${formatCategorySummary(categorySummary)}`,
    ].join(' '),
  );

  const parallelStatus = runPlaywright(parallelFiles, Math.max(parallelWorkers, 1), `${label}:parallel`);
  if (parallelStatus !== 0) {
    return parallelStatus;
  }
  return runPlaywright(serialFiles, 1, `${label}:serial`);
}

function runHostModuleMode(entries, label) {
  const { files, parallelFiles, serialFiles } = splitHostOnlyBySerial(entries, manifest);
  if (files.length === 0) {
    return null;
  }
  const categorySummary = summarizeIsolationCategories(serialFiles, manifest);
  console.log(
    [
      `[${label}] selected=${files.length}`,
      `parallel=${parallelFiles.length}`,
      `serial=${serialFiles.length}`,
      `parallelWorkers=${Math.max(parallelWorkers, 1)}`,
      `serialIsolation=${formatCategorySummary(categorySummary)}`,
    ].join(' '),
  );

  const parallelStatus = runPlaywright(parallelFiles, Math.max(parallelWorkers, 1), `${label}:parallel`);
  if (parallelStatus !== 0) {
    return parallelStatus;
  }
  return runPlaywright(serialFiles, 1, `${label}:serial`);
}

function resolveModuleEntries(scope) {
  const configuredEntries = manifest.moduleScopes[scope];
  if (configuredEntries) {
    return configuredEntries;
  }

  const pluginMatch = /^plugin:([^/]+)$/u.exec(scope);
  if (!pluginMatch) {
    return null;
  }
  return [`plugins/${pluginMatch[1]}`];
}

let exitCode = 0;
if (mode === 'full') {
  try {
    requirePluginWorkspace();
  } catch (error) {
    console.error(error instanceof Error ? error.message : String(error));
    process.exit(1);
  }
  exitCode = runMode(['e2e', 'plugins'], 'full');
} else if (mode === 'host') {
  exitCode = runHostMode(['e2e'], 'host');
} else if (mode === 'smoke') {
  exitCode = runMode(manifest.smoke ?? [], 'smoke');
} else if (mode === 'ci-shard') {
  const kind = process.argv[3];
  const name = process.argv[4];
  const passthroughArgs = normalizePassthroughArgs(process.argv.slice(5));
  extraArgs.splice(0, extraArgs.length, ...passthroughArgs);
  if (!kind || !name) {
    console.error(
      'Usage: node ./scripts/run-suite.mjs ci-shard <host|plugin|plugin-full-extra> <shard-name>',
    );
    process.exit(1);
  }
  try {
    if (kind === 'host') {
      const shard = findHostShard(name, manifest);
      console.log(`[ci-shard:host:${name}] entries=${shard.entries.join(',')} files=${shard.files.length}`);
      exitCode = runHostMode(shard.entries, `ci-shard:host:${name}`);
    } else if (kind === 'plugin') {
      requirePluginWorkspace();
      const shard = findPluginShard(name, manifest);
      console.log(
        `[ci-shard:plugin:${name}] plugins=${shard.pluginIds.join(',')} files=${shard.files.length} weight=${shard.weight}`,
      );
      exitCode = runMode(shard.entries, `ci-shard:plugin:${name}`);
    } else if (kind === 'plugin-full-extra') {
      const shard = findPluginFullExtraShard(name, manifest);
      console.log(
        `[ci-shard:plugin-full-extra:${name}] scope=${shard.scope} files=${shard.files.length}`,
      );
      exitCode = runMode(shard.entries, `ci-shard:plugin-full-extra:${name}`);
    } else {
      console.error(`Unknown ci-shard kind: ${kind}`);
      process.exit(1);
    }
  } catch (error) {
    console.error(error instanceof Error ? error.message : String(error));
    process.exit(1);
  }
} else if (mode === 'module' || mode === 'host-module') {
  const rawModuleArgs = process.argv.slice(3);
  const moduleArgs = normalizePassthroughArgs(rawModuleArgs);
  const scope = moduleArgs[0];
  const passthroughArgs = normalizePassthroughArgs(moduleArgs.slice(1));
  if (!scope) {
    const command = mode === 'host-module' ? 'pnpm test:host:module' : 'pnpm test:module';
    console.error(`Missing module scope. Example: ${command} -- iam:user`);
    process.exit(1);
  }
  const entries = resolveModuleEntries(scope);
  if (!entries) {
    console.error(`Unknown module scope: ${scope}`);
    console.error(
      `Available scopes: ${Object.keys(manifest.moduleScopes).sort().join(', ')}, plugin:<plugin-id>`,
    );
    process.exit(1);
  }
  if (mode === 'host-module' && moduleHasPluginWorkspaceEntries(scope, manifest)) {
    console.error(`Host-only module scope cannot require apps/lina-plugins: ${scope}`);
    process.exit(1);
  }
  const resolvedEntries = resolveEntries(entries);
  const isPluginScope = /^plugin:[^/:]+$/u.test(scope);
  if (moduleRequiresPluginWorkspace(scope, manifest)) {
    try {
      requirePluginWorkspace();
    } catch (error) {
      console.error(error instanceof Error ? error.message : String(error));
      process.exit(1);
    }
  }
  if (mode === 'module' && resolvedEntries.length === 0 && scope !== pluginTestEntry) {
    console.error(`Module scope has no matching test files: ${scope}`);
    if (isPluginScope) {
      console.error('Expected plugin-owned tests under apps/lina-plugins/<plugin-id>/hack/tests/e2e/.');
    }
    process.exit(1);
  }

  extraArgs.splice(0, extraArgs.length, ...passthroughArgs);
  if (mode === 'host-module') {
    const hostModuleStatus = runHostModuleMode(entries, `host-module:${scope}`);
    if (hostModuleStatus === null) {
      console.error(`Host-only module scope has no matching test files after plugin-environment exclusions: ${scope}`);
      process.exit(1);
    }
    exitCode = hostModuleStatus;
  } else {
    exitCode = runMode(entries, `module:${scope}`);
  }
} else {
  console.error(`Unknown run mode: ${mode}`);
  process.exit(1);
}

process.exit(exitCode);
