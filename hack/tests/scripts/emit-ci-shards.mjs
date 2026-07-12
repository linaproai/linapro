#!/usr/bin/env node
/**
 * Emit CI E2E shard matrices for GitHub Actions or local inspection.
 *
 * Usage:
 *   node ./scripts/emit-ci-shards.mjs
 *   node ./scripts/emit-ci-shards.mjs --format json
 *   node ./scripts/emit-ci-shards.mjs --format github-output
 *   node ./scripts/emit-ci-shards.mjs --format summary
 */
import { appendFileSync } from 'node:fs';

import {
  planAllCiShards,
  toGitHubMatrix,
} from './ci-shards.mjs';
import { isPluginWorkspaceReady } from './execution-governance.mjs';

function parseArgs(argv) {
  let format = 'summary';
  let githubOutput = process.env.GITHUB_OUTPUT ?? '';
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === '--format') {
      format = argv[index + 1] ?? format;
      index += 1;
      continue;
    }
    if (arg.startsWith('--format=')) {
      format = arg.slice('--format='.length);
      continue;
    }
    if (arg === '--github-output') {
      githubOutput = argv[index + 1] ?? githubOutput;
      index += 1;
      continue;
    }
  }
  return { format, githubOutput };
}

function writeGitHubOutput(filePath, key, value) {
  // Use heredoc form so JSON quotes/brackets are safe in GITHUB_OUTPUT.
  const serialized = typeof value === 'string' ? value : JSON.stringify(value);
  const delimiter = `GHEXOF_${key.replace(/[^A-Za-z0-9_]/g, '_')}`;
  const payload = `${key}<<${delimiter}\n${serialized}\n${delimiter}\n`;
  if (filePath) {
    appendFileSync(filePath, payload, 'utf8');
  } else {
    process.stdout.write(payload);
  }
}

const { format, githubOutput } = parseArgs(process.argv.slice(2));
const plan = planAllCiShards(undefined, {
  requireWorkspace: isPluginWorkspaceReady(),
});

if (plan.errors.length > 0) {
  console.error('CI shard planning failed:');
  for (const error of plan.errors) {
    console.error(`- ${error}`);
  }
  process.exit(1);
}

const hostMatrix = toGitHubMatrix(plan.host);
const pluginFullMatrix = toGitHubMatrix(plan.pluginFull);
const pluginMatrix = toGitHubMatrix(plan.plugin);

if (format === 'json') {
  console.log(
    JSON.stringify(
      {
        defaults: plan.defaults,
        host: hostMatrix,
        plugin: pluginMatrix,
        pluginFullExtra: toGitHubMatrix(plan.pluginFullExtra),
        pluginFull: pluginFullMatrix,
        hostDetail: plan.host.map((shard) => ({
          name: shard.name,
          files: shard.files.length,
          entries: shard.entries,
        })),
        pluginDetail: plan.plugin.map((shard) => ({
          name: shard.name,
          files: shard.files.length,
          weight: shard.weight,
          pluginIds: shard.pluginIds,
        })),
      },
      null,
      2,
    ),
  );
  process.exit(0);
}

if (format === 'github-output') {
  writeGitHubOutput(githubOutput, 'host-shards', hostMatrix);
  writeGitHubOutput(githubOutput, 'plugin-full-shards', pluginFullMatrix);
  writeGitHubOutput(githubOutput, 'parallel-workers', plan.defaults.parallelWorkers);
  if (!githubOutput) {
    // already wrote to stdout
  } else {
    console.log(
      `Wrote host-shards (${hostMatrix.length}), plugin-full-shards (${pluginFullMatrix.length}), parallel-workers=${plan.defaults.parallelWorkers}`,
    );
  }
  process.exit(0);
}

// summary (default)
console.log(`parallelWorkers=${plan.defaults.parallelWorkers}`);
console.log(`host shards (${hostMatrix.length}):`);
for (const shard of plan.host) {
  console.log(`  - ${shard.name}: files=${shard.files.length} entries=${shard.entries.join(',')}`);
}
console.log(`plugin-full-extra shards (${plan.pluginFullExtra.length}):`);
for (const shard of plan.pluginFullExtra) {
  console.log(`  - ${shard.name}: scope=${shard.scope} files=${shard.files.length}`);
}
if (plan.pluginSkipped) {
  console.log('plugin shards: skipped (plugin workspace not ready)');
} else {
  console.log(`plugin shards (${plan.plugin.length}):`);
  for (const shard of plan.plugin) {
    console.log(
      `  - ${shard.name}: files=${shard.files.length} weight=${shard.weight} plugins=${shard.pluginIds.join(',')}`,
    );
  }
}
console.log('commands:');
for (const shard of [...plan.host, ...plan.pluginFull]) {
  console.log(`  ${shard.command}`);
}
