import { readdirSync, readFileSync, statSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const testsDir = path.resolve(scriptDir, '..');
const repoRoot = path.resolve(testsDir, '..', '..');
const baselinePath = path.resolve(
  testsDir,
  'config/service-dependency-baseline.json',
);

const configuredScanRoots = process.env.LINAPRO_SERVICE_DEP_SCAN_ROOTS
  ? process.env.LINAPRO_SERVICE_DEP_SCAN_ROOTS.split(path.delimiter).filter(Boolean)
  : [];
const baselineOverridePath = process.env.LINAPRO_SERVICE_DEP_BASELINE;

const scanRoots = configuredScanRoots.length > 0 ? configuredScanRoots : [
  'apps/lina-core/internal',
  'apps/lina-core/pkg/pluginservice',
  'apps/lina-plugins',
];

const ignoredPathFragments = [
  '/internal/dao/',
  '/backend/internal/dao/',
  '/internal/model/',
  '/backend/internal/model/',
];

function toPosix(value) {
  return value.split(path.sep).join('/');
}

function exists(value) {
  try {
    statSync(value);
    return true;
  } catch {
    return false;
  }
}

function walk(directory) {
  if (!exists(directory)) {
    return [];
  }
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
    if (current.endsWith('.go') && !current.endsWith('_test.go')) {
      result.push(current);
    }
  }
  return result.sort();
}

function isProductionPath(relativePath) {
  return !ignoredPathFragments.some((fragment) =>
    relativePath.includes(fragment),
  );
}

function importIsCritical(importPath) {
  return (
    importPath.startsWith('lina-core/internal/controller/') ||
    importPath.startsWith('lina-core/internal/service/') ||
    importPath.startsWith('lina-core/pkg/pluginservice/') ||
    importPath.startsWith('lina-core/pkg/pluginhost') ||
    /^lina-plugin-[^/]+\/backend\/internal\/controller\//u.test(importPath) ||
    /^lina-plugin-[^/]+\/backend\/internal\/service\//u.test(importPath)
  );
}

function sourcePluginRegistrationCall(line) {
  return /\bpluginhost\.NewSourcePlugin\s*\(/u.test(line);
}

function defaultImportName(importPath) {
  return importPath.split('/').at(-1);
}

function parseCriticalImports(source) {
  const imports = new Map();
  const importSpecPattern =
    /(?:^|\n)\s*(?:import\s+)?(?:(?<alias>[A-Za-z_][A-Za-z0-9_]*)\s+)?["'](?<path>[^"']+)["']/gu;

  for (const match of source.matchAll(importSpecPattern)) {
    const importPath = match.groups?.path;
    if (!importPath || !importIsCritical(importPath)) {
      continue;
    }
    const alias = match.groups?.alias;
    if (alias === '_' || alias === '.') {
      continue;
    }
    imports.set(alias || defaultImportName(importPath), importPath);
  }
  return [...imports.entries()]
    .map(([name, importPath]) => ({ name, importPath }))
    .sort((a, b) => a.name.localeCompare(b.name));
}

function explicitConstructorCall(line, name, method) {
  if (method.includes('WithDependencies') || method.startsWith('NewWith')) {
    return true;
  }
  const allowedExactConstructors = new Set([
    'NewCoordinationKVProvider',
    'NewCronRegistrar',
    'NewDBStore',
    'NewGlobalMiddlewareRegistrar',
    'NewHTTPRegistrar',
    'NewLocalStorage',
    'NewRedis',
    'NewRouteMiddlewares',
    'NewRouteRegistrar',
    'NewScheduler',
    'NewSQLTableProvider',
    'NewSourcePlugin',
  ]);
  if (allowedExactConstructors.has(method)) {
    return true;
  }
  return method === 'New' && line.includes(`${name}.Dependencies{`);
}

function wasmDefaultHostServiceFallback(relativePath, line) {
  if (!relativePath.includes('/internal/service/plugin/internal/wasm/hostfn_service_')) {
    return false;
  }
  return /^\s*var\s+/u.test(line) || /^\s*[A-Za-z0-9_]+\s*=\s*/u.test(line);
}

function countCriticalConstructors(filePath, relativePath) {
  const source = readFileSync(filePath, 'utf8');
  const imports = parseCriticalImports(source);
  if (imports.length === 0) {
    return { count: 0, calls: [] };
  }

  const calls = [];
  const lines = source.split(/\r?\n/u);
  for (const [index, line] of lines.entries()) {
    for (const { name, importPath } of imports) {
      const pattern = new RegExp(
        String.raw`\b${name.replace(/[.*+?^${}()|[\]\\]/gu, '\\$&')}\.(?<method>New[A-Za-z0-9_]*)\s*\(`,
        'u',
      );
      const match = pattern.exec(line);
      if (!match || sourcePluginRegistrationCall(line)) {
        continue;
      }
      const method = match.groups?.method || '';
      if (explicitConstructorCall(line, name, method)) {
        continue;
      }
      if (wasmDefaultHostServiceFallback(relativePath, line)) {
        continue;
      }
      calls.push({
        line: index + 1,
        importPath,
        constructor: `${name}.${method}`,
        source: line.trim(),
      });
    }
  }
  return { count: calls.length, calls };
}

function loadBaseline() {
  const parsed = JSON.parse(
    readFileSync(baselineOverridePath || baselinePath, 'utf8'),
  );
  const entries = Array.isArray(parsed.entries) ? parsed.entries : [];
  return new Map(entries.map((entry) => [entry.path, entry]));
}

function scan() {
  const results = [];
  for (const root of scanRoots) {
    const absoluteRoot = path.resolve(repoRoot, root);
    for (const file of walk(absoluteRoot)) {
      const relativePath = toPosix(path.relative(repoRoot, file));
      if (!isProductionPath(relativePath)) {
        continue;
      }
      const result = countCriticalConstructors(file, relativePath);
      if (result.count > 0) {
        results.push({ path: relativePath, ...result });
      }
    }
  }
  return results.sort((a, b) => a.path.localeCompare(b.path));
}

const baseline = loadBaseline();
const results = scan();
const errors = [];

for (const result of results) {
  const allowed = baseline.get(result.path);
  if (!allowed) {
    errors.push(
      `New critical service construction found in non-baseline file ${result.path}: ${result.count}`,
    );
    continue;
  }
  if (typeof allowed.reason !== 'string' || allowed.reason.trim() === '') {
    errors.push(`Baseline entry for ${result.path} must include a reason.`);
  }
  if (result.count > allowed.maxCalls) {
    errors.push(
      [
        `Critical service construction count increased in ${result.path}: ${result.count} > ${allowed.maxCalls}.`,
        ...result.calls.map((call) => `  line ${call.line} ${call.constructor}: ${call.source}`),
      ].join('\n'),
    );
  }
}

for (const [relativePath, entry] of baseline.entries()) {
  if (typeof entry.maxCalls !== 'number' || entry.maxCalls < 0) {
    errors.push(`Baseline entry for ${relativePath} has invalid maxCalls.`);
  }
}

if (errors.length > 0) {
  console.error(errors.join('\n\n'));
  process.exit(1);
}

const totalCalls = results.reduce((sum, result) => sum + result.count, 0);
console.log(
  `Service dependency governance passed: ${results.length} files, ${totalCalls} baseline critical constructor calls.`,
);
