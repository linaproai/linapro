import { flushPromises, mount } from '@vue/test-utils';

import { afterEach, describe, expect, it, vi } from 'vitest';

import ThemeButton from './theme-button.vue';

describe('theme toggle button', () => {
  const originalStartViewTransition = Object.getOwnPropertyDescriptor(
    document,
    'startViewTransition',
  );
  const originalAnimate = Object.getOwnPropertyDescriptor(
    document.documentElement,
    'animate',
  );

  afterEach(() => {
    vi.restoreAllMocks();
    restoreProperty(
      document,
      'startViewTransition',
      originalStartViewTransition,
    );
    restoreProperty(document.documentElement, 'animate', originalAnimate);
    delete document.documentElement.dataset.themeTransitionDirection;
    document.documentElement.style.removeProperty('--theme-transition-x');
    document.documentElement.style.removeProperty('--theme-transition-y');
    document.documentElement.style.removeProperty('--theme-transition-radius');
  });

  it.each([
    [
      'light to dark',
      false,
      true,
      'to-dark',
      '::view-transition-old(root)',
      'ease-in',
    ],
    [
      'dark to light',
      true,
      false,
      'to-light',
      '::view-transition-new(root)',
      'ease-in',
    ],
  ])(
    'uses the required circular direction for %s',
    async (
      _name,
      initialDark,
      expectedDark,
      expectedDirection,
      expectedPseudoElement,
      expectedEasing,
    ) => {
      const never = new Promise<void>(() => {});
      const animate = vi.fn(() => ({ cancel: vi.fn(), finished: never }));
      let geometryAtTransitionStart:
        | undefined
        | {
            direction: string | undefined;
            radius: string;
            x: string;
            y: string;
          };
      const startViewTransition = vi.fn((updateCallback: () => unknown) => {
        const root = document.documentElement;
        geometryAtTransitionStart = {
          direction: root.dataset.themeTransitionDirection,
          radius: root.style.getPropertyValue('--theme-transition-radius'),
          x: root.style.getPropertyValue('--theme-transition-x'),
          y: root.style.getPropertyValue('--theme-transition-y'),
        };
        const updateCallbackDone = Promise.resolve().then(updateCallback);
        return {
          finished: never,
          ready: updateCallbackDone,
          updateCallbackDone,
        };
      });
      Object.defineProperty(document, 'startViewTransition', {
        configurable: true,
        value: startViewTransition,
      });
      Object.defineProperty(document.documentElement, 'animate', {
        configurable: true,
        value: animate,
      });

      const wrapper = mount(ThemeButton, {
        props: {
          modelValue: initialDark,
        },
      });

      try {
        const button = wrapper.get('button');
        vi.spyOn(button.element, 'getBoundingClientRect').mockReturnValue({
          bottom: 70,
          height: 40,
          left: 80,
          right: 120,
          toJSON: () => ({}),
          top: 30,
          width: 40,
          x: 80,
          y: 30,
        });
        expect(button.classes()).toContain(
          expectedDark ? 'is-dark' : 'is-light',
        );
        expect(button.find('svg').exists()).toBeTruthy();

        await button.trigger('click', {
          clientX: 5,
          clientY: 5,
        });
        await flushPromises();

        expect(wrapper.emitted('update:modelValue')).toEqual([[expectedDark]]);
        expect(wrapper.emitted('themeChangeIntent')).toEqual([[expectedDark]]);
        expect(startViewTransition).toHaveBeenCalledOnce();
        expect(animate).toHaveBeenCalledOnce();

        const animateCall = animate.mock.calls[0];
        if (!animateCall) {
          throw new Error('Circular reveal animation was not started');
        }
        const [keyframes, options] = animateCall;
        const endRadius = Math.hypot(
          Math.max(100, window.innerWidth - 100),
          Math.max(50, window.innerHeight - 50),
        );
        const fullCircle = `circle(${endRadius}px at 100px 50px)`;
        expect(keyframes).toEqual({
          clipPath: expectedDark
            ? [fullCircle, 'circle(0px at 100px 50px)']
            : ['circle(0px at 100px 50px)', fullCircle],
        });
        expect(options).toEqual({
          duration: 450,
          easing: expectedEasing,
          fill: 'both',
          pseudoElement: expectedPseudoElement,
        });
        expect(geometryAtTransitionStart).toEqual({
          direction: expectedDirection,
          radius: `${endRadius}px`,
          x: '100px',
          y: '50px',
        });
      } finally {
        wrapper.unmount();
      }
    },
  );

  it('cleans the animation only after the transition tree is removed', async () => {
    const animationFinished = createDeferred();
    const transitionFinished = createDeferred();
    const cancel = vi.fn();
    const animate = vi.fn(() => ({
      cancel,
      finished: animationFinished.promise,
    }));
    const startViewTransition = vi.fn((updateCallback: () => unknown) => {
      const updateCallbackDone = Promise.resolve().then(updateCallback);
      return {
        finished: transitionFinished.promise,
        ready: updateCallbackDone,
        updateCallbackDone,
      };
    });
    Object.defineProperty(document, 'startViewTransition', {
      configurable: true,
      value: startViewTransition,
    });
    Object.defineProperty(document.documentElement, 'animate', {
      configurable: true,
      value: animate,
    });

    const wrapper = mount(ThemeButton, {
      props: {
        modelValue: false,
      },
    });

    try {
      const button = wrapper.get('button');
      vi.spyOn(button.element, 'getBoundingClientRect').mockReturnValue({
        bottom: 70,
        height: 40,
        left: 80,
        right: 120,
        toJSON: () => ({}),
        top: 30,
        width: 40,
        x: 80,
        y: 30,
      });

      await button.trigger('click');
      await flushPromises();

      expect(document.documentElement.dataset.themeTransitionDirection).toBe(
        'to-dark',
      );
      expect(cancel).not.toHaveBeenCalled();

      animationFinished.resolve();
      await flushPromises();

      expect(cancel).not.toHaveBeenCalled();

      transitionFinished.resolve();
      await flushPromises();

      expect(cancel).toHaveBeenCalledOnce();
      expect(
        document.documentElement.dataset.themeTransitionDirection,
      ).toBeUndefined();
      expect(
        document.documentElement.style.getPropertyValue('--theme-transition-x'),
      ).toBe('');
      expect(
        document.documentElement.style.getPropertyValue('--theme-transition-y'),
      ).toBe('');
      expect(
        document.documentElement.style.getPropertyValue(
          '--theme-transition-radius',
        ),
      ).toBe('');
    } finally {
      wrapper.unmount();
    }
  });

  it('switches directly when reduced motion is requested', async () => {
    const startViewTransition = vi.fn();
    Object.defineProperty(document, 'startViewTransition', {
      configurable: true,
      value: startViewTransition,
    });
    vi.spyOn(window, 'matchMedia').mockReturnValue({
      matches: true,
    } as MediaQueryList);

    const wrapper = mount(ThemeButton, {
      props: {
        modelValue: false,
      },
    });

    try {
      await wrapper.get('button').trigger('click');

      expect(wrapper.emitted('update:modelValue')).toEqual([[true]]);
      expect(startViewTransition).not.toHaveBeenCalled();
    } finally {
      wrapper.unmount();
    }
  });

  it('keeps the last requested theme when clicks overlap', async () => {
    const updateCallbacks: Array<() => unknown> = [];
    const animate = vi.fn(() => ({ finished: Promise.resolve() }));
    const startViewTransition = vi.fn((updateCallback: () => unknown) => {
      updateCallbacks.push(updateCallback);
      return {
        finished: Promise.resolve(),
        ready: new Promise<void>(() => {}),
      };
    });
    Object.defineProperty(document, 'startViewTransition', {
      configurable: true,
      value: startViewTransition,
    });
    Object.defineProperty(document.documentElement, 'animate', {
      configurable: true,
      value: animate,
    });

    const wrapper = mount(ThemeButton, {
      props: {
        modelValue: false,
      },
    });

    try {
      await wrapper.get('button').trigger('click');
      await wrapper.get('button').trigger('click');

      expect(wrapper.emitted('themeChangeIntent')).toEqual([[true], [false]]);
      expect(startViewTransition).toHaveBeenCalledTimes(2);

      await updateCallbacks[1]?.();
      await updateCallbacks[0]?.();
      await flushPromises();

      expect(wrapper.emitted('update:modelValue')).toBeUndefined();
      expect(animate).not.toHaveBeenCalled();
    } finally {
      wrapper.unmount();
    }
  });
});

function restoreProperty(
  target: object,
  key: PropertyKey,
  descriptor: PropertyDescriptor | undefined,
) {
  if (descriptor) {
    Object.defineProperty(target, key, descriptor);
  } else {
    Reflect.deleteProperty(target, key);
  }
}

function createDeferred() {
  let resolve!: () => void;
  const promise = new Promise<void>((resolvePromise) => {
    resolve = () => resolvePromise();
  });
  return { promise, resolve };
}
