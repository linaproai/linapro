<script lang="ts" setup>
import { computed, nextTick } from 'vue';

import { VbenButton } from '@vben-core/shadcn-ui';

interface Props {
  /**
   * 类型
   */
  type?: 'icon' | 'normal';
}

interface ThemeViewTransition {
  finished: Promise<void>;
  ready: Promise<void>;
}

type ThemeTransitionDocument = Document & {
  startViewTransition?: (
    updateCallback: () => Promise<void>,
  ) => ThemeViewTransition;
};

defineOptions({
  name: 'ThemeToggleButton',
});

const props = withDefaults(defineProps<Props>(), {
  type: 'normal',
});

const emit = defineEmits<{
  themeChangeIntent: [isDark: boolean];
}>();

const isDark = defineModel<boolean>();

const targetTheme = computed(() => {
  return isDark.value ? 'light' : 'dark';
});

let latestThemeChange = 0;
let pendingIsDark: boolean | undefined;

const THEME_TRANSITION_RADIUS = '--theme-transition-radius';
const THEME_TRANSITION_X = '--theme-transition-x';
const THEME_TRANSITION_Y = '--theme-transition-y';

const bindProps = computed(() => {
  const type = props.type;

  return type === 'normal'
    ? {
        variant: 'heavy' as const,
      }
    : {
        class: 'rounded-full',
        size: 'icon' as const,
        style: { padding: '7px' },
        variant: 'icon' as const,
      };
});

function toggleTheme(event: MouseEvent) {
  const targetIsDark = !(pendingIsDark ?? isDark.value);
  const themeChange = ++latestThemeChange;
  pendingIsDark = targetIsDark;
  emit('themeChangeIntent', targetIsDark);
  const startViewTransition = (document as ThemeTransitionDocument)
    .startViewTransition;
  const prefersReducedMotion = window.matchMedia(
    '(prefers-reduced-motion: reduce)',
  ).matches;

  if (!startViewTransition || prefersReducedMotion) {
    isDark.value = targetIsDark;
    pendingIsDark = undefined;
    return;
  }

  const buttonRect = (
    event.currentTarget as HTMLElement
  ).getBoundingClientRect();
  const x = buttonRect.left + buttonRect.width / 2;
  const y = buttonRect.top + buttonRect.height / 2;
  const endRadius = Math.hypot(
    Math.max(x, window.innerWidth - x),
    Math.max(y, window.innerHeight - y),
  );
  const root = document.documentElement;
  prepareThemeTransitionGeometry(root, targetIsDark, x, y, endRadius);

  let transition: ThemeViewTransition;
  try {
    transition = startViewTransition.call(document, async () => {
      if (themeChange !== latestThemeChange) {
        return;
      }
      isDark.value = targetIsDark;
      await nextTick();
      if (themeChange === latestThemeChange) {
        pendingIsDark = undefined;
      }
    });
  } catch (error) {
    pendingIsDark = undefined;
    clearThemeTransitionGeometry(root, themeChange);
    throw error;
  }

  let animation: Animation | undefined;

  const finishTransition = () => {
    animation?.cancel();
    clearThemeTransitionGeometry(root, themeChange);
  };

  void transition.finished.then(finishTransition, finishTransition);

  void transition.ready.then(
    () => {
      if (themeChange !== latestThemeChange) {
        return;
      }
      animation = root.animate(
        {
          clipPath: targetIsDark
            ? [
                `circle(${endRadius}px at ${x}px ${y}px)`,
                `circle(0px at ${x}px ${y}px)`,
              ]
            : [
                `circle(0px at ${x}px ${y}px)`,
                `circle(${endRadius}px at ${x}px ${y}px)`,
              ],
        },
        {
          duration: 450,
          easing: 'ease-in',
          // CSS 负责伪元素创建前的首帧，both 负责动画接管后的首尾连续性。
          fill: 'both',
          pseudoElement: targetIsDark
            ? '::view-transition-old(root)'
            : '::view-transition-new(root)',
        },
      );
    },
    () => undefined,
  );
}

function prepareThemeTransitionGeometry(
  root: HTMLElement,
  targetIsDark: boolean,
  x: number,
  y: number,
  radius: number,
) {
  root.dataset.themeTransitionDirection = targetIsDark ? 'to-dark' : 'to-light';
  root.style.setProperty(THEME_TRANSITION_X, `${x}px`);
  root.style.setProperty(THEME_TRANSITION_Y, `${y}px`);
  root.style.setProperty(THEME_TRANSITION_RADIUS, `${radius}px`);
}

function clearThemeTransitionGeometry(root: HTMLElement, themeChange: number) {
  if (themeChange !== latestThemeChange) {
    return;
  }
  delete root.dataset.themeTransitionDirection;
  root.style.removeProperty(THEME_TRANSITION_X);
  root.style.removeProperty(THEME_TRANSITION_Y);
  root.style.removeProperty(THEME_TRANSITION_RADIUS);
}
</script>

<template>
  <VbenButton
    :aria-label="targetTheme"
    :class="[`is-${targetTheme}`]"
    aria-live="polite"
    class="theme-toggle cursor-pointer border-none bg-none motion-safe:hover:animate-[shrink_0.3s_ease-in-out]"
    v-bind="bindProps"
    @click.stop="toggleTheme"
  >
    <svg aria-hidden="true" height="24" viewBox="0 0 24 24" width="24">
      <mask id="theme-toggle-moon" class="theme-toggle__moon">
        <rect fill="white" height="100%" width="100%" x="0" y="0" />
        <circle cx="40" cy="8" fill="black" r="11" />
      </mask>
      <circle
        id="sun"
        class="theme-toggle__sun fill-foreground/90"
        cx="12"
        cy="12"
        mask="url(#theme-toggle-moon)"
        r="11"
      />
      <g class="theme-toggle__sun-beams stroke-foreground/90 stroke-2">
        <line x1="12" x2="12" y1="1" y2="3" />
        <line x1="12" x2="12" y1="21" y2="23" />
        <line x1="4.22" x2="5.64" y1="4.22" y2="5.64" />
        <line x1="18.36" x2="19.78" y1="18.36" y2="19.78" />
        <line x1="1" x2="3" y1="12" y2="12" />
        <line x1="21" x2="23" y1="12" y2="12" />
        <line x1="4.22" x2="5.64" y1="19.78" y2="18.36" />
        <line x1="18.36" x2="19.78" y1="5.64" y2="4.22" />
      </g>
    </svg>
  </VbenButton>
</template>

<style scoped>
@reference "@vben-core/design/theme";

.theme-toggle__moon > circle {
  transition: transform 0.5s cubic-bezier(0, 0, 0.3, 1);
}

.theme-toggle__sun {
  stroke: none;
  transform-origin: center center;
  transition: transform 1.6s cubic-bezier(0.25, 0, 0.2, 1);
}

.theme-toggle__sun-beams {
  transform-origin: center center;
  transition:
    transform 1.6s cubic-bezier(0.5, 1.5, 0.75, 1.25),
    opacity 0.6s cubic-bezier(0.25, 0, 0.3, 1);
}

.theme-toggle.is-light .theme-toggle__sun {
  @apply scale-50;
}

.theme-toggle.is-light .theme-toggle__sun-beams {
  transform: rotateZ(0.25turn);
}

.theme-toggle.is-dark .theme-toggle__moon > circle {
  transform: translateX(-20px);
}

.theme-toggle.is-dark .theme-toggle__sun-beams {
  @apply opacity-0;
}

.theme-toggle:hover > svg .theme-toggle__sun {
  fill: hsl(var(--foreground));
}

.theme-toggle:hover > svg .theme-toggle__sun-beams {
  stroke: hsl(var(--foreground));
}

@media (prefers-reduced-motion: reduce) {
  .theme-toggle__moon > circle,
  .theme-toggle__sun,
  .theme-toggle__sun-beams {
    transition: none;
  }
}
</style>
