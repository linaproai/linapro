import { execFileSync } from 'node:child_process';
import { existsSync, readFileSync } from 'node:fs';

const postgresAppPsql = '/Applications/Postgres.app/Contents/Versions/18/bin/psql';
const defaultPsqlBin = existsSync(postgresAppPsql) ? postgresAppPsql : 'psql';

export const psqlBin = process.env.E2E_PSQL_BIN ?? defaultPsqlBin;

type PostgresConnection = {
  database: string;
  host: string;
  password: string;
  port: string;
  sslmode: string;
  user: string;
};

function postgresConnection(): PostgresConnection {
  if (process.env.E2E_PG_URL) {
    const url = new URL(process.env.E2E_PG_URL);
    return {
      database: decodeURIComponent(url.pathname.replace(/^\//u, '')) || 'linapro',
      host: url.hostname || '127.0.0.1',
      password: decodeURIComponent(url.password || ''),
      port: url.port || '5432',
      sslmode: url.searchParams.get('sslmode') ?? 'disable',
      user: decodeURIComponent(url.username || 'postgres'),
    };
  }
  return {
    database: process.env.E2E_DB_NAME ?? 'linapro',
    host: process.env.E2E_DB_HOST ?? '127.0.0.1',
    password: process.env.E2E_DB_PASSWORD ?? '12345678',
    port: process.env.E2E_DB_PORT ?? '5432',
    sslmode: process.env.E2E_DB_SSLMODE ?? 'disable',
    user: process.env.E2E_DB_USER ?? 'postgres',
  };
}

const pgConnection = postgresConnection();

function assertSafePostgresTarget() {
  if (process.env.E2E_ALLOW_DESTRUCTIVE_PG === '1') {
    return;
  }
  const allowedHosts = new Set(['127.0.0.1', 'localhost', '::1']);
  if (!allowedHosts.has(pgConnection.host)) {
    throw new Error(
      `Refusing destructive PostgreSQL E2E helper on non-local host ${pgConnection.host}. Set E2E_ALLOW_DESTRUCTIVE_PG=1 only for an isolated test database.`,
    );
  }
  if (!/^(linapro|linapro_e2e|linapro_test|test_)/u.test(pgConnection.database)) {
    throw new Error(
      `Refusing destructive PostgreSQL E2E helper on database ${pgConnection.database}. Use a linapro/test database or set E2E_ALLOW_DESTRUCTIVE_PG=1 for an isolated test database.`,
    );
  }
}

function psqlEnv() {
  return {
    ...process.env,
    PGDATABASE: pgConnection.database,
    PGHOST: pgConnection.host,
    PGPASSWORD: pgConnection.password,
    PGPORT: pgConnection.port,
    PGSSLMODE: pgConnection.sslmode,
    PGUSER: pgConnection.user,
  };
}

function psqlArgs(extraArgs: string[]) {
  return ['-X', '-v', 'ON_ERROR_STOP=1', ...extraArgs];
}

export function execPgSQL(sql: string) {
  assertSafePostgresTarget();
  execFileSync(psqlBin, psqlArgs(['-q', '-c', sql]), {
    env: psqlEnv(),
    stdio: ['ignore', 'ignore', 'inherit'],
  });
}

export function execPgSQLStatements(statements: string[]) {
  execPgSQL(statements.join('\n'));
}

export function execPgSQLFile(filePath: string) {
  assertSafePostgresTarget();
  execFileSync(psqlBin, psqlArgs(['-q']), {
    env: psqlEnv(),
    input: readFileSync(filePath, 'utf8'),
    stdio: ['pipe', 'ignore', 'inherit'],
  });
}

export function queryPgRows(sql: string): string[] {
  const output = execFileSync(
    psqlBin,
    psqlArgs(['-A', '-t', '-q', '-c', sql]),
    {
      encoding: 'utf8',
      env: psqlEnv(),
    },
  );
  return output
    .split(/\r?\n/u)
    .map((line) => line.trim())
    .filter(Boolean);
}

export function queryPgScalar(sql: string) {
  return queryPgRows(sql)[0] ?? '';
}

export function pgEscapeLiteral(value: string) {
  return value.replaceAll("'", "''");
}

export function pgEscapeIdentifier(value: string) {
  return value.replaceAll('"', '""');
}

export function pgIdentifier(value: string) {
  return `"${pgEscapeIdentifier(value)}"`;
}
