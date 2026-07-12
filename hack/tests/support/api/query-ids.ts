/**
 * Build query-string fragments for GoFrame []int query array parameters.
 *
 * GoFrame binds `ids[]=1&ids[]=2` to []int, but collapses plain repeated keys
 * (`ids=1&ids=2`) to the last value only.
 *
 * Related issue: https://github.com/linaproai/linapro/issues/89
 */

/** Encode positive integer IDs as `key[]=a&key[]=b`. */
export function buildBracketArrayQuery(
  key: string,
  ids: Array<number | string>,
): string {
  const params = new URLSearchParams();
  for (const id of ids) {
    params.append(`${key}[]`, String(id));
  }
  return params.toString();
}

/** Encode role/user batch-delete IDs as `ids[]=...`. */
export function buildBatchIdsQuery(ids: Array<number | string>): string {
  return buildBracketArrayQuery("ids", ids);
}

/**
 * Read `ids[]` (and legacy bare `ids` single values) from a request URL.
 * Used by E2E assertions that inspect the browser network call.
 */
export function getBracketArrayIds(
  requestUrl: string,
  key = "ids",
): number[] {
  const params = new URL(requestUrl).searchParams;
  const raw = [
    ...params.getAll(`${key}[]`),
    ...params.getAll(key),
  ];
  const ids: number[] = [];
  for (const value of raw) {
    const id = Number(value.trim());
    if (Number.isFinite(id) && id > 0) {
      ids.push(id);
    }
  }
  return ids.sort((a, b) => a - b);
}
