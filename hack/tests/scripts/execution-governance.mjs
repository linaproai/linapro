import { readdirSync, readFileSync, statSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));

export const testsDir = path.resolve(scriptDir, '..');
export const repoRoot = path.resolve(testsDir, '..', '..');
export const e2eDir = path.resolve(testsDir, 'e2e');
export const pluginsDir = path.resolve(repoRoot, 'apps/lina-plugins');
export const manifestPath = path.resolve(testsDir, 'config/execution-manifest.json');
export const pluginTestEntry = 'plugins';
export const pluginWorkspaceInitCommand = 'git submodule update --init --recursive';

export const isolationCategories = [
  {
    category: 'authSession',
    label: 'shared authenticated browser session',
  },
  {
    category: 'dictionaryData',
    label: 'dictionary data mutation',
  },
  {
    category: 'filesystemArtifact',
    label: 'filesystem-backed runtime artifact mutation',
  },
  {
    category: 'permissionMatrix',
    label: 'menu or permission matrix mutation',
  },
  {
    category: 'pluginLifecycle',
    label: 'plugin lifecycle or artifact mutation',
  },
  {
    category: 'runtimeI18nCache',
    label: 'runtime i18n cache or ETag dependency',
  },
  {
    category: 'sharedDatabaseSeed',
    label: 'shared database seed or mock data dependency',
  },
  {
    category: 'systemConfig',
    label: 'system configuration mutation',
  },
];

export const highRiskRules = [
  {
    category: 'pluginLifecycle',
    label: 'plugin lifecycle or artifact mutation',
    patterns: [
      /\b(?:installPlugin|enablePlugin|disablePlugin|uninstallPlugin|syncPlugins|setPluginEnabled|uploadDynamicPlugin|uploadDynamicPluginViaAPI|uninstallPluginWithOptions|ensureSourcePluginDisabled|ensureSourcePluginUninstalled)\b/u,
      /\b(?:ensureSourcePluginInstalled|ensureSourcePluginEnabled|ensureSourcePluginsEnabled|loadSourcePluginMockData)\b/u,
      /plugins\/sync/u,
      /plugins\/[^`'"\s]+\/(?:install|enable|disable)/u,
    ],
  },
  {
    category: 'runtimeI18nCache',
    label: 'runtime i18n cache or ETag dependency',
    patterns: [
      /If-None-Match/u,
      /runtimePersistentCacheKey/u,
      /linapro:i18n:runtime/u,
    ],
  },
  {
    category: 'systemConfig',
    label: 'system configuration mutation',
    patterns: [
      /\bupdateConfigValue\b/u,
      /config\/import/u,
      /sys\.auth\.loginPanelLayout/u,
    ],
  },
  {
    category: 'dictionaryData',
    label: 'dictionary data mutation',
    patterns: [
      /dict\/type\/import/u,
      /dict\/data\/import/u,
      /\bdictPage\.(?:createType|editType|deleteType|clickCurrentTypeDeleteAction|createData|editData|deleteData)\b/u,
      /字典管理导入/u,
      /字典类型级联删除/u,
    ],
  },
  {
    category: 'permissionMatrix',
    label: 'menu or permission matrix mutation',
    patterns: [
      /createRootMenu/u,
      /updateMenuVisibility/u,
      /menuIds:\s*\[/u,
      /api\.(?:post|put|delete)\(["'`]menu/u,
      /sys_role_menu/u,
    ],
  },
  {
    category: 'filesystemArtifact',
    label: 'filesystem-backed runtime artifact mutation',
    patterns: [
      /\b(?:uploadDynamicPlugin|uploadDynamicPluginViaAPI|ensureRuntimePluginArtifact|ensureBundledRuntimePluginArtifact)\b/u,
      /runtimePluginArtifact/u,
      /DynamicPlugin/u,
    ],
  },
];

export function knownIsolationCategorySet() {
  return new Set(isolationCategories.map((item) => item.category));
}

export function loadManifest() {
  return JSON.parse(readFileSync(manifestPath, 'utf8'));
}

export function toPosix(relativePath) {
  return relativePath.split(path.sep).join('/');
}

export function exists(value) {
  try {
    statSync(value);
    return true;
  } catch {
    return false;
  }
}

export function pluginWorkspaceState() {
  if (!exists(pluginsDir)) {
    return { manifestCount: 0, state: 'missing' };
  }
  const manifests = readdirSync(pluginsDir)
    .map((name) => path.join(pluginsDir, name, 'plugin.yaml'))
    .filter(exists);
  if (manifests.length === 0) {
    return { manifestCount: 0, state: 'empty' };
  }
  return { manifestCount: manifests.length, state: 'ready' };
}

export function requirePluginWorkspace() {
  const workspace = pluginWorkspaceState();
  if (workspace.state === 'ready') {
    return;
  }
  throw new Error(
    `Official plugin workspace is ${workspace.state} at apps/lina-plugins. Initialize it with \`${pluginWorkspaceInitCommand}\`.`,
  );
}

export function isPluginWorkspaceReady() {
  return pluginWorkspaceState().state === 'ready';
}

export function walk(directory) {
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
    result.push(current);
  }
  return result.sort();
}

// listPluginE2EDirs returns source-plugin-owned E2E directories following the
// `apps/lina-plugins/<plugin-id>/hack/tests/e2e` convention.
export function listPluginE2EDirs() {
  if (!exists(pluginsDir)) {
    return [];
  }
  return readdirSync(pluginsDir)
    .map((name) => path.join(pluginsDir, name, 'hack', 'tests', 'e2e'))
    .filter(exists)
    .sort();
}

// listLegacyPluginE2EDirs returns pre-governance plugin E2E directories that
// should no longer contain plugin-owned Playwright assets.
export function listLegacyPluginE2EDirs() {
  if (!exists(pluginsDir)) {
    return [];
  }
  return readdirSync(pluginsDir)
    .flatMap((name) => [
      path.join(pluginsDir, name, 'e2e'),
      path.join(pluginsDir, name, 'e2e-pages'),
      path.join(pluginsDir, name, 'e2e-support'),
    ])
    .filter(exists)
    .sort();
}

// pluginTestRelativePath formats a plugin E2E path relative to the repository
// root. The governance manifest uses this canonical form for plugin-owned tests.
export function pluginTestRelativePath(absolutePath) {
  return toPosix(path.relative(repoRoot, absolutePath));
}

// playwrightFileArg keeps the canonical path shape used by the repo-root
// Playwright config while leaving room for future path normalization.
export function playwrightFileArg(relativePath) {
  return relativePath;
}

// listPluginE2EFiles lists every file under source-plugin-owned E2E
// directories so the validator can reject misplaced helpers or bad names.
export function listPluginE2EFiles() {
  return listPluginE2EDirs()
    .flatMap((directory) => walk(directory))
    .map(pluginTestRelativePath)
    .sort();
}

// listPluginTcFiles lists all source-plugin-owned E2E TC files.
export function listPluginTcFiles() {
  return listPluginE2EFiles()
    .filter(isPluginTcFile)
    .sort();
}

export function listTcFiles(entry) {
  if (entry === pluginTestEntry) {
    return listPluginTcFiles();
  }
  if (entry.startsWith('apps/lina-plugins/')) {
    const absoluteEntry = path.resolve(repoRoot, entry);
    if (!exists(absoluteEntry)) {
      return [];
    }
    const stat = statSync(absoluteEntry);
    if (stat.isFile()) {
      const relativePath = pluginTestRelativePath(absoluteEntry);
      return isPluginTcFile(relativePath) ? [relativePath] : [];
    }
    return walk(absoluteEntry)
      .map(pluginTestRelativePath)
      .filter(isPluginTcFile)
      .sort();
  }
  if (entry.startsWith('plugins/')) {
    const parts = entry.split('/').filter(Boolean);
    if (parts.length >= 2) {
      const absoluteEntry = path.resolve(pluginsDir, parts[1], 'hack', 'tests', 'e2e', ...parts.slice(2));
      if (!exists(absoluteEntry)) {
        return [];
      }
      const stat = statSync(absoluteEntry);
      if (stat.isFile()) {
        const relativePath = pluginTestRelativePath(absoluteEntry);
        return isPluginTcFile(relativePath) ? [relativePath] : [];
      }
      return walk(absoluteEntry)
        .map(pluginTestRelativePath)
        .filter(isPluginTcFile)
        .sort();
    }
  }

  const absoluteEntry = path.resolve(testsDir, entry);
  if (!exists(absoluteEntry)) {
    return [];
  }

  const stat = statSync(absoluteEntry);
  if (stat.isFile()) {
    const relativePath = toPosix(path.relative(testsDir, absoluteEntry));
    return isTcFile(relativePath) ? [relativePath] : [];
  }

  return walk(absoluteEntry)
    .map((item) => toPosix(path.relative(testsDir, item)))
    .filter(isTcFile)
    .sort();
}

export function isTcFile(relativePath) {
  return isHostTcFile(relativePath) || isPluginTcFile(relativePath);
}

// isHostTcFile reports whether a path is a host-owned E2E TC file.
export function isHostTcFile(relativePath) {
  return /^e2e\/(?:.*\/)?TC\d{4}-[^/.]+\.ts$/u.test(relativePath);
}

// isPluginTcFile reports whether a path is a source-plugin-owned E2E TC file.
export function isPluginTcFile(relativePath) {
  return /^apps\/lina-plugins\/[^/]+\/hack\/tests\/e2e\/(?:.*\/)?TC\d{4}-[^/.]+\.ts$/u.test(relativePath);
}

export function unique(values) {
  return [...new Set(values)].sort();
}

export function resolveEntries(entries) {
  return unique((entries ?? []).flatMap((entry) => listTcFiles(entry)));
}

export function serialFileSet(manifest = loadManifest()) {
  return new Set(resolveEntries(manifest.serial ?? []));
}

export function serialIsolationEntries(manifest = loadManifest()) {
  return manifest.serialIsolation ?? [];
}

export function isolationAllowlist(manifest = loadManifest()) {
  return manifest.parallelIsolationAllowlist ?? [];
}

export function serialCategoryMap(manifest = loadManifest()) {
  const result = new Map();
  for (const item of serialIsolationEntries(manifest)) {
    for (const file of listTcFiles(item.entry)) {
      const categories = result.get(file) ?? new Set();
      for (const category of item.categories ?? []) {
        categories.add(category);
      }
      result.set(file, categories);
    }
  }
  return result;
}

export function splitBySerial(entries, manifest = loadManifest()) {
  const files = resolveEntries(entries);
  const serial = serialFileSet(manifest);
  const serialFiles = files.filter((file) => serial.has(file));
  const parallelFiles = files.filter((file) => !serial.has(file));
  return { files, parallelFiles, serialFiles };
}

export function summarizeIsolationCategories(files, manifest = loadManifest()) {
  const categoryMap = serialCategoryMap(manifest);
  const counts = new Map();
  for (const file of files) {
    for (const category of categoryMap.get(file) ?? []) {
      counts.set(category, (counts.get(category) ?? 0) + 1);
    }
  }
  return [...counts.entries()].sort(([left], [right]) => left.localeCompare(right));
}

export function detectRiskCategories(sourceText) {
  const result = new Set();
  for (const rule of highRiskRules) {
    if (rule.patterns.some((pattern) => pattern.test(sourceText))) {
      result.add(rule.category);
    }
  }
  return result;
}

export function allowlistCategoriesForFile(file, manifest = loadManifest()) {
  const result = new Set();
  for (const item of isolationAllowlist(manifest)) {
    if (item.file === file) {
      for (const category of item.categories ?? []) {
        result.add(category);
      }
    }
  }
  return result;
}
