import assert from 'node:assert/strict';

import {
  binPackByWeight,
  planAllCiShards,
  planHostShards,
  toGitHubMatrix,
  validateCiShards,
} from './ci-shards.mjs';
import { isPluginWorkspaceReady, resolveHostOnlyEntries } from './execution-governance.mjs';

function testBinPackBalancesHeavyItems() {
  const packed = binPackByWeight(
    [
      { id: 'heavy', weight: 50, files: Array.from({ length: 50 }, (_, i) => `h${i}`) },
      { id: 'a', weight: 10, files: ['a1'] },
      { id: 'b', weight: 10, files: ['b1'] },
      { id: 'c', weight: 10, files: ['c1'] },
      { id: 'd', weight: 5, files: ['d1'] },
      { id: 'e', weight: 5, files: ['e1'] },
    ],
    3,
  );
  assert.equal(packed.length, 3);
  assert.ok(packed.some((bin) => bin.pluginIds.includes('heavy')));
  const weights = packed.map((bin) => bin.weight).sort((left, right) => right - left);
  // Heavy item alone should not share with all others; remaining weight is split.
  assert.ok(weights[0] >= 50);
  assert.equal(
    packed.reduce((sum, bin) => sum + bin.pluginIds.length, 0),
    6,
  );
}

function testBinPackIsDeterministic() {
  const items = [
    { id: 'z', weight: 3, files: ['z'] },
    { id: 'a', weight: 3, files: ['a'] },
    { id: 'm', weight: 2, files: ['m'] },
  ];
  const first = JSON.stringify(binPackByWeight(items, 2));
  const second = JSON.stringify(binPackByWeight(items, 2));
  assert.equal(first, second);
}

function testHostShardsPartitionHostOnlyFiles() {
  const { shards, errors } = planHostShards();
  assert.deepEqual(errors, []);
  assert.ok(shards.length >= 4);
  const covered = new Set(shards.flatMap((shard) => shard.files));
  const expected = new Set(resolveHostOnlyEntries(['e2e']));
  assert.deepEqual([...covered].sort(), [...expected].sort());
  for (const shard of shards) {
    assert.ok(shard.command.includes(`ci-shard host ${shard.name}`));
  }
}

function testValidateCiShardsPasses() {
  const errors = validateCiShards(undefined, {
    requireWorkspace: isPluginWorkspaceReady(),
  });
  assert.deepEqual(errors, []);
}

function testGitHubMatrixShape() {
  const plan = planAllCiShards(undefined, {
    requireWorkspace: isPluginWorkspaceReady(),
  });
  assert.deepEqual(plan.errors, []);
  const matrix = toGitHubMatrix(plan.host);
  assert.ok(matrix.length > 0);
  for (const item of matrix) {
    assert.equal(typeof item.name, 'string');
    assert.equal(typeof item.command, 'string');
    assert.ok(item.command.startsWith('node ./scripts/run-suite.mjs ci-shard '));
  }
  if (isPluginWorkspaceReady()) {
    assert.ok(plan.plugin.length >= 1);
    assert.ok(plan.pluginFull.length > plan.plugin.length || plan.pluginFullExtra.length === 0);
    const pluginNames = new Set(plan.plugin.map((shard) => shard.name));
    assert.equal(pluginNames.size, plan.plugin.length);
  }
}

testBinPackBalancesHeavyItems();
testBinPackIsDeterministic();
testHostShardsPartitionHostOnlyFiles();
testValidateCiShardsPasses();
testGitHubMatrixShape();

console.log('ci-shards unit tests passed.');
