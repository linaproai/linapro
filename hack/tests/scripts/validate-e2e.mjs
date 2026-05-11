import { readFileSync } from 'node:fs';
import path from 'node:path';

import {
  allowlistCategoriesForFile,
  detectRiskCategories,
  e2eDir,
  exists,
  highRiskRules,
  isolationAllowlist,
  isHostTcFile,
  isPluginTcFile,
  knownIsolationCategorySet,
  listLegacyPluginE2EDirs,
  listPluginE2EFiles,
  listTcFiles,
  loadManifest,
  pluginTestEntry,
  pluginTestRelativePath,
  repoRoot,
  resolveEntries,
  serialCategoryMap,
  serialFileSet,
  serialIsolationEntries,
  testsDir,
  toPosix,
  walk,
} from './execution-governance.mjs';

const manifest = loadManifest();
const errors = [];
const highRiskRuleByCategory = new Map(
  highRiskRules.map((rule) => [rule.category, rule]),
);
const knownCategories = knownIsolationCategorySet();

function addError(message) {
  errors.push(message);
}

function readTestFile(relativePath) {
  if (relativePath.startsWith('apps/lina-plugins/')) {
    return readFileSync(path.resolve(repoRoot, relativePath), 'utf8');
  }
  return readFileSync(path.resolve(testsDir, relativePath), 'utf8');
}

function requireArray(value, label) {
  if (!Array.isArray(value)) {
    addError(`${label} must be an array.`);
    return [];
  }
  return value;
}

function validateCategories(categories, ownerLabel) {
  const values = requireArray(categories, `${ownerLabel}.categories`);
  if (values.length === 0) {
    addError(`${ownerLabel}.categories must contain at least one category.`);
    return [];
  }

  const seen = new Set();
  for (const category of values) {
    if (typeof category !== 'string' || category.trim() === '') {
      addError(`${ownerLabel}.categories contains a non-string or empty category.`);
      continue;
    }
    if (!knownCategories.has(category)) {
      addError(
        `${ownerLabel}.categories contains unknown category "${category}". Known categories: ${[...knownCategories].sort().join(', ')}`,
      );
    }
    if (seen.has(category)) {
      addError(`${ownerLabel}.categories contains duplicate category "${category}".`);
    }
    seen.add(category);
  }
  return values;
}

function validateReason(reason, ownerLabel) {
  if (typeof reason !== 'string' || reason.trim() === '') {
    addError(`${ownerLabel}.reason must explain why the isolation decision is safe.`);
  }
}

const allFiles = [
  ...walk(e2eDir).map((item) => toPosix(path.relative(testsDir, item))),
  ...listPluginE2EFiles(),
];
const legacyPluginE2EDirs = listLegacyPluginE2EDirs();
const testFiles = [];
const tcRegistry = new Map();
const allowedFiles = new Set(
  Object.values(manifest.moduleScopes)
    .flat()
    .flatMap((entry) => listTcFiles(entry)),
);

function entryExistsOrResolves(entry) {
  if (listTcFiles(entry).length > 0) {
    return true;
  }
  if (entry === pluginTestEntry) {
    return true;
  }
  if (
    entry.startsWith('plugins/') ||
    entry.startsWith('apps/lina-plugins/')
  ) {
    return false;
  }
  return exists(path.resolve(testsDir, entry));
}

for (const directory of legacyPluginE2EDirs) {
  const relativePath = pluginTestRelativePath(directory);
  addError(
    `Legacy plugin E2E directory found: ${relativePath}. Use apps/lina-plugins/<plugin-id>/hack/tests/{e2e,pages,support}/ instead.`,
  );
}

for (const file of allFiles) {
  if (!file.endsWith('.ts')) {
    addError(`Non-TypeScript file found under e2e: ${file}`);
    continue;
  }

  if (!isHostTcFile(file) && !isPluginTcFile(file)) {
    addError(`Non-test file found under e2e: ${file}`);
    continue;
  }

  testFiles.push(file);
  const tcId = file.match(/TC(\d{4})/u)?.[1];
  if (!tcId) {
    addError(`Unable to parse TC ID from ${file}`);
    continue;
  }
  const items = tcRegistry.get(tcId) ?? [];
  items.push(file);
  tcRegistry.set(tcId, items);

  if (!allowedFiles.has(file)) {
    addError(`File is not under an allowed module scope: ${file}`);
  }
}

for (const [tcId, files] of tcRegistry.entries()) {
  if (files.length > 1) {
    addError(`Duplicate TC${tcId}: ${files.join(', ')}`);
  }
}

for (const [scope, entries] of Object.entries(manifest.moduleScopes)) {
  const files = entries.flatMap((entry) => listTcFiles(entry));
  if (files.length === 0 && scope !== pluginTestEntry) {
    addError(`Module scope has no matching test files: ${scope}`);
  }
}

for (const entry of manifest.smoke ?? []) {
  if (!entryExistsOrResolves(entry)) {
    addError(`Smoke entry does not exist: ${entry}`);
  }
}

const serialEntries = requireArray(manifest.serial ?? [], 'serial');
for (const entry of serialEntries) {
  if (!entryExistsOrResolves(entry)) {
    addError(`Serial entry does not exist: ${entry}`);
  }
}

const isolationEntries = requireArray(serialIsolationEntries(manifest), 'serialIsolation');
const serialEntrySet = new Set(serialEntries);
const isolationEntrySet = new Set();

for (const [index, item] of isolationEntries.entries()) {
  const owner = `serialIsolation[${index}]`;
  if (!item || typeof item !== 'object') {
    addError(`${owner} must be an object.`);
    continue;
  }
  if (typeof item.entry !== 'string' || item.entry.trim() === '') {
    addError(`${owner}.entry must be a non-empty string.`);
    continue;
  }
  if (!entryExistsOrResolves(item.entry)) {
    addError(`${owner}.entry does not exist: ${item.entry}`);
  }
  if (!serialEntrySet.has(item.entry)) {
    addError(`${owner}.entry is not listed in serial: ${item.entry}`);
  }
  if (isolationEntrySet.has(item.entry)) {
    addError(`${owner}.entry is duplicated: ${item.entry}`);
  }
  isolationEntrySet.add(item.entry);
  validateCategories(item.categories, owner);
  validateReason(item.reason, owner);

  const resolvedFiles = listTcFiles(item.entry);
  if (resolvedFiles.length === 0 && item.entry !== pluginTestEntry) {
    addError(`${owner}.entry does not resolve to any TC file: ${item.entry}`);
  }
}

for (const entry of serialEntries) {
  if (!isolationEntrySet.has(entry)) {
    addError(`Serial entry is missing serialIsolation metadata: ${entry}`);
  }
}

const serialFiles = serialFileSet(manifest);
const categoryMap = serialCategoryMap(manifest);
for (const file of serialFiles) {
  if (!categoryMap.has(file) || categoryMap.get(file).size === 0) {
    addError(`Serial file has no resolved isolation category: ${file}`);
  }
}

const allowlistEntries = requireArray(isolationAllowlist(manifest), 'parallelIsolationAllowlist');
for (const [index, item] of allowlistEntries.entries()) {
  const owner = `parallelIsolationAllowlist[${index}]`;
  if (!item || typeof item !== 'object') {
    addError(`${owner} must be an object.`);
    continue;
  }
  if (typeof item.file !== 'string' || item.file.trim() === '') {
    addError(`${owner}.file must be a non-empty string.`);
    continue;
  }
  if (listTcFiles(item.file).length !== 1) {
    addError(`${owner}.file must reference one existing TC file: ${item.file}`);
  }
  if (serialFiles.has(item.file)) {
    addError(`${owner}.file is already serial and does not need a parallel allowlist: ${item.file}`);
  }
  validateCategories(item.categories, owner);
  validateReason(item.reason, owner);
}

for (const file of testFiles) {
  const detectedCategories = detectRiskCategories(readTestFile(file));
  if (detectedCategories.size === 0) {
    continue;
  }

  const declaredSerialCategories = categoryMap.get(file) ?? new Set();
  const allowedParallelCategories = allowlistCategoriesForFile(file, manifest);
  for (const category of detectedCategories) {
    const rule = highRiskRuleByCategory.get(category);
    const label = rule?.label ?? category;
    if (serialFiles.has(file)) {
      if (!declaredSerialCategories.has(category)) {
        addError(
          `High-risk ${label} detected in serial file ${file}, but serialIsolation does not declare "${category}".`,
        );
      }
      continue;
    }

    if (!allowedParallelCategories.has(category)) {
      addError(
        `High-risk ${label} detected in parallel file ${file}. Add the file to serial with "${category}" isolation or add a documented parallelIsolationAllowlist entry.`,
      );
    }
  }
}

const resolvedSmoke = resolveEntries(manifest.smoke ?? []);
const unresolvedSmokeEntries = (manifest.smoke ?? []).filter((entry) => listTcFiles(entry).length === 0);
for (const entry of unresolvedSmokeEntries) {
  addError(`Smoke entry does not resolve to any TC file: ${entry}`);
}

if (errors.length > 0) {
  console.error('E2E suite validation failed:');
  for (const error of errors) {
    console.error(`- ${error}`);
  }
  process.exit(1);
}

console.log(
  [
    `Validated ${testFiles.length} E2E test files`,
    `across ${Object.keys(manifest.moduleScopes).length} scopes.`,
    `Smoke files: ${resolvedSmoke.length}.`,
    `Serial files: ${serialFiles.size}.`,
  ].join(' '),
);
