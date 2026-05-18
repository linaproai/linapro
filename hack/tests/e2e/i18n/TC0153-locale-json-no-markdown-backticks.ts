import { existsSync, readFileSync, readdirSync } from "node:fs";
import { dirname, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { test, expect } from "../../fixtures/auth";

type Finding = {
  file: string;
  keyPath: string;
  value: string;
};

const currentDir = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(currentDir, "../../../..");
const directLocaleRoots = [
  "apps/lina-core/manifest/i18n",
  "apps/lina-vben/apps/web-antd/src/locales/langs",
  "apps/lina-vben/packages/locales/src/langs",
];
const pluginWorkspaceRoot = "apps/lina-plugins";

function collectJsonFiles(rootPath: string): string[] {
  if (!existsSync(rootPath)) {
    return [];
  }

  return readdirSync(rootPath, { withFileTypes: true }).flatMap((entry) => {
    const entryPath = join(rootPath, entry.name);
    if (entry.isDirectory()) {
      return collectJsonFiles(entryPath);
    }
    if (entry.isFile() && entry.name.endsWith(".json")) {
      return [entryPath];
    }
    return [];
  });
}

function findBacktickStrings(
  value: unknown,
  file: string,
  keyPath: string[] = [],
): Finding[] {
  if (typeof value === "string") {
    if (!value.includes("`")) {
      return [];
    }
    return [
      {
        file,
        keyPath: keyPath.join("."),
        value,
      },
    ];
  }

  if (Array.isArray(value)) {
    return value.flatMap((item, index) =>
      findBacktickStrings(item, file, [...keyPath, String(index)]),
    );
  }

  if (value && typeof value === "object") {
    return Object.entries(value).flatMap(([key, item]) =>
      findBacktickStrings(item, file, [...keyPath, key]),
    );
  }

  return [];
}

function isPluginLocaleJson(file: string): boolean {
  const parts = relative(resolve(repoRoot, pluginWorkspaceRoot), file).split(
    /[\\/]/u,
  );
  const manifestIndex = parts.indexOf("manifest");
  return manifestIndex >= 1 && parts[manifestIndex + 1] === "i18n";
}

function collectLocaleJsonFiles(): string[] {
  const directFiles = directLocaleRoots.flatMap((root) =>
    collectJsonFiles(resolve(repoRoot, root)),
  );
  const pluginFiles = collectJsonFiles(resolve(repoRoot, pluginWorkspaceRoot)).filter(
    isPluginLocaleJson,
  );
  return [...directFiles, ...pluginFiles];
}

test.describe("TC-153 Locale JSON markdown backtick audit", () => {
  test(
    "TC-153a: frontend and delivery locale JSON values do not contain markdown-style backticks",
    async () => {
      const findings = collectLocaleJsonFiles().flatMap((file) => {
        const content = JSON.parse(readFileSync(file, "utf8")) as unknown;
        return findBacktickStrings(content, relative(repoRoot, file));
      });

      expect(
        findings.map(
          (finding) => `${finding.file}#${finding.keyPath}: ${finding.value}`,
        ),
      ).toEqual([]);
    },
  );
});
