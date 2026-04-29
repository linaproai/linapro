import { describe, expect, it, vi } from 'vitest';

import { buildJobHandlerSchemaFields } from './form';

vi.mock('@vben/locales', () => ({
  $t: (key: string, params?: Record<string, string>) =>
    params
      ? `${key}:${Object.entries(params)
          .map(([paramKey, value]) => `${paramKey}=${value}`)
          .join(',')}`
      : key,
}));

describe('buildJobHandlerSchemaFields', () => {
  it('accepts schema property names under properties', () => {
    const fields = buildJobHandlerSchemaFields(
      JSON.stringify({
        properties: {
          seconds: {
            description: '等待秒数',
            type: 'integer',
          },
        },
        required: ['seconds'],
        type: 'object',
      }),
    );

    expect(fields).toEqual([
      expect.objectContaining({
        component: 'InputNumber',
        fieldName: 'seconds',
        label: '等待秒数',
        required: true,
      }),
    ]);
  });

  it('rejects unsupported keywords inside property schema', () => {
    expect(() =>
      buildJobHandlerSchemaFields(
        JSON.stringify({
          properties: {
            seconds: {
              minimum: 1,
              type: 'integer',
            },
          },
          type: 'object',
        }),
      ),
    ).toThrow('key=minimum');
  });
});
