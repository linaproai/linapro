<script setup lang="ts">
import type { PluginSlotKey } from '#/plugins/plugin-slots';
import type { RegisteredPluginSlotModule } from '#/plugins/slot-registry';

import {
  shallowRef,
  computed,
  onBeforeUnmount,
  onMounted,
  useAttrs,
  watch,
} from 'vue';

import {
  getPluginSlots,
  getPluginStateMap,
  onPluginRegistryChanged,
} from '#/plugins/slot-registry';

defineOptions({
  name: 'PluginSlotOutlet',
  inheritAttrs: false,
});

const props = defineProps<{
  slotKey: PluginSlotKey;
}>();

const attrs = useAttrs();
const items = shallowRef<RegisteredPluginSlotModule[]>([]);

let disposeListener: (() => void) | null = null;

const shouldRender = computed(() => items.value.length > 0);

function isEnabled(value: unknown) {
  return value === 1 || value === '1' || value === true;
}

async function refresh(force = false) {
  try {
    const slotDefinitions = getPluginSlots(props.slotKey);
    if (slotDefinitions.length === 0) {
      items.value = [];
      return;
    }

    const pluginStateMap = await getPluginStateMap(force);
    items.value = slotDefinitions.filter((item) => {
      const pluginState = pluginStateMap.get(item.pluginId);
      return (
        isEnabled(pluginState?.installed) && isEnabled(pluginState?.enabled)
      );
    });
  } catch (error) {
    console.error(
      `[plugin-slot] failed to refresh outlet for slot ${props.slotKey}`,
      error,
    );
    items.value = [];
  }
}

onMounted(() => {
  void refresh();
  disposeListener = onPluginRegistryChanged(() => refresh(true));
});

onBeforeUnmount(() => {
  disposeListener?.();
});

watch(
  () => props.slotKey,
  () => {
    void refresh(true);
  },
);
</script>

<template>
  <div
    v-if="shouldRender"
    v-bind="attrs"
    :data-plugin-slot-key="props.slotKey"
    class="plugin-slot-outlet"
  >
    <template v-for="item in items" :key="item.key">
      <div :data-plugin-slot-item="item.key" class="plugin-slot-outlet__item">
        <component :is="item.component" />
      </div>
    </template>
  </div>
</template>
