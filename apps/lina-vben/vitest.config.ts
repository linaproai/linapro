import Vue from '@vitejs/plugin-vue';
import VueJsx from '@vitejs/plugin-vue-jsx';
import { configDefaults, defineConfig } from 'vitest/config';

const virtualLocaleModules: Record<string, string> = {
  'virtual:lina-app-third-party-locales': [
    'export const antdLocaleLoaders = {};',
    'export const dayjsLocaleLoaders = {};',
  ].join('\n'),
  'virtual:lina-plugin-pages': 'export const pluginPageModules = [];',
  'virtual:lina-plugin-slots': 'export const pluginSlotModules = [];',
  'virtual:lina-vxe-locales': 'export const vxeLocaleLoaders = {};',
};

export default defineConfig({
  plugins: [
    {
      name: 'lina-test-third-party-locales',
      resolveId(source) {
        if (source in virtualLocaleModules) {
          return `\0${source}`;
        }
        return null;
      },
      load(id) {
        const source = id.startsWith('\0') ? id.slice(1) : id;
        return virtualLocaleModules[source] ?? null;
      },
    },
    Vue(),
    VueJsx(),
  ],
  test: {
    environment: 'happy-dom',
    environmentOptions: {
      happyDOM: {
        settings: {
          // happy-dom v20+ disables JS evaluation by default (security fix).
          // Treat disabled script loading as success to preserve test behavior.
          handleDisabledFileLoadingAsSuccess: true,
        },
      },
    },
    exclude: [
      ...configDefaults.exclude,
      '**/e2e/**',
      '**/dist/**',
      '**/.{idea,git,cache,output,temp}/**',
      '**/node_modules/**',
      '**/{stylelint,eslint}.config.*',
      '.prettierrc.mjs',
    ],
  },
});
