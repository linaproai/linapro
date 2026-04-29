import os from 'node:os';
import { spawnSync } from 'node:child_process';

import {
  loadManifest,
  splitBySerial,
  summarizeIsolationCategories,
  testsDir,
} from './execution-governance.mjs';

const manifest = loadManifest();

const mode = process.argv[2] ?? 'full';
const extraArgs = process.argv.slice(3);
const parallelWorkers = Number.parseInt(
  process.env.E2E_PARALLEL_WORKERS ?? `${Math.min(Math.max(os.cpus().length - 1, 2), 4)}`,
  10,
);

function runPlaywright(files, workers, label) {
  if (files.length === 0) {
    return 0;
  }

  const args = ['exec', 'playwright', 'test', ...files, `--workers=${workers}`, ...extraArgs];
  console.log(`\n[${label}] playwright test ${files.length} file(s), workers=${workers}`);
  const result = spawnSync('pnpm', args, {
    cwd: testsDir,
    stdio: 'inherit',
    env: process.env,
  });
  return result.status ?? 1;
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

let exitCode = 0;
if (mode === 'full') {
  exitCode = runMode(['e2e'], 'full');
} else if (mode === 'smoke') {
  exitCode = runMode(manifest.smoke ?? [], 'smoke');
} else if (mode === 'module') {
  const rawModuleArgs = process.argv.slice(3);
  const moduleArgs = rawModuleArgs[0] === '--' ? rawModuleArgs.slice(1) : rawModuleArgs;
  const scope = moduleArgs[0];
  const passthroughArgs = moduleArgs.slice(1);
  if (!scope) {
    console.error('Missing module scope. Example: pnpm test:module -- iam:user');
    process.exit(1);
  }
  if (!manifest.moduleScopes[scope]) {
    console.error(`Unknown module scope: ${scope}`);
    console.error(`Available scopes: ${Object.keys(manifest.moduleScopes).sort().join(', ')}`);
    process.exit(1);
  }

  extraArgs.splice(0, extraArgs.length, ...passthroughArgs);
  exitCode = runMode(manifest.moduleScopes[scope], `module:${scope}`);
} else {
  console.error(`Unknown run mode: ${mode}`);
  process.exit(1);
}

process.exit(exitCode);
