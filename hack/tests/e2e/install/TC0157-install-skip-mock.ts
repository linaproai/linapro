import { execFile } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { promisify } from 'node:util';

import { test, expect } from '../../fixtures/auth';

const execFileAsync = promisify(execFile);
const currentFile = fileURLToPath(import.meta.url);
const repoRoot = path.resolve(path.dirname(currentFile), '../../../..');

test.describe('TC-157 Install skip mock', () => {
  test('TC-157a: LINAPRO_SKIP_MOCK is passed through to the platform script', async () => {
    const result = await execFileAsync('bash', [
      path.join(repoRoot, 'hack/tests/scripts/install-bootstrap.sh'),
      'skip-mock',
    ]);

    expect(result.stdout).toContain('PASS install-bootstrap skip-mock');
  });
});
