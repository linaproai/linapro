<script setup lang="ts">
import type { HostServiceCardView } from './plugin-host-service-view';

import { Tag } from 'ant-design-vue';

interface Props {
  cards: HostServiceCardView[];
}

defineProps<Props>();
</script>

<template>
  <div class="flex flex-col gap-4">
    <div
      v-for="card in cards"
      :key="card.service"
      class="rounded-md border border-[var(--ant-color-border)] p-4"
    >
      <div class="mb-3 flex flex-wrap items-center gap-2">
        <span class="text-[15px] font-medium">
          {{ card.title }}
        </span>
        <Tag color="blue">{{ card.service }}</Tag>
      </div>

      <div class="flex flex-col gap-3">
        <div
          v-for="scope in card.scopes"
          :key="scope.key"
          class="flex flex-col gap-2"
        >
          <div class="flex flex-wrap items-center gap-2">
            <Tag
              :data-testid="`plugin-host-service-scope-label-${card.service}-${scope.key}`"
              :color="scope.badgeColor"
            >
              {{ scope.label }}
            </Tag>
            <Tag v-for="method in scope.methods" :key="`${scope.key}-${method}`">
              {{ method }}
            </Tag>
            <span
              v-if="scope.hint"
              class="text-[12px] text-[var(--ant-color-text-secondary)]"
            >
              {{ scope.hint }}
            </span>
          </div>

          <div
            v-if="scope.targets.length > 0"
            class="flex flex-wrap items-center gap-2"
          >
            <Tag
              :data-testid="`plugin-host-service-summary-label-${card.service}-${scope.key}`"
              :color="scope.targetSummaryBadgeColor"
            >
              {{ scope.targetSummaryLabel }}
            </Tag>
            <div
              :data-testid="scope.containerTestId"
              class="flex flex-wrap items-center gap-2"
            >
              <Tag
                v-for="target in scope.targets"
                :key="`${scope.key}-${target.testIdValue}`"
                :data-testid="
                  scope.itemTestIdPrefix
                    ? `${scope.itemTestIdPrefix}-${target.testIdValue}`
                    : undefined
                "
              >
                {{ target.label }}
              </Tag>
            </div>
          </div>

          <div
            v-else
            class="text-[13px] text-[var(--ant-color-text-secondary)]"
          >
            {{ scope.emptyText }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
