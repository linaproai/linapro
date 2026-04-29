<script setup lang="ts">
import type { WorkbenchProjectItem } from '../typing';

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  VbenIcon,
} from '@vben-core/shadcn-ui';

interface Props {
  items?: WorkbenchProjectItem[];
  title: string;
}

defineOptions({
  name: 'WorkbenchProject',
});

withDefaults(defineProps<Props>(), {
  items: () => [],
});

defineEmits(['click']);
</script>

<template>
  <Card>
    <CardHeader class="py-4">
      <CardTitle class="text-lg">{{ title }}</CardTitle>
    </CardHeader>
    <CardContent class="flex flex-wrap p-0">
      <template v-for="(item, index) in items" :key="item.title">
        <div
          :class="{
            'border-r-0': index % 3 === 2,
            'border-b-0': index < 3,
            'pb-4': index > 2,
            'rounded-bl-xl': index === items.length - 3,
            'rounded-br-xl': index === items.length - 1,
          }"
          class="group w-full cursor-pointer border-t border-r border-border p-4 transition-all hover:shadow-xl md:w-1/2 lg:w-1/3"
        >
          <div class="flex items-center">
            <img
              v-if="item.logo"
              :alt="item.title"
              class="size-8 shrink-0 rounded-sm object-contain transition-all duration-300 group-hover:scale-110"
              :src="item.logo"
              @click="$emit('click', item)"
            />
            <VbenIcon
              v-else-if="item.icon"
              :color="item.color"
              :icon="item.icon"
              class="size-8 shrink-0 transition-all duration-300 group-hover:scale-110"
              @click="$emit('click', item)"
            />
            <span class="ml-4 text-lg font-medium">{{ item.title }}</span>
          </div>
          <div class="mt-4 h-10 min-w-0 text-foreground/80">
            <span class="block truncate">{{ item.content }}</span>
          </div>
          <div class="flex justify-between text-foreground/80">
            <span>{{ item.group }}</span>
            <span>{{ item.date }}</span>
          </div>
        </div>
      </template>
    </CardContent>
  </Card>
</template>
