import type { IconifyIconStructure } from '@vben/icons';

import { addIcon } from '@vben/icons';

import { resolvePluginSvgIconId } from './icon-registry-path';

export { resolvePluginSvgIconId } from './icon-registry-path';

function parseSvg(svgData: string): IconifyIconStructure {
  const parser = new DOMParser();
  const xmlDoc = parser.parseFromString(svgData, 'image/svg+xml');
  const svgElement = xmlDoc.documentElement;

  const getAttrs = (el: Element, attrs: string[]) =>
    attrs
      .map((attr) =>
        el.hasAttribute(attr) ? `${attr}="${el.getAttribute(attr)}"` : '',
      )
      .filter(Boolean)
      .join(' ');

  const rootAttrs = getAttrs(svgElement, [
    'fill',
    'stroke',
    'fill-rule',
    'stroke-width',
  ]);

  const svgContent = [...svgElement.childNodes]
    .filter((node) => node.nodeType === Node.ELEMENT_NODE)
    .map((node) => new XMLSerializer().serializeToString(node))
    .join('');
  const body = rootAttrs ? `<g ${rootAttrs}>${svgContent}</g>` : svgContent;

  const viewBoxValue = svgElement.getAttribute('viewBox') || '';
  const [left, top, width, height] = viewBoxValue.split(' ').map((val) => {
    const num = Number(val);
    return Number.isNaN(num) ? undefined : num;
  });

  return {
    body,
    height,
    left,
    top,
    width,
  };
}

function readSvgBody(body: string | { default: string }): string {
  return typeof body === 'object' ? body.default : body;
}

/**
 * Eagerly registers every source-plugin SVG under frontend/icons/.
 * Import from bootstrap so sidebar menus can resolve icons before any
 * plugin page is opened.
 */
export function loadPluginSvgIcons(): void {
  // From apps/lina-vben/apps/web-antd/src/plugins → apps/lina-plugins.
  const svgEagers = import.meta.glob(
    '../../../../../lina-plugins/*/frontend/icons/*.svg',
    {
      eager: true,
      query: '?raw',
    },
  );

  for (const [key, body] of Object.entries(svgEagers)) {
    const iconId = resolvePluginSvgIconId(key);
    if (!iconId) {
      console.warn(`[plugin-icon] skip unrecognized icon path: ${key}`);
      continue;
    }
    addIcon(iconId, {
      ...parseSvg(readSvgBody(body as string | { default: string })),
    });
  }
}
