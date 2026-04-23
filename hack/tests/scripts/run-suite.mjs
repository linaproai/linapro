import { readdirSync, readFileSync, statSync } from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const testsDir = path.resolve(scriptDir, '..');
const manifestPath = path.resolve(testsDir, 'config/execution-manifest.json');
const manifest = JSON.parse(readFileSync(manifestPath, 'utf8'));

const mode = process.argv[2] ?? 'full';
const extraArgs = process.argv.slice(3);
const parallelWorkers = Number.parseInt(
  process.env.E2E_PARALLEL_WORKERS ?? `${Math.min(Math.max(os.cpus().length - 1, 2), 4)}`,
  10,
);

function toPosix(relativePath) {
  return relativePath.split(path.sep).join('/');
}

function listTcFiles(entry) {
  const absoluteEntry = path.resolve(testsDir, entry);
  if (!statExists(absoluteEntry)) {
    return [];
  }

  const result = [];
  const stack = [absoluteEntry];
  while (stack.length > 0) {
    const current = stack.pop();
    const stat = statSync(current);
    if (stat.isDirectory()) {
      const children = readdirSync(current)
        .map((child) => path.join(current, child))
        .sort()
        .reverse();
      stack.push(...children);
      continue;
    }

    const relativePath = toPosix(path.relative(testsDir, current));
    if (/^e2e\/.*\/TC\d{4}.*\.ts$/.test(relativePath) || /^e2e\/TC\d{4}.*\.ts$/.test(relativePath)) {
      result.push(relativePath);
    }
  }

  return result.sort();
}

function statExists(value) {
  try {
    statSync(value);
    return true;
  } catch {
    return false;
  }
}

function unique(values) {
  return [...new Set(values)].sort();
}

function resolveEntries(entries) {
  return unique(entries.flatMap((entry) => listTcFiles(entry)));
}

function serialFileSet() {
  return new Set(resolveEntries(manifest.serial ?? []));
}

function splitBySerial(entries) {
  const files = resolveEntries(entries);
  const serial = serialFileSet();
  const serialFiles = files.filter((file) => serial.has(file));
  const parallelFiles = files.filter((file) => !serial.has(file));
  return { parallelFiles, serialFiles };
}

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

function runMode(entries, label) {
  const { parallelFiles, serialFiles } = splitBySerial(entries);
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
