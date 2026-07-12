// Resolves whether a Playwright suite label represents a host-only run.
// Host-only runs must set E2E_HOST_ONLY_PLUGINS=1 so tests skip plugin-owned fixtures.

/**
 * @param {string | undefined | null} label Playwright run label from run-suite.
 * @returns {boolean} true when the label is a host-only host / host-module / ci-shard host path.
 */
export function isHostOnlyRunLabel(label) {
  const value = String(label ?? '');
  return (
    value === 'host' ||
    value.startsWith('host:') ||
    value.startsWith('host-module') ||
    value.startsWith('ci-shard:host:')
  );
}

/**
 * @param {string | undefined | null} label Playwright run label from run-suite.
 * @param {string | undefined} existingEnv Current process env value when not host-only.
 * @returns {'1' | string} Env value for E2E_HOST_ONLY_PLUGINS.
 */
export function resolveHostOnlyPluginsEnv(label, existingEnv) {
  if (isHostOnlyRunLabel(label)) {
    return '1';
  }
  return existingEnv ?? '0';
}
