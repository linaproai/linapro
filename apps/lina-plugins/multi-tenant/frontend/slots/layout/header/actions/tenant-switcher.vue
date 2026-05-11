<script lang="ts">
import { pluginCapabilityKeys } from '#/plugins/plugin-capabilities';
import { pluginSlotKeys } from '#/plugins/plugin-slots';

export const pluginSlotMeta = {
  capabilities: [pluginCapabilityKeys.tenantManagement],
  order: 0,
  slotKey: pluginSlotKeys.layoutHeaderActionsBefore,
};
</script>

<script setup lang="ts">
import { computed, watch } from 'vue';
import { useRouter } from 'vue-router';

import { useUserStore } from '@vben/stores';

import { Select, Spin } from 'ant-design-vue';

import { $t } from '#/locales';
import { useTenantStore } from '#/store';

const router = useRouter();
const userStore = useUserStore();
const tenantStore = useTenantStore();

const showTenantSwitcher = computed(() => tenantStore.enabled);

async function handleTenantSwitch(value: unknown) {
  const rawTenantId =
    typeof value === 'object' && value !== null && 'value' in value
      ? (value as { value: unknown }).value
      : value;
  const tenantId = Number(rawTenantId);
  if (
    !Number.isFinite(tenantId) ||
    tenantStore.currentTenant?.id === tenantId
  ) {
    return;
  }
  await tenantStore.switchTenant(tenantId, router);
}

async function handleExitImpersonation() {
  await tenantStore.exitImpersonation(router);
}

watch(
  () => ({
    enabled: tenantStore.enabled,
    isPlatform: tenantStore.isPlatform,
    userId: Number(userStore.userInfo?.userId || 0),
  }),
  ({ enabled, isPlatform, userId }) => {
    if (!enabled) {
      return;
    }
    void tenantStore.ensureTenantOptions({ isPlatform, userId });
  },
  { immediate: true },
);
</script>

<template>
  <div class="hidden items-center md:flex">
    <div
      v-if="tenantStore.isImpersonation"
      class="mr-2 flex h-8 max-w-[220px] shrink-0 items-center justify-center gap-1.5 rounded border border-red-300 bg-red-50 px-2.5 text-xs font-medium text-red-700 xl:max-w-[240px] dark:border-red-500/60 dark:bg-red-500/15 dark:text-red-200"
      data-testid="impersonation-banner"
    >
      <span
        :title="
          $t('pages.multiTenant.impersonation.banner', {
            tenant: tenantStore.currentTenant?.name || '',
          })
        "
        class="min-w-0 flex-1 truncate"
        data-testid="impersonation-banner-text"
      >
        {{
          $t('pages.multiTenant.impersonation.banner', {
            tenant: tenantStore.currentTenant?.name || '',
          })
        }}
      </span>
      <a-button
        danger
        ghost
        size="small"
        class="shrink-0 whitespace-nowrap"
        data-testid="impersonation-exit"
        @click="handleExitImpersonation"
      >
        {{ $t('pages.multiTenant.impersonation.exit') }}
      </a-button>
    </div>
    <div v-if="showTenantSwitcher" data-testid="tenant-switcher">
      <Select
        :value="tenantStore.currentTenant?.id"
        :disabled="tenantStore.isImpersonation"
        :field-names="{ label: 'name', value: 'id' }"
        :filter-option="
          (input, option) =>
            String(option?.name || '')
              .toLowerCase()
              .includes(input.toLowerCase()) ||
            String(option?.code || '')
              .toLowerCase()
              .includes(input.toLowerCase())
        "
        :not-found-content="$t('pages.multiTenant.empty.tenants')"
        :options="tenantStore.tenants"
        :placeholder="$t('pages.multiTenant.switcher.placeholder')"
        class="w-60"
        data-testid="tenant-switcher-select"
        show-search
        @select="handleTenantSwitch"
      >
        <template #suffixIcon>
          <Spin
            v-if="tenantStore.switching || tenantStore.loadingTenants"
            size="small"
            spinning
          />
          <span v-else class="icon-[lucide--building-2] size-4"></span>
        </template>
        <template #option="{ name, code }">
          <div class="flex min-w-0 flex-col">
            <span class="truncate">{{ name }}</span>
            <span class="truncate text-xs text-muted-foreground">
              {{ code }}
            </span>
          </div>
        </template>
      </Select>
    </div>
  </div>
</template>
