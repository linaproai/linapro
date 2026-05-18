import type { Buffer } from 'node:buffer';
import type { PathOrFileDescriptor, WriteFileOptions } from 'node:fs';

declare module 'node:fs' {
  // Node accepts Buffer for writeFileSync at runtime; this keeps the E2E suite
  // type-stable on newer Node runtimes whose Buffer type may expose ArrayBufferLike.
  export function writeFileSync(
    file: PathOrFileDescriptor,
    data: string | Buffer,
    options?: WriteFileOptions,
  ): void;
}
