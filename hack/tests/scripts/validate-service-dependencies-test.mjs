import { mkdtempSync, mkdirSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..', '..', '..');
const scriptPath = path.resolve(scriptDir, 'validate-service-dependencies.mjs');

function runValidator(root, baselinePath) {
  return spawnSync('node', [scriptPath], {
    cwd: repoRoot,
    encoding: 'utf8',
    env: {
      ...process.env,
      LINAPRO_SERVICE_DEP_SCAN_ROOTS: root,
      LINAPRO_SERVICE_DEP_BASELINE: baselinePath,
    },
  });
}

function writeFixture(root) {
  const serviceDir = path.join(root, 'apps/lina-core/internal/service/auth');
  const controllerDir = path.join(root, 'apps/lina-core/internal/controller/demo');
  mkdirSync(serviceDir, { recursive: true });
  mkdirSync(controllerDir, { recursive: true });
  writeFileSync(
    path.join(serviceDir, 'auth.go'),
    'package auth\n\nfunc New() any { return nil }\n',
  );
  writeFileSync(
    path.join(controllerDir, 'demo.go'),
    [
      'package demo',
      '',
      'import authsvc "lina-core/internal/service/auth"',
      '',
      'func Build() any {',
      '\treturn authsvc.New()',
      '}',
      '',
    ].join('\n'),
  );
  return path.relative(repoRoot, path.join(controllerDir, 'demo.go'));
}

const tmpRoot = mkdtempSync(path.join(tmpdir(), 'linapro-service-deps-'));
const fixturePath = writeFixture(tmpRoot);
const emptyBaseline = path.join(tmpRoot, 'empty-baseline.json');
const allowedBaseline = path.join(tmpRoot, 'allowed-baseline.json');

writeFileSync(
  emptyBaseline,
  JSON.stringify({ version: 1, entries: [] }, null, 2),
);
writeFileSync(
  allowedBaseline,
  JSON.stringify(
    {
      version: 1,
      entries: [
        {
          path: fixturePath,
          maxCalls: 1,
          reason: 'fixture baseline',
        },
      ],
    },
    null,
    2,
  ),
);

const failing = runValidator(tmpRoot, emptyBaseline);
if (failing.status === 0) {
  console.error('expected validator to fail for non-baseline critical constructor');
  process.exit(1);
}
if (!failing.stderr.includes('non-baseline file')) {
  console.error(failing.stderr);
  console.error('expected failure to mention non-baseline file');
  process.exit(1);
}

const passing = runValidator(tmpRoot, allowedBaseline);
if (passing.status !== 0) {
  console.error(passing.stdout);
  console.error(passing.stderr);
  process.exit(passing.status ?? 1);
}

console.log('Service dependency governance validator self-test passed.');
