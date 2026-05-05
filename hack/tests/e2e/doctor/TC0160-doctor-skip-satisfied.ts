import { execFile } from 'node:child_process';
import { mkdtemp, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { promisify } from 'node:util';

import { test, expect } from '../../fixtures/auth';

const execFileAsync = promisify(execFile);
const currentFile = fileURLToPath(import.meta.url);
const repoRoot = path.resolve(path.dirname(currentFile), '../../../..');

const satisfiedCheck = {
  os: 'macos',
  package_manager: 'brew',
  shell: 'zsh',
  repo_root_detected: true,
  tools: {
    go: { present: true, version: '1.22.0', min_version: '1.22.0', ok: true },
    node: { present: true, version: '20.19.0', min_version: '20.19.0', ok: true },
    pnpm: { present: true, version: '8.0.0', min_version: '8.0.0', ok: true },
    git: { present: true, version: '2.45.0', min_version: null, ok: true },
    make: { present: true, version: '3.81', min_version: null, ok: true },
    openspec: { present: true, version: '1.3.1', min_version: null, ok: true },
    gf: { present: true, version: '2.10.0', min_version: null, ok: true },
    playwright: { present: true, version: '1.58.2', min_version: null, ok: true },
    'goframe-v2': { present: true, version: null, min_version: null, ok: true },
  },
  path_issues: [],
  mirror_hints: [],
};

test.describe('TC-160 Doctor skip satisfied', () => {
  test('TC-160a: plan is empty when all tools are satisfied', async () => {
    const tempDir = await mkdtemp(path.join(tmpdir(), 'lina-doctor-'));
    const input = path.join(tempDir, 'check.json');
    await writeFile(input, JSON.stringify(satisfiedCheck));

    const result = await execFileAsync('bash', [
      '.claude/skills/lina-doctor/scripts/doctor-plan.sh',
      '--input',
      input,
    ], { cwd: repoRoot });

    const plan = JSON.parse(result.stdout);
    expect(plan.items).toEqual([]);
  });
});
