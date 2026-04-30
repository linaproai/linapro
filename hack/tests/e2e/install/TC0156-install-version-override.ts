import { execFile } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { promisify } from 'node:util';

import { test, expect } from '../../fixtures/auth';

const execFileAsync = promisify(execFile);
const currentFile = fileURLToPath(import.meta.url);
const repoRoot = path.resolve(path.dirname(currentFile), '../../../..');

test.describe('TC-156 Install version override', () => {
  test('TC-156a: LINAPRO_VERSION selects the requested fixture tag', async () => {
    const result = await execFileAsync('bash', [
      path.join(repoRoot, 'hack/tests/scripts/install-bootstrap.sh'),
      'version-override',
    ]);

    expect(result.stdout).toContain('PASS install-bootstrap version-override');
  });
});
