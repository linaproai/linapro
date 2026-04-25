/* eslint-disable @typescript-eslint/no-non-null-assertion */
import type { TourProps } from 'ant-design-vue';

import { defineComponent, ref } from 'vue';

import { useLocalStorage } from '@vueuse/core';
import { Tour } from 'ant-design-vue';

import { $t } from '#/locales';

/**
 * 全屏引导
 * @returns value
 */
export function useFullScreenGuide() {
  const open = ref(false);
  /**
   * 是否已读 只显示一次
   */
  const read = useLocalStorage('menu_select_fullscreen_read', false);

  function openGuide() {
    if (!read.value) {
      open.value = true;
    }
  }

  function closeGuide() {
    open.value = false;
    read.value = true;
  }

  const steps: TourProps['steps'] = [
    {
      title: $t('pages.common.confirmTitle'),
      description: $t('pages.tree.messages.fullscreenHint'),
      target: () =>
        document.querySelector('div#menu-select-table .vxe-tools--operate > button')!,
    },
  ];

  const FullScreenGuide = defineComponent({
    name: 'FullScreenGuide',
    inheritAttrs: false,
    setup() {
      return () => (
        <Tour
          onClose={closeGuide}
          open={open.value}
          steps={steps}
          zIndex={9999}
        />
      );
    },
  });

  return {
    FullScreenGuide,
    openGuide,
    closeGuide,
  };
}
