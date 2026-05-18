import { spawnSync } from "node:child_process";
import {
  copyFileSync,
  existsSync,
  mkdirSync,
  readFileSync,
  renameSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const testsDir = path.resolve(scriptDir, "..");
const repoRoot = path.resolve(testsDir, "..", "..");
const coreDir = path.join(repoRoot, "apps", "lina-core");
const configPath = path.join(coreDir, "manifest", "config", "config.yaml");
const configTemplatePath = path.join(
  coreDir,
  "manifest",
  "config",
  "config.template.yaml",
);
const backupConfigPath = path.join(
  repoRoot,
  "temp",
  "sqlite-e2e",
  "config.yaml.backup",
);
const sqliteLink = "sqlite::@file(./temp/sqlite/linapro.db)";
const sqliteDbPath = path.join(coreDir, "temp", "sqlite", "linapro.db");
const backendLogPath = path.join(repoRoot, "temp", "lina-core.log");
const sqliteMode = process.argv[2] ?? "full";
const extraArgs = process.argv.slice(3);
const sqliteTestFiles = [
  "e2e/dialect/TC0164-sqlite-mode-startup.ts",
  "e2e/dialect/TC0165-sqlite-mode-business-zero-impact.ts",
  "e2e/dialect/TC0166-sqlite-mode-rebuild-and-reseed.ts",
];
const sqliteSmokeTestFiles = ["e2e/dialect/TC0164-sqlite-mode-startup.ts"];

function selectedTestFiles() {
  if (sqliteMode === "full") {
    return sqliteTestFiles;
  }
  if (sqliteMode === "smoke") {
    return sqliteSmokeTestFiles;
  }
  throw new Error(
    `[sqlite-e2e] unknown mode ${sqliteMode}; expected "full" or "smoke"`,
  );
}

function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: options.cwd ?? repoRoot,
    env: options.env ?? process.env,
    stdio: "inherit",
  });
  if (result.status !== 0) {
    throw new Error(
      `[sqlite-e2e] command failed with exit code ${result.status ?? 1}: ${command} ${args.join(" ")}`,
    );
  }
}

function runCleanup(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: options.cwd ?? repoRoot,
    env: options.env ?? process.env,
    stdio: "inherit",
  });
  if (result.status !== 0) {
    console.warn(
      `[sqlite-e2e] cleanup command failed: ${command} ${args.join(" ")}`,
    );
  }
}

function writeSQLiteConfig() {
  mkdirSync(path.dirname(backupConfigPath), { recursive: true });
  if (existsSync(configPath)) {
    copyFileSync(configPath, backupConfigPath);
  } else {
    copyFileSync(configTemplatePath, backupConfigPath);
  }

  let content = readFileSync(configTemplatePath, "utf8");
  content = content.replace(
    /link:\s*"[^"]+"/,
    `link: "${sqliteLink}"`,
  );
  content = content.replace(
    /cluster:\n(\s+)enabled:\s*(?:true|false)/,
    "cluster:\n$1enabled: true",
  );
  writeFileSync(configPath, content);
}

function restoreConfig() {
  if (existsSync(backupConfigPath)) {
    renameSync(backupConfigPath, configPath);
  }
}

function cleanSQLiteFiles() {
  mkdirSync(path.dirname(sqliteDbPath), { recursive: true });
  for (const file of [sqliteDbPath, `${sqliteDbPath}-shm`, `${sqliteDbPath}-wal`]) {
    rmSync(file, { force: true });
  }
}

let restored = false;
function restoreOnce() {
  if (restored) {
    return;
  }
  restored = true;
  restoreConfig();
}

process.on("exit", restoreOnce);
process.on("SIGINT", () => {
  restoreOnce();
  process.exit(130);
});
process.on("SIGTERM", () => {
  restoreOnce();
  process.exit(143);
});

writeSQLiteConfig();
cleanSQLiteFiles();

try {
  const testFiles = selectedTestFiles();
  run("make", ["init", "confirm=init", "rebuild=true"], { cwd: repoRoot });
  run("make", ["mock", "confirm=mock"], { cwd: repoRoot });
  run("make", ["dev"], { cwd: repoRoot });
  run(
    "pnpm",
    [
      "exec",
      "playwright",
      "test",
      ...testFiles,
      "--workers=1",
      ...extraArgs,
    ],
    {
      cwd: testsDir,
      env: {
        ...process.env,
        LINAPRO_E2E_SQLITE_MODE: "1",
        LINAPRO_E2E_SQLITE_BACKEND_LOG: backendLogPath,
      },
    },
  );
} catch (error) {
  console.error(error instanceof Error ? error.message : error);
  process.exitCode = 1;
} finally {
  runCleanup("make", ["stop"], { cwd: repoRoot });
  restoreOnce();
}
