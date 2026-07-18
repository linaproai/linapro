import { describe, expect, it } from 'vitest';

import {
  SAFE_FALLBACK_HOME_PATH,
  findFirstNavigableMenuPath,
  resolvePostLoginLandingPath,
} from './post-login-landing';

describe('post-login landing path', () => {
  it('prefers explicit redirect when routes are not ready yet', () => {
    expect(
      resolvePostLoginLandingPath({
        preferredPaths: ['/system/user', '/dashboard/analytics'],
      }),
    ).toBe('/system/user');
  });

  it('falls back to homePath without using hardcoded dashboard default', () => {
    expect(
      resolvePostLoginLandingPath({
        preferredPaths: ['', null, '/system/role'],
      }),
    ).toBe('/system/role');
  });

  it('when workbench is missing, lands on first accessible menu', () => {
    const menus = [
      {
        path: '/dashboard',
        children: [
          { path: '/dashboard/analytics', disabled: true },
          { path: '/dashboard/workspace', show: false },
        ],
      },
      {
        path: '/system',
        children: [
          { path: '/system/user' },
          { path: '/system/role' },
        ],
      },
    ];

    expect(
      resolvePostLoginLandingPath({
        preferredPaths: ['/dashboard/analytics'],
        accessibleMenus: menus,
        accessibleRoutes: [
          {
            path: '/system',
            children: [
              { path: 'user' },
              { path: 'role' },
            ],
          },
        ],
      }),
    ).toBe('/system/user');
  });

  it('drops inaccessible preferred paths after routes are assembled', () => {
    expect(
      resolvePostLoginLandingPath({
        preferredPaths: [
          '/dashboard/analytics',
          '/gone',
          '/system/user',
        ],
        accessibleRoutes: [
          {
            path: '/system',
            children: [{ path: 'user' }],
          },
        ],
      }),
    ).toBe('/system/user');
  });

  it('uses safe fallback when no menus or routes exist', () => {
    expect(
      resolvePostLoginLandingPath({
        preferredPaths: ['/dashboard/analytics'],
        accessibleMenus: [],
        accessibleRoutes: [],
      }),
    ).toBe(SAFE_FALLBACK_HOME_PATH);
  });

  it('finds first navigable leaf menu in sidebar order', () => {
    expect(
      findFirstNavigableMenuPath([
        {
          path: '/dashboard',
          children: [{ path: '/dashboard/analytics', show: false }],
        },
        {
          path: '/content',
          children: [{ path: '/content/notice' }],
        },
      ]),
    ).toBe('/content/notice');
  });
});
