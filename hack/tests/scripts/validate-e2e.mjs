import { readdirSync, readFileSync, statSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const testsDir = path.resolve(scriptDir, '..');
const e2eDir = path.resolve(testsDir, 'e2e');
const manifest = JSON.parse(
  readFileSync(path.resolve(testsDir, 'config/execution-manifest.json'), 'utf8'),
);

function toPosix(relativePath) {
  return relativePath.split(path.sep).join('/');
}

function walk(directory) {
  const result = [];
  const stack = [directory];
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
    result.push(current);
  }
  return result.sort();
}

function listTcFiles(entry) {
  const absoluteEntry = path.resolve(testsDir, entry);
  if (!exists(absoluteEntry)) {
    return [];
  }
  const stat = statSync(absoluteEntry);
  if (stat.isFile()) {
    return [toPosix(path.relative(testsDir, absoluteEntry))];
  }
  return walk(absoluteEntry)
    .map((item) => toPosix(path.relative(testsDir, item)))
    .filter((item) => /\/TC\d{4}.*\.ts$/.test(item) || /^e2e\/TC\d{4}.*\.ts$/.test(item));
}

function exists(value) {
  try {
    statSync(value);
    return true;
  } catch {
    return false;
  }
}

const errors = [];
const allFiles = walk(e2eDir).map((item) => toPosix(path.relative(testsDir, item)));
const testFiles = [];
const tcRegistry = new Map();
const allowedPrefixes = new Set(
  Object.values(manifest.moduleScopes)
    .flat()
    .map((entry) => entry.replace(/\/$/, '')),
);

for (const file of allFiles) {
  if (!file.endsWith('.ts')) {
    errors.push(`Non-TypeScript file found under e2e: ${file}`);
    continue;
  }

  if (!/\/TC\d{4}[-][^.]+\.ts$/.test(file) && !/^e2e\/TC\d{4}[-][^.]+\.ts$/.test(file)) {
    errors.push(`Non-test file found under e2e: ${file}`);
    continue;
  }

  testFiles.push(file);
  const tcId = file.match(/TC(\d{4})/)?.[1];
  if (!tcId) {
    errors.push(`Unable to parse TC ID from ${file}`);
    continue;
  }
  const items = tcRegistry.get(tcId) ?? [];
  items.push(file);
  tcRegistry.set(tcId, items);

  const fileDir = path.posix.dirname(file);
  const matchesAllowedPrefix = [...allowedPrefixes].some((prefix) => {
    return fileDir === prefix || fileDir.startsWith(`${prefix}/`);
  });
  if (!matchesAllowedPrefix) {
    errors.push(`File is not under an allowed module scope: ${file}`);
  }
}

for (const [tcId, files] of tcRegistry.entries()) {
  if (files.length > 1) {
    errors.push(`Duplicate TC${tcId}: ${files.join(', ')}`);
  }
}

for (const [scope, entries] of Object.entries(manifest.moduleScopes)) {
  const files = entries.flatMap((entry) => listTcFiles(entry));
  if (files.length === 0) {
    errors.push(`Module scope has no matching test files: ${scope}`);
  }
}

for (const entry of manifest.smoke ?? []) {
  if (!exists(path.resolve(testsDir, entry))) {
    errors.push(`Smoke entry does not exist: ${entry}`);
  }
}

for (const entry of manifest.serial ?? []) {
  if (!exists(path.resolve(testsDir, entry))) {
    errors.push(`Serial entry does not exist: ${entry}`);
  }
}

if (errors.length > 0) {
  console.error('E2E suite validation failed:');
  for (const error of errors) {
    console.error(`- ${error}`);
  }
  process.exit(1);
}

console.log(`Validated ${testFiles.length} E2E test files across ${Object.keys(manifest.moduleScopes).length} scopes.`);
