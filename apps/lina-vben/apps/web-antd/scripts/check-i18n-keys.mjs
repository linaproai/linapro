import { readdirSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const appRoot = path.resolve(scriptDir, '..');
const localesRoot = path.resolve(appRoot, 'src/locales/langs');
const referenceLocale = 'en-US';
const requiredLocales = ['zh-CN'];
const errors = [];

function readJson(filePath) {
  try {
    return JSON.parse(readFileSync(filePath, 'utf8'));
  } catch (error) {
    errors.push(
      `Unable to parse ${path.relative(appRoot, filePath)}: ${error.message}`,
    );
    return {};
  }
}

function listJsonFiles(locale) {
  const localeDir = path.resolve(localesRoot, locale);
  return readdirSync(localeDir, { withFileTypes: true })
    .filter((entry) => entry.isFile() && entry.name.endsWith('.json'))
    .map((entry) => entry.name)
    .toSorted();
}

function collectKeys(value, prefix = '') {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) {
    return [prefix];
  }

  return Object.entries(value).flatMap(([key, child]) =>
    collectKeys(child, prefix ? `${prefix}.${key}` : key),
  );
}

const referenceFiles = listJsonFiles(referenceLocale);

for (const locale of requiredLocales) {
  const localeFiles = listJsonFiles(locale);
  const localeFileSet = new Set(localeFiles);

  for (const file of referenceFiles) {
    if (!localeFileSet.has(file)) {
      errors.push(`${locale} is missing ${file}`);
      continue;
    }

    const referenceMessages = readJson(
      path.resolve(localesRoot, referenceLocale, file),
    );
    const localeMessages = readJson(path.resolve(localesRoot, locale, file));
    const referenceKeys = new Set(collectKeys(referenceMessages));
    const localeKeys = new Set(collectKeys(localeMessages));

    for (const key of referenceKeys) {
      if (!localeKeys.has(key)) {
        errors.push(`${locale}/${file} is missing key ${key}`);
      }
    }
    for (const key of localeKeys) {
      if (!referenceKeys.has(key)) {
        errors.push(`${locale}/${file} has extra key ${key}`);
      }
    }
  }

  for (const file of localeFiles) {
    if (!referenceFiles.includes(file)) {
      errors.push(`${locale} has extra locale file ${file}`);
    }
  }
}

if (errors.length > 0) {
  console.error('Frontend i18n key validation failed:');
  for (const error of errors) {
    console.error(`- ${error}`);
  }
  process.exit(1);
}

console.log(
  `Validated ${referenceFiles.length} frontend locale files across ${
    requiredLocales.length + 1
  } locales.`,
);
