import os from 'node:os';
import { spawnSync } from 'node:child_process';

import {
  listTcFiles,
  loadManifest,
  playwrightFileArg,
  pluginTestEntry,
  resolveEntries,
  splitBySerial,
  summarizeIsolationCategories,
  testsDir,
} from './execution-governance.mjs';

const manifest = loadManifest();

const mode = process.argv[2] ?? 'full';
const extraArgs = normalizePassthroughArgs(process.argv.slice(3));
const parallelWorkers = Number.parseInt(
  process.env.E2E_PARALLEL_WORKERS ?? `${Math.min(Math.max(os.cpus().length - 1, 2), 4)}`,
  10,
);

function runPlaywright(files, workers, label) {
  if (files.length === 0) {
    return 0;
  }

  const runnableFiles = files.map(playwrightFileArg);
  const args = ['exec', 'playwright', 'test', ...runnableFiles, `--workers=${workers}`, ...extraArgs];
  console.log(`\n[${label}] playwright test ${files.length} file(s), workers=${workers}`);
  const result = spawnSync('pnpm', args, {
    cwd: testsDir,
    stdio: 'inherit',
    env: process.env,
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
  exitCode = runMode(['e2e', 'plugins'], 'full');
} else if (mode === 'smoke') {
  exitCode = runMode(manifest.smoke ?? [], 'smoke');
} else if (mode === 'module') {
  const rawModuleArgs = process.argv.slice(3);
  const moduleArgs = normalizePassthroughArgs(rawModuleArgs);
  const scope = moduleArgs[0];
  const passthroughArgs = normalizePassthroughArgs(moduleArgs.slice(1));
  if (!scope) {
    console.error('Missing module scope. Example: pnpm test:module -- iam:user');
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
  if (resolveEntries(entries).length === 0 && scope !== pluginTestEntry) {
    const pluginHint = listTcFiles(`plugins/${scope.replace(/^plugin:/u, '')}`).length === 0;
    console.error(`Module scope has no matching test files: ${scope}`);
    if (/^plugin:/u.test(scope) && pluginHint) {
      console.error('Expected plugin-owned tests under apps/lina-plugins/<plugin-id>/e2e/.');
    }
    process.exit(1);
  }

  extraArgs.splice(0, extraArgs.length, ...passthroughArgs);
  exitCode = runMode(entries, `module:${scope}`);
} else {
  console.error(`Unknown run mode: ${mode}`);
  process.exit(1);
}

process.exit(exitCode);
