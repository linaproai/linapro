import { describe, expect, it } from 'vitest';

import { resolvePluginSvgIconId } from './icon-registry-path';

describe('resolvePluginSvgIconId', () => {
  it('maps a flat plugin icon path to a namespaced svg icon id', () => {
    expect(
      resolvePluginSvgIconId(
        '/repo/apps/lina-plugins/linapro-storage-qiniu/frontend/icons/mark.svg',
      ),
    ).toBe('svg:linapro-storage-qiniu-mark');
  });

  it('accepts windows-style separators', () => {
    expect(
      resolvePluginSvgIconId(
        String.raw`C:\repo\apps\lina-plugins\linapro-storage-cos\frontend\icons\mark.svg`,
      ),
    ).toBe('svg:linapro-storage-cos-mark');
  });

  it('rejects nested paths under frontend/icons', () => {
    expect(
      resolvePluginSvgIconId(
        '/apps/lina-plugins/linapro-storage-cos/frontend/icons/nested/mark.svg',
      ),
    ).toBeNull();
  });

  it('rejects non-plugin icon paths', () => {
    expect(
      resolvePluginSvgIconId(
        '/apps/lina-vben/packages/icons/src/svg/icons/qiniu.svg',
      ),
    ).toBeNull();
  });
});
