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

test.describe('TC-161 Doctor skip Playwright', () => {
  test('TC-161a: skip env removes Playwright from plan', async () => {
    const tempDir = await mkdtemp(path.join(tmpdir(), 'lina-doctor-'));
    const input = path.join(tempDir, 'check.json');
    await writeFile(input, JSON.stringify({
      os: 'macos',
      package_manager: 'brew',
      shell: 'zsh',
      repo_root_detected: true,
      tools: {
        go: { ok: true },
        node: { ok: true },
        pnpm: { ok: true },
        git: { ok: true },
        make: { ok: true },
        openspec: { ok: true },
        gf: { ok: true },
        playwright: { ok: false },
        'goframe-v2': { ok: true },
      },
      path_issues: [],
      mirror_hints: [],
    }));

    const result = await execFileAsync('bash', [
      '.claude/skills/lina-doctor/scripts/doctor-plan.sh',
      '--input',
      input,
    ], {
      cwd: repoRoot,
      env: { ...process.env, LINAPRO_DOCTOR_SKIP_PLAYWRIGHT: '1' },
    });

    const plan = JSON.parse(result.stdout);
    expect(plan.items.map((item: { tool: string }) => item.tool)).not.toContain('playwright');
  });
});
