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

test.describe('TC-163 Doctor goframe-v2 nonblocking failure', () => {
  test('TC-163a: optional goframe-v2 failure does not stop install runner', async () => {
    const tempDir = await mkdtemp(path.join(tmpdir(), 'lina-doctor-'));
    const planFile = path.join(tempDir, 'plan.json');
    await writeFile(planFile, [
      '{',
      '  "mirror_hints": [],',
      '  "items": [',
      '    { "tool": "goframe-v2", "command": "false", "package_manager": "test", "requires_sudo": false, "depends_on": ["node"], "optional": true }',
      '  ]',
      '}',
      '',
    ].join('\n'));

    const result = await execFileAsync('bash', [
      '.claude/skills/lina-doctor/scripts/doctor-install.sh',
      '--plan',
      planFile,
    ], {
      cwd: repoRoot,
      env: { ...process.env, LINAPRO_DOCTOR_NON_INTERACTIVE: '1' },
    });

    expect(result.stdout).toContain('Optional tool goframe-v2 failed; continuing.');
    expect(result.stdout).toContain('Lina Doctor verification report');
  });
});
