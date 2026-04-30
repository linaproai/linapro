import { execFile } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { promisify } from 'node:util';

import { test, expect } from '../../fixtures/auth';

const execFileAsync = promisify(execFile);
const currentFile = fileURLToPath(import.meta.url);
const repoRoot = path.resolve(path.dirname(currentFile), '../../../..');

test.describe('TC-155 Install default bootstrap', () => {
  test('TC-155a: bootstrap clones the default target directory from a fixture repository', async () => {
    const result = await execFileAsync('bash', [
      path.join(repoRoot, 'hack/tests/scripts/install-bootstrap.sh'),
      'default',
    ]);

    expect(result.stdout).toContain('PASS install-bootstrap default');
  });
});
