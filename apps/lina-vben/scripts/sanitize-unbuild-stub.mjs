#!/usr/bin/env node
import { readdir, readFile, writeFile } from 'node:fs/promises';
import { dirname, extname, join, relative, resolve, sep } from 'node:path';
import { fileURLToPath } from 'node:url';

const jsExtensions = new Set(['.cjs', '.js', '.mjs']);
const declarationExtensions = ['.d.cts', '.d.mts', '.d.ts'];

const scriptDir = dirname(fileURLToPath(import.meta.url));
const frontendRoot = resolve(scriptDir, '..');
const packageRootArg = process.argv
  .slice(2)
  .find((arg) => !arg.startsWith('-'));
const packageRoot = packageRootArg ? resolve(packageRootArg) : process.cwd();
const packageJson = JSON.parse(
  await readFile(join(packageRoot, 'package.json'), 'utf8'),
);
const distDir = join(packageRoot, 'dist');
const absolutePackageRootPattern = new RegExp(
  packageRoot
    .replaceAll(/[.*+?^${}()|[\]\\]/g, '\\$&')
    .replaceAll('\\', '[\\\\/]'),
);

function toPosixPath(value) {
  return value.split(sep).join('/');
}

function relativeImport(fromFile, targetFile) {
  const rel = toPosixPath(relative(dirname(fromFile), targetFile));
  return rel.startsWith('.') ? rel : `./${rel}`;
}

async function collectDistFiles(dir) {
  let entries = [];
  try {
    entries = await readdir(dir, { withFileTypes: true });
  } catch {
    return [];
  }

  const files = [];
  for (const entry of entries) {
    const fullPath = join(dir, entry.name);
    if (entry.isDirectory()) {
      files.push(...(await collectDistFiles(fullPath)));
    } else {
      files.push(fullPath);
    }
  }
  return files;
}

function isDeclarationFile(filePath) {
  return declarationExtensions.some((ext) => filePath.endsWith(ext));
}

function sourceRelFromAbsolute(sourcePath) {
  const marker = '/src/';
  const normalized = sourcePath.replaceAll('\\', '/');
  const markerIndex = normalized.lastIndexOf(marker);
  if (markerIndex === -1) {
    return '';
  }
  return `src/${normalized.slice(markerIndex + marker.length)}`;
}

function sanitizeDeclaration(content, filePath) {
  return content.replaceAll(/"([^"]+\/src\/[^"]+\.js)"/g, (_, sourcePath) => {
    const sourceRel = sourceRelFromAbsolute(sourcePath);
    if (!sourceRel) {
      return `"${sourcePath}"`;
    }
    return `"${relativeImport(filePath, join(packageRoot, sourceRel))}"`;
  });
}

function sanitizeEsmStub(content, filePath) {
  const sourceMatch = content.match(
    /jiti\.import\("([^"]+\/src\/[^"]+\.ts)"\)/,
  );
  if (!sourceMatch) {
    return content;
  }

  const sourceRel = sourceRelFromAbsolute(sourceMatch[1]);
  if (!sourceRel) {
    return content;
  }

  const packageRootRel = relativeImport(filePath, packageRoot);
  const sourceTypeRel = relativeImport(
    filePath,
    join(packageRoot, sourceRel.replace(/(\.m?)ts$/, '$1js')),
  );
  const sourceEntryRel = toPosixPath(sourceRel);
  const importLine = content.match(/^import \{ createJiti \} from .+;$/m)?.[0];
  if (!importLine) {
    return content;
  }

  let next = content.replace(
    importLine,
    [
      importLine,
      'import { dirname, resolve } from "node:path";',
      'import { fileURLToPath } from "node:url";',
    ].join('\n'),
  );

  next = next.replace(
    '\nconst jiti = createJiti(',
    [
      '',
      'const __filename = fileURLToPath(import.meta.url);',
      'const __dirname = dirname(__filename);',
      `const packageRoot = resolve(__dirname, ${JSON.stringify(packageRootRel)});`,
      `const sourceEntry = resolve(packageRoot, ${JSON.stringify(sourceEntryRel)});`,
      '',
      'const jiti = createJiti(',
    ].join('\n'),
  );

  if (packageJson.name) {
    const aliasPattern = new RegExp(
      `("${packageJson.name.replaceAll(/[.*+?^${}()|[\]\\]/g, '\\$&')}"\\s*:\\s*)"[^"]+"`,
      'g',
    );
    next = next.replaceAll(aliasPattern, '$1packageRoot');
  }

  next = next.replace(
    /\/\*\* @type \{import\("[^"]+\/src\/[^"]+\.js"\)\} \*\//,
    `/** @type {import("${sourceTypeRel}")} */`,
  );
  next = next.replace(
    /const _module = await jiti\.import\("[^"]+\/src\/[^"]+\.ts"\);/,
    'const _module = await jiti.import(sourceEntry);',
  );

  return next;
}

function sanitizeFile(content, filePath) {
  if (isDeclarationFile(filePath)) {
    return sanitizeDeclaration(content, filePath);
  }
  if (jsExtensions.has(extname(filePath))) {
    return sanitizeEsmStub(content, filePath);
  }
  return content;
}

for (const filePath of await collectDistFiles(distDir)) {
  const content = await readFile(filePath, 'utf8');
  const sanitized = sanitizeFile(content, filePath);
  if (absolutePackageRootPattern.test(sanitized)) {
    throw new Error(
      `unbuild stub still contains an absolute package path after sanitizing ${filePath}`,
    );
  }
  if (sanitized !== content) {
    await writeFile(filePath, sanitized);
  }
}

if (packageRoot === frontendRoot) {
  process.exit(0);
}
