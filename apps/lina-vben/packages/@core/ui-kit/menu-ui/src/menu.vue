<script setup lang="ts">
import type { MenuRecordRaw } from '@vben-core/typings';

import type { MenuProps } from './types';

import { computed } from 'vue';

import { useForwardProps } from '@vben-core/composables';

import { Menu } from './components';
import SubMenu from './sub-menu.vue';

interface Props extends MenuProps {
  menus: MenuRecordRaw[];
}

defineOptions({
  name: 'MenuView',
});

const props = withDefaults(defineProps<Props>(), {
  collapse: false,
});

const forward = useForwardProps(props);

function buildMenuStructureKey(menus: MenuRecordRaw[]): string {
  return menus
    .map((menu) => {
      return `${menu.path}:${buildMenuStructureKey(menu.children ?? [])}`;
    })
    .join('|');
}

const menuStructureKey = computed(() => buildMenuStructureKey(props.menus));
</script>

<template>
  <Menu :key="menuStructureKey" v-bind="forward">
    <template v-for="menu in menus" :key="menu.path">
      <SubMenu :menu="menu" />
    </template>
  </Menu>
</template>
