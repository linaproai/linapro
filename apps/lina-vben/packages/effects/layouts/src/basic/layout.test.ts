/* eslint-disable vue/one-component-per-file -- 测试展示外壳用于隔离无关 UI，不属于生产组件。 */
import type { PropType } from 'vue';

import type { MenuRecordRaw } from '@vben/types';

import { shallowMount } from '@vue/test-utils';
import { createApp, defineComponent, h, nextTick } from 'vue';
import { createMemoryHistory, createRouter } from 'vue-router';

import { i18n } from '@vben/locales';
import { preferences, updatePreferences } from '@vben/preferences';
import { initStores, useAccessStore, useTabbarStore } from '@vben/stores';
import { cloneDeep } from '@vben/utils';

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import BasicLayout from './layout.vue';
import { LayoutExtraMenu, LayoutMenu, LayoutMixedMenu } from './menu';

// 只替换布局展示外壳，保留真实菜单来源、偏好状态和国际化实现。
const LayoutShell = defineComponent({
  props: { sidebarHidden: Boolean },
  emits: ['toggleSidebar'],
  setup(_, { slots }) {
    return () =>
      h('div', [
        slots.header?.(),
        slots.menu?.(),
        slots['mixed-menu']?.(),
        slots['side-extra']?.(),
      ]);
  },
});

const HeaderShell = defineComponent({
  setup(_, { slots }) {
    return () => h('header', slots.menu?.());
  },
});

function sourceMenus(): MenuRecordRaw[] {
  return [
    {
      children: [
        {
          i18nKey: 'menuTest.role',
          name: 'menuTest.originalRole',
          parent: '/system',
          parents: ['/system'],
          path: '/system/role',
        },
      ],
      i18nKey: 'menuTest.system',
      name: 'menuTest.originalSystem',
      path: '/system',
    },
  ];
}

async function mountLayout() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ component: { render: () => null }, path: '/system/role' }],
  });
  await router.push('/system/role');
  const wrapper = shallowMount(BasicLayout, {
    global: {
      plugins: [router],
      stubs: {
        LayoutHeader: HeaderShell,
        Menu: defineComponent({
          name: 'MenuProbe',
          props: {
            menus: { required: true, type: Array as PropType<MenuRecordRaw[]> },
          },
          render: () => null,
        }),
        VbenLayout: LayoutShell,
      },
    },
  });
  const menus = () => [
    ...wrapper.findAllComponents(LayoutMenu).map((menu) => menu.props('menus')),
    wrapper.getComponent(LayoutMixedMenu).props('menus'),
    wrapper.getComponent(LayoutExtraMenu).props('menus'),
  ];
  return { menus, wrapper };
}

describe('布局菜单投影', () => {
  let mounted: Awaited<ReturnType<typeof mountLayout>> | undefined;
  let restoreState: () => void;
  let namespaceIndex = 0;

  beforeEach(async () => {
    const originalPreferences = cloneDeep({
      app: {
        isMobile: preferences.app.isMobile,
        layout: preferences.app.layout,
      },
      navigation: { split: preferences.navigation.split },
      sidebar: {
        collapsed: preferences.sidebar.collapsed,
        hidden: preferences.sidebar.hidden,
      },
    });
    const originalLocale = i18n.global.locale.value;
    const originalChinese = cloneDeep(i18n.global.getLocaleMessage('zh-CN'));
    const originalEnglish = cloneDeep(i18n.global.getLocaleMessage('en-US'));
    restoreState = () => {
      updatePreferences(originalPreferences);
      i18n.global.setLocaleMessage('zh-CN', originalChinese);
      i18n.global.setLocaleMessage('en-US', originalEnglish);
      i18n.global.locale.value = originalLocale;
    };
    await initStores(createApp({}), {
      namespace: `layout-menu-projection-${namespaceIndex++}`,
    });
    updatePreferences({
      app: { isMobile: false, layout: 'header-nav' },
      navigation: { split: false },
      sidebar: { collapsed: false, hidden: false },
    });
    i18n.global.setLocaleMessage('zh-CN', {
      menuTest: {
        originalRole: '原始角色',
        role: '角色管理',
        system: '系统管理',
      },
    });
    i18n.global.setLocaleMessage('en-US', {
      menuTest: {
        originalRole: 'Original role',
        role: 'Roles',
        system: 'System',
      },
    });
    i18n.global.locale.value = 'zh-CN';
    useAccessStore().setAccessMenus(sourceMenus());
    // 页面刷新不是此测试目标，避免产生无关的进度条和延时任务。
    vi.spyOn(useTabbarStore(), 'refresh').mockResolvedValue(undefined);
    mounted = await mountLayout();
  });

  afterEach(() => {
    mounted?.wrapper.unmount();
    mounted = undefined;
    vi.restoreAllMocks();
    restoreState();
  });

  function current() {
    if (!mounted) throw new Error('布局尚未挂载');
    return mounted;
  }

  it('隐藏、展开和折叠侧栏时复用四类菜单的投影结果', async () => {
    const { menus, wrapper } = current();
    let previous = menus();
    let replacements = [0, 0, 0, 0];
    expect(previous).toHaveLength(4);
    const recordReplacements = () => {
      const next = menus();
      replacements = replacements.map(
        (count, index) => count + Number(next[index] !== previous[index]),
      );
      previous = next;
    };
    for (let index = 0; index < 20; index += 1) {
      wrapper.getComponent(LayoutShell).vm.$emit('toggleSidebar');
      await nextTick();
      expect(preferences.sidebar.hidden).toBe(index % 2 === 0);
      recordReplacements();
    }
    for (const collapsed of [true, false]) {
      updatePreferences({ sidebar: { collapsed } });
      await nextTick();
      recordReplacements();
    }
    // 展示外壳同时消费四类菜单；这里衡量投影重建，不代表浏览器帧率。
    console.info('[菜单投影重建次数]', replacements);
    expect(replacements).toEqual([0, 0, 0, 0]);
  });

  it('保留首层翻译语义且不修改原始菜单树', () => {
    const [header, sidebar, mixed, extra] = current().menus();
    expect(header?.[0]?.name).toBe('系统管理');
    expect(sidebar?.[0]?.children?.[0]?.name).toBe('角色管理');
    expect(mixed?.[0]?.name).toBe('系统管理');
    expect(mixed?.[0]?.children?.[0]?.name).toBe('menuTest.originalRole');
    expect(extra?.[0]?.name).toBe('角色管理');
    expect(useAccessStore().accessMenus).toEqual(sourceMenus());
  });

  it('权限菜单整体替换及节点原地更新后重新投影', async () => {
    useAccessStore().setAccessMenus([
      ...sourceMenus(),
      { name: 'menuTest.role', path: '/new' },
    ]);
    await nextTick();
    expect(
      current()
        .menus()[1]
        ?.map((menu) => menu.name),
    ).toEqual(['系统管理', '角色管理']);
    const child = useAccessStore().accessMenus[0]?.children?.[0];
    if (!child) throw new Error('测试子菜单不存在');
    child.i18nKey = 'menuTest.system';
    await nextTick();
    expect(current().menus()[1]?.[0]?.children?.[0]?.name).toBe('系统管理');
    expect(current().menus()[3]?.[0]?.name).toBe('系统管理');
  });

  it('语言切换后更新所有菜单名称', async () => {
    i18n.global.locale.value = 'en-US';
    await nextTick();
    const [header, sidebar, mixed, extra] = current().menus();
    expect(header?.[0]?.name).toBe('System');
    expect(sidebar?.[0]?.children?.[0]?.name).toBe('Roles');
    expect(mixed?.[0]?.name).toBe('System');
    expect(extra?.[0]?.name).toBe('Roles');
  });

  it('同语言资源新增、合并和替换后立即更新菜单名称', async () => {
    const child = useAccessStore().accessMenus[0]?.children?.[0];
    if (!child) throw new Error('测试子菜单不存在');
    child.i18nKey = 'menuTest.dynamicRole';
    await nextTick();
    expect(current().menus()[3]?.[0]?.name).toBe('原始角色');
    i18n.global.mergeLocaleMessage('zh-CN', {
      menuTest: { dynamicRole: '动态角色', system: '新系统' },
    });
    await nextTick();
    expect(
      current()
        .menus()
        .map((menu) => menu?.[0]?.name),
    ).toEqual(['新系统', '新系统', '新系统', '动态角色']);
    i18n.global.setLocaleMessage('zh-CN', {
      menuTest: { dynamicRole: '替换角色', system: '替换系统' },
    });
    await nextTick();
    expect(
      current()
        .menus()
        .map((menu) => menu?.[0]?.name),
    ).toEqual(['替换系统', '替换系统', '替换系统', '替换角色']);
  });
});
