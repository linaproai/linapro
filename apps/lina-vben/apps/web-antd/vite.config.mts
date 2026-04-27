import { existsSync, readdirSync, readFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { dirname, join, relative, sep } from 'node:path';

import { defineConfig } from '@vben/vite-config';
import type { ViteDevServer } from 'vite';

// Cache the HTML content to avoid repeated synchronous file reads
let cachedApidocsHtml: string | undefined;

const pluginPageModuleId = 'virtual:lina-plugin-pages';
const pluginSlotModuleId = 'virtual:lina-plugin-slots';
const appThirdPartyLocaleModuleId = 'virtual:lina-app-third-party-locales';
const vxeLocaleModuleId = 'virtual:lina-vxe-locales';
const appRequire = createRequire(import.meta.url);

function collectPluginSourceFiles(pluginRoot: string) {
  const pageFiles: string[] = [];
  const slotFiles: string[] = [];

  if (!existsSync(pluginRoot)) {
    return { pageFiles, slotFiles };
  }

  const walk = (currentPath: string) => {
    for (const entry of readdirSync(currentPath, { withFileTypes: true })) {
      const fullPath = join(currentPath, entry.name);
      if (entry.isDirectory()) {
        walk(fullPath);
        continue;
      }
      if (!entry.isFile() || !entry.name.endsWith('.vue')) {
        continue;
      }
      const normalizedPath = fullPath.split(sep).join('/');
      if (normalizedPath.includes('/frontend/pages/')) {
        pageFiles.push(fullPath);
      }
      if (normalizedPath.includes('/frontend/slots/')) {
        slotFiles.push(fullPath);
      }
    }
  };

  walk(pluginRoot);
  return { pageFiles, slotFiles };
}

function normalizeFsPath(filePath: string) {
  return filePath.split(sep).join('/');
}

function normalizeImporterPath(filePath: string) {
  return normalizeFsPath(filePath.split('?')[0]?.split('#')[0] || filePath);
}

function toViteFsPath(filePath: string) {
  const normalizedPath = normalizeFsPath(filePath);
  return normalizedPath.startsWith('/@fs/')
    ? normalizedPath
    : `/@fs${normalizedPath}`;
}

function isPluginFrontendSourceFile(pluginRoot: string, filePath: string) {
  const normalizedPluginRoot = normalizeFsPath(pluginRoot);
  const normalizedFilePath = normalizeImporterPath(filePath);

  if (!normalizedFilePath.startsWith(normalizedPluginRoot)) {
    return false;
  }

  if (!normalizedFilePath.endsWith('.vue')) {
    return false;
  }

  return (
    normalizedFilePath.includes('/frontend/pages/') ||
    normalizedFilePath.includes('/frontend/slots/')
  );
}

function isBareModuleImport(source: string) {
  return (
    !!source &&
    !source.startsWith('.') &&
    !source.startsWith('/') &&
    !source.startsWith('#') &&
    !source.startsWith('\0') &&
    !source.startsWith('virtual:')
  );
}

function invalidatePluginVirtualModules(server: ViteDevServer) {
  for (const moduleId of [pluginPageModuleId, pluginSlotModuleId]) {
    const module = server.moduleGraph.getModuleById(`\0${moduleId}`);
    if (module) {
      server.moduleGraph.invalidateModule(module);
    }
  }
}

function buildPluginPageModuleCode(pluginRoot: string) {
  const { pageFiles } = collectPluginSourceFiles(pluginRoot);
  const imports: string[] = [];
  const records: string[] = [];

  pageFiles.toSorted().forEach((filePath, index) => {
    const relativePath = normalizeFsPath(relative(pluginRoot, filePath));
    const match = relativePath.match(/^([^/]+)\/frontend\/pages\/(.+)\.vue$/);
    if (!match?.[1] || !match[2]) {
      return;
    }

    imports.push(
      `import * as pluginPageModule${index} from ${JSON.stringify(toViteFsPath(filePath))};`,
    );
    records.push(
      `  { filePath: ${JSON.stringify(normalizeFsPath(filePath))}, module: pluginPageModule${index} },`,
    );
  });

  return [
    ...imports,
    'export const pluginPageModules = [',
    ...records,
    '];',
  ].join('\n');
}

function buildPluginSlotModuleCode(pluginRoot: string) {
  const { slotFiles } = collectPluginSourceFiles(pluginRoot);
  const imports: string[] = [];
  const records: string[] = [];

  slotFiles.toSorted().forEach((filePath, index) => {
    const relativePath = normalizeFsPath(relative(pluginRoot, filePath));
    const match = relativePath.match(/^([^/]+)\/frontend\/slots\/(.+)\.vue$/);
    if (!match?.[1] || !match[2]) {
      return;
    }

    imports.push(
      `import * as pluginSlotModule${index} from ${JSON.stringify(toViteFsPath(filePath))};`,
    );
    records.push(
      `  { filePath: ${JSON.stringify(normalizeFsPath(filePath))}, module: pluginSlotModule${index} },`,
    );
  });

  return [
    ...imports,
    'export const pluginSlotModules = [',
    ...records,
    '];',
  ].join('\n');
}

function collectLocaleNamesFromDir(localeDir: string, options = {}) {
  const { exclude = [] } = options as { exclude?: string[] };
  const excluded = new Set(exclude);

  return readdirSync(localeDir, { withFileTypes: true })
    .filter((entry) => {
      if (!entry.isFile() || !entry.name.endsWith('.js')) {
        return false;
      }
      if (entry.name.endsWith('.min.js') || entry.name.endsWith('.umd.js')) {
        return false;
      }
      const localeName = entry.name.slice(0, -'.js'.length);
      return !excluded.has(localeName);
    })
    .map((entry) => entry.name.slice(0, -'.js'.length))
    .toSorted();
}

function collectPackageLocaleNames(sampleImport: string, options = {}) {
  return collectLocaleNamesFromDir(
    dirname(appRequire.resolve(sampleImport)),
    options,
  );
}

function collectRuntimeLocaleNames(localeDirs: string[]) {
  const locales = new Set<string>();
  for (const localeDir of localeDirs) {
    if (!existsSync(localeDir)) {
      continue;
    }
    for (const entry of readdirSync(localeDir, { withFileTypes: true })) {
      if (entry.isDirectory()) {
        locales.add(entry.name);
      }
    }
  }
  return [...locales].toSorted();
}

function uniqueItems(items: string[]) {
  return [...new Set(items.map((item) => item.trim()).filter(Boolean))];
}

function splitLocaleCode(locale: string) {
  const segments = locale.trim().split('-').filter(Boolean);
  const language = String(segments[0] || '').toLowerCase();
  const region = String(segments[segments.length - 1] || '').toUpperCase();
  return { language, region };
}

function buildDayjsLocaleCandidates(locale: string) {
  const { language } = splitLocaleCode(locale);
  return uniqueItems([locale.trim().toLowerCase(), language, 'en']);
}

function buildUnderscoreLocaleCandidates(
  locale: string,
  availableLocaleNames: string[],
) {
  const { language, region } = splitLocaleCode(locale);
  return uniqueItems([
    language && region ? `${language}_${region}` : '',
    findLanguageLocaleName(availableLocaleNames, language, '_'),
    'en_US',
  ]);
}

function buildHyphenLocaleCandidates(
  locale: string,
  availableLocaleNames: string[],
) {
  const { language, region } = splitLocaleCode(locale);
  return uniqueItems([
    language && region ? `${language}-${region}` : '',
    findLanguageLocaleName(availableLocaleNames, language, '-'),
    'en-US',
  ]);
}

function findLanguageLocaleName(
  availableLocaleNames: string[],
  language: string,
  separator: string,
) {
  if (!language) {
    return '';
  }
  const languagePrefix = `${language}${separator}`;
  return (
    availableLocaleNames.find((candidate) => {
      const normalizedCandidate = candidate.toLowerCase();
      return (
        normalizedCandidate === language ||
        normalizedCandidate.startsWith(languagePrefix)
      );
    }) || ''
  );
}

function selectAvailableLocaleNames(
  runtimeLocales: string[],
  availableLocaleNames: string[],
  buildCandidates: (locale: string, availableLocaleNames: string[]) => string[],
) {
  const available = new Set(availableLocaleNames);
  const selected = new Set<string>();
  for (const locale of runtimeLocales) {
    for (const candidate of buildCandidates(locale, availableLocaleNames)) {
      if (available.has(candidate)) {
        selected.add(candidate);
      }
    }
  }
  return [...selected].toSorted();
}

function buildLocaleLoaderMapCode(
  exportName: string,
  packagePathPrefix: string,
  localeNames: string[],
) {
  const entries = localeNames.map(
    (localeName) =>
      `  ${JSON.stringify(localeName)}: () => import(${JSON.stringify(`${packagePathPrefix}/${localeName}`)}),`,
  );

  return [`export const ${exportName} = {`, ...entries, '};'].join('\n');
}

function buildAppThirdPartyLocaleModuleCode(runtimeLocales: string[]) {
  const dayjsLocaleNames = selectAvailableLocaleNames(
    runtimeLocales,
    collectPackageLocaleNames('dayjs/locale/en'),
    buildDayjsLocaleCandidates,
  );
  const antdLocaleNames = selectAvailableLocaleNames(
    runtimeLocales,
    collectPackageLocaleNames('ant-design-vue/es/locale/en_US', {
      exclude: ['index', 'LocaleReceiver'],
    }),
    buildUnderscoreLocaleCandidates,
  );

  return [
    buildLocaleLoaderMapCode(
      'dayjsLocaleLoaders',
      'dayjs/locale',
      dayjsLocaleNames,
    ),
    buildLocaleLoaderMapCode(
      'antdLocaleLoaders',
      'ant-design-vue/es/locale',
      antdLocaleNames,
    ),
  ].join('\n\n');
}

function buildVxeLocaleModuleCode(
  vxeLocaleDir: string,
  runtimeLocales: string[],
) {
  const vxeLocaleNames = selectAvailableLocaleNames(
    runtimeLocales,
    collectLocaleNamesFromDir(vxeLocaleDir),
    buildHyphenLocaleCandidates,
  );

  return buildLocaleLoaderMapCode(
    'vxeLocaleLoaders',
    'vxe-pc-ui/lib/language',
    vxeLocaleNames,
  );
}

export default defineConfig(async () => {
  const vbenRoot = join(import.meta.dirname, '../..');
  const pluginRoot = join(import.meta.dirname, '../../../lina-plugins');
  const runtimeLocales = collectRuntimeLocaleNames([
    join(vbenRoot, 'packages/locales/src/langs'),
    join(import.meta.dirname, 'src/locales/langs'),
  ]);
  const vxeLocaleDir = join(
    vbenRoot,
    'packages/effects/plugins/node_modules/vxe-pc-ui/lib/language',
  );

  return {
    application: {
      printInfoMap: {
        'LinaPro Repository': 'https://github.com/gqcn/lina',
      },
      pwaOptions: {
        manifest: {
          description:
            'LinaPro is an AI-driven full-stack development framework with core host services, a default management workspace, plugin extensibility, and AI-assisted delivery workflows.',
        },
      },
    },
    vite: {
      plugins: [
        {
          name: 'lina-plugin-source-deps',
          resolveId(source, importer) {
            if (
              !importer ||
              !isPluginFrontendSourceFile(pluginRoot, importer) ||
              !isBareModuleImport(source)
            ) {
              return null;
            }

            try {
              return appRequire.resolve(source);
            } catch {
              return null;
            }
          },
        },
        {
          name: 'lina-plugin-registry',
          configureServer(server) {
            server.watcher.add(pluginRoot);

            const handlePluginSourceChange = (filePath: string) => {
              if (!isPluginFrontendSourceFile(pluginRoot, filePath)) {
                return;
              }
              invalidatePluginVirtualModules(server);
              server.ws.send({ type: 'full-reload' });
            };

            server.watcher.on('add', handlePluginSourceChange);
            server.watcher.on('change', handlePluginSourceChange);
            server.watcher.on('unlink', handlePluginSourceChange);
          },
          resolveId(source) {
            if (
              source === pluginPageModuleId ||
              source === pluginSlotModuleId
            ) {
              return `\0${source}`;
            }
            return null;
          },
          load(id) {
            if (id === `\0${pluginPageModuleId}`) {
              return buildPluginPageModuleCode(pluginRoot);
            }
            if (id === `\0${pluginSlotModuleId}`) {
              return buildPluginSlotModuleCode(pluginRoot);
            }
            return null;
          },
        },
        {
          name: 'lina-third-party-locales',
          resolveId(source) {
            if (source === appThirdPartyLocaleModuleId) {
              return `\0${source}`;
            }
            if (source === vxeLocaleModuleId) {
              return `\0${source}`;
            }
            return null;
          },
          load(id) {
            if (id === `\0${appThirdPartyLocaleModuleId}`) {
              return buildAppThirdPartyLocaleModuleCode(runtimeLocales);
            }
            if (id === `\0${vxeLocaleModuleId}`) {
              return buildVxeLocaleModuleCode(vxeLocaleDir, runtimeLocales);
            }
            return null;
          },
        },
      ],
      resolve: {
        alias: [
          {
            find: /^#\//,
            replacement: `${join(import.meta.dirname, 'src')}/`,
          },
        ],
      },
      server: {
        fs: {
          allow: [vbenRoot, pluginRoot],
        },
        proxy: {
          '/api': {
            changeOrigin: true,
            // Forward /api/* to backend at localhost:8080/api/*
            target: 'http://localhost:8080',
            ws: true,
          },
          '/plugin-assets': {
            changeOrigin: true,
            // Runtime plugin static assets are hosted by the backend even in
            // dev mode, so the frontend dev server must proxy these requests.
            target: 'http://localhost:8080',
          },
          '/stoplight/apidocs.html': {
            target: 'http://localhost:8080',
            bypass(_req, res) {
              // Serve the static HTML file directly, bypassing Vite's SPA fallback
              if (!cachedApidocsHtml) {
                const filePath = join(
                  import.meta.dirname,
                  'public/stoplight/apidocs.html',
                );
                cachedApidocsHtml = readFileSync(filePath, 'utf8');
              }
              res.setHeader('Content-Type', 'text/html; charset=utf-8');
              res.end(cachedApidocsHtml);
              // Return false to prevent proxy from connecting to the target
              return false;
            },
          },
        },
        watch: {
          // Exclude directories that don't need HMR watching
          ignored: [
            '**/public/stoplight/**',
            '**/node_modules/**',
            '**/dist/**',
            '**/.vite/**',
          ],
        },
      },
    },
  };
});
