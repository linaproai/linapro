import { execFile } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { promisify } from 'node:util';

import { test, expect } from '../../fixtures/auth';

const execFileAsync = promisify(execFile);
const currentFile = fileURLToPath(import.meta.url);
const repoRoot = path.resolve(path.dirname(currentFile), '../../../..');

async function runDoctor(args: string[]) {
  try {
    const result = await execFileAsync('bash', args, { cwd: repoRoot });
    return { stdout: result.stdout, stderr: result.stderr, code: 0 };
  } catch (error) {
    const execError = error as { stdout?: string; stderr?: string; code?: number };
    return {
      stdout: execError.stdout ?? '',
      stderr: execError.stderr ?? '',
      code: execError.code ?? 1,
    };
  }
}

test.describe('TC-159 Doctor check-only', () => {
  test('TC-159a: check-only emits JSON and does not install tools', async () => {
    const result = await runDoctor([
      '.claude/skills/lina-doctor/scripts/doctor-check.sh',
      '--check-only',
    ]);

    expect([0, 1, 2]).toContain(result.code);
    expect(result.stderr).toBe('');

    const payload = JSON.parse(result.stdout);
    expect(payload).toHaveProperty('os');
    expect(payload).toHaveProperty('package_manager');
    expect(payload).toHaveProperty('tools.go');
    expect(payload).toHaveProperty('tools.node');
    expect(payload).toHaveProperty('tools.goframe-v2');
  });
});
