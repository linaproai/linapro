import { defineComponent, h } from 'vue';

import { Button } from 'ant-design-vue';

/**
 * 表格操作列按钮专用
 * 固定为 primary + ghost + small 样式
 */
export const GhostButton = defineComponent({
  name: 'GhostButton',
  inheritAttrs: false,
  setup(_props, { attrs, slots }) {
    return () =>
      h(
        Button,
        { ...attrs, type: 'primary', ghost: true, size: 'small' },
        slots,
      );
  },
});
