import type { CAC } from 'cac';

import { mkdtemp, readFile, rm } from 'node:fs/promises';
import { createRequire } from 'node:module';
import { tmpdir } from 'node:os';
import { extname } from 'node:path';
import { dirname, join } from 'node:path';

import { execa, getStagedFiles } from '@vben/node-utils';

// 默认配置
const DEFAULT_CONFIG = {
  allowedExtensions: ['.cjs', '.js', '.jsx', '.mjs', '.ts', '.tsx', '.vue'],
  ignoreDirs: [
    'dist',
    '.turbo',
    'output',
    '.cache',
    'scripts',
    'internal',
    'packages/effects/request/src/',
    'packages/@core/ui-kit/menu-ui/src/',
    'packages/@core/ui-kit/popup-ui/src/',
  ],
  threshold: 0, // 循环依赖的阈值
} as const;

// 类型定义
type CircularDependencyResult = string[];

interface CircularDepsDetectOptions {
  absolute: boolean;
  cwd: string;
  ignore: string[];
  verbose: boolean;
}

interface CheckCircularConfig {
  allowedExtensions?: string[];
  ignoreDirs?: string[];
  threshold?: number;
}

interface CommandOptions {
  config?: CheckCircularConfig;
  staged: boolean;
  verbose: boolean;
}

// 缓存机制
const cache = new Map<string, CircularDependencyResult[]>();
const require = createRequire(import.meta.url);

function isNotFoundError(error: unknown): boolean {
  return (
    typeof error === 'object' &&
    error !== null &&
    'code' in error &&
    (error as { code?: string }).code === 'ENOENT'
  );
}

function resolveScannerCliPath(): string {
  const packageJsonPath =
    require.resolve('circular-dependency-scanner/package.json');

  return join(dirname(packageJsonPath), 'dist/cli.js');
}

async function readScannerOutput(
  outputPath: string,
): Promise<CircularDependencyResult[]> {
  try {
    const raw = await readFile(outputPath, 'utf8');
    return JSON.parse(raw) as CircularDependencyResult[];
  } catch (error) {
    if (isNotFoundError(error)) {
      return [];
    }

    throw error;
  }
}

async function detectCircularDependencies({
  absolute,
  cwd,
  ignore,
  verbose,
}: CircularDepsDetectOptions): Promise<CircularDependencyResult[]> {
  const outputDir = await mkdtemp(join(tmpdir(), 'vsh-circular-'));
  const outputPath = join(outputDir, 'circles.json');

  try {
    const args = [
      resolveScannerCliPath(),
      ...(absolute ? ['--absolute'] : []),
      '--ignore',
      ...ignore,
      '--output',
      outputPath,
    ];

    await execa(process.execPath, args, {
      cwd,
      stdio: verbose ? 'inherit' : 'pipe',
    });

    return await readScannerOutput(outputPath);
  } finally {
    await rm(outputDir, { force: true, recursive: true });
  }
}

/**
 * 格式化循环依赖的输出
 * @param circles - 循环依赖结果
 */
function formatCircles(circles: CircularDependencyResult[]): void {
  if (circles.length === 0) {
    console.log('✅ No circular dependencies found');
    return;
  }

  console.log('⚠️ Circular dependencies found:');
  circles.forEach((circle, index) => {
    console.log(`\nCircular dependency #${index + 1}:`);
    circle.forEach((file) => console.log(`  → ${file}`));
  });
}

/**
 * 检查项目中的循环依赖
 * @param options - 检查选项
 * @param options.staged - 是否只检查暂存区文件
 * @param options.verbose - 是否显示详细信息
 * @param options.config - 自定义配置
 * @returns Promise<void>
 */
async function checkCircular({
  config = {},
  staged,
  verbose,
}: CommandOptions): Promise<void> {
  try {
    // 合并配置
    const finalConfig = {
      ...DEFAULT_CONFIG,
      ...config,
    };

    // 生成忽略模式
    const ignorePattern = `**/{${finalConfig.ignoreDirs.join(',')}}/**`;

    // 检查缓存
    const cacheKey = `${staged}-${process.cwd()}-${ignorePattern}`;
    if (cache.has(cacheKey)) {
      const cachedResults = cache.get(cacheKey);
      if (cachedResults) {
        verbose && formatCircles(cachedResults);
      }
      return;
    }

    // 检测循环依赖
    const results = await detectCircularDependencies({
      absolute: staged,
      cwd: process.cwd(),
      ignore: [ignorePattern],
      verbose,
    });

    if (staged) {
      let files = await getStagedFiles();
      const allowedExtensions = new Set(finalConfig.allowedExtensions);

      // 过滤文件列表
      files = files.filter((file) => allowedExtensions.has(extname(file)));

      const circularFiles: CircularDependencyResult[] = [];

      for (const file of files) {
        for (const result of results) {
          const resultFiles = result.flat();
          if (resultFiles.includes(file)) {
            circularFiles.push(result);
          }
        }
      }

      // 更新缓存
      cache.set(cacheKey, circularFiles);
      verbose && formatCircles(circularFiles);
    } else {
      // 更新缓存
      cache.set(cacheKey, results);
      verbose && formatCircles(results);
    }

    // 如果发现循环依赖，只输出警告信息
    if (results.length > 0) {
      console.log(
        '\n⚠️ Warning: Circular dependencies found, please check and fix',
      );
    }
  } catch (error) {
    console.error(
      '❌ Error checking circular dependencies:',
      error instanceof Error ? error.message : error,
    );
  }
}

/**
 * 定义检查循环依赖的命令
 * @param cac - CAC实例
 */
function defineCheckCircularCommand(cac: CAC): void {
  cac
    .command('check-circular')
    .option('--staged', 'Only check staged files')
    .option('--verbose', 'Show detailed information')
    .option('--threshold <number>', 'Threshold for circular dependencies', {
      default: 0,
    })
    .option('--ignore-dirs <dirs>', 'Directories to ignore, comma separated')
    .usage('Analyze project circular dependencies')
    .action(async ({ ignoreDirs, staged, threshold, verbose }) => {
      const config: CheckCircularConfig = {
        threshold: Number(threshold),
        ...(ignoreDirs && { ignoreDirs: ignoreDirs.split(',') }),
      };

      await checkCircular({
        config,
        staged,
        verbose: verbose ?? true,
      });
    });
}

export { type CheckCircularConfig, defineCheckCircularCommand };
