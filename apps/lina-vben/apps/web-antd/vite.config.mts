import { existsSync, readdirSync, readFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { join, relative, sep } from 'node:path';

import { defineConfig } from '@vben/vite-config';
import type { ViteDevServer } from 'vite';

// Cache the HTML content to avoid repeated synchronous file reads
let cachedApidocsHtml: string | undefined;

const pluginPageModuleId = 'virtual:lina-plugin-pages';
const pluginSlotModuleId = 'virtual:lina-plugin-slots';
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

export default defineConfig(async () => {
  const vbenRoot = join(import.meta.dirname, '../..');
  const pluginRoot = join(import.meta.dirname, '../../../lina-plugins');

  return {
    application: {
      printInfoMap: {
        'Lina Repository': 'https://github.com/gqcn/lina',
      },
      pwaOptions: {
        manifest: {
          description:
            'Lina is an AI-driven full-stack development framework with core host services, a default management workspace, plugin extensibility, and AI-assisted delivery workflows.',
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
