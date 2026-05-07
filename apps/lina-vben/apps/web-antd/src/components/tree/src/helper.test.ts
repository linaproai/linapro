import { describe, expect, it } from 'vitest';

import { filterPersistedMenuIds } from './helper';

describe('filterPersistedMenuIds', () => {
  it('drops synthetic display nodes from submitted menu ids', () => {
    expect(filterPersistedMenuIds([-1000001, 1, '2', 0, '-3'])).toEqual([
      1,
      '2',
    ]);
  });
});
