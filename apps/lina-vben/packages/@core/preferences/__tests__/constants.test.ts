import { describe, expect, it } from 'vitest';

import { DEFAULT_TIME_ZONE_OPTIONS } from '../src/constants';

describe('DEFAULT_TIME_ZONE_OPTIONS', () => {
  it('uses UTC labels for built-in timezone options', () => {
    expect(DEFAULT_TIME_ZONE_OPTIONS).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          label: 'Asia/Shanghai(UTC+8)',
          timezone: 'Asia/Shanghai',
        }),
        expect.objectContaining({
          label: 'Europe/London(UTC+0)',
          timezone: 'Europe/London',
        }),
      ]),
    );
    expect(
      DEFAULT_TIME_ZONE_OPTIONS.some((item) => item.label.includes('GMT')),
    ).toBe(false);
  });
});
