import type { DeepPartial } from '@vben-core/typings';

import type { Preferences, ThemePreferences } from './types';

import { generatorColorVariables } from '@vben-core/shared/color';

import { BUILT_IN_THEME_PRESETS } from './constants';

type ThemePatch = DeepPartial<ThemePreferences>;

interface ColorVariableItem {
  alias?: string;
  color: string;
  name: string;
}

const COLOR_VARIABLE_GROUPS = {
  colorDestructive: ['--destructive', '--red'],
  colorPrimary: ['--primary'],
  colorSuccess: ['--green', '--success'],
  colorWarning: ['--warning', '--yellow'],
} as const;

let themeTransitionCleanupFrame: number | undefined;
let themeTransitionFrame: number | undefined;

function hasOwnProperty<T extends object>(value: T, key: PropertyKey) {
  return Object.prototype.hasOwnProperty.call(value, key);
}

/**
 * 主题变量变化时暂停页面级颜色过渡，避免不同组件分帧换肤产生闪屏。
 */
function beginThemeTransition(root: HTMLElement) {
  if (themeTransitionFrame !== undefined) {
    window.cancelAnimationFrame(themeTransitionFrame);
  }
  if (themeTransitionCleanupFrame !== undefined) {
    window.cancelAnimationFrame(themeTransitionCleanupFrame);
  }

  root.dataset.themeTransition = 'updating';
  themeTransitionFrame = window.requestAnimationFrame(() => {
    themeTransitionFrame = undefined;
    themeTransitionCleanupFrame = window.requestAnimationFrame(() => {
      delete root.dataset.themeTransition;
      themeTransitionCleanupFrame = undefined;
    });
  });
}

/**
 * 更新主题的 CSS 变量以及其他 CSS 变量
 * @param preferences - 当前完整偏好设置
 * @param themePatch - 本次主题更新补丁，用于限制 DOM 更新范围
 */
function updateCSSVariables(
  preferences: Preferences,
  themePatch: ThemePatch = preferences.theme,
) {
  const root = document.documentElement;
  if (!root) {
    return;
  }

  const theme = preferences.theme;

  // html 设置 dark 类
  if (hasOwnProperty(themePatch, 'mode')) {
    beginThemeTransition(root);
    root.classList.toggle('dark', isDarkTheme(theme.mode));
  }

  // html 设置 data-theme=[builtinType]
  if (
    hasOwnProperty(themePatch, 'builtinType') &&
    root.dataset.theme !== theme.builtinType
  ) {
    root.dataset.theme = theme.builtinType;
  }

  clearEmptyColorVariables(themePatch, root);
  const colorItems = resolveColorItems(theme, themePatch);
  if (colorItems.length > 0) {
    updateMainColorVariables(colorItems, root);
  }

  // 更新圆角
  if (hasOwnProperty(themePatch, 'radius')) {
    root.style.setProperty('--radius', `${theme.radius}rem`);
  }

  // 更新字体大小
  if (hasOwnProperty(themePatch, 'fontSize')) {
    const fontSize = theme.fontSize;
    root.style.setProperty('--font-size-base', `${fontSize}px`);
    root.style.setProperty('--menu-font-size', `calc(${fontSize}px * 0.875)`);
  }
}

/**
 * 空颜色表示移除运行时覆盖，重新使用设计系统基础变量。
 */
function clearEmptyColorVariables(themePatch: ThemePatch, root: HTMLElement) {
  Object.entries(COLOR_VARIABLE_GROUPS).forEach(([field, prefixes]) => {
    const colorField = field as keyof typeof COLOR_VARIABLE_GROUPS;
    if (
      !hasOwnProperty(themePatch, colorField) ||
      themePatch[colorField] !== ''
    ) {
      return;
    }

    for (let index = root.style.length - 1; index >= 0; index--) {
      const variableName = root.style.item(index);
      if (
        prefixes.some(
          (prefix) =>
            variableName === prefix || variableName.startsWith(`${prefix}-`),
        )
      ) {
        root.style.removeProperty(variableName);
      }
    }
  });
}

/**
 * 根据本次补丁解析真正需要重新生成的颜色。
 */
function resolveColorItems(
  theme: ThemePreferences,
  themePatch: ThemePatch,
): ColorVariableItem[] {
  const colorItems: ColorVariableItem[] = [];
  const preset = BUILT_IN_THEME_PRESETS.find(
    (item) => item.type === theme.builtinType,
  );
  const lightPrimary = preset?.primaryColor || preset?.color;
  const darkPrimary = preset?.darkPrimaryColor || lightPrimary;
  const hasModeSpecificPrimary =
    !!lightPrimary && !!darkPrimary && lightPrimary !== darkPrimary;

  let primaryColor: string | undefined;
  if (
    hasModeSpecificPrimary &&
    (hasOwnProperty(themePatch, 'mode') ||
      hasOwnProperty(themePatch, 'builtinType'))
  ) {
    primaryColor = isDarkTheme(theme.mode) ? darkPrimary : lightPrimary;
  } else if (hasOwnProperty(themePatch, 'colorPrimary')) {
    primaryColor = theme.colorPrimary;
  } else if (hasOwnProperty(themePatch, 'builtinType')) {
    const presetPrimary = isDarkTheme(theme.mode) ? darkPrimary : lightPrimary;
    primaryColor =
      preset?.type === 'custom' ? theme.colorPrimary : presetPrimary;
  }

  if (primaryColor) {
    colorItems.push({ color: primaryColor, name: 'primary' });
  }
  if (hasOwnProperty(themePatch, 'colorWarning')) {
    colorItems.push({
      alias: 'warning',
      color: theme.colorWarning,
      name: 'yellow',
    });
  }
  if (hasOwnProperty(themePatch, 'colorSuccess')) {
    colorItems.push({
      alias: 'success',
      color: theme.colorSuccess,
      name: 'green',
    });
  }
  if (hasOwnProperty(themePatch, 'colorDestructive')) {
    colorItems.push({
      alias: 'destructive',
      color: theme.colorDestructive,
      name: 'red',
    });
  }

  return colorItems;
}

/**
 * 只把本次生成的颜色变量写入根节点，避免替换整张运行时样式表。
 */
function updateMainColorVariables(
  colorItems: ColorVariableItem[],
  root: HTMLElement,
) {
  const colorVariables = generatorColorVariables(colorItems);

  // 要设置的 CSS 变量映射
  const colorMappings: Record<string, string> = {
    '--green-500': '--success',
    '--primary-500': '--primary',
    '--red-500': '--destructive',
    '--yellow-500': '--warning',
  };

  Object.entries(colorVariables).forEach(([name, value]) => {
    root.style.setProperty(name, value);
  });

  Object.entries(colorMappings).forEach(([sourceVar, targetVar]) => {
    const colorValue = colorVariables[sourceVar];
    if (colorValue) {
      root.style.setProperty(targetVar, colorValue);
    }
  });
}

function isDarkTheme(theme: string) {
  let dark = theme === 'dark';
  if (theme === 'auto') {
    dark = window.matchMedia('(prefers-color-scheme: dark)').matches;
  }
  return dark;
}

export { isDarkTheme, updateCSSVariables };
