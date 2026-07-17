import { mount } from '@vue/test-utils';

import { describe, expect, it } from 'vitest';

import BuiltinTheme from './builtin.vue';

describe('preference builtin theme', () => {
  it('does not emit a second primary-color update when only mode changes', async () => {
    const wrapper = mount(BuiltinTheme, {
      props: {
        isDark: false,
        modelValue: 'zinc',
        themeColorPrimary: 'hsl(240 5.9% 10%)',
      },
    });

    await wrapper.setProps({ isDark: true });

    expect(wrapper.emitted('update:themeColorPrimary')).toBeUndefined();
  });

  it('keeps the current primary color when custom mode is selected', async () => {
    const wrapper = mount(BuiltinTheme, {
      props: {
        isDark: false,
        modelValue: 'zinc',
        themeColorPrimary: 'hsl(240 5.9% 10%)',
      },
    });

    await wrapper.setProps({ modelValue: 'custom' });

    expect(wrapper.emitted('update:themeColorPrimary')).toBeUndefined();
  });
});
