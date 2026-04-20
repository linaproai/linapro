<script setup lang="ts">
import type { HostServiceCardView } from './plugin-host-service-view';

import { Tag } from 'ant-design-vue';

interface Props {
  cards: HostServiceCardView[];
}

defineProps<Props>();

function hasPanelTargets(card: HostServiceCardView['scopes'][number]) {
  return card.targets.some((target) => target.variant === 'panel');
}
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
            :class="hasPanelTargets(scope) ? 'flex flex-col gap-2' : 'flex flex-wrap items-center gap-2'"
          >
            <Tag
              v-if="scope.targetSummaryLabel"
              :data-testid="`plugin-host-service-summary-label-${card.service}-${scope.key}`"
              :color="scope.targetSummaryBadgeColor"
            >
              {{ scope.targetSummaryLabel }}
            </Tag>
            <div
              :data-testid="scope.containerTestId"
              :class="
                hasPanelTargets(scope)
                  ? 'flex flex-wrap items-start gap-2'
                  : 'flex flex-wrap items-center gap-2'
              "
            >
              <template
                v-for="target in scope.targets"
                :key="`${scope.key}-${target.testIdValue}`"
              >
                <div
                  v-if="target.variant === 'panel'"
                  :data-testid="
                    scope.itemTestIdPrefix
                      ? `${scope.itemTestIdPrefix}-${target.testIdValue}`
                      : undefined
                  "
                  class="min-w-[260px] max-w-full rounded-md border border-[var(--ant-color-border-secondary)] bg-[var(--ant-color-fill-quaternary)] px-3 py-2"
                >
                  <div class="text-[13px] font-medium text-[var(--ant-color-text)]">
                    {{ target.label }}
                  </div>
                  <div
                    v-if="target.details?.length"
                    class="mt-1 flex flex-col gap-1 text-[12px] leading-6 text-[var(--ant-color-text-secondary)]"
                  >
                    <div
                      v-for="detail in target.details"
                      :key="`${target.testIdValue}-${detail.label}`"
                      class="break-words"
                    >
                      <span class="font-semibold text-[var(--ant-color-text)]">
                        {{ detail.label }}：
                      </span>
                      <span>{{ detail.value }}</span>
                    </div>
                  </div>
                </div>
                <Tag
                  v-else
                  :data-testid="
                    scope.itemTestIdPrefix
                      ? `${scope.itemTestIdPrefix}-${target.testIdValue}`
                      : undefined
                  "
                >
                  {{ target.label }}
                </Tag>
              </template>
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
