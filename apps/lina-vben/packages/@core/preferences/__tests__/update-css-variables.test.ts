import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { defaultPreferences } from '../src/config';
import { updateCSSVariables } from '../src/update-css-variables';

const { generatorColorVariables } = vi.hoisted(() => ({
  generatorColorVariables: vi.fn(() => ({})),
}));

vi.mock('@vben-core/shared/color', () => ({
  generatorColorVariables,
}));

function createPreferences() {
  return structuredClone(defaultPreferences);
}

describe('updateCSSVariables', () => {
  let animationFrames: FrameRequestCallback[];

  beforeEach(() => {
    animationFrames = [];
    vi.spyOn(window, 'requestAnimationFrame').mockImplementation((callback) => {
      animationFrames.push(callback);
      return animationFrames.length;
    });
    generatorColorVariables.mockClear();
    document.documentElement.className = '';
    delete document.documentElement.dataset.theme;
    delete document.documentElement.dataset.themeTransition;
    document.documentElement.removeAttribute('style');
  });

  afterEach(() => {
    while (animationFrames.length > 0) {
      animationFrames.shift()?.(0);
    }
    vi.restoreAllMocks();
  });

  it('only toggles the root class for a default mode update', () => {
    const preferences = createPreferences();
    preferences.theme.mode = 'light';
    const root = document.documentElement;
    root.style.setProperty('--primary-500', 'preserved-primary');
    root.style.setProperty('--radius', 'preserved-radius');
    root.style.setProperty('--font-size-base', 'preserved-font');

    updateCSSVariables(preferences, { mode: 'light' });

    expect(root.classList.contains('dark')).toBe(false);
    expect(generatorColorVariables).not.toHaveBeenCalled();
    expect(root.style.getPropertyValue('--primary-500')).toBe(
      'preserved-primary',
    );
    expect(root.style.getPropertyValue('--radius')).toBe('preserved-radius');
    expect(root.style.getPropertyValue('--font-size-base')).toBe(
      'preserved-font',
    );
    expect(root.dataset.themeTransition).toBe('updating');

    animationFrames.shift()?.(0);
    expect(root.dataset.themeTransition).toBe('updating');

    animationFrames.shift()?.(16);
    expect(root.dataset.themeTransition).toBeUndefined();
  });

  it('regenerates only the effective primary color for a mode-aware preset', () => {
    const preferences = createPreferences();
    preferences.theme.builtinType = 'zinc';
    preferences.theme.mode = 'dark';

    updateCSSVariables(preferences, { mode: 'dark' });

    expect(generatorColorVariables).toHaveBeenCalledOnce();
    expect(generatorColorVariables).toHaveBeenCalledWith([
      { color: 'hsl(0 0% 98%)', name: 'primary' },
    ]);
  });

  it('regenerates only color fields present in the patch', () => {
    const preferences = createPreferences();
    preferences.theme.colorSuccess = 'hsl(120 50% 50%)';

    updateCSSVariables(preferences, {
      colorSuccess: preferences.theme.colorSuccess,
    });

    expect(generatorColorVariables).toHaveBeenCalledOnce();
    expect(generatorColorVariables).toHaveBeenCalledWith([
      {
        alias: 'success',
        color: 'hsl(120 50% 50%)',
        name: 'green',
      },
    ]);
  });

  it('updates radius and font variables independently', () => {
    const preferences = createPreferences();
    preferences.theme.radius = '0.75';
    const root = document.documentElement;
    root.style.setProperty('--font-size-base', 'preserved-font');

    updateCSSVariables(preferences, { radius: '0.75' });

    expect(root.style.getPropertyValue('--radius')).toBe('0.75rem');
    expect(root.style.getPropertyValue('--font-size-base')).toBe(
      'preserved-font',
    );
    expect(generatorColorVariables).not.toHaveBeenCalled();
  });

  it('removes an inline palette when its explicit color is cleared', () => {
    const preferences = createPreferences();
    preferences.theme.colorPrimary = '';
    const root = document.documentElement;
    root.style.setProperty('--primary', 'old-primary');
    root.style.setProperty('--primary-50', 'old-primary-50');
    root.style.setProperty('--radius', 'preserved-radius');

    updateCSSVariables(preferences, { colorPrimary: '' });

    expect(root.style.getPropertyValue('--primary')).toBe('');
    expect(root.style.getPropertyValue('--primary-50')).toBe('');
    expect(root.style.getPropertyValue('--radius')).toBe('preserved-radius');
    expect(generatorColorVariables).not.toHaveBeenCalled();
  });
});
