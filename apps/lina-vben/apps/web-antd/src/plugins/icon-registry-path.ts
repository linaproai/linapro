/**
 * Source-plugin custom menu icons live under
 * `apps/lina-plugins/<plugin-id>/frontend/icons/*.svg`.
 * The workbench registers them at build time as
 * `svg:<plugin-id>-<file-stem>` so brand marks stay closed-loop in the plugin.
 *
 * Prefer Iconify collection icons in plugin.yaml when a suitable glyph exists.
 * Do not add business brand SVGs to `packages/icons/src/svg/icons/`.
 */

const pluginSvgIconPathPattern =
  /\/lina-plugins\/([^/]+)\/frontend\/icons\/([^/]+)\.svg$/;

/**
 * Derives the Iconify id for one plugin-owned SVG path.
 * Returns null when the path is not a flat file under frontend/icons/.
 */
export function resolvePluginSvgIconId(filePath: string): null | string {
  const normalized = filePath.replaceAll('\\', '/');
  const match = normalized.match(pluginSvgIconPathPattern);
  if (!match?.[1] || !match[2]) {
    return null;
  }
  const pluginId = match[1].trim();
  const stem = match[2].trim();
  if (!pluginId || !stem || stem.includes('/') || stem.includes('..')) {
    return null;
  }
  return `svg:${pluginId}-${stem}`;
}
